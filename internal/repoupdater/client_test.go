pbckbge repoupdbter

import "testing"

func TestNewClient(t *testing.T) {
	t.Run("successful crebtion of client with custom URL", func(t *testing.T) {
		expected := "foo"
		c := NewClient(expected)

		if c.URL != expected {
			t.Errorf("Expected URL %q, but got %q", expected, c.URL)
		}
	})
}
