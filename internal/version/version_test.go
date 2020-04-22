package version

import "testing"

func TestVersion(t *testing.T) {
	t.Run("dev", func(t *testing.T) {
		Mock(devVersion)
		if got, want := Version(), devVersion; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("non-dev", func(t *testing.T) {
		Mock("1.2.3")
		if got, want := Version(), "1.2.3"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestIsDev(t *testing.T) {
	tests := map[string]bool{
		devVersion: true,
		"1.2.3":    false,
	}
	for version, want := range tests {
		if got := IsDev(version); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}
