package repos

import "testing"

func TestCheckConnection(t *testing.T) {
	t.Run("bad URL", func(t *testing.T) {
		if err := checkConnection("foo"); err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("good URL", func(t *testing.T) {
		if err := checkConnection("https://sourcegraph.com"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL with port", func(t *testing.T) {
		if err := checkConnection("https://sourcegraph.com:80"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL without protocol", func(t *testing.T) {
		if err := checkConnection("sourcegraph.com"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL with port but without protocol", func(t *testing.T) {
		if err := checkConnection("sourcegraph.com:80"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL with username:password", func(t *testing.T) {
		if err := checkConnection("https://username:password@sourcegraph.com"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})
}
