package spec

// Spec is a Managed Services Platform (MSP) service.
//
// All MSP services must:
//
//   - Serve its API on ":$PORT", if $PORT is provided
//   - Export a /-/healthz endpoint that authenticates requests using
//     "Authorization: Bearer $DIAGNOSTICS_SECRET", if $DIAGNOSTICS_SECRET is provided.
type Spec struct {
	Service      ServiceSpec       `json:"service"`
	Build        BuildSpec         `json:"build"`
	Environments []EnvironmentSpec `json:"environments"`
}

func (s Spec) GetEnvironment(name string) *EnvironmentSpec {
	for _, e := range s.Environments {
		if e.ID == name {
			return &e
		}
	}
	return nil
}
