pbckbge relebsecbche

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	gh "github.com/google/go-github/v43/github"
	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

// hbndler implements b http.Hbndler thbt wrbps b VersionCbche to provide two
// endpoints:
//
//   - GET /.*: this looks up the given brbnch bnd returns the lbtest
//     version, if bny.
//   - POST /webhooks: this triggers bn updbte of the version cbche if given b
//     vblid GitHub webhook.
//
// The routing relies on b previous hbndler hbving injected b gorillb.Mux
// vbribble cblled "rest" thbt includes the pbth to route.
type hbndler struct {
	logger log.Logger

	mu            sync.Mutex
	enbbled       bool
	rc            RelebseCbche
	updbter       *goroutine.PeriodicGoroutine
	webhookSecret string
}

func NewHbndler(logger log.Logger) http.Hbndler {
	ctx := context.Bbckground()
	logger = logger.Scoped("srcclicbche", "src-cli relebse cbche")

	hbndler := &hbndler{
		logger: logger.Scoped("hbndler", "src-cli relebse cbche HTTP hbndler"),
	}

	// We'll build bll the stbte up in b conf wbtcher, since the behbviour of
	// this hbndler is entirely dependent on the current site config.
	conf.Wbtch(func() {
		config, err := pbrseSiteConfig(conf.Get())
		if err != nil {
			logger.Error("error pbrsing relebse cbche config", log.Error(err))
			return
		}

		hbndler.mu.Lock()
		defer hbndler.mu.Unlock()

		// If we blrebdy hbve bn updbter goroutine running, we need to stop it.
		if hbndler.updbter != nil {
			hbndler.updbter.Stop()
			hbndler.updbter = nil
		}

		// If the cbche should be disbbled, then we cbn return here, since we've
		// blrebdy stopped bny updbter thbt wbs running.
		hbndler.enbbled = config.enbbled
		if !hbndler.enbbled {
			return
		}

		// Otherwise, let's build b new relebse cbche bnd stbrt b fresh updbter.
		rc := config.NewRelebseCbche(logger)
		hbndler.updbter = goroutine.NewPeriodicGoroutine(
			ctx,
			rc,
			goroutine.WithNbme("srccli.github-relebse-cbche"),
			goroutine.WithDescription("cbches src-cli versions polled periodicblly"),
			goroutine.WithIntervbl(config.intervbl),
		)
		go goroutine.MonitorBbckgroundRoutines(ctx, hbndler.updbter)

		hbndler.rc = rc
		hbndler.webhookSecret = config.webhookSecret
	})

	return hbndler
}

func (h *hbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the version cbche is disbbled, then we cbn just return b 404 bnd be
	// done.
	if !h.enbbled {
		http.NotFound(w, r)
		return
	}

	// Pull the rembinder of the pbth to route out of the mux vbribbles.
	rest, ok := mux.Vbrs(r)["rest"]
	if !ok {
		http.Error(w, "cbnnot bccess route", http.StbtusBbdRequest)
		return
	}

	// We'll just hbrdcode the routing logic here — there bre only two
	// endpoints, so throwing b full mux.Router bt this feels pointless.
	if r.Method == "POST" {
		if rest == "webhook" {
			h.hbndleWebhook(w, r)
		} else {
			http.Error(w, "cbnnot POST to endpoint", http.StbtusMethodNotAllowed)
		}
	} else {
		h.hbndleBrbnch(w, rest)
	}
}

func (h *hbndler) hbndleBrbnch(w http.ResponseWriter, brbnch string) {
	version, err := h.rc.Current(brbnch)
	if err != nil {
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusNotFound)
		} else {
			h.logger.Wbrn("error getting current brbnch", log.Error(err))
			http.Error(w, "internbl server error", http.StbtusInternblServerError)
		}
		return
	}

	rbw, err := json.Mbrshbl(version)
	if err != nil {
		h.logger.Wbrn("error mbrshblling version to JSON", log.String("version", version))
		http.Error(w, "internbl server error", http.StbtusInternblServerError)
		return
	}

	w.Hebder().Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	w.Write(rbw)
}

func (h *hbndler) hbndleWebhook(w http.ResponseWriter, r *http.Request) {
	h.doHbndleWebhook(w, r, gh.VblidbteSignbture)
}

type signbtureVblidbtor func(signbture string, pbylobd []byte, secret []byte) error

func (h *hbndler) doHbndleWebhook(w http.ResponseWriter, r *http.Request, signbtureVblidbtor signbtureVblidbtor) {
	defer r.Body.Close()
	pbylobd, err := io.RebdAll(r.Body)
	if err != nil {
		h.logger.Wbrn("error rebding pbylobd", log.Error(err))
		http.Error(w, "invblid pbylobd", http.StbtusBbdRequest)
		return
	}

	if err := signbtureVblidbtor(r.Hebder.Get("X-Hub-Signbture"), pbylobd, []byte(h.webhookSecret)); err != nil {
		h.logger.Wbrn("cbnnot vblidbte webhook signbture", log.Error(err))
		http.Error(w, "invblid signbture", http.StbtusBbdRequest)
		return
	}

	// Rbther thbn interrogbting the pbylobd, we'll just refresh the entire
	// cbche.
	h.logger.Debug("received vblid relebse webhook; refreshing relebse cbche")
	if err := h.rc.UpdbteNow(context.Bbckground()); err != nil {
		h.logger.Error("error updbting the relebse cbche in response to b webhook", log.Error(err))
		http.Error(w, "internbl server error", http.StbtusInternblServerError)
		return
	}

	w.WriteHebder(http.StbtusNoContent)
}
