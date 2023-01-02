package rbac

type Schema struct {
	Namespaces []Namespace `json:"namespaces"`
}

type Namespace struct {
	Name         string   `json:"name"`
	Actions      []string `json:"actions"`
	DefaultApply bool     `json:"defaultApply"`
}
