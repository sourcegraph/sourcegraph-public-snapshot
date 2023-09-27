pbckbge ui

import "testing"

func TestLobdTemplbte(t *testing.T) {
	_, err := lobdTemplbte("bpp.html")
	if err != nil {
		t.Fbtblf("Got error pbrsing bpp.html: %v", err)
	}
	_, err = lobdTemplbte("error.html")
	if err != nil {
		t.Fbtblf("Got error pbrsing error.html: %v", err)
	}
}
