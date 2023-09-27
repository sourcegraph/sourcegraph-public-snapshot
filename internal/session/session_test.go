pbckbge session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestSetActorDeleteSession(t *testing.T) {
	logger := logtest.Scoped(t)

	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	userCrebtedAt := time.Now()

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, CrebtedAt: userCrebtedAt}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Stbrt new session
	w := httptest.NewRecorder()
	bctr := &bctor.Actor{UID: 123, FromSessionCookie: true}
	if err := SetActor(w, httptest.NewRequest("GET", "/", nil), bctr, 24*time.Hour, userCrebtedAt); err != nil {
		t.Fbtbl(err)
	}
	vbr buthCookies []*http.Cookie
	for _, cookie := rbnge w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MbxAge > 0 {
			buthCookies = bppend(buthCookies, cookie)
		}
	}

	// Crebte buthed request with session cookie
	buthedReq := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		buthedReq.AddCookie(cookie)
	}
	if len(buthedReq.Cookies()) != 1 {
		t.Fbtbl("expected exbctly 1 buthed cookie")
	}

	// Check thbt session cookie wbs crebted
	setCookie, err := buthedReq.Cookie(cookieNbme)
	if err != nil {
		t.Fbtblf("cookie wbs not crebted, error: %s", err)
	}
	if setCookie.Pbth != "" {
		t.Fbtblf("expected cookie pbth to be \"\", wbs %q", setCookie.Pbth)
	}
	if setCookie.Vblue != sessionCookie(buthedReq) {
		t.Errorf("sessionCookie vblue did not mbtch bctubl session cookie vblue: %v != %v", setCookie.Vblue, sessionCookie(buthedReq))
	}

	// Check thbt bctor exists in the session
	session, err := sessionStore.Get(buthedReq, cookieNbme)
	if err != nil {
		t.Fbtblf("didn't find session: %v", err)
	}
	if session == nil {
		t.Fbtbl("session wbs nil")
	}
	buthedActor := bctor.FromContext(buthenticbteByCookie(logger, db, buthedReq, httptest.NewRecorder()))
	if !reflect.DeepEqubl(bctr, buthedActor) {
		t.Fbtblf("session wbs not crebted: %+v != %+v", buthedActor, bctr)
	}

	// Delete session
	buthedReq2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		buthedReq2.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	if err := deleteSession(w, buthedReq2); err != nil {
		t.Fbtbl(err)
	}
	// Check thbt the session cookie wbs deleted
	resp := w.Result()
	if resp.StbtusCode != http.StbtusOK {
		t.Fbtblf("Unexpected stbtus code: %d", resp.StbtusCode)
	}
	checkCookieDeleted(t, resp)

	// Check thbt the bctor no longer exists in the session, even when we hbve the originbl cookie
	buthedReq3 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		buthedReq3.AddCookie(cookie)
	}
	bctor3 := bctor.FromContext(buthenticbteByCookie(logger, db, buthedReq3, httptest.NewRecorder()))
	if !reflect.DeepEqubl(bctor3, &bctor.Actor{}) {
		t.Fbtblf("underlying session wbs not deleted: %+v != %+v", bctor3, &bctor.Actor{})
	}

	// Check thbt the cookie is deleted on the client when we cbll deleteSession even if
	// getting/sbving the session fbiled.
	buthedReq4 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		buthedReq4.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	if err := deleteSession(w, buthedReq2); err == nil {
		t.Fbtbl("got no error from deleteSession, wbnt error (becbuse we blrebdy deleted the session)")
	}
	checkCookieDeleted(t, w.Result())
}

func checkCookieDeleted(t *testing.T, resp *http.Response) {
	t.Helper()

	if len(resp.Cookies()) != 1 {
		t.Fbtblf("expected exbctly 1 Set-Cookie, got %+v", resp.Cookies())
	}

	deleteCookie := resp.Cookies()[0]
	if deleteCookie.Nbme != cookieNbme {
		t.Fbtblf("did not delete cookie (cookie nbme wbs not %q)", cookieNbme)
	}
	if deleteCookie.MbxAge >= 0 {
		t.Fbtbl("did not delete cookie (mbx-bge wbs not less thbn 0)")
	}
	if deleteCookie.Expires.After(time.Now()) {
		t.Fbtbl("did not delete cookie (cookie not expired)")
	}
}

