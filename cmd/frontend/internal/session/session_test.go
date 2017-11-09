package session

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

func TestStartDeleteSession(t *testing.T) {
	cleanup := ResetMockSessionStore(t)
	defer cleanup()

	// Start new session
	w := httptest.NewRecorder()
	actr := &actor.Actor{UID: "test-actor-123"}
	if err := StartNewSession(w, httptest.NewRequest("GET", "/", nil), actr, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	var authCookies []*http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MaxAge > 0 {
			authCookies = append(authCookies, cookie)
		}
	}

	// Create authed request with session cookie
	authedReq := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range authCookies {
		authedReq.AddCookie(cookie)
	}
	if len(authedReq.Cookies()) != 1 {
		t.Fatal("expected exactly 1 authed cookie")
	}

	// Check that session cookie was created
	setCookie, err := authedReq.Cookie("sg-session")
	if err != nil {
		t.Fatalf("cookie was not created, error: %s", err)
	}
	if setCookie.Path != "" {
		t.Fatalf("expected cookie path to be \"\", was %q", setCookie.Path)
	}
	if setCookie.Value != SessionCookie(authedReq) {
		t.Errorf("SessionCookie value did not match actual session cookie value: %v != %v", setCookie.Value, SessionCookie(authedReq))
	}

	// Check that actor exists in the session
	session, err := sessionStore.Get(authedReq, "sg-session")
	if err != nil {
		t.Fatalf("didn't find session: %v", err)
	}
	if session == nil {
		t.Fatal("session was nil")
	}
	authedActor := actor.FromContext(authenticateByCookie(authedReq))
	if !reflect.DeepEqual(actr, authedActor) {
		t.Fatalf("session was not created: %+v != %+v", authedActor, actr)
	}

	// Delete session
	authedReq2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range authCookies {
		authedReq2.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	DeleteSession(w, authedReq2)

	// Check that the session cookie was deleted
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}
	if len(resp.Cookies()) != 1 {
		t.Fatal("expected exactly 1 Set-Cookie")
	}
	deleteCookie := resp.Cookies()[0]
	if deleteCookie.Name != "sg-session" {
		t.Fatal("did not delete cookie (cookie name was not \"sg-session\")")
	}
	if deleteCookie.MaxAge >= 0 {
		t.Fatal("did not delete cookie (max-age was not less than 0)")
	}
	if deleteCookie.Expires.After(time.Now()) {
		t.Fatal("did not delete cookie (cookie not expired)")
	}

	// Check that the actor no longer exists in the session, even when we have the original cookie
	authedReq3 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range authCookies {
		authedReq3.AddCookie(cookie)
	}
	actor3 := actor.FromContext(authenticateByCookie(authedReq3))
	if !reflect.DeepEqual(actor3, &actor.Actor{}) {
		t.Fatalf("underlying session was not deleted: %+v != %+v", actor3, &actor.Actor{})
	}
}

