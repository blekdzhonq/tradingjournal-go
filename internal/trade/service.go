package trade

import (
	"context"
)

type Repository interface{
	GetActiveTrade(ctx context.Context, Symbol Symbol) (*Trade, error)
	SaveTrade(ctx context.Context, t *Trade) error
}

type Service struct{
	repo Repository
}

func NewService (repo Repository) *Service{
	return &Service{
		repo: repo,
	}
}


func (s *Service) AddExecution (ctx context.Context, exec Execution) error{

	if err := exec.Validate(); err != nil{
		return err
	}

	t, err := s.repo.GetActiveTrade(ctx, exec.Symbol)

	if err != nil{
		return err
	}

	if t == nil{
		t = &Trade{
			Symbol: exec.Symbol,
			Quantity: exec.Quantity,
			EntryTime: exec.Time,
			Direction: exec.Direction,
		}
	}

	t.Executions = append(t.Executions, exec)
	t.AggregateExecutions()
	t.CalculatePnl()

	s.repo.SaveTrade(ctx, t)

	return nil
}