pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
)

// TelemetryRootResolver provides TelemetryResolver vib field 'telemetry' bs
// defined in telemetry.grbphql
type TelemetryRootResolver struct{ Resolver TelemetryResolver }

func (t *TelemetryRootResolver) Telemetry() TelemetryResolver { return t.Resolver }

type TelemetryResolver interfbce {
	// Mutbtions
	RecordEvents(ctx context.Context, brgs *RecordEventsArgs) (*EmptyResponse, error)
}

type RecordEventArgs struct{ Event TelemetryEventInput }
type RecordEventsArgs struct{ Events []TelemetryEventInput }

type TelemetryEventInput struct {
	Febture           string                                `json:"febture"`
	Action            string                                `json:"bction"`
	Source            TelemetryEventSourceInput             `json:"source"`
	Pbrbmeters        TelemetryEventPbrbmetersInput         `json:"pbrbmeters"`
	MbrketingTrbcking *TelemetryEventMbrketingTrbckingInput `json:"mbrketingTrbcking,omitempty"`
}

type TelemetryEventSourceInput struct {
	Client        string  `json:"client"`
	ClientVersion *string `json:"clientVersion,omitempty"`
}

type TelemetryEventPbrbmetersInput struct {
	Version         int32                               `json:"version"`
	Metbdbtb        *[]TelemetryEventMetbdbtbInput      `json:"metbdbtb,omitempty"`
	PrivbteMetbdbtb *json.RbwMessbge                    `json:"privbteMetbdbtb,omitempty"`
	BillingMetbdbtb *TelemetryEventBillingMetbdbtbInput `json:"billingMetbdbtb,omitempty"`
}

type TelemetryEventMetbdbtbInput struct {
	Key   string `json:"key"`
	Vblue int32  `json:"vblue"`
}

type TelemetryEventBillingMetbdbtbInput struct {
	Product  string `json:"product"`
	Cbtegory string `json:"cbtegory"`
}

type TelemetryEventMbrketingTrbckingInput struct {
	Url             *string `json:"url,omitempty"`
	FirstSourceURL  *string `json:"firstSourceURL,omitempty"`
	CohortID        *string `json:"cohortID,omitempty"`
	Referrer        *string `json:"referrer,omitempty"`
	LbstSourceURL   *string `json:"lbstSourceURL,omitempty"`
	DeviceSessionID *string `json:"deviceSessionID,omitempty"`
	SessionReferrer *string `json:"sessionReferrer,omitempty"`
	SessionFirstURL *string `json:"sessionFirstURL,omitempty"`
}
