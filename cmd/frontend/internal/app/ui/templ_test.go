package ui

import "testing"

func TestLoadTemplate(t *testing.T) {
	_, err := loadTemplate("app.html")
	if err != nil {
		t.Fatalf("Got error parsing app.html: %v", err)
	}
	_, err = loadTemplate("error.html")
	if err != nil {
		t.Fatalf("Got error parsing error.html: %v", err)
	}
}
