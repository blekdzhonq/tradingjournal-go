package main

import (
	"fmt"
	"tradingjournal-go/internal/trade"

	"github.com/shopspring/decimal"
)

func main() {
	trade := trade.Trade{
		ID:         "trade-001",
		Symbol:     "ES",
		EntryPrice: decimal.RequireFromString("100"),
		ExitPrice:  decimal.RequireFromString("105"),
		Quantity:   20,
		FeeCents:   trade.Money{Currency: "USD", Amount: decimal.RequireFromString("500")},
		IsLong:     false,
		PointValue: decimal.RequireFromString("10")}

	fmt.Println(trade.CalculatePnl())
}
