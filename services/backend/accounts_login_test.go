package backend

import "testing"

func TestIsValidLogin(t *testing.T) {
	tests := map[string]bool{
		"":      false,
		"a":     false,
		"aa":    false,
		"aaa":   true,
		"admin": false,
		"Admin": false,
		"αdmin": false,
		"sqs":   true,
		"e2etestuserx4FF3register_flow12345678901234567890": true,
	}
	for login, valid := range tests {
		got := isValidLogin(login)
		if got != valid {
			t.Errorf("login %q: got valid == %v, want %v", login, got, valid)
		}
	}
}
