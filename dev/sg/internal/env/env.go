package env

type Priority int

const (
	GlobalEnvPriority Priority = iota
	BaseCommandsetEnvPriority
	BaseCommandEnvPriority
	SecretEnvPriority
	OverrideGlobalEnvPriority
	OverrideCommandsetEnvPriority
	OverrideCommandEnvPriority
)

type EnvVar struct {
	Name  string
	Value string

	Priority Priority
}

func (e EnvVar) GetValue() string {
	return e.Value
}

func New(name, value string, priority Priority) EnvVar {
	return EnvVar{
		Name:     name,
		Value:    value,
		Priority: priority,
	}
}

// ConvertEnvMap converts a map of strings to map[string]EnvVar.
func ConvertEnvMap(ev map[string]string, priority Priority) map[string]EnvVar {
	em := make(map[string]EnvVar, len(ev))
	for name, val := range ev {
		em[name] = New(name, val, priority)
	}
	return em
}

func ConvertToMap(ev map[string]EnvVar) map[string]string {
	em := make(map[string]string, len(ev))
	for name, val := range ev {
		em[name] = val.Value
	}
	return em
}

// CompareByPriority returns the EnvVar with the higher priority.
func CompareByPriority(ev1 EnvVar, ev2 EnvVar) EnvVar {
	if ev1.Priority > ev2.Priority {
		return ev1
	}
	return ev2
}

// Merge merges two env maps, prioritizing the second map if there is a conflict.
func Merge(ev1 map[string]EnvVar, ev2 map[string]EnvVar) map[string]EnvVar {
	if ev1 == nil {
		return ev2
	}
	for k := range ev2 {
		ev1[k] = CompareByPriority(ev1[k], ev2[k])
	}
	return ev1
}
