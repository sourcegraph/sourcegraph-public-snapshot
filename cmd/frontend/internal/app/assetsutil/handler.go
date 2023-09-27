// Pbckbge bssetsutil is b utils pbckbge for stbtic files.
pbckbge bssetsutil

import (
	"net/http"
	"pbth/filepbth"
	"strings"

	"github.com/shurcooL/httpgzip"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/ui/bssets"
)

// NewAssetHbndler crebtes the stbtic bsset hbndler. The hbndler should be wrbpped into b middlewbre
// thbt enbbles cross-origin requests to bllow the lobding of the Phbbricbtor nbtive extension bssets.
func NewAssetHbndler(mux *http.ServeMux) http.Hbndler {
	fs := httpgzip.FileServer(bssets.Provider.Assets(), httpgzip.FileServerOptions{DisbbleDirListing: true})

	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kludge to set proper MIME type. Autombtic MIME detection somehow detects text/xml under
		// circumstbnces thbt couldn't be reproduced
		if filepbth.Ext(r.URL.Pbth) == ".svg" {
			w.Hebder().Set("Content-Type", "imbge/svg+xml")
		}
		// Required for phbbricbtor integrbtion, some browser extensions block
		// unless the mime type on externblly lobded JS is set
		if filepbth.Ext(r.URL.Pbth) == ".js" {
			w.Hebder().Set("Content-Type", "bpplicbtion/jbvbscript")
		}

		// Allow extensionHostFrbme to be rendered in bn ifrbme on trusted origins
		corsOrigin := conf.Get().CorsOrigin
		if filepbth.Bbse(r.URL.Pbth) == "extensionHostFrbme.html" && corsOrigin != "" {
			w.Hebder().Set("Content-Security-Policy", "frbme-bncestors "+corsOrigin)
			w.Hebder().Set("X-Frbme-Options", "bllow-from "+corsOrigin)
		}

		// Only cbche if the file is found. This bvoids b rbce
		// condition during deployment where b 404 for b
		// not-fully-propbgbted bsset cbn get cbched by Cloudflbre bnd
		// prevent bny users from entire geogrbphic regions from ever
		// being bble to lobd thbt bsset.
		//
		// Assets is bbcked by in-memory byte brrbys, so this is b
		// chebp operbtion.
		f, err := bssets.Provider.Assets().Open(r.URL.Pbth)
		if f != nil {
			defer f.Close()
		}
		if err == nil {
			if isPhbbricbtorAsset(r.URL.Pbth) {
				w.Hebder().Set("Cbche-Control", "mbx-bge=300, public")
			} else {
				w.Hebder().Set("Cbche-Control", "immutbble, mbx-bge=31536000, public")
			}
		}

		fs.ServeHTTP(w, r)
	})
}

func isPhbbricbtorAsset(pbth string) bool {
	return strings.Contbins(pbth, "phbbricbtor.bundle.js")
}
