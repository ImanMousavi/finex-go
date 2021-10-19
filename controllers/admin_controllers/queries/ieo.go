package queries

import (
	"github.com/shopspring/decimal"
	"github.com/zsmartex/finex/models/datatypes"
	"github.com/zsmartex/finex/types"
)

type IEOPayload struct {
	ID                  uint64                         `json:"id"`
	CurrencyID          string                         `json:"currency_id"`
	MainPaymentCurrency string                         `json:"main_payment_currency"`
	Price               decimal.Decimal                `json:"price"`
	PaymentCurrencies   datatypes.IEOPaymentCurrencies `json:"payment_currencies"`
	MinAmount           decimal.Decimal                `json:"min_amount"`
	State               types.MarketState              `json:"state"`
	StartTime           int64                          `json:"start_time"`
	EndTime             int64                          `json:"end_time"`
}