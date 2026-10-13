package trade

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) GetActiveTrade(ctx context.Context, symbol string) (*Trade, error) {

	row := r.pool.QueryRow(ctx, `
			SELECT id, symbol_name, point_value, direction, entry_price, exit_price, quantity, pnl, fee_amount, fee_currency, is_closed, entry_time, exit_time 
			FROM trades
			WHERE symbol_name = $1 AND is_closed = FALSE
			LIMIT 1;
			`, symbol)

	var t Trade

	var symbolName string
	var pointValue, feeAmount decimal.Decimal
	var feeCurrency string

	err := row.Scan(
		&t.ID, &symbolName, &pointValue, &t.Direction,
		&t.EntryPrice, &t.ExitPrice, &t.Quantity, &t.PnL,
		&feeAmount, &feeCurrency, &t.IsClosed, &t.EntryTime, &t.ExitTime,
	)

	if err != nil {
		return nil, err
	}

	t.Symbol = Symbol{Name: symbolName, PointValue: t.Symbol.PointValue}
	t.Fee = Money{Currency: feeCurrency, Amount: feeAmount}

	rows, err := r.pool.Query(ctx, `
				SELECT id, symbol_name, direction, pric, quantity, fee_amount, fee_currency, time
				FROM executions
				WHERE trade_id = $1
				ORDER BY time ASC;
				`, t.ID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var exec Execution
		var execSymbolName string

		err := rows.Scan(
			&exec.ID, &execSymbolName, &exec.Direction,
			&exec.Price, &exec.Quantity, &exec.Fee.Amount, &exec.Fee.Currency, &exec.Time,
		)

		if err != nil {
			return nil, err
		}

		exec.Symbol = Symbol{Name: execSymbolName, PointValue: t.Symbol.PointValue}
		t.Executions = append(t.Executions, exec)

	}

	return &t, err

}

func (r *PostgresRepository) SaveTrade(ctx context.Context, t *Trade) error {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
			INSERT INTO trades (id, symbol_name, point_value, direction, entry_price, exit_price, quantity, pnl, fee_amount, fee_currency, is_closed, entry_time, exit_time)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (id) DO UPDATE SET
				entry_price = EXCLUDED.entry_price
				exit_price = EXCLUDED.exit_price
				quantity = EXCLUDED.quantity
				pnl = EXCLUDED.pnl
				fee_amount = EXCLUDED.fee_amount
				is_closed = EXCLUDED.is_closed
				exit_time = EXCLUDED.exit_time;
	`, t.ID, t.Symbol.Name, t.Symbol.PointValue, t.Direction, t.EntryPrice, t.ExitPrice, t.Quantity, t.PnL, t.Fee.Amount, t.Fee.Currency, t.IsClosed, t.EntryTime, t.ExitTime)

	if err != nil {
		return err
	}

	for _, exec := range t.Executions {
		_, err = tx.Exec(ctx, `
					INSERT INTO executions (id, trade_id, symbol_name, direction, price, quantity, fee_amount, fee_currency, time)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
					ON CONFLICT (id) DO NOTHING
		`, exec.ID, t.ID, exec.Symbol.Name, exec.Direction, exec.Price, exec.Quantity, exec.Fee.Amount, exec.Fee.Currency, exec.Time)

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
