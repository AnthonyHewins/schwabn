package socket

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AnthonyHewins/schwabn/gen/go/schwabn/stream/v0"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
)

type metrics struct {
	marshalFail, publishFail prometheus.Counter
}

func newMetrics(appName, system string) metrics {
	fn := func(name, help string) prometheus.Counter {
		return prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: appName,
			Subsystem: system,
			Name:      name + "_count",
			Help:      help,
		})
	}

	return metrics{
		marshalFail: fn("marshal_fail", "Count of failed proto marshals"),
		publishFail: fn("publish_fail", "Count of publish failures"),
	}
}

type Controller struct {
	future forwarder[*future, *stream.Future]
}

func New(appName string, logger *slog.Logger, js jetstream.JetStream, timeout time.Duration) *Controller {
	return &Controller{
		future: forwarder[*future, *stream.Future]{
			metrics: newMetrics(appName, "futures"),
			timeout: timeout,
			js:      js,
			logger:  logger,
			conv:    futureToProto,
		},
	}
}

type jetstreamMsg interface {
	subject() string
}

type forwarder[X jetstreamMsg, Y proto.Message] struct {
	metrics
	timeout time.Duration
	js      jetstream.JetStream
	logger  *slog.Logger
	conv    func(X) Y
}

func (f forwarder[X, Y]) Forward(x X) {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()

	y := f.conv(x)

	buf, err := proto.Marshal(y)
	if err != nil {
		f.marshalFail.Inc()
		f.logger.Error("failed proto marshal", "type", fmt.Sprintf("%T", y), "err", err, "raw", x)
		return
	}

	if _, err = f.js.Publish(ctx, x.subject(), buf); err != nil {
		f.publishFail.Inc()
		f.logger.Error("failed publishing msg async", "err", err)
		return
	}
}
