package main

import (
	"context"
	"time"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"github.com/AnthonyHewins/schwabn/internal/socket"
	"github.com/AnthonyHewins/td"
	"github.com/caarlos0/env/v11"
	"github.com/nats-io/nats.go/jetstream"
)

type app struct {
	*conf.Server

	ws            *td.WS
	c             *conf.Schwab
	keepaliveErrs chan error

	maxConnAttempts uint8
	connBackoff     time.Duration

	handler                     *socket.Controller
	chartFutures, chartEquities []string
	futures                     []td.FutureID
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

	futureIDs, err := c.getFutureIDs()
	if err != nil {
		b.Logger.ErrorContext(ctx, "failed getting future IDs", "err", err, "got", c.Futures)
		return nil, err
	}

	a := app{
		Server:          (*conf.Server)(b),
		c:               &c.Schwab,
		keepaliveErrs:   make(chan error),
		futures:         futureIDs,
		chartFutures:    c.symbolList(c.ChartFutures),
		chartEquities:   c.symbolList(c.ChartEquities),
		maxConnAttempts: c.ConnAttempts,
		connBackoff:     c.Backoff,
	}
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

	a.handler = socket.New(appName, a.Logger, js, c.Prefix, c.Schwab.Timeout)
	return &a, nil
}
