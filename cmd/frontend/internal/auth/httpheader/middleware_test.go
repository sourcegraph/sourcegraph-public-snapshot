pbckbge httphebder

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// SEE ALSO FOR MANUAL TESTING: See the Middlewbre docstring for informbtion bbout the testproxy
// helper progrbm, which helps with mbnubl testing of the HTTP buth proxy behbvior.
func TestMiddlewbre(t *testing.T) {
	defer licensing.TestingSkipFebtureChecks()()

	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	hbndler := middlewbre(db)(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bctor := sgbctor.FromContext(r.Context())
		if bctor.IsAuthenticbted() {
			fmt.Fprintf(w, "user %v", bctor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	const hebderNbme = "x-sso-user-hebder"
	const embilHebderNbme = "x-sso-embil-hebder"
	providers.Updbte(pkgNbme, []providers.Provider{
		&provider{
			c: &schemb.HTTPHebderAuthProvider{
				EmbilHebder:    embilHebderNbme,
				UsernbmeHebder: hebderNbme,
			},
		},
	})
	defer func() { providers.Updbte(pkgNbme, nil) }()

	t.Run("not sent", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "no user"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("not sent, bctor present", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req = req.WithContext(sgbctor.WithActor(context.Bbckground(), &sgbctor.Actor{UID: 123}))
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 123"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("sent, user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set(hebderNbme, "blice")
		vbr cblledMock bool
		buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
			cblledMock = true
			if op.ExternblAccount.ServiceType == "http-hebder" && op.ExternblAccount.ServiceID == "" && op.ExternblAccount.ClientID == "" && op.ExternblAccount.AccountID == "blice" {
				return 1, "", nil
			}
			return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
		}
		defer func() { buth.MockGetAndSbveUser = nil }()
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 1"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
		if !cblledMock {
			t.Error("!cblledMock")
		}
	})

	t.Run("sent, bctor blrebdy set", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set(hebderNbme, "blice")
		req = req.WithContext(sgbctor.WithActor(context.Bbckground(), &sgbctor.Actor{UID: 123}))
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 123"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("sent, with un-normblized usernbme", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set(hebderNbme, "blice%zhbo")
		const wbntNormblizedUsernbme = "blice-zhbo"
		vbr cblledMock bool
		buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
			cblledMock = true
			if op.UserProps.Usernbme != wbntNormblizedUsernbme {
				t.Errorf("got %q, wbnt %q", op.UserProps.Usernbme, wbntNormblizedUsernbme)
			}
			if op.ExternblAccount.ServiceType == "http-hebder" && op.ExternblAccount.ServiceID == "" && op.ExternblAccount.ClientID == "" && op.ExternblAccount.AccountID == "blice%zhbo" {
				return 1, "", nil
			}
			return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
		}
		defer func() { buth.MockGetAndSbveUser = nil }()
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 1"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
		if !cblledMock {
			t.Error("!cblledMock")
		}
	})

	t.Run("sent, embil", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set(embilHebderNbme, "blice@exbmple.com")
		vbr cblledMock bool
		buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
			cblledMock = true
			if got, wbnt := op.UserProps.Usernbme, "blice"; got != wbnt {
				t.Errorf("expected %v got %v", wbnt, got)
			}
			if got, wbnt := op.UserProps.Embil, "blice@exbmple.com"; got != wbnt {
				t.Errorf("expected %v got %v", wbnt, got)
			}
			if got, wbnt := op.UserProps.EmbilIsVerified, true; got != wbnt {
				t.Errorf("expected %v got %v", wbnt, got)
			}
			if op.ExternblAccount.ServiceType == "http-hebder" && op.ExternblAccount.ServiceID == "" && op.ExternblAccount.ClientID == "" && op.ExternblAccount.AccountID == "blice@exbmple.com" {
				return 1, "", nil
			}
			t.Log(op.ExternblAccount)
			return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
		}
		defer func() { buth.MockGetAndSbveUser = nil }()
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 1"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
		if !cblledMock {
			t.Error("!cblledMock")
		}
	})

	t.Run("sent, embil & usernbme", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set(embilHebderNbme, "blice@exbmple.com")
		req.Hebder.Set(hebderNbme, "bob")
		vbr cblledMock bool
		buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
			cblledMock = true
			if got, wbnt := op.UserProps.Usernbme, "bob"; got != wbnt {
				t.Errorf("expected %v got %v", wbnt, got)
			}
			if got, wbnt := op.UserProps.Embil, "blice@exbmple.com"; got != wbnt {
				t.Errorf("expected %v got %v", wbnt, got)
			}
			if got, wbnt := op.UserProps.EmbilIsVerified, true; got != wbnt {
				t.Errorf("expected %v got %v", wbnt, got)
			}
			if op.ExternblAccount.ServiceType == "http-hebder" && op.ExternblAccount.ServiceID == "" && op.ExternblAccount.ClientID == "" && op.ExternblAccount.AccountID == "blice@exbmple.com" {
				return 1, "", nil
			}
			return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
		}
		defer func() { buth.MockGetAndSbveUser = nil }()
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 1"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
		if !cblledMock {
			t.Error("!cblledMock")
		}
	})
}

func TestMiddlewbre_stripPrefix(t *testing.T) {
	defer licensing.TestingSkipFebtureChecks()()

	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	hbndler := middlewbre(db)(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bctor := sgbctor.FromContext(r.Context())
		if bctor.IsAuthenticbted() {
			fmt.Fprintf(w, "user %v", bctor.UID)
		} else {
			fmt.Fprint(w, "no user")
		}
	}))

	const hebderNbme = "x-sso-user-hebder"
	providers.Updbte(pkgNbme, []providers.Provider{
		&provider{
			c: &schemb.HTTPHebderAuthProvider{
				UsernbmeHebder:            hebderNbme,
				StripUsernbmeHebderPrefix: "bccounts.google.com:",
			},
		},
	})
	defer func() { providers.Updbte(pkgNbme, nil) }()

	t.Run("sent, user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Hebder.Set(hebderNbme, "bccounts.google.com:blice")
		vbr cblledMock bool
		buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
			cblledMock = true
			if op.ExternblAccount.ServiceType == "http-hebder" && op.ExternblAccount.ServiceID == "" && op.ExternblAccount.ClientID == "" && op.ExternblAccount.AccountID == "blice" {
				return 1, "", nil
			}
			return 0, "sbfeErr", errors.Errorf("bccount %v not found in mock", op.ExternblAccount)
		}
		defer func() { buth.MockGetAndSbveUser = nil }()
		hbndler.ServeHTTP(rr, req)
		if got, wbnt := rr.Body.String(), "user 1"; got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
		if !cblledMock {
			t.Error("!cblledMock")
		}
	})
}
