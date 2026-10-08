package main

import (
	"fmt"
	"tradingjournal-go/internal/trade"

	"github.com/shopspring/decimal"
)

func main() {
	trade := trade.Trade{
		ID:         "trade-001",
		Symbol: trade.Symbol{Name: "ES", PointValue:decimal.RequireFromString("50")},
		EntryPrice: decimal.RequireFromString("100"),
		ExitPrice:  decimal.RequireFromString("105"),
		Quantity:   decimal.RequireFromString("20"),
		Fee:   trade.Money{Currency: "USD", Amount: decimal.RequireFromString("5")},
		Direction:     trade.Short,}

	fmt.Println(trade.CalculatePnl())
}
