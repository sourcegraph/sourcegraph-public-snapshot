package spec

// Spec is a Managed Services Platform service.
type Spec struct {
	Service      ServiceSpec       `json:"service"`
	Build        BuildSpec         `json:"build"`
	Environments []EnvironmentSpec `json:"environments"`
}

func (s Spec) GetEnvironment(name string) *EnvironmentSpec {
	for _, e := range s.Environments {
		if e.Name == name {
			return &e
		}
	}
	return nil
}
