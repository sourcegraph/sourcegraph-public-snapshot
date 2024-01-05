package runtime

type ConfigLoader[ConfigT any] interface {
	*ConfigT

	// Load should populate ConfigT with values from env. Errors should be added
	// to env using env.AddError().
	Load(env *Env)
}