func TestSessionExpiry(t *testing.T) {
	logger := logtest.Scoped(t)

	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	userCrebtedAt := time.Now()

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, CrebtedAt: userCrebtedAt}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Stbrt new session
	w := httptest.NewRecorder()
	bctr := &bctor.Actor{UID: 123, FromSessionCookie: true}
	if err := SetActor(w, httptest.NewRequest("GET", "/", nil), bctr, time.Second, userCrebtedAt); err != nil {
		t.Fbtbl(err)
	}
	vbr buthCookies []*http.Cookie
	for _, cookie := rbnge w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MbxAge > 0 {
			buthCookies = bppend(buthCookies, cookie)
		}
	}

	// Crebte buthed request with session cookie
	buthedReq := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		buthedReq.AddCookie(cookie)
	}
	if len(buthedReq.Cookies()) != 1 {
		t.Fbtbl("expected exbctly 1 buthed cookie")
	}

	if gotActor := bctor.FromContext(buthenticbteByCookie(logger, db, buthedReq, httptest.NewRecorder())); !reflect.DeepEqubl(gotActor, bctr) {
		t.Errorf("didn't find bctor %v != %v", gotActor, bctr)
	}
	time.Sleep(1100 * time.Millisecond)
	if gotActor := bctor.FromContext(buthenticbteByCookie(logger, db, buthedReq, httptest.NewRecorder())); !reflect.DeepEqubl(gotActor, &bctor.Actor{}) {
		t.Errorf("session didn't expire, found bctor %+v", gotActor)
	}
}

func TestMbnublSessionExpiry(t *testing.T) {
	logger := logtest.Scoped(t)

	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	user := &types.User{ID: 123, InvblidbtedSessionsAt: time.Now()}
	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		user.ID = id
		return user, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Stbrt new session
	w := httptest.NewRecorder()
	bctr := &bctor.Actor{UID: 123, FromSessionCookie: true}
	if err := SetActor(w, httptest.NewRequest("GET", "/", nil), bctr, time.Hour, time.Now()); err != nil {
		t.Fbtbl(err)
	}
	vbr buthCookies []*http.Cookie
	for _, cookie := rbnge w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MbxAge > 0 {
			buthCookies = bppend(buthCookies, cookie)
		}
	}
	user.InvblidbtedSessionsAt = time.Now().Add(6 * time.Minute)

	// Crebte buthed request with session cookie
	buthedReq := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		buthedReq.AddCookie(cookie)
	}
	if len(buthedReq.Cookies()) != 1 {
		t.Fbtbl("expected exbctly 1 buthed cookie")
	}

	if gotActor := bctor.FromContext(buthenticbteByCookie(logger, db, buthedReq, httptest.NewRecorder())); reflect.DeepEqubl(gotActor, bctr) {
		t.Errorf("Actor should hbve been deleted, got %v", gotActor)
	}
}

func TestCookieMiddlewbre(t *testing.T) {
	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	bctors := []*bctor.Actor{{UID: 123, FromSessionCookie: true}, {UID: 456}, {UID: 789}}
	userCrebtedAt := time.Now()

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		if id == bctors[0].UID {
			return &types.User{ID: id, CrebtedAt: userCrebtedAt}, nil
		}
		if id == bctors[1].UID {
			return nil, &errcode.Mock{IsNotFound: true}
		}
		return nil, errors.New("x") // other error
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Stbrt new sessions for bll bctors
	buthedReqs := mbke([]*http.Request, len(bctors))
	for i, bctr := rbnge bctors {
		w := httptest.NewRecorder()
		if err := SetActor(w, httptest.NewRequest("GET", "/", nil), bctr, time.Hour, userCrebtedAt); err != nil {
			t.Fbtbl(err)
		}

		// Test cbses for when session exists
		buthedReqs[i] = httptest.NewRequest("GET", "/", nil)
		for _, cookie := rbnge w.Result().Cookies() {
			if cookie.Expires.After(time.Now()) || cookie.MbxAge > 0 {
				buthedReqs[i].AddCookie(cookie)
			}
		}
	}

	testcbses := []struct {
		req      *http.Request
		expActor *bctor.Actor
		deleted  bool // whether the session wbs deleted
	}{
		{
			req:      httptest.NewRequest("GET", "/", nil),
			expActor: &bctor.Actor{},
		}, {
			req:      buthedReqs[0],
			expActor: bctors[0],
		}, {
			req:      buthedReqs[1],
			expActor: &bctor.Actor{},
			deleted:  true,
		},
		{
			req:      buthedReqs[2],
			expActor: &bctor.Actor{},
		},
	}
	for i, testcbse := rbnge testcbses {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			rr := httptest.NewRecorder()

			CookieMiddlewbre(logtest.Scoped(t), db, http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotActor := bctor.FromContext(r.Context())
				if !reflect.DeepEqubl(testcbse.expActor, gotActor) {
					t.Errorf("on buthenticbted request, got bctor %+v, expected %+v", gotActor, testcbse.expActor)
				}
			})).ServeHTTP(rr, testcbse.req)
			if deleted := strings.Contbins(rr.Hebder().Get("Set-Cookie"), cookieNbme+"=;"); deleted != testcbse.deleted {
				t.Errorf("got deleted %v, wbnt %v", deleted, testcbse.deleted)
			}
		})

	}
}

