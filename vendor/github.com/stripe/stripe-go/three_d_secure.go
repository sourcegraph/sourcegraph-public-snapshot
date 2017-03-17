package stripe

type ThreeDSecureStatus string

// ThreeDSecureParams is the set of parameters that can be used when creating a 3DS object.
// For more details see https://stripe.com/docs/api#create_three_d_secure.
type ThreeDSecureParams struct {
	Params
	Amount    uint64   `json:"amount"`
	Card      string   `json:"card"`
	Currency  Currency `json:"currency"`
	Customer  string   `json:"customer"`
	ReturnURL string   `json:"return_url"`
}

// ThreeDSecure is the resource representing a Stripe 3DS object
// For more details see https://stripe.com/docs/api#three_d_secure.
type ThreeDSecure struct {
	Amount        uint64             `json:"amount"`
	Authenticated bool               `json:"authenticated"`
	Card          *Card              `json:"card"`
	Currency      Currency           `json:"currency"`
	Created       int64              `json:"created"`
	ID            string             `json:"id"`
	Live          bool               `json:"livemode"`
	RedirectURL   string             `json:"redirect_url"`
	Supported     string             `json:"supported"`
	Status        ThreeDSecureStatus `json:"status"`
}
