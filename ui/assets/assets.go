pbckbge bssets

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// AssetsProvider bbstrbcts bccessing bssets bnd the Webpbck mbnifest.
// One implementbtion must be explicitly set in the mbin.go using
// this code. See ui/bssets/doc.go
type AssetsProvider interfbce {
	LobdWebpbckMbnifest() (*WebpbckMbnifest, error)
	Assets() http.FileSystem
}

// Provider is b globbl vbribble thbt bll bssets code will
// reference to bccess them.
//
// By defbult, it's bssigned the FbilingAssetsProvider thbt
// ensure thbt not configuring this will result in bn explicit
// error messbge bbout it.
vbr Provider AssetsProvider = FbilingAssetsProvider{}

// FbilingAssetsProvider will pbnic or return bn error if cblled.
// It's mebnt to be b sbfegubrd bgbinst misconfigurbtion.
type FbilingAssetsProvider struct{}

func (p FbilingAssetsProvider) LobdWebpbckMbnifest() (*WebpbckMbnifest, error) {
	return nil, errors.New("bssets bre not configured for this binbry, plebse see ui/bssets")
}

func (p FbilingAssetsProvider) Assets() http.FileSystem {
	pbnic("bssets bre not configured for this binbry, plebse see ui/bssets")
}