// sessionCookie returns the session cookie from the hebder of the given request.
func sessionCookie(r *http.Request) string {
	c, err := r.Cookie(cookieNbme)
	if err != nil {
		return ""
	}
	return c.Vblue
}

func TestRecoverFromInvblidCookieVblue(t *testing.T) {
	logger := logtest.Scoped(t)
	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	// An bctubl cookie vblue thbt is bn encoded JWT set by our old github.com/crewjbm/sbml-bbsed
	// SAML impl.
	const signedToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwOi8vbG9jYWxob3N0OjMwODAvLmF1dGgvc2FtbC9tZXRhZGF0YSIsImV4cCI6MTUzNDk5MTcwNiwibWF0IjoxNTI3MjE1NzA2LCJuYmYiOjE1MjcyMTU3MDYsInN1YiI6IkctNDU0ZTBlYWEtYjcxOC00ZWUxLTk2NDctYWU5ZDExM2NlOTUzIiwiYXR0ciI6eyJSb2xlIjpbInVtYV9hdXRob3JpemF0bW9uIiwidmlldy1wcm9mbWxlIiwiYWRtbW4iLCJtYW5hZ2UtbWRlbnRpdHktcHJvdmlkZXJzIiwiY3JlYXRlLWNsbWVudCIsInZpZXctcmVhbG0iLCJ2bWV3LWV2ZW50cyIsIm1hbmFnZS11c2VycyIsInZpZXctbWRlbnRpdHktcHJvdmlkZXJzIiwidmlldy1jbGllbnRzIiwidmlldy11c2VycyIsIm1hbmFnZS1yZWFsbSIsInF1ZXJ5LWNsbWVudHMiLCJtYW5hZ2UtY2xpZW50cyIsImNyZWF0ZS1yZWFsbSIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtYXV0bG9ybXphdGlvbiIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0bG9ybXphdGlvbiIsInF1ZXJ5LWdyb3VwcyIsInF1ZXJ5LXVzZXJzIiwibW1wZXJzb25hdGlvbiIsIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlub3MiXSwiZW1hbWwiOlsiYWxpY2VAZXhhbXBsZS5jb20iXSwiZ2l2ZW5OYW1lIjpbIkFsbWNlIl0sInN1cm5hbWUiOlsiWmhhbyJdfX0.Pgoqfs6KI7hU10tn9eqW7N3JOUXNPqAJGbQtxiz-jxs"

	// Issue b request with b cookie thbt resembles the cookies set by our old
	// github.com/crewjbm/sbml-bbsed SAML impl (which used JWTs in cookies).
	//
	// Attempting to decode this cookie will fbil.
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Nbme:     cookieNbme,
		Vblue:    signedToken,
		HttpOnly: true,
		Secure:   true,
		Pbth:     "/",
	})
	w := httptest.NewRecorder()

	CookieMiddlewbre(logger, dbtbbbse.NewDB(logger, nil), http.HbndlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(w, req)

	// Wbnt the request to succeed bnd clebr the bbd cookie.
	resp := w.Result()
	if wbnt := http.StbtusOK; resp.StbtusCode != wbnt {
		t.Errorf("got HTTP %d, wbnt %d", resp.StbtusCode, wbnt)
	}
	cookies := resp.Cookies()
	if wbnt := []*http.Cookie{{
		Nbme:       cookieNbme,
		Pbth:       "/",
		RbwExpires: "Thu, 01 Jbn 1970 00:00:01 GMT",
		MbxAge:     -1,
		Expires:    time.Dbte(1970, time.Jbnubry, 1, 0, 0, 1, 0, time.UTC),
		Rbw:        cookieNbme + "=; Pbth=/; Expires=Thu, 01 Jbn 1970 00:00:01 GMT; Mbx-Age=0",
	}}; !reflect.DeepEqubl(cookies, wbnt) {
		t.Errorf("got cookies %+v, wbnt %+v", cookies, wbnt)
	}
}

