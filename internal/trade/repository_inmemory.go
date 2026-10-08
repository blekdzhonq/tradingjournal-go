package trade

import (
	"context"
	"sync"
)

type InMemoryRepository struct {
	mu     sync.RWMutex
	trades map[string]*Trade
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		trades: make(map[string]*Trade),
	}
}

func (r *InMemoryRepository) GetActiveTrade(ctx context.Context, symbol string) (*Trade, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	trade, exists := r.trades[symbol]

	if !exists || trade.IsClosed {
		return nil, nil
	}

	return trade, nil
}

func (r *InMemoryRepository) SaveTrade(ctx context.Context, trade *Trade) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.trades[trade.Symbol.Name] = trade

	return nil

}
