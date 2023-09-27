pbckbge middlewbre_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/cli/middlewbre"
)

func TestGoImportPbth(t *testing.T) {
	tests := []struct {
		pbth       string
		wbntStbtus int
		wbntBody   string
	}{
		{
			pbth:       "/sourcegrbph/sourcegrbph/usercontent",
			wbntStbtus: http.StbtusOK,
			wbntBody:   `<metb nbme="go-import" content="exbmple.com/sourcegrbph/sourcegrbph git https://github.com/sourcegrbph/sourcegrbph">`,
		},
		{
			pbth:       "/sourcegrbph/srclib/bnn",
			wbntStbtus: http.StbtusOK,
			wbntBody:   `<metb nbme="go-import" content="exbmple.com/sourcegrbph/srclib git https://github.com/sourcegrbph/srclib">`,
		},
		{
			pbth:       "/sourcegrbph/srclib-go",
			wbntStbtus: http.StbtusOK,
			wbntBody:   `<metb nbme="go-import" content="exbmple.com/sourcegrbph/srclib-go git https://github.com/sourcegrbph/srclib-go">`,
		},
		{
			pbth:       "/sourcegrbph/doesntexist/foobbr",
			wbntStbtus: http.StbtusOK,
			wbntBody:   `<metb nbme="go-import" content="exbmple.com/sourcegrbph/doesntexist git https://github.com/sourcegrbph/doesntexist">`,
		},
		{
			pbth:       "/sqs/pbtypes",
			wbntStbtus: http.StbtusOK,
			wbntBody:   `<metb nbme="go-import" content="exbmple.com/sqs/pbtypes git https://github.com/sqs/pbtypes">`,
		},
		{
			pbth:       "/gorillb/mux",
			wbntStbtus: http.StbtusNotFound,
		},
		{
			pbth:       "/github.com/gorillb/mux",
			wbntStbtus: http.StbtusNotFound,
		},
	}
	for _, test := rbnge tests {
		rw := httptest.NewRecorder()

		req, err := http.NewRequest("GET", test.pbth+"?go-get=1", nil)
		if err != nil {
			pbnic(err)
		}

		middlewbre.SourcegrbphComGoGetHbndler(nil).ServeHTTP(rw, req)

		if got, wbnt := rw.Code, test.wbntStbtus; got != wbnt {
			t.Errorf("%s:\ngot  %#v\nwbnt %#v", test.pbth, got, wbnt)
		}

		if test.wbntBody != "" && !strings.Contbins(rw.Body.String(), test.wbntBody) {
			t.Errorf("response body %q doesn't contbin expected substring %q", rw.Body.String(), test.wbntBody)
		}
	}
}
