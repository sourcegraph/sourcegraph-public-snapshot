package nosurf

import (
	"bytes"
	"errors"
	"testing"
)

func TestSetsReasonCorrectly(t *testing.T) {
	req := dummyGet()

	// set token first, as it's required for ctxSetReason
	ctxSetToken(req, []byte("abcdef"))

	err := errors.New("universe imploded")
	ctxSetReason(req, err)

	got := contextMap[req].reason

	if got != err {
		t.Errorf("Reason set incorrectly: expected %v, got %v", err, got)
	}
}

func TestSettingReasonFailsWithoutContext(t *testing.T) {
	req := dummyGet()
	err := errors.New("universe imploded")

	defer func() {
		r := recover()
		if r == nil {
			t.Error("ctxSetReason() didn't panic on no context")
		}
	}()

	ctxSetReason(req, err)
}

func TestSetsTokenCorrectly(t *testing.T) {
	req := dummyGet()
	token := []byte("12345678901234567890123456789012")
	ctxSetToken(req, token)

	got := contextMap[req].token

	if !bytes.Equal(token, unmaskToken(b64decode(got))) {
		t.Errorf("Token set incorrectly: expected %v, got %v", token, got)
	}
}

func TestGetsTokenCorrectly(t *testing.T) {
	req := dummyGet()
	token := Token(req)

	if len(token) != 0 {
		t.Errorf("Token hasn't been set yet, but it's not an empty slice, it's %v", token)
	}

	intended := []byte("12345678901234567890123456789012")
	ctxSetToken(req, intended)

	token = Token(req)
	decToken := unmaskToken(b64decode(token))
	if !bytes.Equal(intended, decToken) {
		t.Errorf("Token has been set to %v, but it's %v", intended, token)
	}
}

func TestGetsReasonCorrectly(t *testing.T) {
	req := dummyGet()

	reason := Reason(req)
	if reason != nil {
		t.Errorf("Reason hasn't been set yet, but it's not nil, it's %v", reason)
	}

	// again, needed for ctxSetReason() to work
	ctxSetToken(req, []byte("dummy"))

	intended := errors.New("universe imploded")
	ctxSetReason(req, intended)

	reason = Reason(req)
	if reason != intended {
		t.Errorf("Reason has been set to %v, but it's %v", intended, reason)
	}
}

func TestClearsContextEntry(t *testing.T) {
	req := dummyGet()

	ctxSetToken(req, []byte("dummy"))
	ctxSetReason(req, errors.New("some error"))

	ctxClear(req)

	entry, found := contextMap[req]

	if found {
		t.Errorf("Context entry %v found for the request %v, even though"+
			" it should have been cleared.", entry, req)
	}
}

func TestClearsContextEntryEvenIfNotSet(t *testing.T) {
	r := dummyGet()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ctxClear(r) panicked with %v", r)
		}
	}()
	ctxClear(r)
}
