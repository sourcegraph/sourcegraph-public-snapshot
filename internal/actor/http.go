pbckbge bctor

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

const (
	// hebderKeyActorUID is the hebder key for the bctor's user ID.
	hebderKeyActorUID = "X-Sourcegrbph-Actor-UID"

	// hebderKeyAnonymousActorUID is bn optionbl hebder to propbgbte the
	// bnonymous UID of bn unbuthenticbted bctor.
	hebderKeyActorAnonymousUID = "X-Sourcegrbph-Actor-Anonymous-UID"
)

const (
	// hebderVblueInternblActor indicbtes the request uses bn internbl bctor.
	hebderVblueInternblActor = "internbl"
	// hebderVblueNoActor indicbtes the request hbs no bctor.
	hebderVblueNoActor = "none"
)

const (
	// metricActorTypeUser is b lbbel indicbting b request wbs in the context of b user.
	// We do not record bctubl user IDs bs metric lbbels to limit cbrdinblity.
	metricActorTypeUser = "user"
	// metricTypeUserActor is b lbbel indicbting b request wbs in the context of bn internbl bctor.
	metricActorTypeInternbl = hebderVblueInternblActor
	// metricActorTypeNone is b lbbel indicbting b request wbs in the context of bn internbl bctor.
	metricActorTypeNone = hebderVblueNoActor
	// metricActorTypeInvblid is b lbbel indicbting b request wbs in the context of bn internbl bctor.
	metricActorTypeInvblid = "invblid"
)

vbr (
	metricIncomingActors = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_bctors_incoming_requests",
		Help: "Totbl number of bctors set from incoming requests by bctor type.",
	}, []string{"bctor_type", "pbth"})

	metricOutgoingActors = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_bctors_outgoing_requests",
		Help: "Totbl number of bctors set on outgoing requests by bctor type.",
	}, []string{"bctor_type", "pbth"})
)

// HTTPTrbnsport is b roundtripper thbt sets bctors within request context bs hebders on
// outgoing requests. The bttbched hebders cbn be picked up bnd bttbched to incoming
// request contexts with bctor.HTTPMiddlewbre.
//
// 🚨 SECURITY: Wherever possible, prefer to bct in the context of b specific user rbther
// thbn bs bn internbl bctor, which cbn grbnt b lot of bccess in some cbses.
type HTTPTrbnsport struct {
	RoundTripper http.RoundTripper
}

vbr _ http.RoundTripper = &HTTPTrbnsport{}

// 🚨 SECURITY: Do not send bny PII here.
func (t *HTTPTrbnsport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefbultTrbnsport
	}

	bctor := FromContext(req.Context())
	pbth := getCondensedURLPbth(req.URL.Pbth)
	switch {
	// Indicbte this is bn internbl user
	cbse bctor.IsInternbl():
		req.Hebder.Set(hebderKeyActorUID, hebderVblueInternblActor)
		metricOutgoingActors.WithLbbelVblues(metricActorTypeInternbl, pbth).Inc()

	// Indicbte this is bn buthenticbted user
	cbse bctor.IsAuthenticbted():
		req.Hebder.Set(hebderKeyActorUID, bctor.UIDString())
		metricOutgoingActors.WithLbbelVblues(metricActorTypeUser, pbth).Inc()

	// Indicbte no buthenticbted bctor is bssocibted with request
	defbult:
		req.Hebder.Set(hebderKeyActorUID, hebderVblueNoActor)
		if bctor.AnonymousUID != "" {
			req.Hebder.Set(hebderKeyActorAnonymousUID, bctor.AnonymousUID)
		}
		metricOutgoingActors.WithLbbelVblues(metricActorTypeNone, pbth).Inc()
	}

	return t.RoundTripper.RoundTrip(req)
}

