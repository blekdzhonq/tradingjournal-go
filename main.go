package main

import (
	"context"
	"fmt"
	"tradingjournal-go/internal/trade"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

func main() {

	repo := trade.NewInMemoryRepository()

	service := trade.NewService(repo)

	ctx := context.Background()

	Symbol := trade.Symbol{
		Name:       "AAPL",
		PointValue: decimal.NewFromInt(1),
	}

	exec := trade.Execution{
		Symbol:    Symbol,
		Direction: trade.Long,
		Price:     decimal.NewFromInt(100),
		Quantity:  decimal.NewFromInt(10),
		Fee:       trade.Money{Currency: "USD", Amount: decimal.NewFromInt(2)},
	}

	service.AddExecution(
		ctx,
		exec)

	t1, err := repo.GetActiveTrade(ctx, Symbol.Name)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Trade 1")
	fmt.Printf("trade1 IsClosed= %v Pnl= %q ", t1.IsClosed, t1.PnL)

	exec2 := trade.Execution{
		Symbol:    Symbol,
		Direction: trade.Short,
		Price:     decimal.NewFromInt(90),
		Quantity:  decimal.NewFromInt(20),
		Fee:       trade.Money{Currency: "USD", Amount: decimal.NewFromInt(1)},
	}

	service.AddExecution(ctx, exec2)

}

func InitDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return nil
	}

	return pgxpool.NewWithConfig(ctx, config)
}
