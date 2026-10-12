package trade

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type mockRepository struct {
	tradeToReturn *Trade
	errToReturn   error
	savedTrades   []*Trade
}

func (m *mockRepository) GetActiveTrade(ctx context.Context, Symbol string) (*Trade, error) {
	return m.tradeToReturn, m.errToReturn
}

func (m *mockRepository) SaveTrade(ctx context.Context, t *Trade) error {
	m.savedTrades = append(m.savedTrades, t)
	return nil
}

func TestAddExecution(t *testing.T) {

	appleSymbol := Symbol{Name: "AAPL", PointValue: decimal.NewFromInt(1)}
	usd := "USD"

	tests := []struct {
		name         string
		activeTrade  *Trade
		newExecution Execution
		wantErrText  string
		wantIsClosed bool
		wantPnL      decimal.Decimal
	}{
		{
			name:        "1. Open new long position. Pnl should be -2",
			activeTrade: nil,
			newExecution: Execution{
				Symbol:    appleSymbol,
				Direction: Long,
				Price:     decimal.NewFromInt(150),
				Quantity:  decimal.NewFromInt(10),
				Fee:       Money{Amount: decimal.NewFromInt(2), Currency: usd},
				Time:      time.Now(),
			},

			wantErrText:  "",
			wantIsClosed: false,
			wantPnL:      decimal.NewFromInt(-2),
		},
		{
			name: "2. Full close LONG position (Should calculate profit minus fee)",
			activeTrade: &Trade{
				Symbol:    appleSymbol,
				Direction: Long,
				IsClosed:  false,
				Executions: []Execution{
					{
						Symbol:    appleSymbol,
						Price:     decimal.NewFromInt(150),
						Quantity:  decimal.NewFromInt(10),
						Direction: Long,
						Fee:       Money{Amount: decimal.NewFromInt(2), Currency: usd},
					},
				},
			},
			newExecution: Execution{
				Symbol:    appleSymbol,
				Direction: Short,
				Price:     decimal.NewFromInt(170),
				Quantity:  decimal.NewFromInt(10),
				Fee:       Money{Amount: decimal.NewFromInt(3), Currency: usd},
				Time:      time.Now(),
			},

			wantIsClosed: true,
			wantPnL:      decimal.NewFromInt(195),
			wantErrText:  "",
		},
		{
			name: "3. Partial close LONG position (Should calculate profit minus fee)",
			activeTrade: &Trade{
				Symbol:    appleSymbol,
				Direction: Long,
				IsClosed:  false,
				Executions: []Execution{
					{
						Symbol:    appleSymbol,
						Price:     decimal.NewFromInt(150),
						Quantity:  decimal.NewFromInt(10),
						Direction: Long,
						Fee:       Money{Amount: decimal.NewFromInt(2), Currency: usd},
					},
				},
			},
			newExecution: Execution{
				Symbol:    appleSymbol,
				Direction: Short,
				Price:     decimal.NewFromInt(170),
				Quantity:  decimal.NewFromInt(5),
				Fee:       Money{Amount: decimal.NewFromInt(1), Currency: usd},
				Time:      time.Now(),
			},

			wantIsClosed: false,
			wantPnL:      decimal.NewFromInt(97),
			wantErrText:  "",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			var activeTradeForTest *Trade

			if tt.activeTrade != nil {
				activeTradeForTest = &Trade{
					Symbol:     tt.activeTrade.Symbol,
					Direction:  tt.activeTrade.Direction,
					IsClosed:   tt.activeTrade.IsClosed,
					Quantity:   tt.activeTrade.Quantity,
					EntryPrice: tt.activeTrade.EntryPrice,
				}
				activeTradeForTest.Executions = append([]Execution{}, tt.activeTrade.Executions...)
			}

			if activeTradeForTest != nil {
				activeTradeForTest.AggregateExecutions()
			}

			repo := &mockRepository{tradeToReturn: activeTradeForTest}
			service := NewService(repo)

			err := service.AddExecution(context.Background(), tt.newExecution)

			if tt.wantErrText != "" {
				if err == nil || err.Error() != tt.wantErrText {
					t.Fatalf("%q error expected but recieved %v", tt.wantErrText, err)
					return
				}
			}

			if err != nil {
				t.Fatalf("Didn't expected error but recieved %v", err)
				return
			}

			savedTrade := repo.savedTrades[0]

			if savedTrade == nil {
				t.Fatal("Service sucessfuly completed, but trade wasn't saved")
			}

			if savedTrade.IsClosed != tt.wantIsClosed {
				t.Fatalf("Status IsClosed is %v but we expect IsClosed = %v", savedTrade.IsClosed, tt.wantIsClosed)
			}

			if !savedTrade.PnL.Equal(tt.wantPnL) {
				t.Errorf("=== ТЕСТ УПАЛ: %s ===", tt.name)
				t.Errorf("Ожидали PnL: %s, но получили: %s", tt.wantPnL, savedTrade.PnL)
				t.Errorf("--- Отладочная информация по структуре Trade ---")
				t.Errorf("Направление трейда: %s", savedTrade.Direction)
				t.Errorf("Статус IsClosed:    %v", savedTrade.IsClosed)
				t.Errorf("Посчитанная EntryPrice: %s", savedTrade.EntryPrice)
				t.Errorf("Посчитанная ExitPrice:  %s", savedTrade.ExitPrice)
				t.Errorf("Посчитанный Quantity:   %s", savedTrade.Quantity)
				t.Errorf("Суммарная Fee (Комиссия): %s", savedTrade.Fee.Amount)
				t.Errorf("Количество сделок в массиве: %d", len(savedTrade.Executions))
				t.FailNow()
			}

		})
	}
}

