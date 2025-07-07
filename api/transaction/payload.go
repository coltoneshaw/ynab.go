package transaction

import (
	"github.com/coltoneshaw/ynab.go/api"
)

// PayloadTransaction is the payload contract for saving a transaction, new or existent
type PayloadTransaction struct {
	ID        string   `json:"id"`
	AccountID string   `json:"account_id"`
	Date      api.Date `json:"date"`
	// Amount The transaction amount in milliunits format
	Amount   int64          `json:"amount"`
	Cleared  ClearingStatus `json:"cleared"`
	Approved bool           `json:"approved"`

	// PayeeID Transfer payees are not permitted and will be ignored if supplied
	PayeeID *string `json:"payee_id"`
	// PayeeName If the payee name is provided and payee ID has a null value, the
	// payee name value will be used to resolve the payee by either (1) a matching
	// payee rename rule (only if import_id is also specified) or (2) a payee with
	// the same name or (3) creation of a new payee
	PayeeName *string `json:"payee_name"`
	// CategoryID Split and Credit Card Payment categories are not permitted and
	// will be ignored if supplied.
	CategoryID *string    `json:"category_id"`
	Memo       *string    `json:"memo"`
	FlagColor  *FlagColor `json:"flag_color"`
	// ImportID If the Transaction was imported, this field is a unique (by account) import
	// identifier. If this transaction was imported through File Based Import or
	// Direct Import and not through the API, the import_id will have the format:
	// 'YNAB:[milliunit_amount]:[iso_date]:[occurrence]'. For example, a transaction
	// dated 2015-12-30 in the amount of -$294.23 USD would have an import_id of
	// 'YNAB:-294230:2015-12-30:1’. If a second transaction on the same account
	// was imported and had the same date and same amount, its import_id would
	// be 'YNAB:-294230:2015-12-30:2’.
	ImportID *string `json:"import_id"`
}

// PayloadScheduledTransaction is the payload contract for saving a scheduled transaction, new or existent
type PayloadScheduledTransaction struct {
	AccountID string   `json:"account_id"`
	Date      api.Date `json:"date"`
	// Amount The scheduled transaction amount in milliunits format
	Amount    int64              `json:"amount"`
	Frequency ScheduledFrequency `json:"frequency"`

	// PayeeID Transfer payees are not permitted and will be ignored if supplied
	PayeeID *string `json:"payee_id"`
	// PayeeName If the payee name is provided and payee ID has a null value, the
	// payee name value will be used to resolve the payee by either (1) a matching
	// payee rename rule or (2) a payee with the same name or (3) creation of a new payee
	PayeeName *string `json:"payee_name"`
	// CategoryID Credit Card Payment categories are not permitted and will be ignored if supplied.
	CategoryID *string    `json:"category_id"`
	Memo       *string    `json:"memo"`
	FlagColor  *FlagColor `json:"flag_color"`
}
