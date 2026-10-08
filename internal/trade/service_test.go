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
	savedTrade    *Trade
}

func (m *mockRepository) GetActiveTrade(ctx context.Context, Symbol string) (*Trade, error) {
	return m.tradeToReturn, m.errToReturn
}

func (m *mockRepository) SaveTrade(ctx context.Context, t *Trade) error {
	m.savedTrade = t
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

			repo := &mockRepository{tradeToReturn: tt.activeTrade}
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

			savedTrade := repo.savedTrade

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
