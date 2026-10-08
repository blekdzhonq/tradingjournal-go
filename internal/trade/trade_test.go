package trade

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculatePnl(t *testing.T) {
	type testCase struct {
		name  string
		trade Trade
		want  decimal.Decimal
	}

	

	tests := []testCase{
		{
			name: "Profitable Long for 1 contract (ES)",
			trade: Trade{
				EntryPrice: decimal.RequireFromString("4500"),
				ExitPrice:  decimal.RequireFromString("4510"),
				Quantity:   decimal.RequireFromString("10"),
				Symbol: Symbol{Name: "ES", PointValue:decimal.RequireFromString("50")},
				Fee:   Money{Currency: "USD", Amount: decimal.RequireFromString("2.5")},
				Direction:  Long,
				IsClosed: true,
			},
			want: decimal.RequireFromString("4997.5"),
		},

		{
			name: "Loss Short for 2 contracts",
			trade: Trade{
				EntryPrice: decimal.RequireFromString("100"),
				ExitPrice:  decimal.RequireFromString("105"),
				Quantity:  	decimal.RequireFromString("20"),
				Symbol: Symbol{Name: "ES", PointValue:decimal.RequireFromString("10")},
				Fee:   Money{Currency: "USD", Amount: decimal.RequireFromString("5")},
				Direction:  Short,
				IsClosed: true,
			},

			want: decimal.RequireFromString("-1005"),
		},

		{
			name: "Small price change with large quantity",
			trade: Trade{
				EntryPrice: decimal.RequireFromString("1000000"),
				ExitPrice:  decimal.RequireFromString("1000001"),
				Quantity:  	decimal.RequireFromString("2000"),
				Symbol: Symbol{Name: "ES", PointValue:decimal.RequireFromString("50")},
				Fee:   Money{Currency: "USD", Amount: decimal.RequireFromString("0")},
				Direction:     Long,
				IsClosed: true,
			},

			want: decimal.RequireFromString("100000"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.trade.CalculatePnl()

			if !got.Equal(tc.want) {
				t.Errorf("CalculatePnl() = %v; want to get %v", got, tc.want)
			}
		})
	}

}