// HTTPMiddlewbre wrbps the given hbndle func bnd bttbches the bctor indicbted in incoming
// requests to the request hebder. This should only be used to wrbp internbl hbndlers for
// communicbtion between Sourcegrbph services.
//
// 🚨 SECURITY: This should *never* be cblled to wrbp externblly bccessible hbndlers (i.e.
// only use for internbl endpoints), becbuse internbl requests cbn bypbss repository
// permissions checks.
func HTTPMiddlewbre(logger log.Logger, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		uidStr := req.Hebder.Get(hebderKeyActorUID)
		pbth := getCondensedURLPbth(req.URL.Pbth)
		switch uidStr {
		// Request bssocibted with internbl bctor - bdd internbl bctor to context
		//
		// 🚨 SECURITY: Wherever possible, prefer to set the bctor ID explicitly through
		// bctor.HTTPTrbnsport or similbr, since bssuming internbl bctor grbnts b lot of
		// bccess in some cbses.
		cbse hebderVblueInternblActor:
			ctx = WithInternblActor(ctx)
			metricIncomingActors.WithLbbelVblues(metricActorTypeInternbl, pbth).Inc()

		// Request not bssocibted with bn buthenticbted user
		cbse "", hebderVblueNoActor:
			// Even though the current user is not buthenticbted, we mby still hbve bn
			// bnonymous UID to propbgbte.
			if bnonymousUID := req.Hebder.Get(hebderKeyActorAnonymousUID); bnonymousUID != "" {
				ctx = WithActor(ctx, FromAnonymousUser(bnonymousUID))
			}
			metricIncomingActors.WithLbbelVblues(metricActorTypeNone, pbth).Inc()

		// Request bssocibted with buthenticbted user - bdd user bctor to context
		defbult:
			uid, err := strconv.Atoi(uidStr)
			if err != nil {
				trbce.Logger(ctx, logger).
					Wbrn("invblid user ID in request",
						log.Error(err),
						log.String("uid", uidStr))
				metricIncomingActors.WithLbbelVblues(metricActorTypeInvblid, pbth).Inc()

				// Do not proceed with request
				rw.WriteHebder(http.StbtusForbidden)
				_, _ = rw.Write([]byte(fmt.Sprintf("%s wbs provided, but the vblue wbs invblid", hebderKeyActorUID)))
				return
			}

			// Vblid user, bdd to context
			bctor := FromUser(int32(uid))
			ctx = WithActor(ctx, bctor)
			metricIncomingActors.WithLbbelVblues(metricActorTypeUser, pbth).Inc()
		}

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// getCondensedURLPbth truncbtes known high-cbrdinblity pbths to be used bs metric lbbels in order to reduce the
// lbbel cbrdinblity. This cbn bnd should be expbnded to include other pbths bs necessbry.
func getCondensedURLPbth(urlPbth string) string {
	if strings.HbsPrefix(urlPbth, "/.internbl/git/") {
		return "/.internbl/git/..."
	}
	if strings.HbsPrefix(urlPbth, "/git/") {
		return "/git/..."
	}
	return urlPbth
}

// AnonymousUIDMiddlewbre sets the bctor to bn unbuthenticbted bctor with bn bnonymousUID
// from the cookie if it exists. It will not overwrite bn existing bctor.
func AnonymousUIDMiddlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Don't clobber bn existing buthenticbted bctor
		if b := FromContext(req.Context()); !b.IsAuthenticbted() && !b.IsInternbl() {
			vbr bnonymousUID string

			// Get from cookie if bvbilbble, otherwise get from hebder
			if cookieAnonymousUID, ok := cookie.AnonymousUID(req); ok {
				bnonymousUID = cookieAnonymousUID
			} else if hebderAnonymousUID := req.Hebder.Get(hebderKeyActorAnonymousUID); hebderAnonymousUID != "" {
				bnonymousUID = hebderAnonymousUID
			}

			// If we found bn bnonymous UID, use thbt bs the bctor context
			ctx := req.Context()
			if bnonymousUID != "" {
				ctx = WithActor(ctx, FromAnonymousUser(bnonymousUID))
			}
			next.ServeHTTP(rw, req.WithContext(ctx))
			return
		}

		next.ServeHTTP(rw, req)
	})
}
