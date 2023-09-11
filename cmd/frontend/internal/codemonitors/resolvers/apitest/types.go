// Package apitest provided types used in testing.
package apitest

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Response struct {
	User User
}

type UpdateCodeMonitorResponse struct {
	UpdateCodeMonitor Monitor
}

type User struct {
	Monitors MonitorConnection
}

type Node struct {
	Node Monitor
}

type MonitorConnection struct {
	Nodes      []Monitor
	TotalCount int
	PageInfo   PageInfo
}

type Monitor struct {
	Id          string
	Description string
	Enabled     bool
	Owner       UserOrg
	CreatedBy   UserOrg
	CreatedAt   string
	Trigger     Trigger
	Actions     ActionConnection
}

type UserOrg struct {
	Name string `json:"username"`
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   *string
}

type ActionConnection struct {
	Nodes      []Action
	TotalCount int
	PageInfo   PageInfo
}

type Action struct {
	Email        *ActionEmail
	Webhook      *ActionWebhook
	SlackWebhook *ActionSlackWebhook
}

func (a *Action) UnmarshalJSON(b []byte) error {
	type typeUnmarshaller struct {
		TypeName string `json:"__typename"`
	}
	var t typeUnmarshaller
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}

	switch t.TypeName {
	case "MonitorEmail":
		a.Email = &ActionEmail{}
		return json.Unmarshal(b, &a.Email)
	case "MonitorWebhook":
		a.Webhook = &ActionWebhook{}
		return json.Unmarshal(b, &a.Webhook)
	case "MonitorSlackWebhook":
		a.SlackWebhook = &ActionSlackWebhook{}
		return json.Unmarshal(b, &a.SlackWebhook)
	default:
		return errors.Errorf("unexpected typename %q", t.TypeName)
	}
}

type ActionEmail struct {
	Id         string
	Enabled    bool
	Priority   string
	Recipients RecipientsConnection
	Header     string
	Events     ActionEventConnection
}

type ActionWebhook struct {
	Id      string
	Enabled bool
	URL     string
	Events  ActionEventConnection
}

type ActionSlackWebhook struct {
	Id      string
	Enabled bool
	URL     string
	Events  ActionEventConnection
}

type RecipientsConnection struct {
	Nodes      []UserOrg
	TotalCount int
	PageInfo   PageInfo
}

type Trigger struct {
	Id     string
	Query  string
	Events TriggerEventConnection
}

type TriggerEventConnection struct {
	Nodes      []TriggerEvent
	TotalCount int
	PageInfo   PageInfo
}

type TriggerEvent struct {
	Id        string
	Status    string
	Timestamp string
	Message   *string
}

type ActionEventConnection struct {
	Nodes      []ActionEvent
	TotalCount int
	PageInfo   PageInfo
}

type ActionEvent struct {
	Id        string
	Status    string
	Timestamp string
	Message   *string
}
