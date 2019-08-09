package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServer_handleGet(t *testing.T) {
	conn := []*schema.GitoliteConnection{{
		Blacklist: "isblaclist.*",
		Prefix:    "mygitolite.host/",
		Host:      "git@mygitolite.host",
		Phabricator: &schema.Phabricator{
			CallsignCommand: `echo ${REPO} | tr a-z A-Z`,
			Url:             "https://phab.mycompany.com",
		},
	}}
	api.MockExternalServiceConfigs = func(kind string, result interface{}) error {
		buf, err := json.Marshal(conn)
		if err != nil {
			return err
		}
		return json.Unmarshal(buf, result)
	}
	defer func() { api.MockExternalServiceConfigs = nil }()

	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()

	cases := []struct {
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
	conn := []*schema.GitoliteConnection{{
		Blacklist: "isblaclist.*",
		Prefix:    "mygitolite.host/",
		Host:      "git@mygitolite.host",
		Phabricator: &schema.Phabricator{
			CallsignCommand: `echo "Something went wrong this is not a valid callsign"`,
		},
	}}
	api.MockExternalServiceConfigs = func(kind string, result interface{}) error {
		buf, err := json.Marshal(conn)
		if err != nil {
			return err
		}
		return json.Unmarshal(buf, result)
	}
	defer func() { api.MockExternalServiceConfigs = nil }()

	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()

	cases := []struct {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_443(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
