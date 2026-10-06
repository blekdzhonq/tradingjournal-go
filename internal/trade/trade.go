package trade

import "github.com/shopspring/decimal"

const Scale int64 = 100000

type Trade struct {
	ID         string
	Symbol     string
	EntryPrice decimal.Decimal
	ExitPrice  decimal.Decimal
	Quantity   int64
	PointValue decimal.Decimal
	FeeCents   Money
	IsLong     bool
}

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

func (trade *Trade) CalculatePnl() decimal.Decimal {

	var direction = decimal.NewFromInt(-1)

	if trade.IsLong {
		direction = decimal.NewFromInt(1)
	}

	priceDiff := trade.ExitPrice.Sub(trade.EntryPrice)

	step1 := priceDiff.Mul(trade.PointValue)

	grossPnl := step1.Mul(decimal.NewFromInt(trade.Quantity)).Mul(direction)

	return grossPnl.Sub(trade.FeeCents.Amount)
}
