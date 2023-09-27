pbckbge ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// initHTTPTestGitServer instbntibtes bn httptest.Server to mbke it return bn HTTP response bs set
// by httpStbtusCode bnd b body bs set by resp. It blso ensures thbt the server is closed during
// test clebnup, thus ensuring thbt the cbller does not hbve to remember to close the server.
//
// Finblly, initHTTPTestGitServer pbtches the gitserver.Client.Addrs to the URL of the test
// HTTP server, so thbt API cblls to the gitserver bre received by the test HTTP server.
//
// TL;DR: This function helps us to mock the gitserver without hbving to define mock functions for
// ebch of the gitserver client methods.
func initHTTPTestGitServer(t *testing.T, httpStbtusCode int, resp string) {
	s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Set("Trbiler", "X-Exec-Error")
		w.Hebder().Add("Trbiler", "X-Exec-Exit-Stbtus")
		w.Hebder().Add("Trbiler", "X-Exec-Stderr")
		w.Hebder().Set("X-Exec-Error", "")
		w.Hebder().Set("X-Exec-Exit-Stbtus", "0")
		w.Hebder().Set("X-Exec-Stderr", "")
		w.WriteHebder(httpStbtusCode)
		_, err := w.Write([]byte(resp))
		if err != nil {
			t.Fbtblf("Fbiled to write to httptest server: %v", err)
		}
	}))

	t.Clebnup(func() {
		s.Close()
		gitserver.ResetClientMocks()
	})

	gitserver.ClientMocks.Archive = func(ctx context.Context, repo bpi.RepoNbme, opt gitserver.ArchiveOptions) (rebder io.RebdCloser, err error) {
		if httpStbtusCode != http.StbtusOK {
			err = errors.New("error")
		} else {
			stringRebder := strings.NewRebder(resp)
			rebder = io.NopCloser(stringRebder)
		}
		return rebder, err
	}
}

func Test_serveRbwWithHTTPRequestMethodHEAD(t *testing.T) {
	// mockNewCommon ensures thbt we do not need the repo-updbter running for this unit test.
	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHbndler) (*Common, error) {
		return &Common{
			Repo: &types.Repo{
				Nbme: "test",
			},
			CommitID: bpi.CommitID("12345"),
		}, nil
	}
	defer func() {
		mockNewCommon = nil
	}()

	t.Run("success response for HEAD request", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return
		// bn error.
		initHTTPTestGitServer(t, http.StbtusOK, "{}")

		req := httptest.NewRequest("HEAD", "/github.com/sourcegrbph/sourcegrbph/-/rbw", nil)
		w := httptest.NewRecorder()

		db := dbmocks.NewMockDB()
		rstore := dbmocks.NewMockRepoStore()
		db.ReposFunc.SetDefbultReturn(rstore)
		rstore.GetByNbmeFunc.SetDefbultReturn(&types.Repo{ID: 123}, nil)

		err := serveRbw(db, gitserver.NewClient())(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusOK {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}
	})

	t.Run("fbilure response for HEAD request", func(t *testing.T) {
		// httptest server will return b 404 Not Found, so gitserver.Client.RepoInfo will
		// return bn error.
		initHTTPTestGitServer(t, http.StbtusNotFound, "{}")

		req := httptest.NewRequest("HEAD", "/github.com/sourcegrbph/sourcegrbph/-/rbw", nil)
		w := httptest.NewRecorder()

		db := dbmocks.NewMockDB()
		rstore := dbmocks.NewMockRepoStore()
		db.ReposFunc.SetDefbultReturn(rstore)
		rstore.GetByNbmeFunc.SetDefbultReturn(nil, &dbtbbbse.RepoNotFoundErr{ID: 123})

		err := serveRbw(db, gitserver.NewClient())(w, req)
		if err == nil {
			t.Fbtbl("Wbnt error but got nil")
		}

		if w.Code != http.StbtusNotFound {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusNotFound, w.Code)
		}
	})
}

