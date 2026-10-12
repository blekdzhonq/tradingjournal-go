package trade

import (
	"context"
)

type Repository interface {
	GetActiveTrade(ctx context.Context, Symbol string) (*Trade, error)
	SaveTrade(ctx context.Context, t *Trade) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) AddExecution(ctx context.Context, exec Execution) error {

	if err := exec.Validate(); err != nil {
		return err
	}

	t, err := s.repo.GetActiveTrade(ctx, exec.Symbol.Name)

	if err != nil {
		return err
	}

	if t == nil {
		t = &Trade{
			Symbol:    exec.Symbol,
			Quantity:  exec.Quantity,
			EntryTime: exec.Time,
			Direction: exec.Direction,
		}

		t.Executions = append(t.Executions, exec)
		t.AggregateExecutions()

		return s.repo.SaveTrade(ctx, t)
	}

	
	if exec.Direction != t.Direction {
		
		if exec.Quantity.GreaterThan(t.Quantity) {
			closingExec := exec
			closingExec.Quantity = t.Quantity 
			
			t.Executions = append(t.Executions, closingExec)
			t.AggregateExecutions()
			
			if err := s.repo.SaveTrade(ctx, t); err != nil {
				return err
			}

			remQty := exec.Quantity.Sub(t.Quantity)

			newTrade := &Trade{
				Symbol:    exec.Symbol,
				Direction: exec.Direction, 
				IsClosed:  false,
			}
			
			flipExec := exec
			flipExec.Quantity = remQty
			
			newTrade.Executions = append(newTrade.Executions, flipExec)
			newTrade.AggregateExecutions()
			
			return s.repo.SaveTrade(ctx, newTrade)
		}
	}

	t.Executions = append(t.Executions, exec)
	t.AggregateExecutions()

	return s.repo.SaveTrade(ctx, t)

}
