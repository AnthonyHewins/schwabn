package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"github.com/AnthonyHewins/td"
	"github.com/coder/websocket"
)

func (a *app) renewWS(ctx context.Context, c *conf.Schwab) error {
	c.APISecret = strings.TrimSpace(c.APISecret)
	l := a.Logger.With(
		"apikey", c.APIKey,
		"baseURL", c.BaseURL,
		"authUrl", c.AuthURL,
		"len(secret)>0 after trimming spaces", len(c.APISecret) > 0,
		"len(accessToken)>0", len(c.AccessToken) > 0,
		"len(refreshToken)>0", len(c.RefreshToken) > 0,
		"timeout", c.Timeout,
	)

	httpClient, err := td.New(
		ctx,
		c.BaseURL,
		c.AuthURL,
		c.APIKey,
		c.APISecret,
		c.RefreshToken,
		td.WithClientLogger(a.Logger.Handler()),
		td.WithHTTPAccessToken(c.AccessToken),
	)

	if err != nil {
		return err
	}

	a.ws, err = td.NewSocket(
		ctx,
		&websocket.DialOptions{HTTPClient: &http.Client{Timeout: c.Timeout}},
		httpClient,
		c.RefreshToken,
		td.WithTimeout(c.Timeout),
		// td.WithEquityHandler(),
		td.WithChartEquityHandler(a.handler.ChartEquity),
		td.WithFutureHandler(a.handler.Future),
		td.WithChartFutureHandler(a.handler.ChartFuture),
		// td.WithOptionHandler(),
		// td.WithFutureOptionHandler(),
		td.WithLogger(a.Logger.Handler()),
		td.WithErrHandler(func(err error) {
			if err == nil {
				return
			}

			select {
			case a.keepaliveErrs <- err:
			case <-ctx.Done():
			}
		}),
	)

	if err != nil {
		l.ErrorContext(ctx, "failed creating schwab socket", "err", err)
		return err
	}

	for _, fn := range []func(context.Context) error{
		a.subFutures,
		a.subChartFutures,
		a.subChartEquities,
	} {
		if err := fn(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *app) subFutures(ctx context.Context) error {
	if len(a.futures) == 0 {
		return nil
	}

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

func (a *app) subChartEquities(ctx context.Context) error {
	if len(a.chartEquities) == 0 {
		return nil
	}

	_, err := a.ws.AddChartEquitySubscription(ctx, &td.ChartEquityReq{
		Symbols: a.chartEquities,
		Fields:  td.ChartEquityFieldValues(),
	})

	if err != nil {
		a.Logger.ErrorContext(ctx, "failed adding chart equity subscription", "err", err)
		return err
	}

	a.Logger.InfoContext(ctx, "subbed to chart equities", "symbols", a.chartEquities)
	return nil
}