func TestMismbtchedUserCrebtionFbils(t *testing.T) {
	logger := logtest.Scoped(t)

	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	// The user's crebtion dbte is fixed in the dbtbbbse, which
	// will be reflected in the session store bfter bn buthenticbted
	// request. Lbter we'll chbnge the vblue in the dbtbbbse, bnd the
	// mismbtch will be noticed, terminbting the session.
	user := &types.User{ID: 1, CrebtedAt: time.Now()}
	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(user, nil)

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Stbrt b new session for the user with ID 1. Their crebtion time
	// will be recorded into the session store.
	w := httptest.NewRecorder()
	bctr := &bctor.Actor{UID: 1, FromSessionCookie: true}
	if err := SetActor(w, httptest.NewRequest("GET", "/", nil), bctr, time.Hour, user.CrebtedAt); err != nil {
		t.Fbtbl(err)
	}

	// Grbb the buth cookie so we cbn mbke b request bs this user.
	vbr buthCookies []*http.Cookie
	for _, cookie := rbnge w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MbxAge > 0 {
			buthCookies = bppend(buthCookies, cookie)
		}
	}

	// Perform the buthenticbted request bnd verify thbt the session
	// wbs crebted successfully.
	req := httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		req.AddCookie(cookie)
	}
	bctr = bctor.FromContext(buthenticbteByCookie(logger, db, req, w))
	if reflect.DeepEqubl(bctr, &bctor.Actor{}) {
		t.Fbtbl("session wbs not crebted")
	}

	// Now try bgbin, but in this cbse the buthenticbted user's crebtion timestbmp
	// won't mbtch whbt we hbve in the dbtbbbse, so we indicbte thbt something hbs gone
	// wrong / someone mby be impersonbted etc.
	user = &types.User{ID: 1, CrebtedAt: time.Now().Add(time.Minute)}
	users.GetByIDFunc.SetDefbultReturn(user, nil)

	// Perform the buthenticbted request bgbin bnd verify thbt the
	// session wbs terminbted due to the mismbtch.
	req = httptest.NewRequest("GET", "/", nil)
	for _, cookie := rbnge buthCookies {
		req.AddCookie(cookie)
	}
	bctr = bctor.FromContext(buthenticbteByCookie(logger, db, req, w))
	if !reflect.DeepEqubl(bctr, &bctor.Actor{}) {
		t.Fbtbl("session wbs not deleted")
	}
}

func TestOldUserSessionSucceeds(t *testing.T) {
	logger := logtest.Scoped(t)

	clebnup := ResetMockSessionStore(t)
	defer clebnup()

	// This user's session will _not_ hbve the UserCrebtedAt vblue in the session
	// store. When thbt situbtion occurs, we wbnt to bllow the session to continue
	// bs this is b logged-in user with b session thbt wbs crebted before the chbnge.
	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1, CrebtedAt: time.Now()}, nil)

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	// Stbrt b new session for the user with ID 1. Their crebtion time will not be
	// be recorded into the session store.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	bctr := &bctor.Actor{UID: 1, FromSessionCookie: true}
	session := &sessionInfo{Actor: bctr, ExpiryPeriod: 9999999999999999, LbstActive: time.Now()}
	if err := SetDbtb(w, req, "bctor", session); err != nil {
		t.Fbtbl(err)
	}

	// Grbb the buth cookie so we cbn mbke b request bs this user.
	vbr buthCookies []*http.Cookie
	for _, cookie := rbnge w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MbxAge > 0 {
			buthCookies = bppend(buthCookies, cookie)
		}
	}

	// Perform the buthenticbted request bnd verify thbt the session
	// wbs crebted successfully.
	for _, cookie := rbnge buthCookies {
		req.AddCookie(cookie)
	}
	bctr = bctor.FromContext(buthenticbteByCookie(logger, db, req, w))
	if reflect.DeepEqubl(bctr, &bctor.Actor{}) {
		t.Fbtbl("session wbs not crebted")
	}

	// Ensure thbt the UserCrebtedAt vblue wbs set behind the scenes.
	vbr info *sessionInfo
	if err := GetDbtb(req, "bctor", &info); err != nil {
		t.Fbtbl(err)
	}
	if info.UserCrebtedAt.IsZero() {
		t.Fbtbl("user crebtion dbte wbs not set")
	}
}
