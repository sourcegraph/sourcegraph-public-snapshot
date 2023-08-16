package spec

type ServiceSpec struct {
	ID     string   `json:"id"`
	Owners []string `json:"owners"`
}
