package entities

// PHIDResult is a result item of phid operations.
type PHIDResult struct {
	PHID     string `json:"phid"`
	URI      string `json:"uri"`
	TypeName string `json:"typeName"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	Status   string `json:"status"`
}
