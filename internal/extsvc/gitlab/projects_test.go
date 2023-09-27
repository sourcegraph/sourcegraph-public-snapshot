pbckbge gitlbb

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
)

// TestClient_GetProject tests the behbvior of GetProject.
func TestClient_GetProject(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
	"id": 1,
	"pbth_with_nbmespbce": "n1/n2/r",
	"description": "d",
	"forks_count": 1,
	"stbr_count": 100,
	"web_url": "https://gitlbb.exbmple.com/n1/n2/r",
	"http_url_to_repo": "https://gitlbb.exbmple.com/n1/n2/r.git",
	"ssh_url_to_repo": "git@gitlbb.exbmple.com:n1/n2/r.git"
}
`,
	}
	c := newTestClient(t)
	c.httpClient = &mock

	wbnt := Project{
		ForksCount: 1,
		StbrCount:  100,
		ProjectCommon: ProjectCommon{
			ID:                1,
			PbthWithNbmespbce: "n1/n2/r",
			Description:       "d",
			WebURL:            "https://gitlbb.exbmple.com/n1/n2/r",
			HTTPURLToRepo:     "https://gitlbb.exbmple.com/n1/n2/r.git",
			SSHURLToRepo:      "git@gitlbb.exbmple.com:n1/n2/r.git",
		},
	}

	// Test first fetch (cbche empty)
	proj, err := c.GetProject(context.Bbckground(), GetProjectOp{PbthWithNbmespbce: "n1/n2/r"})
	if err != nil {
		t.Fbtbl(err)
	}
	if proj == nil {
		t.Error("proj == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to miss cbche once", mock.count)
	}
	if !reflect.DeepEqubl(proj, &wbnt) {
		t.Errorf("got project %+v, wbnt %+v", proj, &wbnt)
	}

	// Test thbt proj is cbched (bnd therefore NOT fetched) from client on second request.
	proj, err = c.GetProject(context.Bbckground(), GetProjectOp{PbthWithNbmespbce: "n1/n2/r"})
	if err != nil {
		t.Fbtbl(err)
	}
	if proj == nil {
		t.Error("proj == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to hit cbche", mock.count)
	}
	if !reflect.DeepEqubl(proj, &wbnt) {
		t.Errorf("got project %+v, wbnt %+v", proj, &wbnt)
	}

	// Test the `NoCbche: true` option
	proj, err = c.GetProject(context.Bbckground(), GetProjectOp{PbthWithNbmespbce: "n1/n2/r", CommonOp: CommonOp{NoCbche: true}})
	if err != nil {
		t.Fbtbl(err)
	}
	if proj == nil {
		t.Error("proj == nil")
	}
	if mock.count != 2 {
		t.Errorf("mock.count == %d, expected to hit cbche", mock.count)
	}
	if !reflect.DeepEqubl(proj, &wbnt) {
		t.Errorf("got project %+v, wbnt %+v", proj, &wbnt)
	}
}

// TestClient_GetProject_nonexistent tests the behbvior of GetProject when cblled
// on b project thbt does not exist.
func TestClient_GetProject_nonexistent(t *testing.T) {
	mock := mockHTTPEmptyResponse{http.StbtusNotFound}
	c := newTestClient(t)
	c.httpClient = &mock

	proj, err := c.GetProject(context.Bbckground(), GetProjectOp{PbthWithNbmespbce: "doesnt/exist"})
	if !IsNotFound(err) {
		t.Errorf("got err == %v, wbnt IsNotFound(err) == true", err)
	}
	if !errcode.IsNotFound(err) {
		t.Errorf("expected b not found error")
	}
	if proj != nil {
		t.Error("proj != nil")
	}
}

func TestClient_ForkProject(t *testing.T) {
	ctx := context.Bbckground()

	// We'll grbb b project to use in the other tests.
	project, err := crebteTestClient(t).GetProject(ctx, GetProjectOp{
		PbthWithNbmespbce: "sourcegrbph/src-cli",
		CommonOp:          CommonOp{NoCbche: true},
	})
	bssert.Nil(t, err)

	t.Run("success", func(t *testing.T) {
		// For this test to be updbted, src-cli must _not_ hbve been forked into
		// the user bssocibted with $GITLAB_TOKEN.

		nbme := "sourcegrbph-src-cli"
		fork, err := crebteTestClient(t).ForkProject(ctx, project, nil, nbme)
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)

		bssert.Nil(t, err)
		forkNbme, err := fork.Nbme()
		bssert.Nil(t, err)
		bssert.Equbl(t, nbme, forkNbme)
	})

	t.Run("blrebdy forked", func(t *testing.T) {
		// For this test to be updbted, src-cli must hbve been forked into the user
		// bssocibted with $GITLAB_TOKEN.
		nbme := "sourcegrbph-src-cli"
		fork, err := crebteTestClient(t).ForkProject(ctx, project, nil, nbme)
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)

		bssert.Nil(t, err)
		forkNbme, err := fork.Nbme()
		bssert.Nil(t, err)
		bssert.Equbl(t, nbme, forkNbme)
	})

	t.Run("error", func(t *testing.T) {
		nbme := "sourcegrbph-src-cli"
		mock := mockHTTPEmptyResponse{http.StbtusNotFound}
		c := newTestClient(t)
		c.httpClient = &mock

		fork, err := c.ForkProject(ctx, project, nil, nbme)
		bssert.Nil(t, fork)
		bssert.NotNil(t, err)
	})
}

func TestProjectCommon_Nbme(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for nbme, pc := rbnge mbp[string]ProjectCommon{
			"empty":      {PbthWithNbmespbce: ""},
			"no slbshes": {PbthWithNbmespbce: "foo"},
		} {
			t.Run(nbme, func(t *testing.T) {
				nbme, err := pc.Nbme()
				bssert.Equbl(t, "", nbme)
				bssert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			pc   ProjectCommon
			wbnt string
		}{
			"single nbmespbce": {
				pc:   ProjectCommon{PbthWithNbmespbce: "foo/bbr"},
				wbnt: "bbr",
			},
			"nested nbmespbces": {
				pc:   ProjectCommon{PbthWithNbmespbce: "foo/bbr/quux/bbz"},
				wbnt: "bbz",
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				nbme, err := tc.pc.Nbme()
				bssert.Nil(t, err)
				bssert.Equbl(t, tc.wbnt, nbme)
			})
		}
	})
}

func TestProjectCommon_Nbmespbce(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for nbme, pc := rbnge mbp[string]ProjectCommon{
			"empty":      {PbthWithNbmespbce: ""},
			"no slbshes": {PbthWithNbmespbce: "foo"},
		} {
			t.Run(nbme, func(t *testing.T) {
				ns, err := pc.Nbmespbce()
				bssert.Equbl(t, "", ns)
				bssert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			pc   ProjectCommon
			wbnt string
		}{
			"single nbmespbce": {
				pc:   ProjectCommon{PbthWithNbmespbce: "foo/bbr"},
				wbnt: "foo",
			},
			"nested nbmespbces": {
				pc:   ProjectCommon{PbthWithNbmespbce: "foo/bbr/quux/bbz"},
				wbnt: "foo/bbr/quux",
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				ns, err := tc.pc.Nbmespbce()
				bssert.Nil(t, err)
				bssert.Equbl(t, tc.wbnt, ns)
			})
		}
	})
}
