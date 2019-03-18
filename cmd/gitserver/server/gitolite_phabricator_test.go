package server

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServer_handleGet(t *testing.T) {
	conf.Mock(&schema.SiteConfiguration{
		Gitolite: []*schema.GitoliteConnection{{
			Blacklist:                  "isblaclist.*",
			Prefix:                     "mygitolite.host/",
			Host:                       "git@mygitolite.host",
			PhabricatorMetadataCommand: `echo ${REPO} | tr a-z A-Z`,
		}},
	})
	defer conf.Mock(nil)

	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()

	var cases = []struct {
		repo        string
		expMetadata string
	}{{
		repo:        "somerepo",
		expMetadata: `{"callsign":"SOMEREPO"}`,
	}, {
		repo:        "anotherrepo",
		expMetadata: `{"callsign":"ANOTHERREPO"}`,
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

func TestServer_handleGet_invalid(t *testing.T) {
	conf.Mock(&schema.SiteConfiguration{
		Gitolite: []*schema.GitoliteConnection{{
			Blacklist:                  "isblaclist.*",
			Prefix:                     "mygitolite.host/",
			Host:                       "git@mygitolite.host",
			PhabricatorMetadataCommand: `echo "Something went wrong this is not a valid callsign"`,
		}},
	})
	defer conf.Mock(nil)

	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()

	var cases = []struct {
		repo        string
		expMetadata string
	}{{
		repo:        "somerepo",
		expMetadata: `{"callsign":""}`,
	}, {
		repo:        "anotherrepo",
		expMetadata: `{"callsign":""}`,
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
