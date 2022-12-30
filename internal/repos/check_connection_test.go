package repos

import (
	"testing"
)

func TestDnsLookup(t *testing.T) {
	t.Run("bad URL", func(t *testing.T) {
		if err := dnsLookup("foo"); err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("good URL", func(t *testing.T) {
		if err := dnsLookup("https://sourcegraph.com"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL with port", func(t *testing.T) {
		if err := dnsLookup("https://sourcegraph.com:80"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL without protocol", func(t *testing.T) {
		if err := dnsLookup("sourcegraph.com"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL with port but without protocol", func(t *testing.T) {
		if err := dnsLookup("sourcegraph.com:80"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("good URL with username:password", func(t *testing.T) {
		if err := dnsLookup("https://username:password@sourcegraph.com"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})
}

func TestPing(t *testing.T) {
	t.Run("bad URL", func(t *testing.T) {
		if err := ping("foo"); err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("bad URL with non HTTP protocol", func(t *testing.T) {
		if err := ping("ftp://foo"); err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("hostname and port", func(t *testing.T) {
		if err := ping("sourcegraph.com:80"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("hostname without port", func(t *testing.T) {
		if err := ping("ghe.sgdev.org"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("HTTP hostname", func(t *testing.T) {
		if err := ping("http://ghe.sgdev.org"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("HTTP hostname and port", func(t *testing.T) {
		if err := ping("http://ghe.sgdev.org:80"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("HTTPS hostname", func(t *testing.T) {
		if err := ping("https://ghe.sgdev.org"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})

	t.Run("HTTPS hostname and port", func(t *testing.T) {
		if err := ping("https://ghe.sgdev.org:443"); err != nil {
			t.Errorf("Expected nil but got error: %v", err)
		}
	})
}
