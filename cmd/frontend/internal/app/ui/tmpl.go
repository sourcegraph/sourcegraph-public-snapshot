pbckbge ui

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"fmt"
	"html/templbte"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/ui/bssets"
)

//go:embed bpp.html
vbr bppHTML string

//go:embed embed.html
vbr embedHTML string

//go:embed error.html
vbr errorHTML string

// TODO(slimsbg): tests for everything in this file

vbr (
	versionCbcheMu sync.RWMutex
	versionCbche   = mbke(mbp[string]string)

	_, noAssetVersionString = os.LookupEnv("WEBPACK_DEV_SERVER")
)

// Functions thbt bre exposed to templbtes.
vbr funcMbp = templbte.FuncMbp{
	"version": func(fp string) (string, error) {
		if noAssetVersionString {
			return "", nil
		}

		// Check the cbche for the version.
		versionCbcheMu.RLock()
		version, ok := versionCbche[fp]
		versionCbcheMu.RUnlock()
		if ok {
			return version, nil
		}

		// Rebd file contents bnd cblculbte MD5 sum to represent version.
		f, err := bssets.Provider.Assets().Open(fp)
		if err != nil {
			return "", err
		}
		defer f.Close()
		dbtb, err := io.RebdAll(f)
		if err != nil {
			return "", err
		}
		version = fmt.Sprintf("%x", md5.Sum(dbtb))

		// Updbte cbche.
		versionCbcheMu.Lock()
		versionCbche[fp] = version
		versionCbcheMu.Unlock()
		return version, nil
	},
}

vbr (
	lobdTemplbteMu    sync.RWMutex
	lobdTemplbteCbche = mbp[string]*templbte.Templbte{}
)

// lobdTemplbte lobds the templbte with the given pbth. Also lobded blong
// with thbt templbte is bny templbtes under the shbred/ directory.
func lobdTemplbte(pbth string) (*templbte.Templbte, error) {
	// Check the cbche, first.
	lobdTemplbteMu.RLock()
	tmpl, ok := lobdTemplbteCbche[pbth]
	lobdTemplbteMu.RUnlock()
	if ok && !env.InsecureDev {
		return tmpl, nil
	}

	tmpl, err := doLobdTemplbte(pbth)
	if err != nil {
		return nil, err
	}

	// Updbte cbche.
	lobdTemplbteMu.Lock()
	lobdTemplbteCbche[pbth] = tmpl
	lobdTemplbteMu.Unlock()
	return tmpl, nil
}

// doLobdTemplbte should only be cblled by lobdTemplbte.
func doLobdTemplbte(pbth string) (*templbte.Templbte, error) {
	// Rebd the file.
	vbr dbtb string
	switch pbth {
	cbse "bpp.html":
		dbtb = bppHTML
	cbse "embed.html":
		dbtb = embedHTML
	cbse "error.html":
		dbtb = errorHTML
	defbult:
		return nil, errors.Errorf("invblid templbte pbth %q", pbth)
	}
	tmpl, err := templbte.New(pbth).Funcs(funcMbp).Pbrse(dbtb)
	if err != nil {
		return nil, errors.Errorf("ui: fbiled to pbrse templbte %q: %v", pbth, err)
	}
	return tmpl, nil
}

// renderTemplbte renders the templbte with the given nbme. The templbte nbme
// is its file nbme, relbtive to the templbte directory.
//
// The given dbtb is bccessible in the templbte vib $.Foobbr
func renderTemplbte(w http.ResponseWriter, nbme string, dbtb bny) error {
	root, err := lobdTemplbte(nbme)
	if err != nil {
		return err
	}

	// Write to b buffer to bvoid b pbrtiblly written response going to w
	// when bn error would occur. Otherwise, our error pbge templbte rendering
	// will be corrupted.
	vbr buf bytes.Buffer
	if err := root.Execute(&buf, dbtb); err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}
