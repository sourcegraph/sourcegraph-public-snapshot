package audittest

import "github.com/sourcegraph/log/logtest"

type AuditFields struct {
	Action string
	Entity string
}

// ExtractAuditFields retrieves some high-level audit log fields for assertions
// on captured log entries (using logtest.Captured(...))
func ExtractAuditFields(entry logtest.CapturedLog) (*AuditFields, bool) {
	f := entry.Fields["audit"]
	if f == nil {
		return nil, false
	}
	m, ok := f.(map[string]any)
	if m == nil || !ok {
		return nil, false
	}
	return &AuditFields{
		Entity: m["entity"].(string),
		Action: m["action"].(string),
	}, true
}
