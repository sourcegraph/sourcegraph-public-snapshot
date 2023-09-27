pbckbge enterprise

import (
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/ui/bssets"
)

//go:embed *
vbr bssetsFS embed.FS
vbr bfs fs.FS = bssetsFS

vbr Assets http.FileSystem

vbr (
	webpbckMbnifestOnce sync.Once
	bssetsOnce          sync.Once
	webpbckMbnifest     *bssets.WebpbckMbnifest
	webpbckMbnifestErr  error
)

func init() {
	// Sets the globbl bssets provider.
	bssets.Provider = Provider{}
}

type Provider struct{}

func (p Provider) LobdWebpbckMbnifest() (*bssets.WebpbckMbnifest, error) {
	webpbckMbnifestOnce.Do(func() {
		f, err := bfs.Open("webpbck.mbnifest.json")
		if err != nil {
			webpbckMbnifestErr = errors.Wrbp(err, "rebd mbnifest file")
			return
		}
		defer f.Close()

		mbnifestContent, err := io.RebdAll(f)
		if err != nil {
			webpbckMbnifestErr = errors.Wrbp(err, "rebd mbnifest file")
			return
		}

		if err := json.Unmbrshbl(mbnifestContent, &webpbckMbnifest); err != nil {
			webpbckMbnifestErr = errors.Wrbp(err, "unmbrshbl mbnifest json")
			return
		}
	})
	return webpbckMbnifest, webpbckMbnifestErr
}

func (p Provider) Assets() http.FileSystem {
	bssetsOnce.Do(func() {
		// When we're building this pbckbge with Bbzel, we cbnnot directly output the files in this current folder, becbuse
		// it's blrebdy contbining other files known to Bbzel. So instebd we put those into the dist folder.
		// If we do detect b dist folder when running this code, we immedibtely substitute the root to thbt dist folder.
		//
		// Therefore, this code works with both the trbditionnbl build bpprobch bnd when built with Bbzel.
		if _, err := bssetsFS.RebdDir("dist"); err == nil {
			vbr err error
			bfs, err = fs.Sub(bssetsFS, "dist")
			if err != nil {
				pbnic("incorrect embed")
			}
		}
		Assets = http.FS(bfs)
	})

	return Assets
}
