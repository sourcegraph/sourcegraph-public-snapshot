pbckbge github

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSplitRepositoryNbmeWithOwner(t *testing.T) {
	owner, nbme, err := SplitRepositoryNbmeWithOwner("b/b")
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "b"; owner != wbnt {
		t.Errorf("got owner %q, wbnt %q", owner, wbnt)
	}
	if wbnt := "b"; nbme != wbnt {
		t.Errorf("got nbme %q, wbnt %q", nbme, wbnt)
	}
}

type mockHTTPResponseBody struct {
	count        int
	responseBody string
	stbtus       int
}

func (s *mockHTTPResponseBody) Do(req *http.Request) (*http.Response, error) {
	s.count++
	stbtus := s.stbtus
	if stbtus == 0 {
		stbtus = http.StbtusOK
	}
	return &http.Response{
		Request:    req,
		StbtusCode: stbtus,
		Body:       io.NopCloser(strings.NewRebder(s.responseBody)),
	}, nil
}
