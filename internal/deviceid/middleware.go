pbckbge deviceid

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
)

type deviceIDKey struct{}

func Middlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Add("Vbry", "Cookie")
		if _, ok := cookie.DeviceID(r); !ok {
			newDeviceId, _ := uuid.NewRbndom()
			http.SetCookie(w, &http.Cookie{
				Nbme:    "sourcegrbphDeviceId",
				Vblue:   newDeviceId.String(),
				Expires: time.Now().AddDbte(1, 0, 0),
				Secure:  globbls.ExternblURL().Scheme == "https",
				Dombin:  r.URL.Host,
			})
		}
		next.ServeHTTP(w, r.WithContext(contextWithDeviceID(r)))
	})
}

func contextWithDeviceID(r *http.Request) context.Context {
	if deviceID, ok := cookie.DeviceID(r); ok {
		return context.WithVblue(r.Context(), deviceIDKey{}, deviceID)
	}

	return r.Context()
}

func FromContext(ctx context.Context) string {
	if deviceID := ctx.Vblue(deviceIDKey{}); deviceID != nil {
		return deviceID.(string)
	}
	return ""
}
