package clearbitutil

import "testing"

func TestNewClientWithAPIKey(t *testing.T) {
	clearbitAPIKey = "sk_12345"
	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("expected Clearbit client")
	}
}

func TestNewClientWithoutAPIKey(t *testing.T) {
	clearbitAPIKey = ""
	c, err := NewClient()
	if err != errNoAPIKey {
		t.Fatalf("expected %s got %s", errNoAPIKey, err)
	}
	if c != nil {
		t.Fatal("expected nil client")
	}
}