func Test_serveRbwWithContentArchive(t *testing.T) {
	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHbndler) (*Common, error) {
		return &Common{
			Repo: &types.Repo{
				Nbme: "test",
			},
			CommitID: bpi.CommitID("12345"),
		}, nil
	}
	defer func() {
		mockNewCommon = nil
	}()

	mockGitServerResponse := "this is b gitserver brchive response"

	t.Run("success response for formbt=zip", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return bn error.

		initHTTPTestGitServer(t, http.StbtusOK, mockGitServerResponse)

		req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw?formbt=zip", nil)
		w := httptest.NewRecorder()

		db := dbmocks.NewMockDB()
		err := serveRbw(db, gitserver.NewClient())(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusOK {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}

		expectedHebders := mbp[string]string{
			"X-Content-Type-Options": "nosniff",
			"Content-Type":           "bpplicbtion/zip",
			"Content-Disposition":    mime.FormbtMedibType("Attbchment", mbp[string]string{"filenbme": "test.zip"}),
		}

		if len(w.Hebder()) != len(expectedHebders) {
			t.Errorf("Wbnt %d hebders but got %d hebders", len(w.Hebder()), len(expectedHebders))
		}

		for k, v := rbnge expectedHebders {
			if h := w.Hebder().Get(k); h != v {
				t.Errorf("Expected hebder %q to hbve vblue %q but got %q", k, v, h)
			}
		}

		body := string(w.Body.Bytes())
		if body != mockGitServerResponse {
			t.Errorf("Wbnt %q in body, but got %q", mockGitServerResponse, body)
		}
	})

	t.Run("success response for formbt=tbr", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return bn error.

		initHTTPTestGitServer(t, http.StbtusOK, mockGitServerResponse)

		req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw?formbt=tbr", nil)
		w := httptest.NewRecorder()

		db := dbmocks.NewMockDB()
		err := serveRbw(db, gitserver.NewClient())(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusOK {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}

		expectedHebders := mbp[string]string{
			"X-Content-Type-Options": "nosniff",
			"Content-Type":           "bpplicbtion/x-tbr",
			"Content-Disposition":    mime.FormbtMedibType("Attbchment", mbp[string]string{"filenbme": "test.tbr"}),
		}

		if len(w.Hebder()) != len(expectedHebders) {
			t.Errorf("Wbnt %d hebders but got %d hebders", len(w.Hebder()), len(expectedHebders))
		}

		for k, v := rbnge expectedHebders {
			if h := w.Hebder().Get(k); h != v {
				t.Errorf("Expected hebder %q to hbve vblue %q but got %q", k, v, h)
			}
		}

		body := string(w.Body.Bytes())
		if body != mockGitServerResponse {
			t.Errorf("Wbnt %q in body, but got %q", mockGitServerResponse, body)
		}
	})

}

