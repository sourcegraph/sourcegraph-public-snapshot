package refstate

import "github.com/sourcegraph/zap"

// RefState is all of the server state for a ref. The RefState value
// is typically obtained in the (refdb.Ref).Object field of a ref.
type RefState struct {
	// zap.RefState is the data that makes up a ref's persistent
	// state. It excludes ephemeral state that depends on
	// configuration or upstream server communication.
	zap.RefState

	Upstream *Upstream // remote upstream server connection state (NOTE: assumes at most 1 remote)
}

// Target returns r's target ref. If r is a non-symbolic ref, it
// returns "".
func (r RefState) Target() string { return r.RefState.Target }