func TestSessionExpiry(t *testing.T) {
	cleanup := ResetMockSessionStore(t)
	defer cleanup()

	// Start new session
	w := httptest.NewRecorder()
	actr := &actor.Actor{UID: "test-actor-123"}
	if err := StartNewSession(w, httptest.NewRequest("GET", "/", nil), actr, time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var authCookies []*http.Cookie
	for _, cookie := range w.Result().Cookies() {
		if cookie.Expires.After(time.Now()) || cookie.MaxAge > 0 {
			authCookies = append(authCookies, cookie)
		}
	}

	// Create authed request with session cookie
	authedReq := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range authCookies {
		authedReq.AddCookie(cookie)
	}
	if len(authedReq.Cookies()) != 1 {
		t.Fatal("expected exactly 1 authed cookie")
	}

	if gotActor := actor.FromContext(authenticateByCookie(authedReq)); !reflect.DeepEqual(gotActor, actr) {
		t.Errorf("didn't find actor %v != %v", gotActor, actr)
	}
	time.Sleep(1100 * time.Millisecond)
	if gotActor := actor.FromContext(authenticateByCookie(authedReq)); !reflect.DeepEqual(gotActor, &actor.Actor{}) {
		t.Errorf("session didn't expire, found actor %+v", gotActor)
	}
}

func TestCookieMiddleware(t *testing.T) {
	cleanup := ResetMockSessionStore(t)
	defer cleanup()

	actors := []*actor.Actor{{UID: "test-actor-123"}, {UID: "test-actor-456"}}

	// Start new sessions for all actors
	authedReqs := make([]*http.Request, len(actors))
	for i, actr := range actors {
		w := httptest.NewRecorder()
		if err := StartNewSession(w, httptest.NewRequest("GET", "/", nil), actr, time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}

		// Test cases for when session exists
		authedReqs[i] = httptest.NewRequest("GET", "/", nil)
		for _, cookie := range w.Result().Cookies() {
			if cookie.Expires.After(time.Now()) || cookie.MaxAge > 0 {
				authedReqs[i].AddCookie(cookie)
			}
		}
	}

	testcases := []struct {
		req      *http.Request
		expActor *actor.Actor
	}{{
		req:      httptest.NewRequest("GET", "/", nil),
		expActor: &actor.Actor{},
	}, {
		req:      authedReqs[0],
		expActor: actors[0],
	}, {
		req:      authedReqs[1],
		expActor: actors[1],
	}}
	for _, testcase := range testcases {
		CookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotActor := actor.FromContext(r.Context())
			if !reflect.DeepEqual(testcase.expActor, gotActor) {
				t.Errorf("on authenticated request, got actor %+v, expected %+v", gotActor, testcase.expActor)
			}
		})).ServeHTTP(httptest.NewRecorder(), testcase.req)
	}
}

func TestCookieOrSessionMiddleware(t *testing.T) {
	cleanup := ResetMockSessionStore(t)
	defer cleanup()

	actors := []*actor.Actor{{UID: "test-actor-123"}, {UID: "test-actor-456"}}

	// Start new sessions for all actors and record authentication info
	cookieAuthedReqs := make([]*http.Request, len(actors))
	for i, actr := range actors {
		w := httptest.NewRecorder()
		if err := StartNewSession(w, httptest.NewRequest("GET", "/", nil), actr, time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}

		// Test cases for when session exists
		cookieAuthedReqs[i] = httptest.NewRequest("GET", "/", nil)
		for _, cookie := range w.Result().Cookies() {
			if cookie.Expires.After(time.Now()) || cookie.MaxAge > 0 {
				cookieAuthedReqs[i].AddCookie(cookie)
			}
		}
	}

	headerAuthedReqs := make([]*http.Request, len(cookieAuthedReqs))
	for i, cookieAuthedReq := range cookieAuthedReqs {
		headerAuthedReqs[i] = httptest.NewRequest("GET", "/", nil)
		cookie, err := cookieAuthedReq.Cookie("sg-session")
		if err != nil {
			t.Fatal(err)
		}
		headerAuthedReqs[i].Header.Set("Authorization", "session "+cookie.Value)
	}

	testcases := []struct {
		req      *http.Request
		expActor *actor.Actor
	}{{
		req:      httptest.NewRequest("GET", "/", nil),
		expActor: &actor.Actor{},
	}, {
		req:      cookieAuthedReqs[0],
		expActor: actors[0],
	}, {
		req:      cookieAuthedReqs[1],
		expActor: actors[1],
	}, {
		req:      headerAuthedReqs[0],
		expActor: actors[0],
	}, {
		req:      headerAuthedReqs[1],
		expActor: actors[1],
	}}
	for _, testcase := range testcases {
		CookieOrSessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotActor := actor.FromContext(r.Context())
			if !reflect.DeepEqual(testcase.expActor, gotActor) {
				t.Errorf("on authenticated request, got actor %+v, expected %+v", gotActor, testcase.expActor)
			}
		})).ServeHTTP(httptest.NewRecorder(), testcase.req)
	}
}
