package graphqlbackend

type threadType string

const (
	threadTypeThread threadType = "THREAD"
	threadTypeCheck             = "CHECK"
)

type threadStatus string

const (
	threadStatusOpenActive threadStatus = "OPEN_ACTIVE"
	threadStatusInactive                = "INACTIVE"
	threadStatusClosed                  = "CLOSED"
)
