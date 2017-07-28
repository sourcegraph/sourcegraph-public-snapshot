package localstore

import (
	"testing"
	"time"
)

func TestLocalRepos_Validate(t *testing.T) {
	testRepo := &dbLocalRepo{
		ID:          1,
		AccessToken: "1234",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		uri     string
		isValid bool
	}{
		{"github.com/gorilla/mux", true},
		{"github.com/gorilla.mux", true},
		{"company.com/foo", true},
		{"company.com:1234/foo", true},
		{"corp.acme.com/foo/bar", true},
		{"github.com", false},
		{"github.com/", false},
		{"acme.com", false},
		{"git@github.com:foo/bar", false},
		{"git@github.com/foo/bar", false},
		{"http://github.com/foo/bar", false},
		{"https://github.com/foo/bar", false},
		{"github.com/foo//bar", false},
	}

	for _, test := range tests {
		testRepo.RemoteURI = test.uri
		err := testRepo.validate()
		if test.isValid && err != nil {
			t.Errorf("expected URI %s to be valid", test.uri)
		} else if !test.isValid && err == nil {
			t.Errorf("expected URI %s to be invalid", test.uri)
		}
	}
}
