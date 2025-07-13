package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"golang.org/x/sync/errgroup"
)

const appName = "schwabn"

var version string

type config struct {
	conf.BootstrapConf

	Futures      string `env:"FUTURES"`
	ChartFutures string `env:"CHART_FUTURES"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := newApp(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	if info, ok := debug.ReadBuildInfo(); ok {
		a.Logger.InfoContext(ctx,
			"Starting "+appName,
			"version", info.Main.Version,
			"path", info.Main.Path,
			"checksum", info.Main.Sum,
			"codeVersion", version,
		)
	}

	g, ctx := errgroup.WithContext(ctx)
	a.start(ctx, g)

	select { // watch for signal interruptions or context completion
	case sig := <-interrupt:
		a.Logger.Warn("kill signal received", "sig", sig.String())
		cancel()
		break
	case <-ctx.Done():
		a.Logger.Warn("context canceled", "err", ctx.Err())
		break
	}

	a.shutdown()

	if err = g.Wait(); err == nil || errors.Is(err, http.ErrServerClosed) {
		return
	}

	a.Logger.ErrorContext(ctx, "server goroutines stopped with error", "error", err)
	os.Exit(1)
}

func (a *app) start(ctx context.Context, g *errgroup.Group) {
	if a.Metrics != nil {
		g.Go(func() error {
			a.Logger.InfoContext(ctx, "starting metrics server")
			return a.Metrics.ListenAndServe()
		})
	}

	if a.Health != nil {
		g.Go(func() error {
			a.Logger.InfoContext(ctx, "starting health server")
			return a.Health.Start(ctx)
		})
	}
}

func (a *app) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if a.ws != nil {
		a.ws.Close(ctx)
	}

	a.Server.Shutdown(ctx)
}
