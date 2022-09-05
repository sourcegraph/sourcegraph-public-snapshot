package buildkite

type featureFlags struct {
}

// FeatureFlags are for experimenting with CI pipeline features. Use sparingly!
var FeatureFlags = featureFlags{}

func (f *featureFlags) ApplyEnv(env map[string]string) {
}
