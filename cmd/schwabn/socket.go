package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"github.com/AnthonyHewins/td"
	"github.com/coder/websocket"
)

func (a *app) createWS(ctx context.Context, c *conf.Schwab) (*td.WS, error) {
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
		td.WithFutureHandler(a.handler.Future),
		td.WithChartFutureHandler(a.handler.ChartFuture),
		// td.WithOptionHandler(),
		// td.WithFutureOptionHandler(),
		td.WithLogger(a.Logger.Handler()),
	)

	if err != nil {
		l.ErrorContext(ctx, "failed creating schwab socket", "err", err)
		return nil, err
	}

	for _, fn := range []func(context.Context) error{
		a.subFutures,
		a.subChartFutures,
	} {
		if err := fn(ctx); err != nil {
			return nil, err
		}
	}

	return socket, nil
}

func (a *app) subFutures(ctx context.Context) error {
	_, err := a.ws.AddFutureSubscription(ctx, &td.FutureReq{
		Symbols: a.futures,
		Fields:  td.FutureFieldValues(),
	})

	if err != nil {
		a.Logger.ErrorContext(ctx, "failed adding futures subscription", "err", err)
		return err
	}

	a.Logger.InfoContext(ctx, "subbed to futures", "symbols", a.futures)
	return nil
}

func (a *app) subChartFutures(ctx context.Context) error {
	if len(a.chartFutures) == 0 {
		return nil
	}

	_, err := a.ws.AddChartFutureSubscription(ctx, &td.ChartFutureReq{
		Symbols: a.chartFutures,
		Fields:  td.ChartFutureFieldValues(),
	})

	if err != nil {
		a.Logger.ErrorContext(ctx, "failed adding chart futures subscription", "err", err)
		return err
	}

	a.Logger.InfoContext(ctx, "subbed to chart futures", "symbols", a.chartFutures)
	return nil
}
