package socket

import (
	"fmt"

	"github.com/AnthonyHewins/schwabn/gen/go/schwabn/stream/v0"
	"github.com/AnthonyHewins/td"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type chartEquity td.ChartEquity

func (c *chartEquity) subject() string {
	return fmt.Sprintf("chartequity.%s", c.Symbol)
}

func (c *Controller) ChartEquity(s *td.ChartEquity) {
	c.chartEquity.forward((*chartEquity)(s))
}

func chartEquityToProto(c *chartEquity) *stream.ChartEquity {
	return &stream.ChartEquity{
		Symbol:     c.Symbol,
		OpenPrice:  c.OpenPrice,
		HighPrice:  c.HighPrice,
		LowPrice:   c.LowPrice,
		ClosePrice: c.ClosePrice,
		Volume:     c.Volume,
		Sequence:   int64(c.Sequence),
		Time:       timestamppb.New(c.Time),
		Day:        int64(c.Day),
	}
}
