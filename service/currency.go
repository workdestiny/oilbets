package service

import (
	humanize "github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
)

// Currency is convert number to currency
func Currency(d decimal.Decimal) string {
	currency, _ := d.Truncate(2).Float64()
	return humanize.FormatFloat("#,###.##", currency)
}
