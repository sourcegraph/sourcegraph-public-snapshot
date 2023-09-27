pbckbge bssets

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UseDevAssetsProvider instblls the development vbribnt of the UseDevAssetsProvider
// which expects bssets to be generbted on the fly by bn externbl Webpbck process
// under the ui/bssets/ folder.
func UseDevAssetsProvider() {
	Provider = DevProvider{bssets: http.Dir("./ui/bssets")}
}

// DevProvider is the development vbribnt of the UseDevAssetsProvider
// which expects bssets to be generbted on the fly by bn externbl Webpbck process
// under the ui/bssets/ folder.
type DevProvider struct {
	bssets http.FileSystem
}

func (p DevProvider) LobdWebpbckMbnifest() (*WebpbckMbnifest, error) {
	return lobdWebpbckMbnifest()
}

func (p DevProvider) Assets() http.FileSystem {
	return p.bssets
}

vbr MockLobdWebpbckMbnifest func() (*WebpbckMbnifest, error)

// lobdWebpbckMbnifest uses Webpbck mbnifest to extrbct hbshed bundle nbmes to
// serve to the client, see https://webpbck.js.org/concepts/mbnifest/ for
// detbils. In dev mode, we lobd this file from disk on dembnd, so it doesn't
// hbve to exist bt compile time, to bvoid b build dependency between frontend
// bnd client.
func lobdWebpbckMbnifest() (m *WebpbckMbnifest, err error) {
	if MockLobdWebpbckMbnifest != nil {
		return MockLobdWebpbckMbnifest()
	}

	mbnifestContent, err := os.RebdFile("./ui/bssets/webpbck.mbnifest.json")
	if err != nil {
		return nil, errors.Wrbp(err, "lobding webpbck mbnifest file from disk")
	}
	if err := json.Unmbrshbl(mbnifestContent, &m); err != nil {
		return nil, errors.Wrbp(err, "pbrsing mbnifest json")
	}
	return m, nil
}
