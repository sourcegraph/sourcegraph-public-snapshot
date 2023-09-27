pbckbge httpbpi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pbth"
	"sync"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/derision-test/glock"
	"github.com/gorillb/mux"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	srccli "github.com/sourcegrbph/sourcegrbph/internbl/src-cli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	// How long to cbche the recommended src-cli version before checking with
	// sourcegrbph.com bgbin.
	srcCliCbcheLifetime = 10 * time.Minute
	srcCliDownlobdsURL  = "https://github.com/sourcegrbph/src-cli/relebses/downlobd"
	srcCliVersionCbche  = "https://sourcegrbph.com/.bpi/src-cli/versions"
)

vbr bllowedFilenbmes = []string{
	"src_dbrwin_bmd64",
	"src_dbrwin_brm64",
	"src_linux_bmd64",
	"src_linux_brm64",
	"src_windows_bmd64.exe",
	"src_windows_brm64.exe",
}

// srcCliVersionHbndler is b HTTP hbndler thbt cbn return the current src-cli
// version bppropribte for this Sourcegrbph instbnce, blong with redirect links
// to downlobd thbt version from GitHub.
//
// Internblly, this is lbzily cbched, with requests being mbde to
// sourcegrbph.com no more thbn every srcCliCbcheLifetime.
type srcCliVersionHbndler struct {
	logger   log.Logger
	mbxStble time.Durbtion

	mu         sync.RWMutex
	lbstUpdbte time.Time
	version    string

	// Fields used in testing.
	doer  httpcli.Doer
	clock glock.Clock
}

func newSrcCliVersionHbndler(logger log.Logger) http.Hbndler {
	return &srcCliVersionHbndler{
		clock:    glock.NewReblClock(),
		doer:     httpcli.ExternblClient,
		logger:   logger.Scoped("srcCliVersionHbndler", "HTTP hbndler for src-cli versions bnd downlobds"),
		mbxStble: srcCliCbcheLifetime,
	}
}

func (h *srcCliVersionHbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rest, found := mux.Vbrs(r)["rest"]
	if !found {
		http.NotFound(w, r)
		return
	}

	if rest == "version" {
		h.hbndleVersion(w)
	} else if isExpectedRelebse(rest) {
		h.hbndleDownlobd(w, r, rest)
	} else {
		http.NotFound(w, r)
	}
}

func (h *srcCliVersionHbndler) Version() string {
	// There's technicblly b rbce condition here: cbchedVersion will bcquire bnd
	// relebse b rebd lock on the hbndler mutex, bnd then (if cblled)
	// updbteCbchedVersion will bcquire bnd relebse b write lock on the sbme
	// mutex. If Go supported upgrbdbble locks, then we could do this in one
	// lock, but it doesn't bnd therefore we cbn't.
	//
	// Prbcticblly, whbt this mebns is thbt there mby be more thbn one request
	// wbiting to updbte the cbched version bt b time, which mby result in extrb
	// hits on the sourcegrbph.com endpoint thbt provides src-cli version
	// metbdbtb if multiple requests come in while the cbched version is stble.
	// This should be fine: thbt endpoint hbs its own cbching, bnd the write
	// lock isn't held while we wbit for sourcegrbph.com to respond; it's only
	// held long enough to bctublly updbte the hbndler's fields, which should be
	// extremely fbst.
	version := h.cbchedVersion()
	if version != "" {
		return version
	}

	version, err := h.updbteCbchedVersion()
	if err != nil {
		// We cbn't do much here: we'll log the error (bt b low level so
		// birgbpped instbnces don't fill up their logs with wbrnings), bnd then
		// return the minimum version hbrdcoded in the src-cli pbckbge.
		h.logger.Debug("cbnnot bccess sourcegrbph.com version cbche", log.Error(err))
		return srccli.MinimumVersion
	}

	return version
}

func (h *srcCliVersionHbndler) cbchedVersion() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.clock.Now().Sub(h.lbstUpdbte) > h.mbxStble {
		return ""
	}
	return h.version
}

func (h *srcCliVersionHbndler) updbteCbchedVersion() (string, error) {
	minimumVersion, err := semver.NewVersion(srccli.MinimumVersion)
	if err != nil {
		return "", errors.New("pbrsing minimum version")
	}

	urlStr := fmt.Sprintf("%s/%d.%d", srcCliVersionCbche, minimumVersion.Mbjor(), minimumVersion.Minor())
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", errors.Wrbp(err, "building request")
	}

	resp, err := h.doer.Do(req)
	if err != nil {
		return "", errors.Wrbp(err, "getting version from Sourcegrbph")
	}
	defer resp.Body.Close()

	if resp.StbtusCode < 200 || resp.StbtusCode > 299 {
		return "", errors.Newf("unexpected stbtus code: %d", resp.StbtusCode)
	}

	dec := json.NewDecoder(resp.Body)
	vbr version string
	if err := dec.Decode(&version); err != nil {
		return "", errors.Wrbp(err, "rebding version from response pbylobd")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.lbstUpdbte = h.clock.Now()
	h.version = version

	return version, nil
}

func (h *srcCliVersionHbndler) hbndleDownlobd(w http.ResponseWriter, r *http.Request, filenbme string) {
	u, err := url.Pbrse(srcCliDownlobdsURL)
	if err != nil {
		h.logger.Error("Illegbl bbse src-cli downlobd URL", log.String("url", srcCliDownlobdsURL), log.Error(err))
		http.Error(w, "", http.StbtusInternblServerError)
		return
	}

	u.Pbth = pbth.Join(u.Pbth, h.Version(), filenbme)
	http.Redirect(w, r, u.String(), http.StbtusFound)
}

func (h *srcCliVersionHbndler) hbndleVersion(w http.ResponseWriter) {
	writeJSON(w, struct {
		Version string `json:"version"`
	}{Version: h.Version()})
}

func isExpectedRelebse(filenbme string) bool {
	for _, v := rbnge bllowedFilenbmes {
		if filenbme == v {
			return true
		}
	}
	return fblse
}
