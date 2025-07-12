package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"github.com/AnthonyHewins/schwabn/internal/socket"
	"github.com/AnthonyHewins/td"
	"github.com/caarlos0/env/v11"
	"github.com/coder/websocket"
	"github.com/nats-io/nats.go/jetstream"
)

type app struct {
	*conf.Server
	ws *td.WS
}

func newApp(ctx context.Context) (*app, error) {
	var c config
	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	b, err := c.BootstrapConf.New(ctx, appName)
	if err != nil {
		return nil, err
	}

	a := app{Server: (*conf.Server)(b)}
	defer func() {
		if err != nil {
			a.shutdown()
		}
	}()

	js, err := jetstream.New(a.NC)
	if err != nil {
		a.Logger.ErrorContext(ctx, "failed connecting to jetstream", "err", err)
		return nil, err
	}

	if a.ws, err = a.createWS(ctx, &c.Schwab, js); err != nil {
		return nil, err
	}

	return &a, nil
}

func (a *app) createWS(ctx context.Context, c *conf.Schwab, js jetstream.JetStream) (*td.WS, error) {
	handler := socket.New(appName, a.Logger, js, c.Timeout)

	c.APISecret = strings.TrimSpace(c.APISecret)
	l := a.Logger.With(
		"apikey", c.APIKey,
		"len(secret)>0 after trimming spaces", len(c.APISecret) > 0,
		"baseURL", c.BaseURL,
		"authUrl", c.AuthURL,
		"len(accessToken)>0", c.AccessToken,
		"len(refreshToken)>0", c.RefreshToken,
		"timeout", c.Timeout,
	)

	socket, err := td.NewSocket(
		ctx,
		&websocket.DialOptions{HTTPClient: &http.Client{Timeout: c.Timeout}},
		td.New(
			c.BaseURL,
			c.AuthURL,
			c.APIKey,
			c.APISecret,
			td.WithClientLogger(a.Logger.Handler()),
			td.WithHTTPAccessToken(c.AccessToken),
		),
		c.RefreshToken,
		td.WithTimeout(c.Timeout),
		// td.WithEquityHandler(),
		// td.WithChartEquityHandler(),
		td.WithFutureHandler(handler.Future),
		// td.WithChartFutureHandler(),
		// td.WithOptionHandler(),
		// td.WithFutureOptionHandler(),
		// td.WithLogger(),
	)

	if err != nil {
		l.ErrorContext(ctx, "failed creating schwab socket", "err", err)
		return nil, err
	}

	return socket, nil
}
