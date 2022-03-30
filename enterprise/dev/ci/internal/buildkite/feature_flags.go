package buildkite

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

type featureFlags struct {
	// StatelessBuild triggers a stateless build by overriding the default queue to send the build on the stateless
	// agents and forces a MainDryRun type build to avoid impacting normal builds.
	//
	// It is meant to test the stateless builds without any side effects.
	StatelessBuild bool
}

// FeatureFlags are for experimenting with CI pipeline features. Use sparingly!
var FeatureFlags = featureFlags{
	StatelessBuild: os.Getenv("CI_FEATURE_FLAG_STATELESS") == "true" ||
		// Always process retries on stateless agents.
		// TODO: remove when we switch over entirely to stateless agents
		os.Getenv("BUILDKITE_REBUILT_FROM_BUILD_NUMBER") != "" ||
		// Roll out to 50% of builds
		rand.NewSource(time.Now().UnixNano()).Int63()%100 < 50,
}

func (f *featureFlags) ApplyEnv(env map[string]string) {
	env["CI_FEATURE_FLAG_STATELESS_BUILD"] = fmt.Sprintf("%v", f.StatelessBuild)
}
