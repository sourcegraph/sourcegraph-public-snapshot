package entities

// ConduitMethod is a conduit method representation returned by
// `conduit.query`.
type ConduitMethod struct {
	Description string      `json:"description"`
	Params      interface{} `json:"params"`
	Return      string      `json:"return"`
}
