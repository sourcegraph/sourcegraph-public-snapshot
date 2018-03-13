package server

import (
	"net/http/httptest"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

func TestServer_handleGet(t *testing.T) {
	conf.MockGetData = &schema.SiteConfiguration{
		Gitolite: []schema.GitoliteConnection{{
			Blacklist: "isblaclist.*",
			Prefix:    "mygitolite.host/",
			Host:      "git@mygitolite.host",
			PhabricatorMetadataCommand: "echo \"CALLSIGN:$REPO\"",
		}},
	}

	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()

	var cases = []struct {
		repo        string
		expMetadata string
	}{{
		repo:        "somerepo",
		expMetadata: `{"callsign":"CALLSIGN:somerepo"}`,
	}, {
		repo:        "anotherrepo",
		expMetadata: `{"callsign":"CALLSIGN:anotherrepo"}`,
	}}

	for _, testcase := range cases {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/getGitolitePhabricatorMetadata?gitolite=git@mygitolite.host&repo="+testcase.repo, nil)
		h.ServeHTTP(rr, req)

		result := strings.TrimSpace(rr.Body.String())
		if result != testcase.expMetadata {
			t.Errorf("for repo %q, expected metadata %q, but got %q", testcase.repo, testcase.expMetadata, result)
		}
	}
}
