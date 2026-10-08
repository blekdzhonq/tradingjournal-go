package trade

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type Direction string

const (
	Long  Direction = "LONG"
	Short Direction = "SHORT"
)

type Symbol struct {
	Name       string
	PointValue decimal.Decimal
}

type Money struct {
	Amount   decimal.Decimal
	Currency string
}
type Trade struct {
	ID         string
	Symbol     Symbol
	EntryPrice decimal.Decimal
	ExitPrice  decimal.Decimal
	Quantity   decimal.Decimal
	EntryTime  time.Time
	ExitTime   time.Time
	Fee        Money
	Direction  Direction
	Executions []Execution
	IsClosed   bool
	PnL        decimal.Decimal
}

type Execution struct {
	ID        string
	Symbol    Symbol
	Time      time.Time
	Direction Direction
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Fee       Money
}

func (e Execution) Validate() error {
	if e.Symbol.Name == "" {
		return errors.New("symbol name cannot be empty")
	}
	if !e.Symbol.PointValue.IsPositive() {
		return errors.New("symbol point value must be greater than zero")
	}
	if !e.Price.IsPositive() {
		return errors.New("price must be greater than zero")
	}
	if !e.Quantity.IsPositive() {
		return errors.New("quantity must be greater than zero")
	}
	if e.Direction != Long && e.Direction != Short {
		return errors.New("invalid execution direction")
	}
	return nil
}

func (trade *Trade) CalculatePnl() decimal.Decimal {

	if !trade.IsClosed {
		return decimal.Zero
	}

	var direction = decimal.NewFromInt(-1)

	if trade.Direction == Long {
		direction = decimal.NewFromInt(1)
	}

	priceDiff := trade.ExitPrice.Sub(trade.EntryPrice)

	step1 := priceDiff.Mul(trade.Symbol.PointValue)

	grossPnl := step1.Mul(trade.Quantity).Mul(direction)

	pnl := grossPnl.Sub(trade.Fee.Amount)

	return pnl
}

func (trade *Trade) AggregateExecutions() {
	if len(trade.Executions) == 0 {
		return
	}

	var totalEntryQuantity decimal.Decimal
	var totalEntryCost decimal.Decimal
	var totalExitQuantity decimal.Decimal
	var realizedGrossPnL decimal.Decimal
	var totalFee decimal.Decimal

	trade.EntryTime = trade.Executions[0].Time

	for _, exec := range trade.Executions {
		totalFee = totalFee.Add(exec.Fee.Amount)

		if exec.Direction == trade.Direction {
			totalEntryQuantity = totalEntryQuantity.Add(exec.Quantity)
			totalEntryCost = totalEntryCost.Add(exec.Price.Mul(exec.Quantity))
			trade.EntryPrice = totalEntryCost.Div(totalEntryQuantity)
		} else {
			totalExitQuantity = totalExitQuantity.Add(exec.Quantity)

			var priceDiff decimal.Decimal
			if trade.Direction == Long {
				priceDiff = exec.Price.Sub(trade.EntryPrice)
			} else {
				priceDiff = trade.EntryPrice.Sub(exec.Price)
			}

			execPnL := priceDiff.Mul(trade.Symbol.PointValue).Mul(exec.Quantity)
			realizedGrossPnL = realizedGrossPnL.Add(execPnL)
			trade.ExitPrice = exec.Price
		}
	}

	remainingQty := totalEntryQuantity.Sub(totalExitQuantity)

	if remainingQty.LessThanOrEqual(decimal.Zero) {
		trade.IsClosed = true
		trade.ExitTime = trade.Executions[len(trade.Executions)-1].Time
		trade.Quantity = totalEntryQuantity
	} else {
		trade.IsClosed = false
		trade.Quantity = remainingQty
	}

	// Фиксация PnL: чистый профит от закрытых частей МИНУС абсолютно все комиссии,
	// уплаченные за время жизни этого трейда.
	trade.PnL = realizedGrossPnL.Sub(totalFee)

	trade.Fee = Money{
		Currency: trade.Executions[0].Fee.Currency,
		Amount:   totalFee,
	}
}
