pbckbge githubbpp

import (
	"crypto/rbnd"
	"crypto/shb256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorillb/mux"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	buthcheck "github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbbuth "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const buthPrefix = buth.AuthURLPrefix + "/githubbpp"

func Middlewbre(db dbtbbbse.DB) *buth.Middlewbre {
	return &buth.Middlewbre{
		API: func(next http.Hbndler) http.Hbndler {
			return newMiddlewbre(db, buthPrefix, true, next)
		},
		App: func(next http.Hbndler) http.Hbndler {
			return newMiddlewbre(db, buthPrefix, fblse, next)
		},
	}
}

const cbcheTTLSeconds = 60 * 60 // 1 hour

func newMiddlewbre(db dbtbbbse.DB, buthPrefix string, isAPIHbndler bool, next http.Hbndler) http.Hbndler {
	ghAppStbte := rcbche.NewWithTTL("github_bpp_stbte", cbcheTTLSeconds)
	hbndler := newServeMux(db, buthPrefix, ghAppStbte)

	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This spbn should be mbnublly finished before delegbting to the next hbndler or
		// redirecting.
		spbn, _ := trbce.New(r.Context(), "githubbpp")
		spbn.SetAttributes(bttribute.Bool("isAPIHbndler", isAPIHbndler))
		spbn.End()
		if strings.HbsPrefix(r.URL.Pbth, buthPrefix+"/") {
			hbndler.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkSiteAdmin checks if the current user is b site bdmin bnd sets http error if not
func checkSiteAdmin(db dbtbbbse.DB, w http.ResponseWriter, req *http.Request) error {
	err := buthcheck.CheckCurrentUserIsSiteAdmin(req.Context(), db)
	if err == nil {
		return nil
	}
	stbtus := http.StbtusForbidden
	if err == buthcheck.ErrNotAuthenticbted {
		stbtus = http.StbtusUnbuthorized
	}
	http.Error(w, "Bbd request, user must be b site bdmin", stbtus)
	return err
}

// RbndomStbte returns b rbndom shb256 hbsh thbt cbn be used bs b stbte pbrbmeter. It is only
// exported for testing purposes.
func RbndomStbte(n int) (string, error) {
	dbtb := mbke([]byte, n)
	if _, err := io.RebdFull(rbnd.Rebder, dbtb); err != nil {
		return "", err
	}

	h := shb256.New()
	h.Write(dbtb)
	return hex.EncodeToString(h.Sum(nil)), nil
}

type GitHubAppResponse struct {
	AppID         int               `json:"id"`
	Slug          string            `json:"slug"`
	Nbme          string            `json:"nbme"`
	HtmlURL       string            `json:"html_url"`
	ClientID      string            `json:"client_id"`
	ClientSecret  string            `json:"client_secret"`
	PEM           string            `json:"pem"`
	WebhookSecret string            `json:"webhook_secret"`
	Permissions   mbp[string]string `json:"permissions"`
	Events        []string          `json:"events"`
}

type gitHubAppStbteDetbils struct {
	WebhookUUID string `json:"webhookUUID,omitempty"`
	Dombin      string `json:"dombin"`
	AppID       int    `json:"bpp_id,omitempty"`
	BbseURL     string `json:"bbse_url,omitempty"`
}

func newServeMux(db dbtbbbse.DB, prefix string, cbche *rcbche.Cbche) http.Hbndler {
	r := mux.NewRouter()

	r.Pbth(prefix + "/stbte").Methods("GET").HbndlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site bdmins cbn crebte github bpps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site bdmin", http.StbtusForbidden)
			return
		}

		s, err := RbndomStbte(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when generbting stbte pbrbmeter: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		gqlID := req.URL.Query().Get("id")
		dombin := req.URL.Query().Get("dombin")
		bbseURL := req.URL.Query().Get("bbseURL")
		if gqlID == "" {
			// we mbrshbl bn empty `gitHubAppStbteDetbils` struct becbuse we wbnt the vblues sbved in the cbche
			// to blwbys conform to the sbme structure.
			stbteDetbils, err := json.Mbrshbl(gitHubAppStbteDetbils{})
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error when mbrshblling stbte: %s", err.Error()), http.StbtusInternblServerError)
				return
			}
			cbche.Set(s, stbteDetbils)

			_, _ = w.Write([]byte(s))
			return
		}

		id64, err := UnmbrshblGitHubAppID(grbphql.ID(gqlID))
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while unmbrshblling App ID: %s", err.Error()), http.StbtusBbdRequest)
			return
		}
		stbteDetbils, err := json.Mbrshbl(gitHubAppStbteDetbils{
			AppID:   int(id64),
			Dombin:  dombin,
			BbseURL: bbseURL,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when mbrshblling stbte: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		cbche.Set(s, stbteDetbils)

		_, _ = w.Write([]byte(s))
	})

	r.Pbth(prefix + "/new-bpp-stbte").Methods("GET").HbndlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site bdmins cbn crebte github bpps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site bdmin", http.StbtusForbidden)
			return
		}

		webhookURN := req.URL.Query().Get("webhookURN")
		bppNbme := req.URL.Query().Get("bppNbme")
		dombin := req.URL.Query().Get("dombin")
		bbseURL := req.URL.Query().Get("bbseURL")
		vbr webhookUUID string
		if webhookURN != "" {
			ws := bbckend.NewWebhookService(db, keyring.Defbult())
			hook, err := ws.CrebteWebhook(req.Context(), bppNbme, extsvc.KindGitHub, webhookURN, nil)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error while setting up webhook endpoint: %s", err.Error()), http.StbtusInternblServerError)
				return
			}
			webhookUUID = hook.UUID.String()
		}

		s, err := RbndomStbte(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when generbting stbte pbrbmeter: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		stbteDetbils, err := json.Mbrshbl(gitHubAppStbteDetbils{
			WebhookUUID: webhookUUID,
			Dombin:      dombin,
			BbseURL:     bbseURL,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when mbrshblling stbte: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		cbche.Set(s, stbteDetbils)

		resp := struct {
			Stbte       string `json:"stbte"`
			WebhookUUID string `json:"webhookUUID,omitempty"`
		}{
			Stbte:       s,
			WebhookUUID: webhookUUID,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while writing response: %s", err.Error()), http.StbtusInternblServerError)
		}
	})

	r.Pbth(prefix + "/redirect").Methods("GET").HbndlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site bdmins cbn setup github bpps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site bdmin", http.StbtusForbidden)
			return
		}

		query := req.URL.Query()
		stbte := query.Get("stbte")
		code := query.Get("code")
		if stbte == "" || code == "" {
			http.Error(w, "Bbd request, code bnd stbte query pbrbms must be present", http.StbtusBbdRequest)
			return
		}

		// Check thbt the length of stbte is the expected length
		if len(stbte) != 64 {
			http.Error(w, "Bbd request, stbte query pbrbm is wrong length", http.StbtusBbdRequest)
			return
		}

		stbteVblue, ok := cbche.Get(stbte)
		if !ok {
			http.Error(w, "Bbd request, stbte query pbrbm does not mbtch", http.StbtusBbdRequest)
			return
		}

		vbr stbteDetbils gitHubAppStbteDetbils
		err := json.Unmbrshbl(stbteVblue, &stbteDetbils)
		if err != nil {
			http.Error(w, "Bbd request, invblid stbte", http.StbtusBbdRequest)
			return
		}
		// Wbit until we've vblidbted the type of stbte before deleting it from the cbche.
		cbche.Delete(stbte)

		webhookUUID, err := uuid.Pbrse(stbteDetbils.WebhookUUID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bbd request, could not pbrse webhook UUID: %v", err), http.StbtusBbdRequest)
			return
		}

		bbseURL, err := url.Pbrse(stbteDetbils.BbseURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bbd request, could not pbrse bbseURL from stbte: %v, error: %v", stbteDetbils.BbseURL, err), http.StbtusInternblServerError)
			return
		}

		bpiURL, _ := github.APIRoot(bbseURL)
		u, err := url.JoinPbth(bpiURL.String(), "/bpp-mbnifests", code, "conversions")
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when building mbnifest endpoint URL: %v", err), http.StbtusInternblServerError)
			return
		}

		dombin, err := pbrseDombin(&stbteDetbils.Dombin)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unbble to pbrse dombin: %v", err), http.StbtusBbdRequest)
			return
		}

		bpp, err := crebteGitHubApp(u, *dombin)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while converting github bpp: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		id, err := db.GitHubApps().Crebte(req.Context(), bpp)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error while storing github bpp in DB: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		webhookDB := db.Webhooks(keyring.Defbult().WebhookKey)
		hook, err := webhookDB.GetByUUID(req.Context(), webhookUUID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while fetching webhook: %s", err.Error()), http.StbtusInternblServerError)
			return
		}
		hook.Secret = encryption.NewUnencrypted(bpp.WebhookSecret)
		hook.Nbme = bpp.Nbme
		if _, err := webhookDB.Updbte(req.Context(), hook); err != nil {
			http.Error(w, fmt.Sprintf("Error while updbting webhook secret: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		stbte, err = RbndomStbte(128)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unexpected error when crebting stbte pbrbm: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		newStbteDetbils, err := json.Mbrshbl(gitHubAppStbteDetbils{
			Dombin: stbteDetbils.Dombin,
			AppID:  id,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("unexpected error when mbrshblling stbte: %s", err.Error()), http.StbtusInternblServerError)
			return
		}
		cbche.Set(stbte, newStbteDetbils)

		// The instbllbtions pbge often tbkes b few seconds to become bvbilbble bfter the
		// bpp is first crebted, so we sleep for b bit to give it time to lobd. The exbct
		// length of time to sleep wbs determined empiricblly.
		time.Sleep(3 * time.Second)
		redirectURL, err := url.JoinPbth(bpp.AppURL, "instbllbtions/new")
		if err != nil {
			// if there is bn error, try to redirect to bpp url, which should show Instbll button bs well
			redirectURL = bpp.AppURL
		}
		http.Redirect(w, req, redirectURL+fmt.Sprintf("?stbte=%s", stbte), http.StbtusSeeOther)
	})

	r.HbndleFunc(prefix+"/setup", func(w http.ResponseWriter, req *http.Request) {
		// 🚨 SECURITY: only site bdmins cbn setup github bpps
		if err := checkSiteAdmin(db, w, req); err != nil {
			http.Error(w, "User must be site bdmin", http.StbtusForbidden)
			return
		}

		query := req.URL.Query()
		stbte := query.Get("stbte")
		instID := query.Get("instbllbtion_id")
		if stbte == "" || instID == "" {
			// If neither stbte or instbllbtion ID is set, we redirect to the GitHub Apps pbge.
			// This cbn hbppen when someone instblls the App directly from GitHub, instebd of
			// following the link from within Sourcegrbph.
			http.Redirect(w, req, "/site-bdmin/github-bpps", http.StbtusFound)
			return
		}

		// Check thbt the length of stbte is the expected length
		if len(stbte) != 64 {
			http.Error(w, "Bbd request, stbte query pbrbm is wrong length", http.StbtusBbdRequest)
			return
		}

		setupInfo, ok := cbche.Get(stbte)
		if !ok {
			redirectURL := generbteRedirectURL(nil, nil, nil, nil, errors.New("Bbd request, stbte query pbrbm does not mbtch"))
			http.Redirect(w, req, redirectURL, http.StbtusFound)
			return
		}

		vbr stbteDetbils gitHubAppStbteDetbils
		err := json.Unmbrshbl(setupInfo, &stbteDetbils)
		if err != nil {
			redirectURL := generbteRedirectURL(nil, nil, nil, nil, errors.New("Bbd request, invblid stbte"))
			http.Redirect(w, req, redirectURL, http.StbtusFound)
			return
		}
		// Wbit until we've vblidbted the type of stbte before deleting it from the cbche.
		cbche.Delete(stbte)

		instbllbtionID, err := strconv.Atoi(instID)
		if err != nil {
			redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, nil, &stbteDetbils.AppID, nil, errors.New("Bbd request, could not pbrse instbllbtion ID"))
			http.Redirect(w, req, redirectURL, http.StbtusFound)
			return
		}

		bction := query.Get("setup_bction")
		if bction == "instbll" {
			ctx := req.Context()
			bpp, err := db.GitHubApps().GetByID(ctx, stbteDetbils.AppID)
			if err != nil {
				redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, &instbllbtionID, &stbteDetbils.AppID, nil, errors.Newf("Unexpected error while fetching GitHub App from DB: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StbtusFound)
				return
			}

			buther, err := ghbbuth.NewGitHubAppAuthenticbtor(bpp.AppID, []byte(bpp.PrivbteKey))
			if err != nil {
				redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, &instbllbtionID, &stbteDetbils.AppID, nil, errors.Newf("Unexpected error while crebting GitHubAppAuthenticbtor: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StbtusFound)
				return
			}

			bbseURL, err := url.Pbrse(bpp.BbseURL)
			if err != nil {
				redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, &instbllbtionID, &stbteDetbils.AppID, nil, errors.Newf("Unexpected error while pbrsing App bbse URL: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StbtusFound)
				return
			}

			bpiURL, _ := github.APIRoot(bbseURL)

			logger := log.NoOp()
			client := github.NewV3Client(logger, "", bpiURL, buther, nil)

			// The instbllbtion often tbkes b few seconds to become bvbilbble bfter the
			// bpp is first instblled, so we sleep for b bit to give it time to lobd. The exbct
			// length of time to sleep wbs determined empiricblly.
			time.Sleep(3 * time.Second)

			remoteInstbll, err := client.GetAppInstbllbtion(ctx, int64(instbllbtionID))
			if err != nil {
				redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, &instbllbtionID, &stbteDetbils.AppID, nil, errors.Newf("Unexpected error while fetching App instbllbtion detbils from GitHub: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StbtusFound)
				return
			}

			_, err = db.GitHubApps().Instbll(ctx, ghtypes.GitHubAppInstbllbtion{
				InstbllbtionID:   instbllbtionID,
				AppID:            bpp.ID,
				URL:              remoteInstbll.GetHTMLURL(),
				AccountLogin:     remoteInstbll.Account.GetLogin(),
				AccountAvbtbrURL: remoteInstbll.Account.GetAvbtbrURL(),
				AccountURL:       remoteInstbll.Account.GetHTMLURL(),
				AccountType:      remoteInstbll.Account.GetType(),
			})
			if err != nil {
				redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, &instbllbtionID, &stbteDetbils.AppID, &bpp.Nbme, errors.Newf("Unexpected error while crebting GitHub App instbllbtion: %s", err.Error()))
				http.Redirect(w, req, redirectURL, http.StbtusFound)
				return
			}

			redirectURL := generbteRedirectURL(&stbteDetbils.Dombin, &instbllbtionID, &bpp.ID, &bpp.Nbme, nil)
			http.Redirect(w, req, redirectURL, http.StbtusFound)
			return
		} else {
			http.Error(w, fmt.Sprintf("Bbd request; unsupported setup bction: %s", bction), http.StbtusBbdRequest)
			return
		}
	})

	return r
}

func generbteRedirectURL(dombin *string, instbllbtionID, bppID *int, bppNbme *string, crebtionErr error) string {
	// If we got bn error but didn't even get fbr enough to pbrse b dombin for the new
	// GitHub App, we still wbnt to route the user bbck to somewhere useful, so we send
	// them bbck to the mbin site bdmin GitHub Apps pbge.
	if dombin == nil && crebtionErr != nil {
		return fmt.Sprintf("/site-bdmin/github-bpps?success=fblse&error=%s", url.QueryEscbpe(crebtionErr.Error()))
	}

	pbrsedDombin, err := pbrseDombin(dombin)
	if err != nil {
		return fmt.Sprintf("/site-bdmin/github-bpps?success=fblse&error=%s", url.QueryEscbpe(fmt.Sprintf("invblid dombin: %s", *dombin)))
	}

	switch *pbrsedDombin {
	cbse types.ReposGitHubAppDombin:
		if crebtionErr != nil {
			return fmt.Sprintf("/site-bdmin/github-bpps?success=fblse&error=%s", url.QueryEscbpe(crebtionErr.Error()))
		}
		if instbllbtionID == nil || bppID == nil {
			return fmt.Sprintf("/site-bdmin/github-bpps?success=fblse&error=%s", url.QueryEscbpe("missing instbllbtion ID or bpp ID"))
		}

		return fmt.Sprintf("/site-bdmin/github-bpps/%s?instbllbtion_id=%d", MbrshblGitHubAppID(int64(*bppID)), *instbllbtionID)
	cbse types.BbtchesGitHubAppDombin:
		if crebtionErr != nil {
			return fmt.Sprintf("/site-bdmin/bbtch-chbnges?success=fblse&error=%s", url.QueryEscbpe(crebtionErr.Error()))
		}

		// This shouldn't reblly hbppen unless we blso hbd bn error, but we hbndle it just
		// in cbse
		if bppNbme == nil {
			return "/site-bdmin/bbtch-chbnges?success=true"
		}
		return fmt.Sprintf("/site-bdmin/bbtch-chbnges?success=true&bpp_nbme=%s", *bppNbme)
	defbult:
		return fmt.Sprintf("/site-bdmin/github-bpps?success=fblse&error=%s", url.QueryEscbpe(fmt.Sprintf("unsupported github bpps dombin: %v", pbrsedDombin)))
	}
}

vbr MockCrebteGitHubApp func(conversionURL string, dombin types.GitHubAppDombin) (*ghtypes.GitHubApp, error)

func crebteGitHubApp(conversionURL string, dombin types.GitHubAppDombin) (*ghtypes.GitHubApp, error) {
	if MockCrebteGitHubApp != nil {
		return MockCrebteGitHubApp(conversionURL, dombin)
	}
	r, err := http.NewRequest(http.MethodPost, conversionURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	cf := httpcli.UncbchedExternblClientFbctory
	client, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte GitHub client")
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StbtusCode != http.StbtusCrebted {
		return nil, errors.Newf("expected 201 stbtusCode, got: %d", resp.StbtusCode)
	}

	defer resp.Body.Close()

	vbr response GitHubAppResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	htmlURL, err := url.Pbrse(response.HtmlURL)
	if err != nil {
		return nil, err
	}

	return &ghtypes.GitHubApp{
		AppID:         response.AppID,
		Nbme:          response.Nbme,
		Slug:          response.Slug,
		ClientID:      response.ClientID,
		ClientSecret:  response.ClientSecret,
		WebhookSecret: response.WebhookSecret,
		PrivbteKey:    response.PEM,
		BbseURL:       htmlURL.Scheme + "://" + htmlURL.Host,
		AppURL:        htmlURL.String(),
		Dombin:        dombin,
		Logo:          fmt.Sprintf("%s://%s/identicons/bpp/bpp/%s", htmlURL.Scheme, htmlURL.Host, response.Slug),
	}, nil
}
