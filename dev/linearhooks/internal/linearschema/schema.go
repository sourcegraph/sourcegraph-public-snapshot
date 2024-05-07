package linearschema

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Event represents a linear webhook event.
// Learn more from https://developers.linear.app/docs/graphql/webhooks#the-webhook-payload
type Event struct {
	Action           EventActionType `json:"action"`
	Type             EventEntityType `json:"type"`
	Url              string          `json:"url"`
	WebhookId        string          `json:"webhookId"`
	WebhookTimestamp int             `json:"webhookTimestamp"`
	Data             any             `json:"data"`

	IssueData *IssueData `json:"-"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type wrapper Event
	var v wrapper
	if err := json.Unmarshal(data, &v); err != nil {
		return errors.Wrap(err, "unmarshal type")

	}
	switch v.Type {
	case EventEntityTypeIssue:
		b, err := json.Marshal(v.Data)
		if err != nil {
			return errors.Wrap(err, "marshal raw issue data")
		}
		v.IssueData = &IssueData{}
		if err := json.Unmarshal(b, v.IssueData); err != nil {
			return errors.Wrap(err, "unmarshal issue data")
		}
	default:
		return errors.Newf("unsupported event type %q", v.Type)
	}

	*e = Event(v)
	return nil
}

type EventEntityType string

const (
	EventEntityTypeIssue EventEntityType = "Issue"
)

type EventActionType string

const (
	ActionTypeCreate EventActionType = "create"
	ActionTypeUpdate EventActionType = "update"
	ActionTypeDelete EventActionType = "remove"
)

type IssueData struct {
	ID          string           `json:"id"`
	Identifier  string           `json:"identifier"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Team        IssueTeamData    `json:"team"`
	Labels      []IssueLabelData `json:"labels"`
}

type IssueLabelData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IssueTeamData struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

func (d IssueData) LabelNames() []string {
	var names []string
	for _, l := range d.Labels {
		names = append(names, l.Name)
	}
	return names
}
