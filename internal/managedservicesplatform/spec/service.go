package spec

type ServiceSpec struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Owners []string `json:"owners"`
}
