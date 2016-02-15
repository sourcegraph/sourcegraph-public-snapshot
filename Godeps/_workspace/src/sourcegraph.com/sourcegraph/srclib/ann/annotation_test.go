package ann

import "testing"

func TestAnn_LinkURL(t *testing.T) {
	want := "http://example.com"

	var a Ann
	if err := a.SetLinkURL(want); err != nil {
		t.Fatal(err)
	}
	u, err := a.LinkURL()
	if err != nil {
		t.Fatal(err)
	}
	if got := u.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnn_LinkURL_badURL(t *testing.T) {
	var a Ann
	if err := a.SetLinkURL(":%:"); err == nil {
		t.Fatal("SetLinkURL returned nil error")
	}

	a.Data = []byte(`":%:"`)
	a.Type = Link
	if _, err := a.LinkURL(); err == nil {
		t.Fatal("LinkURL returned nil error")
	}
}
