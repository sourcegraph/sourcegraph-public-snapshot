package alert

type ResponderType string

const (
	UserResponder       ResponderType = "user"
	TeamResponder       ResponderType = "team"
	EscalationResponder ResponderType = "escalation"
	ScheduleResponder   ResponderType = "schedule"
)

type Responder struct {
	Type     ResponderType `json:"type, omitempty"`
	Name     string        `json:"name,omitempty"`
	Id       string        `json:"id,omitempty"`
	Username string        `json:"username, omitempty"`
}

type Team struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type User struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

type Escalation struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Schedule struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (s *Schedule) SetID(id string) {
	s.ID = id
}

func (s *Schedule) SetUsername(name string) {
	s.Name = name
}
