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
				Quantity:   10,
				PointValue: decimal.RequireFromString("50"),
				FeeCents:   Money{Currency: "USD", Amount: decimal.RequireFromString("250")},
				IsLong:     true,
			},
			want: decimal.RequireFromString("4750"),
		},

		{
			name: "Loss Short for 2 contracts",
			trade: Trade{
				EntryPrice: decimal.RequireFromString("100"),
				ExitPrice:  decimal.RequireFromString("105"),
				Quantity:   20,
				PointValue: decimal.RequireFromString("10"),
				FeeCents:   Money{Currency: "USD", Amount: decimal.RequireFromString("500")},
				IsLong:     false,
			},

			want: decimal.RequireFromString("-1500"),
		},

		{
			name: "Small price change with large quantity",
			trade: Trade{
				EntryPrice: decimal.RequireFromString("1000000"),
				ExitPrice:  decimal.RequireFromString("1000001"),
				Quantity:   2000,
				PointValue: decimal.RequireFromString("50"),
				FeeCents:   Money{Currency: "USD", Amount: decimal.RequireFromString("0")},
				IsLong:     true,
			},

			want: decimal.RequireFromString("100000"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.trade.CalculatePnl()

			if !got.Equal(tc.want) {
				t.Errorf("CalculatePnl() = %d; want to get %d", got, tc.want)
			}
		})
	}

}
