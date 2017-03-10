package stripe

import "encoding/json"

// TransactionStatus is the list of allowed values for the transaction's status.
// Allowed values are "available", "pending".
type TransactionStatus string

// TransactionType is the list of allowed values for the transaction's type.
// Allowed values are "charge", "refund", "adjustment", "application_fee",
// "application_fee_refund", "transfer", "transfer_cancel", "transfer_failure".
type TransactionType string

// TransactionSourceType consts represent valid balance transaction sources.
type TransactionSourceType string

const (
	// TransactionSourceCharge is a constant representing a transaction source of charge
	TransactionSourceCharge TransactionSourceType = "charge"

	// TransactionSourceDispute is a constant representing a transaction source of dispute
	TransactionSourceDispute TransactionSourceType = "dispute"

	// TransactionSourceFee is a constant representing a transaction source of application_fee
	TransactionSourceFee TransactionSourceType = "application_fee"

	// TransactionSourceRefund is a constant representing a transaction source of refund
	TransactionSourceRefund TransactionSourceType = "refund"

	// TransactionSourceReversal is a constant representing a transaction source of reversal
	TransactionSourceReversal TransactionSourceType = "reversal"

	// TransactionSourceTransfer is a constant representing a transaction source of transfer
	TransactionSourceTransfer TransactionSourceType = "transfer"
)

// TransactionSource describes the source of a balance Transaction.
// The Type should indicate which object is fleshed out.
// For more details see https://stripe.com/docs/api#retrieve_balance_transaction
type TransactionSource struct {
	Type     TransactionSourceType `json:"object"`
	ID       string                `json:"id"`
	Charge   *Charge               `json:"-"`
	Dispute  *Dispute              `json:"-"`
	Fee      *Fee                  `json:"-"`
	Refund   *Refund               `json:"-"`
	Reversal *Reversal             `json:"-"`
	Transfer *Transfer             `json:"-"`
}

// BalanceParams is the set of parameters that can be used when retrieving a balance.
// For more details see https://stripe.com/docs/api#balance.
type BalanceParams struct {
	Params
}

// TxParams is the set of parameters that can be used when retrieving a transaction.
// For more details see https://stripe.com/docs/api#retrieve_balance_transaction.
type TxParams struct {
	Params
}

// TxListParams is the set of parameters that can be used when listing balance transactions.
// For more details see https://stripe.com/docs/api/#balance_history.
type TxListParams struct {
	ListParams
	Created, Available      int64
	Currency, Src, Transfer string
	Type                    TransactionType
}

// Balance is the resource representing your Stripe balance.
// For more details see https://stripe.com/docs/api/#balance.
type Balance struct {
	// Live indicates the live mode.
	Live      bool     `json:"livemode"`
	Available []Amount `json:"available"`
	Pending   []Amount `json:"pending"`
}

// Transaction is the resource representing the balance transaction.
// For more details see https://stripe.com/docs/api/#balance.
type Transaction struct {
	ID         string            `json:"id"`
	Amount     int64             `json:"amount"`
	Currency   Currency          `json:"currency"`
	Available  int64             `json:"available_on"`
	Created    int64             `json:"created"`
	Fee        int64             `json:"fee"`
	FeeDetails []TxFee           `json:"fee_details"`
	Net        int64             `json:"net"`
	Status     TransactionStatus `json:"status"`
	Type       TransactionType   `json:"type"`
	Desc       string            `json:"description"`
	Src        TransactionSource `json:"source"`
	Recipient  string            `json:"recipient"`
}

// TransactionList is a list of transactions as returned from a list endpoint.
type TransactionList struct {
	ListMeta
	Values []*Transaction `json:"data"`
}

// Amount is a structure wrapping an amount value and its currency.
type Amount struct {
	Value    int64    `json:"amount"`
	Currency Currency `json:"currency"`
}

// TxFee is a structure that breaks down the fees in a transaction.
type TxFee struct {
	Amount      int64    `json:"amount"`
	Currency    Currency `json:"currency"`
	Type        string   `json:"type"`
	Desc        string   `json:"description"`
	Application string   `json:"application"`
}

// UnmarshalJSON handles deserialization of a Transaction.
// This custom unmarshaling is needed because the resulting
// property may be an id or the full struct if it was expanded.
func (t *Transaction) UnmarshalJSON(data []byte) error {
	type transaction Transaction
	var tt transaction
	err := json.Unmarshal(data, &tt)
	if err == nil {
		*t = Transaction(tt)
	} else {
		// the id is surrounded by "\" characters, so strip them
		t.ID = string(data[1 : len(data)-1])
	}

	return nil
}

// UnmarshalJSON handles deserialization of a TransactionSource.
// This custom unmarshaling is needed because the specific
// type of transaction source it refers to is specified in the JSON
func (s *TransactionSource) UnmarshalJSON(data []byte) error {
	type source TransactionSource
	var ss source
	err := json.Unmarshal(data, &ss)
	if err == nil {
		*s = TransactionSource(ss)

		switch s.Type {
		case TransactionSourceCharge:
			json.Unmarshal(data, &s.Charge)
		case TransactionSourceDispute:
			json.Unmarshal(data, &s.Dispute)
		case TransactionSourceFee:
			json.Unmarshal(data, &s.Fee)
		case TransactionSourceRefund:
			json.Unmarshal(data, &s.Refund)
		case TransactionSourceReversal:
			json.Unmarshal(data, &s.Reversal)
		case TransactionSourceTransfer:
			json.Unmarshal(data, &s.Transfer)
		}
	} else {
		// the id is surrounded by "\" characters, so strip them
		s.ID = string(data[1 : len(data)-1])
	}

	return nil
}

// MarshalJSON handles serialization of a TransactionSource.
func (s *TransactionSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ID)
}