func TestAddExecution_PositionFlip(t *testing.T) {
	appleSymbol := Symbol{Name: "AAPL", PointValue: decimal.NewFromInt(1)}
	usd := "USD"

	activeTrade := &Trade{
		Symbol:    appleSymbol,
		Direction: Long,
		Quantity:  decimal.NewFromInt(10),
		IsClosed:  false,
		Executions: []Execution{
			{
				Symbol:    appleSymbol,
				Direction: Long,
				Price:     decimal.NewFromInt(150),
				Quantity:  decimal.NewFromInt(10),
				Fee:       Money{Amount: decimal.NewFromInt(2), Currency: usd},
			},
		},
	}

	newExecution := Execution{
		Symbol:    appleSymbol,
		Direction: Short,
		Price:     decimal.NewFromInt(170),
		Quantity:  decimal.NewFromInt(25),
		Fee:       Money{Amount: decimal.NewFromInt(3), Currency: usd},
	}

	repo := &mockRepository{tradeToReturn: activeTrade}
	service := NewService(repo)

	wantFirstClosed := true
	wantFirstQty := decimal.NewFromInt(10)
	wantFirstPnL := decimal.NewFromInt(195)

	wantSecondClosed := false
	wantSecondQty := decimal.NewFromInt(15)
	wantSecondPnL := decimal.Zero

	err := service.AddExecution(context.Background(), newExecution)
	if err != nil {
		t.Fatalf("Service returned an error: %v", err)
	}

	if len(repo.savedTrades) != 2 {
		t.Fatalf("2 trades want but got: %d", len(repo.savedTrades))
	}

	firstTrade := repo.savedTrades[0]
	secondTrade := repo.savedTrades[1]

	if firstTrade.IsClosed != wantFirstClosed {
		t.Errorf("1st Trade: IsClosed: got %v, want %v", firstTrade.IsClosed, wantFirstClosed)
	}
	if !firstTrade.Quantity.Equal(wantFirstQty) {
		t.Errorf("1st Trade: Quantity: got %s, want %s", firstTrade.Quantity, wantFirstQty)
	}
	if !firstTrade.PnL.Equal(wantFirstPnL) {
		t.Errorf("1st Trade:  PnL: got %s, want %s", firstTrade.PnL, wantFirstPnL)
	}

	if secondTrade.IsClosed != wantSecondClosed {
		t.Errorf("2nd Trade: IsClosed: got %v, want %v", secondTrade.IsClosed, wantSecondClosed)
	}
	if !secondTrade.Quantity.Equal(wantSecondQty) {
		t.Errorf("2nd Trade: Quantity: got %s, want %s", secondTrade.Quantity, wantSecondQty)
	}
	if !secondTrade.PnL.Equal(wantSecondPnL) {
		t.Errorf("2nd Trade: PnL: got %s, want %s", secondTrade.PnL, wantSecondPnL)
	}
}

