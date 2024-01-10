package vcssyncer

import (
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

func TestP4DepotSyncer_p4CommandEnv(t *testing.T) {
	syncer := &perforceDepotSyncer{
		logger:                  logtest.Scoped(t),
		recordingCommandFactory: wrexec.NewNoOpRecordingCommandFactory(),
		P4Client:                "client",
		P4Home:                  "p4home",
	}
	vars := syncer.p4CommandEnv("host", "username", "password")
	assertEnv := func(key, value string) {
		var match string
		for _, s := range vars {
			parts := strings.SplitN(s, "=", 2)
			if len(parts) != 2 {
				t.Errorf("Expected 2 parts, got %d in %q", len(parts), s)
				continue
			}
			if parts[0] != key {
				continue
			}
			// Last match wins
			if parts[1] == value {
				match = parts[1]
			}
		}
		if match == "" {
			t.Errorf("No match found for %q", key)
		} else if match != value {
			t.Errorf("Want %q, got %q", value, match)
		}
	}
	assertEnv("HOME", "p4home")
	assertEnv("P4CLIENT", "client")
	assertEnv("P4PORT", "host")
	assertEnv("P4USER", "username")
	assertEnv("P4PASSWD", "password")
}
