package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/AnthonyHewins/schwabn/internal/conf"
	"github.com/AnthonyHewins/td"
	"golang.org/x/sync/errgroup"
)

const appName = "schwabn"

var version string

type config struct {
	conf.BootstrapConf

	ConnAttempts uint8 `env:"CONN_ATTEMPTS" envDefault:"30"`

	Prefix string `env:"PREFIX" envDefault:"schwabn"`

	Futures       string `env:"FUTURES"`
	ChartFutures  string `env:"CHART_FUTURES"`
	ChartEquities string `env:"CHART_EQUITIES"`
}

func (c *config) getFutureIDs() ([]td.FutureID, error) {
	x := c.symbolList(c.Futures)
	if len(x) == 0 {
		return nil, nil
	}

	ids := make([]td.FutureID, len(x))
	for i, v := range x {
		var id td.FutureID
		if err := json.Unmarshal([]byte(v), &id); err != nil {
			return nil, fmt.Errorf(
				"failed unmarshal of future ID %s; ID must be '/'+<symbol>+<month>+<last 2 digits of year>; error: %w",
				v, err,
			)
		}
		ids[i] = id
	}

	return ids, nil
}

func (c *config) symbolList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	return strings.Split(s, ",")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := newApp(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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

	if err = g.Wait(); err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, context.Canceled) {
		return
	}

	a.Logger.ErrorContext(ctx, "server goroutines stopped with error", "error", err)
	os.Exit(1)
}

func (a *app) start(ctx context.Context, g *errgroup.Group) {
	g.Go(func() error {
		l := a.Logger.With("maxConnAttempts", a.maxConnAttempts)
		for i := uint8(0); i < uint8(a.maxConnAttempts); i++ {
			scopedCtx, cancel := context.WithCancel(ctx)

			err := a.renewWS(scopedCtx, a.c)
			if err != nil {
				cancel()
				l.ErrorContext(ctx, "failed connection attempt, retrying", "attempt", i)
				continue
			}

			i = 0
			select {
			case <-ctx.Done():
				cancel()
				a.Logger.ErrorContext(ctx, "application context killed; user or system wants to end process", "err", ctx.Err())
				return ctx.Err()
			case err, ok := <-a.keepaliveErrs:
				cancel()
				if !ok {
					l.ErrorContext(ctx, "empty keepalive error sent? keepalive pipe closed? killing keepalive loop")
					return fmt.Errorf("keepalive channel closed?")
				}

				if !errors.Is(err, net.ErrClosed) {
					_ = a.ws.Close(ctx)
				}
			}
		}

		return fmt.Errorf("exceeeded connection attempt maximum %d", a.maxConnAttempts)
	})

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

	a.Server.Shutdown(ctx)
}
