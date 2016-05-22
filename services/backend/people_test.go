package backend

import (
	"strings"
	"testing"
)

func TestNewTransientPerson_invalidEmailStillYieldsObfuscatedLogin(t *testing.T) {
	u := newTransientPerson("invalidemail")
	if !strings.Contains(u.Email, "@-x-") {
		t.Errorf("got email %q, want to be obfuscated email", u.Email)
	}
}
