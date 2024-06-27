package shared

import (
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	config := Config{}
	// Assume nothing is set explicitly in the env.
	config.SetMockGetter(mapGetter(nil))
	config.Load()

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %s", err)
	}

	if have, want := config.ReposDir, "/data/repos"; have != want {
		t.Errorf("invalid value for ReposDir: have=%s want=%s", have, want)
	}
	if have, want := config.CoursierCacheDir, "/data/repos/coursier"; have != want {
		t.Errorf("invalid value for CoursierCacheDir: have=%s want=%s", have, want)
	}
	if have, want := config.SyncRepoStateInterval, 10*time.Minute; have != want {
		t.Errorf("invalid value for SyncRepoStateInterval: have=%s want=%s", have, want)
	}
	if have, want := config.SyncRepoStateBatchSize, 500; have != want {
		t.Errorf("invalid value for SyncRepoStateBatchSize: have=%d want=%d", have, want)
	}
	if have, want := config.SyncRepoStateUpdatePerSecond, 500; have != want {
		t.Errorf("invalid value for SyncRepoStateUpdatePerSecond: have=%d want=%d", have, want)
	}
	if have, want := config.JanitorInterval, time.Minute; have != want {
		t.Errorf("invalid value for JanitorInterval: have=%s want=%s", have, want)
	}
	if have, want := config.JanitorDisableDeleteReposOnWrongShard, false; have != want {
		t.Errorf("invalid value for JanitorDisableDeleteReposOnWrongShard: have=%t want=%t", have, want)
	}
}

func mapGetter(env map[string]string) func(name, defaultValue, description string) string {
	return func(name, defaultValue, description string) string {
		if v, ok := env[name]; ok {
			return v
		}

		return defaultValue
	}
}
