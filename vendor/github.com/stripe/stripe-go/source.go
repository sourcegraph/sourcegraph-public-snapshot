package stripe

import (
	"encoding/json"
	"fmt"
)

// SourceStatus represents the possible statuses of a source object.
type SourceStatus string

const (
	// SourceStatusPending the source is freshly created and not yet
	// chargeable. The flow should indicate how to authenticate it with your
	// customer.
	SourceStatusPending SourceStatus = "pending"
	// SourceStatusChargeable the source is ready to be charged (once if usage
	// is `single-use`, repeatidly otherwise).
	SourceStatusChargeable SourceStatus = "chargeable"
	// SourceStatusConsumed the source is `single-use` usage and has been
	// charged already.
	SourceStatusConsumed SourceStatus = "consumed"
	// SourceStatusFailed the source is no longer usable.
	SourceStatusFailed SourceStatus = "failed"
	// SourceStatusCanceled we canceled the source along with any side-effect
	// it had (returned funds to customers if any were sent).
	SourceStatusCanceled SourceStatus = "canceled"
)

// SourceFlow represents the possible flows of a source object.
type SourceFlow string

const (
	// FlowRedirect a redirect is required to authenticate the source.
	FlowRedirect SourceFlow = "redirect"
	// FlowReceiver a receiver address should be communicated to the customer
	// to push funds to it.
	FlowReceiver SourceFlow = "receiver"
	// FlowVerification a verification code should be communicated by the
	// customer to authenticate the source.
	FlowVerification SourceFlow = "verification"
	// FlowNone no particular authentication is involved the source should
	// become chargeable directly or asyncrhonously.
	FlowNone SourceFlow = "none"
)

// SourceUsage represents the possible usages of a source object.
type SourceUsage string

const (
	// UsageSingleUse the source can only be charged once for the specified
	// amount and currency.
	UsageSingleUse SourceUsage = "single-use"
	// UsageReusable the source can be charged multiple times for arbitrary
	// amounts.
	UsageReusable SourceUsage = "reusable"
)

type SourceOwnerParams struct {
	Address *AddressParams
	Email   string
	Name    string
	Phone   string
}

type RedirectParams struct {
	ReturnURL string
}

type SourceObjectParams struct {
	Params
	Type     string
	Amount   uint64
	Currency Currency
	Flow     SourceFlow
	Owner    *SourceOwnerParams

	Redirect *RedirectParams

	TypeData map[string]string
}

type SourceOwner struct {
	Address         *Address `json:"address,omitempty"`
	Email           string   `json:"email"`
	Name            string   `json:"name"`
	Phone           string   `json:"phone"`
	VerifiedAddress *Address `json:"verified_address,omitempty"`
	VerifiedEmail   string   `json:"verified_email"`
	VerifiedName    string   `json:"verified_name"`
	VerifiedPhone   string   `json:"verified_phone"`
}

// RedirectFlowStatus represents the possible statuses of a redirect flow.
type RedirectFlowStatus string

const (
	RedirectFlowStatusPending   RedirectFlowStatus = "pending"
	RedirectFlowStatusSucceeded RedirectFlowStatus = "succeeded"
	RedirectFlowStatusFailed    RedirectFlowStatus = "failed"
)

// ReceiverFlow informs of the state of a redirect authentication flow.
type RedirectFlow struct {
	URL       string             `json:"url"`
	ReturnURL string             `json:"return_url"`
	Status    RedirectFlowStatus `json:"status"`
}

// RefundAttributesStatus are the possible status of a receiver's refund
// attributes.
type RefundAttributesStatus string

const (
	// RefundAttributesAvailable the refund attributes are available
	RefundAttributesAvailable RefundAttributesStatus = "available"
	// RefundAttributesRequested the refund attributes have been requested
	RefundAttributesRequested RefundAttributesStatus = "requested"
	// RefundAttributesMissing the refund attributes are missing
	RefundAttributesMissing RefundAttributesStatus = "missing"
)

// RefundAttributesMethod are the possible method to retrieve a receiver's
// refund attributes.
type RefundAttributesMethod string

const (
	// RefundAttributesEmail the refund attributes are automatically collected over email
	RefundAttributesEmail RefundAttributesMethod = "email"
	// RefundAttributesManual the refund attributes should be collected by the user
	RefundAttributesManual RefundAttributesMethod = "manual"
)

// ReceiverFlow informs of the state of a receiver authentication flow.
type ReceiverFlow struct {
	RefundAttributesStatus RefundAttributesStatus `json:"refund_attributes_status"`
	RefundAttributesMethod RefundAttributesMethod `json:"refund_attributes_method"`
	Address                string                 `json:"address"`
	AmountReceived         int64                  `json:"amount_received"`
	AmountReturned         int64                  `json:"amount_returned"`
	AmountCharged          int64                  `json:"amount_charged"`
}

// VerificationFlowStatus represents the possible statuses of a verification
// flow.
type VerificationFlowStatus string

const (
	VerificationFlowStatusPending   VerificationFlowStatus = "pending"
	VerificationFlowStatusSucceeded VerificationFlowStatus = "succeeded"
	VerificationFlowStatusFailed    VerificationFlowStatus = "failed"
)

// ReceiverFlow informs of the state of a verification authentication flow.
type VerificationFlow struct {
	AttemptsRemaining uint64             `json:"attempts_remaining"`
	Status            RedirectFlowStatus `json:"status"`
}

type Source struct {
	ID           string       `json:"id"`
	Amount       int64        `json:"amount"`
	ClientSecret string       `json:"client_secret"`
	Created      int64        `json:"created"`
	Currency     Currency     `json:"currency"`
	Flow         SourceFlow   `json:"flow"`
	Status       SourceStatus `json:"status"`
	Type         string       `json:"type"`
	Usage        SourceUsage  `json:"usage"`

	Live bool              `json:"livemode"`
	Meta map[string]string `json:"metadata"`

	Owner        SourceOwner       `json:"owner"`
	Redirect     *RedirectFlow     `json:"redirect,omitempty"`
	Receiver     *ReceiverFlow     `json:"receiver,omitempty"`
	Verification *VerificationFlow `json:"verification,omitempty"`

	TypeData map[string]interface{}
}

// Display human readable representation of a Source.
func (s *Source) Display() string {
	var status string
	switch s.Status {
	case SourceStatusPending:
		status = "Pending"
	case SourceStatusChargeable:
		status = "Chargeable"
	case SourceStatusConsumed:
		status = "Consumed"
	case SourceStatusFailed:
		status = "Failed"
	case SourceStatusCanceled:
		status = "Canceled"
	}

	desc := fmt.Sprintf("%s %s source", status, s.Type)
	if s.Amount > 0 {
		desc += fmt.Sprintf(" (%d %s)", s.Amount, s.Currency)
	}
	return desc
}

// UnmarshalJSON handles deserialization of an Source. This custom unmarshaling
// is needed to extract the type specific data (accessible under `TypeData`)
// but stored in JSON under a hash named after the `type` of the source.
func (s *Source) UnmarshalJSON(data []byte) error {
	type source Source
	var ss source
	err := json.Unmarshal(data, &ss)
	if err != nil {
		return err
	}
	*s = Source(ss)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	if d, ok := raw[s.Type]; ok {
		if m, ok := d.(map[string]interface{}); ok {
			s.TypeData = m
		}
	}

	return nil
}