func TestAddExecution_PositionFlip_AfterPartialClose(t *testing.T) {
	appleSymbol := Symbol{Name: "AAPL", PointValue: decimal.NewFromInt(1)}
	usd := "USD"

	activeTrade := &Trade{
		Symbol:    appleSymbol,
		Direction: Long,
		Quantity:  decimal.NewFromInt(10),
		IsClosed:  false,
		Executions: []Execution{
			{
				Symbol:    appleSymbol,
				Direction: Long,
				Price:     decimal.NewFromInt(150),
				Quantity:  decimal.NewFromInt(10),
				Fee:       Money{Amount: decimal.NewFromInt(2), Currency: usd},
			},
		},
	}

	newExecution1 := Execution{
		Symbol:    appleSymbol,
		Direction: Short,
		Price:     decimal.NewFromInt(170),
		Quantity:  decimal.NewFromInt(5),
		Fee:       Money{Amount: decimal.NewFromInt(3), Currency: usd},
	}

	newExecution2 := Execution{
		Symbol:    appleSymbol,
		Direction: Short,
		Price:     decimal.NewFromInt(170),
		Quantity:  decimal.NewFromInt(8),
		Fee:       Money{Amount: decimal.NewFromInt(3), Currency: usd},
	}

	repo := &mockRepository{tradeToReturn: activeTrade}
	service := NewService(repo)

	wantFirstClosed := true
	wantFirstQty := decimal.NewFromInt(10)
	wantFirstPnL := decimal.NewFromInt(192)

	wantSecondClosed := false
	wantSecondQty := decimal.NewFromInt(3)
	wantSecondPnL := decimal.Zero

	if err := service.AddExecution(context.Background(), newExecution1); err != nil {
		t.Fatalf("partial close failed: %v", err)
	}
	if err := service.AddExecution(context.Background(), newExecution2); err != nil {
		t.Fatalf("Service returned an error: %v", err)
	}

	if len(repo.savedTrades) != 3 {
		t.Fatalf("3 trades want but got: %d", len(repo.savedTrades))
	}

	firstTrade := repo.savedTrades[1]
	secondTrade := repo.savedTrades[2]

	if firstTrade.IsClosed != wantFirstClosed {
		t.Errorf("1st Trade: IsClosed: got %v, want %v", firstTrade.IsClosed, wantFirstClosed)
	}
	if !firstTrade.Quantity.Equal(wantFirstQty) {
		t.Errorf("1st Trade: Quantity: got %s, want %s", firstTrade.Quantity, wantFirstQty)
	}
	if !firstTrade.PnL.Equal(wantFirstPnL) {
		t.Errorf("1st Trade:  PnL: got %s, want %s", firstTrade.PnL, wantFirstPnL)
	}

	if secondTrade.IsClosed != wantSecondClosed {
		t.Errorf("2nd Trade: IsClosed: got %v, want %v", secondTrade.IsClosed, wantSecondClosed)
	}
	if !secondTrade.Quantity.Equal(wantSecondQty) {
		t.Errorf("2nd Trade: Quantity: got %s, want %s", secondTrade.Quantity, wantSecondQty)
	}
	if !secondTrade.PnL.Equal(wantSecondPnL) {
		t.Errorf("2nd Trade: PnL: got %s, want %s", secondTrade.PnL, wantSecondPnL)
	}
}
