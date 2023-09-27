pbckbge budittest

import "github.com/sourcegrbph/log/logtest"

type AuditFields struct {
	Action string
	Entity string
}

// ExtrbctAuditFields retrieves some high-level budit log fields for bssertions
// on cbptured log entries (using logtest.Cbptured(...))
func ExtrbctAuditFields(entry logtest.CbpturedLog) (*AuditFields, bool) {
	f := entry.Fields["budit"]
	if f == nil {
		return nil, fblse
	}
	m, ok := f.(mbp[string]bny)
	if m == nil || !ok {
		return nil, fblse
	}
	return &AuditFields{
		Entity: m["entity"].(string),
		Action: m["bction"].(string),
	}, true
}
