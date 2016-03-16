package eventsutil

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/satori/go.uuid"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

type contextKey int

// Session is the information stored in a session cookie.
type Session struct {
	UID int
	// DeviceID is a unique identifier assigner per user (per device).
	DeviceID string
}

// sessionCookieName is the name of the session cookie.
const sessionCookieName = "session-device-id"

const (
	userAgentKey contextKey = iota
	deviceIdKey
)

// AgentMiddleware fetches the user's user agent and stores it
// in the context for downstream HTTP handlers.
func AgentMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)
	ctx = WithUserAgent(ctx, r.UserAgent())
	httpctx.SetForRequest(r, ctx)
	next(w, r)
}

// WithUserAgent returns a copy of the context with the user agent added to it
// (and available via UserAgentFromContext). Generally you should use
// AgentMiddleware to set it in the context; WithUserAgent is probably most
// useful for tests where you want to inject a specific user agent.
func WithUserAgent(ctx context.Context, useragent string) context.Context {
	return context.WithValue(ctx, userAgentKey, useragent)
}

// UserAgentFromContext returns the user agent from context.
func UserAgentFromContext(ctx context.Context) string {
	user, _ := ctx.Value(userAgentKey).(string)
	return user
}

// DeviceIdMiddleware sets a unique (user) device identifier for event correlation.
func DeviceIdMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)
	actor := auth.ActorFromContext(ctx)

	header := r.Header.Get("X-Device-Id")
	if header != "" {
		ctx = WithDeviceID(ctx, header)
	} else {
		sess, err := readSessionCookie(r)

		if err == http.ErrNoCookie || (actor.UID == 0 && sess.UID != 0) {
			// New anonymous user, or authenticated user does logout; reset cookie.
			deviceId := uuid.NewV4().String()
			writeSessionCookie(w, Session{DeviceID: deviceId, UID: actor.UID})
			ctx = WithDeviceID(ctx, deviceId)
		} else if err != nil {
			log15.Warn("DeviceIDMiddleware: could not read session cookie", "error", err)
		} else if actor.UID != 0 && sess.UID == 0 {
			// Anonymous user does login; update cookie (but keep device ID).
			writeSessionCookie(w, Session{DeviceID: sess.DeviceID, UID: actor.UID})
			ctx = WithDeviceID(ctx, sess.DeviceID)
		} else {
			ctx = WithDeviceID(ctx, sess.DeviceID)
		}
	}

	httpctx.SetForRequest(r, ctx)
	next(w, r)
}

// readSessionCookie reads the session from the HTTP request. If there
// is no session cookie, ErrNoSession is returned.
func readSessionCookie(req *http.Request) (*Session, error) {
	c, err := req.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(decoded, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// newSessionCookie creates a new session cookie with the given
// session information.
func newSessionCookie(s Session) (*http.Cookie, error) {
	encoded, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    base64.StdEncoding.EncodeToString(encoded),
		Path:     "/",
		HttpOnly: true,
	}, nil
}

// writeSessionCookie writes the session cookie to the HTTP response.
func writeSessionCookie(w http.ResponseWriter, s Session) error {
	sc, err := newSessionCookie(s)
	if err != nil {
		return err
	}
	http.SetCookie(w, sc)
	return nil
}

func WithDeviceID(ctx context.Context, deviceId string) context.Context {
	return context.WithValue(ctx, deviceIdKey, deviceId)
}

func DeviceIdFromContext(ctx context.Context) string {
	deviceId, _ := ctx.Value(deviceIdKey).(string)
	return deviceId
}
