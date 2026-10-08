package trade

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

type mockRepository struct{
	tradeToReturn *Trade
	errToReturn error
	savedTrade *Trade
}

func (m *mockRepository) GetActiveTrade(ctx context.Context, Symbol Symbol) (*Trade, error){
	return m.tradeToReturn, m.errToReturn
}

func (m *mockRepository) SaveTrade(ctx context.Context, t *Trade) error{
	m.savedTrade = t
	return nil
}

func TestAddExecution_InvalidPrice(t *testing.T){
	repo := &mockRepository{}
	service := NewService(repo)

	exec := Execution{
		Symbol: Symbol{Name: "ES", PointValue: decimal.NewFromInt(50)},
		Price: decimal.Zero,
		Quantity: decimal.NewFromInt(1),
	}
	err := service.AddExecution(context.Background(), exec)

	if(err == nil){
		t.Error("Expecting price validation error, but method get executed sucessfully")
	}
	
}

func TestAddExecution_CreateTrade(t *testing.T){
	repo := &mockRepository{}
	service := NewService(repo)

	exec := Execution{
		Symbol: Symbol{Name: "ES", PointValue: decimal.NewFromInt(50)},
		Price: decimal.NewFromInt(100),
		Quantity: decimal.NewFromInt(1),
		Direction: Long,
	}

	err := service.AddExecution(context.Background(), exec)

	if(err != nil){
		t.Fatalf("Method return unexpected error: %v", err)
	}

	if(repo.savedTrade == nil){
		t.Error("Service should've create and save the new trade")
	}


}