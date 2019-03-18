package auth

import "encoding/json"

// SetExternalAccountData sets the db.ExternalAccountData field to the (JSON-encoded) value. If an
// error occurs while JSON-encoding value, a JSON object describing the error is written to the
// field instead.
func SetExternalAccountData(field **json.RawMessage, value interface{}) {
	if value == nil {
		*field = nil
		return
	}

	b, err := json.Marshal(value)
	if err != nil {
		b, _ = json.Marshal(struct {
			Error string `json:"__jsonError"`
		}{Error: err.Error()})
	}
	*field = (*json.RawMessage)(&b)
}
