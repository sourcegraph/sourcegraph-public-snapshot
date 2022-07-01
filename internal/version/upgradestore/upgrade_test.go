package upgradestore

import (
	"testing"

	"github.com/Masterminds/semver"
)

func TestIsValidUpgrade(t *testing.T) {
	for _, tc := range []struct {
		name     string
		previous string
		latest   string
		want     bool
	}{
		{
			name:     "no versions",
			previous: "",
			latest:   "",
			want:     true,
		}, {
			name:     "no previous version",
			previous: "",
			latest:   "v3.13.0",
			want:     true,
		}, {
			name:     "no latest version",
			previous: "v3.13.0",
			latest:   "",
			want:     true,
		}, {
			name:     "same version",
			previous: "v3.13.0",
			latest:   "v3.13.0",
			want:     true,
		}, {
			name:     "one minor version up",
			previous: "v3.12.4",
			latest:   "v3.13.1",
			want:     true,
		}, {
			name:     "one patch version up",
			previous: "v3.12.4",
			latest:   "v3.12.5",
			want:     true,
		}, {
			name:     "two patch versions up",
			previous: "v3.12.4",
			latest:   "v3.12.6",
			want:     true,
		}, {
			name:     "one major version up",
			previous: "v3.13.1",
			latest:   "v4.0.0",
			want:     true,
		}, {
			name:     "more than one minor version up",
			previous: "v3.9.4",
			latest:   "v3.11.0",
			want:     false,
		}, {
			name:     "major jump",
			previous: "v3.9.4",
			latest:   "v4.1.0",
			want:     false,
		}, {
			name:     "major rollback",
			previous: "v4.1.0",
			latest:   "v3.9.4",
			want:     true,
		}, {
			name:     "minor rollback",
			previous: "v4.1.0",
			latest:   "v4.0.4",
			want:     true,
		}, {
			name:     "patch rollback",
			previous: "v4.1.4",
			latest:   "v4.1.3",
			want:     true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous, _ := semver.NewVersion(tc.previous)
			latest, _ := semver.NewVersion(tc.latest)

			if got := IsValidUpgrade(previous, latest); got != tc.want {
				t.Errorf(
					"IsValidUpgrade(previous: %s, latest: %s) = %t, want %t",
					tc.previous,
					tc.latest,
					got,
					tc.want,
				)
			}
		})
	}
}
