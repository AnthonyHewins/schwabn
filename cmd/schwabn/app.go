package main

import (
	"context"
	"strings"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"github.com/AnthonyHewins/schwabn/internal/socket"
	"github.com/AnthonyHewins/td"
	"github.com/caarlos0/env/v11"
	"github.com/nats-io/nats.go/jetstream"
)

type app struct {
	*conf.Server
	ws *td.WS

	handler      *socket.Controller
	chartFutures []string
	futures      []td.FutureID
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
		return nil, err
	}

	a := app{
		Server:       (*conf.Server)(b),
		futures:      futureIDs,
		chartFutures: strings.Split(strings.TrimSpace(c.ChartFutures), ","),
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

	a.handler = socket.New(appName, a.Logger, js, c.Schwab.Timeout)

	if a.ws, err = a.createWS(ctx, &c.Schwab); err != nil {
		return nil, err
	}

	return &a, nil
}
