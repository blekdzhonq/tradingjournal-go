package trade

import (
	"context"
	"sync"
)

type InMemoryRepository struct{
	trades map[string]*Trade
	rm sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository{
	return &InMemoryRepository{
		trades: make(map[string]*Trade),
	}
}

 func (r *InMemoryRepository) GetActiveTrade(ctx context.Context, symbolName string) (*Trade, error){
	r.rm.RLock()
	defer r.rm.RUnlock()

	trade, exists := r.trades[symbolName]

	if !exists{
		return nil, nil
	}

	return trade, nil

}

func (r *InMemoryRepository) SaveTrade(ctx context.Context, t *Trade) error{
	r.rm.Lock()
	defer r.rm.Unlock()

	r.trades[t.Symbol.Name] = t

	return nil

}