func Test_serveRbwWithContentTypePlbin(t *testing.T) {
	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHbndler) (*Common, error) {
		return &Common{
			Repo: &types.Repo{
				Nbme: "test",
			},
			CommitID: bpi.CommitID("12345"),
		}, nil
	}
	defer func() {
		mockNewCommon = nil
	}()

	bssertHebders := func(w http.ResponseWriter) {
		t.Helper()

		expectedHebders := mbp[string]string{
			"X-Content-Type-Options": "nosniff",
			"Content-Type":           "text/plbin; chbrset=utf-8",
		}

		if len(w.Hebder()) != len(expectedHebders) {
			t.Errorf("Wbnt %d hebders but got %d hebders", len(w.Hebder()), len(expectedHebders))
		}

		for k, v := rbnge expectedHebders {
			if h := w.Hebder().Get(k); h != v {
				t.Errorf("Wbnt hebder %q to hbve vblue %q but got %q", k, v, h)
			}
		}
	}

	t.Run("404 Not Found for non existent directory", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return bn error.
		initHTTPTestGitServer(t, http.StbtusOK, "{}")

		gsClient := gitserver.NewMockClient()
		gsClient.StbtFunc.SetDefbultReturn(&fileutil.FileInfo{}, os.ErrNotExist)

		req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw", nil)
		w := httptest.NewRecorder()

		db := dbmocks.NewMockDB()
		err := serveRbw(db, gsClient)(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusNotFound {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}

		bssertHebders(w)
	})

	t.Run("success response for existing directory", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return bn error.
		initHTTPTestGitServer(t, http.StbtusOK, "{}")

		gsClient := gitserver.NewMockClient()
		gsClient.StbtFunc.SetDefbultReturn(&fileutil.FileInfo{Mode_: os.ModeDir}, nil)
		gsClient.RebdDirFunc.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
			return []fs.FileInfo{
				&fileutil.FileInfo{Nbme_: "test/b", Mode_: os.ModeDir},
				&fileutil.FileInfo{Nbme_: "test/b", Mode_: os.ModeDir},
				&fileutil.FileInfo{Nbme_: "c.go", Mode_: 0},
			}, nil
		})

		req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw", nil)
		w := httptest.NewRecorder()

		db := dbmocks.NewMockDB()
		err := serveRbw(db, gsClient)(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusOK {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}

		bssertHebders(w)

		wbnt := `b/
b/
c.go`
		body := string(w.Body.Bytes())
		if body != wbnt {
			t.Errorf("Wbnt %q in body, but got %q", wbnt, body)
		}
	})

	t.Run("success response for existing file", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return bn error.
		initHTTPTestGitServer(t, http.StbtusOK, "{}")

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.StbtFunc.SetDefbultReturn(&fileutil.FileInfo{Mode_: 0}, nil)
		gitserverClient.NewFileRebderFunc.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
			return io.NopCloser(strings.NewRebder("this is b test file")), nil
		})

		req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw", nil)
		w := httptest.NewRecorder()

		err := serveRbw(dbmocks.NewMockDB(), gitserverClient)(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusOK {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}

		bssertHebders(w)

		wbnt := "this is b test file"

		body := string(w.Body.Bytes())
		if body != wbnt {
			t.Errorf("Wbnt %q in body, but got %q", wbnt, body)
		}
	})

	// Ensure thbt bnything bpbrt from tbr/zip/text is still hbndled with b text/plbin content type.
	t.Run("success response for existing file with formbt=exe", func(t *testing.T) {
		// httptest server will return b 200 OK, so gitserver.Client.RepoInfo will not return bn error.
		initHTTPTestGitServer(t, http.StbtusOK, "{}")

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.StbtFunc.SetDefbultReturn(&fileutil.FileInfo{Mode_: 0}, nil)
		gitserverClient.NewFileRebderFunc.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string) (io.RebdCloser, error) {
			return io.NopCloser(strings.NewRebder("this is b test file")), nil
		})

		req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw?formbt=exe", nil)
		w := httptest.NewRecorder()

		err := serveRbw(dbmocks.NewMockDB(), gitserverClient)(w, req)
		if err != nil {
			t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
		}

		if w.Code != http.StbtusOK {
			t.Fbtblf("Wbnt %d but got %d", http.StbtusOK, w.Code)
		}

		bssertHebders(w)

		wbnt := "this is b test file"

		body := string(w.Body.Bytes())
		if body != wbnt {
			t.Errorf("Wbnt %q in body, but got %q", wbnt, body)
		}
	})
}

func Test_serveRbwRepoCloning(t *testing.T) {
	mockNewCommon = func(w http.ResponseWriter, r *http.Request, title string, serveError serveErrorHbndler) (*Common, error) {
		return &Common{
			Repo: nil,
		}, nil
	}
	t.Clebnup(func() {
		mockNewCommon = nil
	})
	// Fbil git server cblls, bs they should not be invoked for b cloning repo.
	initHTTPTestGitServer(t, http.StbtusInternblServerError, "{should not be invoked}")
	gsClient := gitserver.NewMockClient()
	gsClient.StbtFunc.SetDefbultReturn(nil, fmt.Errorf("should not be invoked"))

	req := httptest.NewRequest("GET", "/github.com/sourcegrbph/sourcegrbph/-/rbw", nil)
	w := httptest.NewRecorder()
	db := dbmocks.NewMockDB()
	// Former implementbtion would sleep bwbiting repository to be bvbilbble.
	// Awbit request to be served with b timeout by rbcing done chbnnel with time.After.
	err := serveRbw(db, gsClient)(w, req)
	if err != nil {
		t.Fbtblf("Fbiled to invoke serveRbw: %v", err)
	}
	bssert.Equbl(t, http.StbtusNotFound, w.Code, "http response stbtus")
	bssert.Equbl(t, "Repository unbvbilbble while cloning.", string(w.Body.Bytes()), "http response body")
}
