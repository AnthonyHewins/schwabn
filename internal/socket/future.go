package socket

import (
	"fmt"
	"time"

	"github.com/AnthonyHewins/schwabn/gen/go/schwabn/stream/v0"
	"github.com/AnthonyHewins/td"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type future td.Future

func (f *future) subject() string {
	return fmt.Sprintf("futures.%s.%02d.%s", f.Symbol.Symbol, f.Symbol.Year, f.Symbol.Month)
}

func (c *Controller) Future(f *td.Future) {
	c.future.forward((*future)(f))
}

func futureToProto(f *future) *stream.Future {
	return &stream.Future{
		Symbol: &stream.FutureID{
			Symbol: f.Symbol.Symbol,
			Month:  stream.Month(f.Symbol.Month),
			Year:   uint32(f.Symbol.Year),
		},
		Description:     f.Description,
		BidPrice:        f.BidPrice,
		AskPrice:        f.AskPrice,
		LastPrice:       f.LastPrice,
		BidSize:         f.BidSize,
		AskSize:         f.AskSize,
		BidId:           stream.ExchangeID(f.BidID),
		AskId:           stream.ExchangeID(f.AskID),
		ExchangeId:      stream.ExchangeID(f.ExchangeID),
		LastId:          stream.ExchangeID(f.LastID),
		ExchangeName:    f.ExchangeName,
		SecurityStatus:  stream.SecurityStatus(f.SecurityStatus),
		OpenInterest:    int32(f.OpenInterest),
		Mark:            f.Mark,
		Tick:            f.Tick,
		TickAmount:      f.TickAmount,
		Product:         f.Product,
		FuturePriceFmt:  f.FuturePriceFmt,
		TradingHours:    f.TradingHours,
		IsTradable:      f.IsTradable,
		Multiplier:      f.Multiplier,
		IsActive:        f.IsActive,
		SettlementPrice: f.SettlementPrice,
		ActiveSymbol:    f.ActiveSymbol,
		NetChange:       f.NetChange,
		PercentChange:   f.PercentChange,
		HighPrice:       f.HighPrice,
		LowPrice:        f.LowPrice,
		ClosePrice:      f.ClosePrice,
		TotalVolume:     f.TotalVolume,
		LastSize:        f.LastSize,
		QuotedInSession: f.QuotedInSession,
		QuoteTime:       newTS(f.QuoteTime),
		TradeTime:       newTS(f.TradeTime),
		AskTime:         newTS(f.AskTime),
		BidTime:         newTS(f.BidTime),
		ExpirationDate:  newTS(f.ExpirationDate),
		SettlementDate:  newTS(f.SettlementDate),
		ExpirationStyle: f.ExpirationStyle,
	}
}

func newTS(t time.Time) *timestamppb.Timestamp {
	if !t.IsZero() {
		return timestamppb.New(t)
	}
	return nil
}
