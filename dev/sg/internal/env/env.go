package env

type Priority int

const (
	GlobalEnvPriority Priority = iota
	BaseCommandsetEnvPriority
	BaseCommandEnvPriority
	OverrideCommandsetEnvPriority
	OverrideCommandEnvPriority
)

type EnvVar struct {
	Name  string
	Value string

	Priority Priority
}

//func MergeEnvs(... map[string])
