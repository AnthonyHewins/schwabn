package main

import (
	"context"
	"encoding/json"
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

	for _, fn := range []func(context.Context, *config) error{
		a.futures,
		a.chartFutures,
	} {
		if err := fn(ctx, &c); err != nil {
			return nil, err
		}
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

func (a *app) futures(ctx context.Context, c *config) error {
	x := strings.Split(strings.TrimSpace(c.Futures), ",")

	if len(x) == 0 {
		return nil
	}

	ids := make([]td.FutureID, len(x))
	for i, v := range x {
		var id td.FutureID
		if err := json.Unmarshal([]byte(v), &id); err != nil {
			a.Logger.ErrorContext(ctx,
				"failed unmarshal of future ID; ID must be '/'+<symbol>+<month>+<last 2 digits of year>",
				"raw", v,
				"err", err,
			)
			return err
		}
		ids[i] = id
	}

	_, err := a.ws.AddFutureSubscription(ctx, &td.FutureReq{
		Symbols: ids,
		Fields:  td.FutureFieldValues(),
	})

	if err != nil {
		a.Logger.ErrorContext(ctx, "failed adding futures subscription", "err", err)
		return err
	}

	a.Logger.InfoContext(ctx, "subbed to futures", "symbols", x)
	return nil
}

func (a *app) chartFutures(ctx context.Context, c *config) error {
	x := strings.Split(strings.TrimSpace(c.ChartFutures), ",")

	if len(x) == 0 {
		return nil
	}

	_, err := a.ws.AddChartFutureSubscription(ctx, &td.ChartFutureReq{
		Symbols: x,
		Fields:  td.ChartFutureFieldValues(),
	})

	if err != nil {
		a.Logger.ErrorContext(ctx, "failed adding chart futures subscription", "err", err)
		return err
	}

	a.Logger.InfoContext(ctx, "subbed to chart futures", "symbols", x)
	return nil
}
