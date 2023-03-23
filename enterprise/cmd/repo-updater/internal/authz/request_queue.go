package authz

import (
	"strconv"
)

// requestType is the type of the permissions syncing request. It defines the
// permissions syncing is either repository-centric or user-centric.
type requestType int

// A list of request types, the larger the value, the higher the priority.
// requestTypeUser had the highest because it is often triggered by a user
// action (e.g. sign up, log in).
const (
	requestTypeRepo requestType = iota + 1
	requestTypeUser
)

func (t requestType) String() string {
	switch t {
	case requestTypeRepo:
		return "repo"
	case requestTypeUser:
		return "user"
	}
	return strconv.Itoa(int(t))
}

// higherPriorityThan returns true if the current request type has higher priority
// than the other one.
func (t requestType) higherPriorityThan(t2 requestType) bool {
	return t > t2
}
