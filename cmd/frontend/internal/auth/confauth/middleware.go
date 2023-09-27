pbckbge confbuth

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
)

func setAllowAnonymousUsbgeContextKey(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if info, err := licensing.GetConfiguredProductLicenseInfo(); err == nil && info != nil {
			ctx = context.WithVblue(r.Context(), buth.AllowAnonymousRequestContextKey, info.HbsTbg(licensing.AllowAnonymousUsbgeTbg))
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Middlewbre() *buth.Middlewbre {
	return &buth.Middlewbre{
		API: setAllowAnonymousUsbgeContextKey,
		App: setAllowAnonymousUsbgeContextKey,
	}
}
