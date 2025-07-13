package socket

import (
	"fmt"

	"github.com/AnthonyHewins/schwabn/gen/go/schwabn/stream/v0"
	"github.com/AnthonyHewins/td"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type chartFuture td.ChartFuture

func (c *chartFuture) subject() string {
	return fmt.Sprintf("chartfutures.%s", c.Symbol)
}

func (c *Controller) ChartFuture(f *td.ChartFuture) {
	c.chartFuture.forward((*chartFuture)(f))
}

func newChartFuture(f *chartFuture) *stream.ChartFuture {
	return &stream.ChartFuture{
		Symbol:     f.Symbol,
		Time:       timestamppb.New(f.Time),
		OpenPrice:  f.OpenPrice,
		HighPrice:  f.HighPrice,
		LowPrice:   f.LowPrice,
		ClosePrice: f.ClosePrice,
		Volume:     f.Volume,
	}
}
