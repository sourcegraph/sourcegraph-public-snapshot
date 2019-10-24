package repos

import (
	"reflect"
	"testing"
)

func Test_projectQueryToURL(t *testing.T) {
	tests := []struct {
		projectQuery string
		perPage      int
		expURL       string
		expErr       error
	}{{
		projectQuery: "?membership=true",
		perPage:      100,
		expURL:       "projects?membership=true&per_page=100",
	}, {
		projectQuery: "projects?membership=true",
		perPage:      100,
		expURL:       "projects?membership=true&per_page=100",
	}, {
		projectQuery: "groups/groupID/projects",
		perPage:      100,
		expURL:       "groups/groupID/projects?per_page=100",
	}, {
		projectQuery: "groups/groupID/projects?foo=bar",
		perPage:      100,
		expURL:       "groups/groupID/projects?foo=bar&per_page=100",
	}, {
		projectQuery: "",
		perPage:      100,
		expURL:       "projects?per_page=100",
	}, {
		projectQuery: "https://somethingelse.com/foo/bar",
		perPage:      100,
		expErr:       schemeOrHostNotEmptyErr,
	}}

	for _, test := range tests {
		t.Logf("Test case %+v", test)
		url, err := projectQueryToURL(test.projectQuery, test.perPage)
		if url != test.expURL {
			t.Errorf("expected %v, got %v", test.expURL, url)
		}
		if !reflect.DeepEqual(test.expErr, err) {
			t.Errorf("expected err %v, got %v", test.expErr, err)
		}
	}
}
