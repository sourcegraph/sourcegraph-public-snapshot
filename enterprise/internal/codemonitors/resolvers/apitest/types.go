// Package apitest provided types used in testing.
package apitest

type Response struct {
	User User
}

type UpdateCodeMonitorResponse struct {
	UpdateCodeMonitor Monitor
}

type User struct {
	Monitors MonitorConnection
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
	Name string `json:"username" json:"name"`
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
	ActionEmail
}

type ActionEmail struct {
	Id         string
	Enabled    bool
	Priority   string
	Recipients RecipientsConnection
	Header     string
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
