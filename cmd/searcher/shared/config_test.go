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

	if have, want := config.ListenAddress, ":3181"; have != want {
		t.Errorf("invalid value for ListenAddress: have=%s want=%s", have, want)
	}
	if have, want := config.CacheDir, "/tmp"; have != want {
		t.Errorf("invalid value for CacheDir: have=%s want=%s", have, want)
	}
	if have, want := config.CacheSizeMB, 100_000; have != want {
		t.Errorf("invalid value for CacheSizeMB: have=%d want=%d", have, want)
	}
	if have, want := config.BackgroundTimeout, 2*time.Hour; have != want {
		t.Errorf("invalid value for BackgroundTimeout: have=%s want=%s", have, want)
	}
	if have, want := config.MaxTotalGitArchivePathsLength, 100_000; have != want {
		t.Errorf("invalid value for MaxTotalGitArchivePathsLength: have=%d want=%d", have, want)
	}
	if have, want := config.DisableHybridSearch, false; have != want {
		t.Errorf("invalid value for DisableHybridSearch: have=%t want=%t", have, want)
	}
	if have, want := config.ExhaustiveRequestLoggingEnabled, false; have != want {
		t.Errorf("invalid value for ExhaustiveRequestLoggingEnabled: have=%t want=%t", have, want)
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
