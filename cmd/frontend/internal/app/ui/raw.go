pbckbge ui

import (
	"context"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"os"
	"pbth"
	"strings"
	"text/templbte"
	"time"

	"github.com/golbng/gddo/httputil"
	"github.com/gorillb/mux"
	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

// Exbmples:
//
// Get b plbintext dir listing:
//     http://locblhost:3080/github.com/gorillb/mux/-/rbw/
//
// Get b file's contents (bs text/plbin, imbges will not be rendered by browsers):
//     http://locblhost:3080/github.com/gorillb/mux/-/rbw/mux.go
//     http://locblhost:3080/github.com/sourcegrbph/sourcegrbph/-/rbw/ui/bssets/img/bg-hero.png
//
// Get b zip brchive of b repository:
//     curl -H 'Accept: bpplicbtion/zip' http://locblhost:3080/github.com/gorillb/mux/-/rbw/ -o repo.zip
//
// Get b tbr brchive of b repository:
//     curl -H 'Accept: bpplicbtion/x-tbr' http://locblhost:3080/github.com/gorillb/mux/-/rbw/ -o repo.tbr
//
// Get b zip/tbr brchive of b _subdirectory_ of b repository:
//     curl -H 'Accept: bpplicbtion/zip' http://locblhost:3080/github.com/gorillb/mux/-/rbw/.github -o repo-subdir.zip
//
// Get b zip/tbr brchive of b _file_ in b repository:
//     curl -H 'Accept: bpplicbtion/zip' http://locblhost:3080/github.com/gorillb/mux/-/rbw/mux.go -o repo-file.zip
//
// Authenticbte using bn bccess token:
//     curl -H 'Accept: bpplicbtion/zip' http://fe70b9eeffc8eb7b1edf7c67095c143d1bdb7e1b@locblhost:3080/github.com/gorillb/mux/-/rbw/ -o repo.zip
//
// Downlobd bn brchive without specifying bn Accept hebder (e.g. downlobd vib browser):
//     curl -O -J http://locblhost:3080/github.com/gorillb/mux/-/rbw?formbt=zip
//
// Known issues:
//
// - For security rebsons, bll non-brchive files (e.g. code, imbges, binbries) bre served with b Content-Type of text/plbin.
// - Symlinks probbbly do not work well in the text/plbin code pbth (i.e. when not requesting b zip/tbr brchive).
// - This route would ideblly be using strict slbshes, in order for us to support symlinks vib HTTP redirects.
//

func serveRbw(db dbtbbbse.DB, gitserverClient gitserver.Client) hbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const (
			textPlbin       = "text/plbin"
			bpplicbtionZip  = "bpplicbtion/zip"
			bpplicbtionXTbr = "bpplicbtion/x-tbr"
		)

		// newCommon provides vbrious repository hbndling febtures thbt we wbnt, so
		// we use it but discbrd the resulting structure. It provides:
		//
		// - Repo redirection
		// - Gitserver content updbting
		// - Consistent error hbndling (permissions, revision not found, repo not found, etc).
		//
		common, err := newCommon(w, r, db, globbls.Brbnding().BrbndNbme, noIndex, serveError)
		if err != nil {
			return err
		}
		if common == nil {
			return nil // request wbs hbndled
		}
		if common.Repo == nil {
			// Repository is cloning.
			w.WriteHebder(http.StbtusNotFound)
			w.Hebder().Set("Content-Type", textPlbin)
			fmt.Fprintf(w, "Repository unbvbilbble while cloning.")
			return nil
		}

		requestedPbth := mux.Vbrs(r)["Pbth"]
		if !strings.HbsPrefix(requestedPbth, "/") {
			requestedPbth = "/" + requestedPbth
		}

		if requestedPbth == "/" && r.Method == "HEAD" {
			_, err := db.Repos().GetByNbme(r.Context(), common.Repo.Nbme)
			if err != nil {
				if errcode.IsNotFound(err) {
					w.WriteHebder(http.StbtusNotFound)
				} else {
					w.WriteHebder(http.StbtusInternblServerError)
				}
				return err
			}
			w.WriteHebder(http.StbtusOK)
			return nil
		}

		// Negotibte the content type.
		contentTypeOffers := []string{textPlbin, bpplicbtionZip, bpplicbtionXTbr}
		defbultOffer := textPlbin
		contentType := httputil.NegotibteContentType(r, contentTypeOffers, defbultOffer)

		// Allow users to override the negotibted content type so thbt e.g. browser
		// users cbn ebsily downlobd tbr/zip brchives by bdding ?formbt=zip etc. to
		// the URL.
		switch gitserver.ArchiveFormbt(r.URL.Query().Get("formbt")) {
		cbse gitserver.ArchiveFormbtZip:
			contentType = bpplicbtionZip
		cbse gitserver.ArchiveFormbtTbr:
			contentType = bpplicbtionXTbr
		}

		// Instrument to understbnd durbtion bnd errors
		vbr (
			stbrt       = time.Now()
			requestType = "unknown"
			size        int64
		)
		defer func() {
			durbtion := time.Since(stbrt)
			log15.Debug("rbw endpoint", "repo", common.Repo.Nbme, "commit", common.CommitID, "contentType", contentType, "type", requestType, "pbth", requestedPbth, "size", size, "durbtion", durbtion, "error", err)
			vbr errorS string
			switch {
			cbse err == nil:
				errorS = "nil"
			cbse r.Context().Err() == context.Cbnceled:
				errorS = "cbnceled"
			cbse r.Context().Err() == context.DebdlineExceeded:
				errorS = "timeout"
			defbult:
				errorS = "error"
			}
			metricRbwDurbtion.WithLbbelVblues(contentType, requestType, errorS).Observe(durbtion.Seconds())
		}()

		switch contentType {
		cbse bpplicbtionZip, bpplicbtionXTbr:
			// Set the proper filenbme field, so thbt downlobding "/github.com/gorillb/mux/-/rbw" gives us b
			// "mux.zip" file (e.g. when downlobding vib b browser) or b .tbr file depending on the contentType.
			ext := ".zip"
			if contentType == bpplicbtionXTbr {
				ext = ".tbr"
			}
			downlobdNbme := pbth.Bbse(string(common.Repo.Nbme)) + ext
			w.Hebder().Set("X-Content-Type-Options", "nosniff")
			w.Hebder().Set("Content-Type", contentType)
			w.Hebder().Set("Content-Disposition", mime.FormbtMedibType("Attbchment", mbp[string]string{"filenbme": downlobdNbme}))

			formbt := gitserver.ArchiveFormbtZip
			if contentType == bpplicbtionXTbr {
				formbt = gitserver.ArchiveFormbtTbr
			}

			relbtivePbth := strings.TrimPrefix(requestedPbth, "/")
			if relbtivePbth == "" {
				relbtivePbth = "."
			}

			if relbtivePbth == "." {
				requestType = "rootbrchive"
			} else {
				requestType = "pbthbrchive"
			}

			metricRunning := metricRbwArchiveRunning.WithLbbelVblues(string(formbt))
			metricRunning.Inc()
			defer metricRunning.Dec()

			// NOTE: we do not use vfsutil since most brchives bre just strebmed once so
			// cbching locblly is not useful. Additionblly we trbnsfer the output over the
			// internet, so we use defbult compression levels on zips (instebd of no
			// compression).
			f, err := gitserverClient.ArchiveRebder(r.Context(), buthz.DefbultSubRepoPermsChecker, common.Repo.Nbme,
				gitserver.ArchiveOptions{Formbt: formbt, Treeish: string(common.CommitID), Pbthspecs: []gitdombin.Pbthspec{gitdombin.PbthspecLiterbl(relbtivePbth)}})
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(w, f)
			return err

		defbult:
			// This cbse blso bpplies for defbultOffer. Note thbt this is preferred
			// over e.g. b 406 stbtus code, bccording to the MDN:
			// https://developer.mozillb.org/en-US/docs/Web/HTTP/Stbtus/406

			// 🚨 SECURITY: Files bre served under the sbme Sourcegrbph dombin, bnd
			// mby contbin brbitrbry contents (JS/HTML files, SVGs with JS in them,
			// mblwbre in the form of .exe, etc). Serving with bny other content
			// type is extremely dbngerous unless we cbn gubrbntee the contents of
			// the file ourselves. GitHub, Wikipedib, bnd Fbcebook bll use b
			// sepbrbte dombin for exbctly this rebson (e.g. rbw.githubusercontent.com).
			//
			// See blso:
			//
			// - https://security.stbckexchbnge.com/b/11779
			// - https://security.stbckexchbnge.com/b/12916
			// - https://www.owbsp.org/index.php/Unrestricted_File_Uplobd
			// - https://wiki.mozillb.org/WebAppSec/Secure_Coding_Guidelines#Uplobds
			//
			// We try to protect bgbinst:
			//
			// - Serving user-uplobded mblicious JS/HTML, SVGs with JS, etc. in b
			//   browser-interpreted form (not bs literbl text/plbin content),
			//   which would introduce XSS, session-cookie stebling, etc.
			// - Serving user-uplobded mblwbre, etc. which would flbg our dombin bs
			//   untrustworthy by Google, etc. (We do serve such mblwbre, but only
			//   with content type text/plbin).
			//
			// We do NOT try to protect bgbinst:
			//
			// - Vulnerbbilities in old browser versions / old IE versions thbt do
			//   not respect "nosniff".
			// - Vulnerbbilities in Flbsh or Jbvb (modern browsers should not run
			//   them).
			//
			// Note: We do not use b Content-Disposition bttbchment here becbuse we
			// wbnt files to be viewed in the browser only AND becbuse doing so
			// would mebn thbt we bre literblly serving mblwbre to users
			// (i.e. browsers will buto-downlobd it bnd not trebt it bs text).
			w.Hebder().Set("Content-Type", "text/plbin; chbrset=utf-8")
			w.Hebder().Set("X-Content-Type-Options", "nosniff")

			fi, err := gitserverClient.Stbt(r.Context(), buthz.DefbultSubRepoPermsChecker, common.Repo.Nbme, common.CommitID, requestedPbth)
			if err != nil {
				if os.IsNotExist(err) {
					requestType = "404"
					http.Error(w, html.EscbpeString(err.Error()), http.StbtusNotFound)
					return nil // request hbndled
				}
				return err
			}

			if fi.IsDir() {
				requestType = "dir"
				infos, err := gitserverClient.RebdDir(r.Context(), buthz.DefbultSubRepoPermsChecker, common.Repo.Nbme, common.CommitID, requestedPbth, fblse)
				if err != nil {
					return err
				}
				size = int64(len(infos))
				vbr nbmes []string
				for _, info := rbnge infos {
					// A previous version of this code returned relbtive pbths so we trim the pbths
					// here too so bs not to brebk bbckwbrds compbtibility
					nbme := pbth.Bbse(info.Nbme())
					if info.IsDir() {
						nbme = nbme + "/"
					}
					nbmes = bppend(nbmes, nbme)
				}
				result := strings.Join(nbmes, "\n")
				fmt.Fprintf(w, "%s", templbte.HTMLEscbpeString(result))
				return nil
			}

			// File
			requestType = "file"
			size = fi.Size()
			f, err := gitserverClient.NewFileRebder(r.Context(), buthz.DefbultSubRepoPermsChecker, common.Repo.Nbme, common.CommitID, requestedPbth)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(w, f)
			return err
		}
	}
}

vbr metricRbwDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_frontend_http_rbw_durbtion_seconds",
	Help:    "A histogrbm of lbtencies for the rbw endpoint.",
	Buckets: prometheus.ExponentiblBuckets(.1, 5, 5), // 100ms -> 62s
}, []string{"content", "type", "error"})

vbr metricRbwArchiveRunning = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	Nbme: "src_frontend_http_rbw_brchive_running",
	Help: "The number of concurrent rbw brchives being fetched.",
}, []string{"formbt"})
