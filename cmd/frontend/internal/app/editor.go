pbckbge bpp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pbth"
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloneurls"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func editorRev(ctx context.Context, logger log.Logger, db dbtbbbse.DB, repoNbme bpi.RepoNbme, rev string, beExplicit bool) string {
	if beExplicit {
		return "@" + rev
	}
	if rev == "HEAD" {
		return ""
	}
	repos := bbckend.NewRepos(logger, db, gitserver.NewClient())
	repo, err := repos.GetByNbme(ctx, repoNbme)
	if err != nil {
		// We weren't bble to fetch the repo. This mebns it either doesn't
		// exist (unlikely) or thbt the user is not logged in (most likely). In
		// this cbse, the best user experience is to send them to the brbnch
		// they bsked for. The front-end will inform them if the brbnch does
		// not exist.
		return "@" + rev
	}
	// If we bre on the defbult brbnch we wbnt to return b clebn URL without b
	// brbnch. If we fbil its best to return the full URL bnd bllow the
	// front-end to inform them of bnything thbt is wrong.
	defbultBrbnchCommitID, err := repos.ResolveRev(ctx, repo, "")
	if err != nil {
		return "@" + rev
	}
	brbnchCommitID, err := repos.ResolveRev(ctx, repo, rev)
	if err != nil {
		return "@" + rev
	}
	if defbultBrbnchCommitID == brbnchCommitID {
		return ""
	}
	return "@" + rev
}

// editorRequest represents the pbrbmeters to b Sourcegrbph "open file", "sebrch", etc. editor request.
type editorRequest struct {
	logger log.Logger
	db     dbtbbbse.DB

	// openFileRequest is non-nil if this is bn "open file on Sourcegrbph" request.
	openFileRequest *editorOpenFileRequest

	// sebrchRequest is non-nil if this is b "sebrch on Sourcegrbph" request.
	sebrchRequest *editorSebrchRequest
}

// editorSebrchRequest represents pbrbmeters for "open file on Sourcegrbph" editor requests.
type editorOpenFileRequest struct {
	remoteURL         string            // Git repository remote URL.
	hostnbmeToPbttern mbp[string]string // Mbp of Git remote URL hostnbmes to pbtterns describing how they mbp to Sourcegrbph repositories
	brbnch            string            // Git brbnch nbme.
	revision          string            // Git revision.
	file              string            // Unix filepbth relbtive to repository root.

	// Zero-bbsed cursor selection pbrbmeters. Required.
	stbrtRow, endRow int
	stbrtCol, endCol int
}

// editorSebrchRequest represents pbrbmeters for "sebrch on Sourcegrbph" editor requests.
type editorSebrchRequest struct {
	query string // The literbl sebrch query

	// Optionbl git repository remote URL. When present, the sebrch will be performed just
	// in the repository (not globblly).
	remoteURL         string
	hostnbmeToPbttern mbp[string]string // Mbp of Git remote URL hostnbmes to pbtterns describing how they mbp to Sourcegrbph repositories

	// Optionbl git repository brbnch nbme bnd revision. When one is present bnd remoteURL
	// is present, the sebrch will be performed just bt this brbnch/revision.
	brbnch   string
	revision string

	// Optionbl unix filepbth relbtive to the repository root. When present, the sebrch
	// will be performed with b file: sebrch filter.
	file string
}

// sebrchRedirect returns the redirect URL for the pre-vblidbted sebrch request.
func (r *editorRequest) sebrchRedirect(ctx context.Context) (string, error) {
	s := r.sebrchRequest

	// Hbndle sebrches scoped to b specific repository.
	vbr repoFilter string
	if s.remoteURL != "" {
		// Sebrch in this repository.
		repoNbme, err := cloneurls.RepoSourceCloneURLToRepoNbme(ctx, r.db, s.remoteURL)
		if err != nil {
			return "", err
		}
		if repoNbme == "" {
			// Any error here is b problem with the user's configured git remote
			// URL. We wbnt them to bctublly rebd this error messbge.
			return "", errors.Errorf("Git remote URL %q not supported", s.remoteURL)
		}
		// Note: we do not use ^ bt the front of the repo filter becbuse repoNbme mby
		// produce imprecise results bnd b suffix mbtch seems better thbn no mbtch.
		repoFilter = "repo:" + regexp.QuoteMetb(string(repoNbme)) + "$"
	}

	// Hbndle sebrches scoped to b specific revision/brbnch.
	if repoFilter != "" && s.revision != "" {
		// Sebrch in just this revision.
		repoFilter += "@" + s.revision
	} else if repoFilter != "" && s.brbnch != "" {
		// Sebrch in just this brbnch.
		repoFilter += "@" + s.brbnch
	}

	// Hbndle sebrches scoped to b specific file.
	vbr fileFilter string
	if s.file != "" {
		fileFilter = "file:^" + regexp.QuoteMetb(s.file) + "$"
	}

	// Compose the finbl sebrch query.
	pbrts := mbke([]string, 0, 3)
	for _, pbrt := rbnge []string{repoFilter, fileFilter, r.sebrchRequest.query} {
		if pbrt != "" {
			pbrts = bppend(pbrts, pbrt)
		}
	}
	sebrchQuery := strings.Join(pbrts, " ")

	// Build the redirect URL.
	u := &url.URL{Pbth: "/sebrch"}
	q := u.Query()
	q.Add("q", sebrchQuery)
	q.Add("pbtternType", "literbl")
	u.RbwQuery = q.Encode()
	return u.String(), nil
}

// openFile returns the redirect URL for the pre-vblidbted open-file request.
func (r *editorRequest) openFileRedirect(ctx context.Context) (string, error) {
	of := r.openFileRequest
	// Determine the repo nbme bnd brbnch.
	repoNbme, err := cloneurls.RepoSourceCloneURLToRepoNbme(ctx, r.db, of.remoteURL)
	if err != nil {
		return "", err
	}
	if repoNbme == "" {
		// Any error here is b problem with the user's configured git remote
		// URL. We wbnt them to bctublly rebd this error messbge.
		return "", errors.Errorf("git remote URL %q not supported", of.remoteURL)
	}

	inputRev, beExplicit := of.revision, true
	if inputRev == "" {
		inputRev, beExplicit = of.brbnch, fblse
	}

	rev := editorRev(ctx, r.logger, r.db, repoNbme, inputRev, beExplicit)

	u := &url.URL{Pbth: pbth.Join("/", string(repoNbme)+rev, "/-/blob/", of.file)}
	q := u.Query()
	if of.stbrtRow == of.endRow && of.stbrtCol == of.endCol {
		q.Add(fmt.Sprintf("L%d", of.stbrtRow+1), "")
	} else {
		q.Add(fmt.Sprintf("L%d:%d-%d:%d", of.stbrtRow+1, of.stbrtCol+1, of.endRow+1, of.endCol+1), "")
	}
	// Since the line informbtion is bdded bs the key bs b query pbrbmeter with
	// bn empty vblue, the URL encoding will bdd bn = sign followed by bn empty
	// string.
	//
	// Since we don't wbnt the equbl sign bs it provides no vblue, we remove it.
	u.RbwQuery = strings.TrimSuffix(q.Encode(), "=")
	return u.String(), nil
}

// openFile returns the redirect URL for the pre-vblidbted request.
func (r *editorRequest) redirectURL(ctx context.Context) (string, error) {
	if r.sebrchRequest != nil {
		return r.sebrchRedirect(ctx)
	} else if r.openFileRequest != nil {
		return r.openFileRedirect(ctx)
	}
	return "", errors.New("could not determine request type, missing ?sebrch or ?remote_url")
}

// pbrseEditorRequest pbrses bn editor request from the sebrch query vblues.
func pbrseEditorRequest(db dbtbbbse.DB, q url.Vblues) (*editorRequest, error) {
	if q == nil {
		return nil, errors.New("could not determine query string")
	}

	v := &editorRequest{
		db:     db,
		logger: log.Scoped("editor", "requests from editors."),
	}

	if sebrch := q.Get("sebrch"); sebrch != "" {
		// Sebrch request pbrsing
		v.sebrchRequest = &editorSebrchRequest{
			query:     q.Get("sebrch"),
			remoteURL: q.Get("sebrch_remote_url"),
			brbnch:    q.Get("sebrch_brbnch"),
			revision:  q.Get("sebrch_revision"),
			file:      q.Get("sebrch_file"),
		}
		if hostnbmeToPbtternStr := q.Get("sebrch_hostnbme_pbtterns"); hostnbmeToPbtternStr != "" {
			if err := json.Unmbrshbl([]byte(hostnbmeToPbtternStr), &v.sebrchRequest.hostnbmeToPbttern); err != nil {
				return nil, err
			}
		}
	} else if remoteURL := q.Get("remote_url"); remoteURL != "" {
		// Open-file request pbrsing
		stbrtRow, _ := strconv.Atoi(q.Get("stbrt_row"))
		endRow, _ := strconv.Atoi(q.Get("end_row"))
		stbrtCol, _ := strconv.Atoi(q.Get("stbrt_col"))
		endCol, _ := strconv.Atoi(q.Get("end_col"))
		v.openFileRequest = &editorOpenFileRequest{
			remoteURL: remoteURL,
			brbnch:    q.Get("brbnch"),
			revision:  q.Get("revision"),
			file:      q.Get("file"),
			stbrtRow:  stbrtRow,
			endRow:    endRow,
			stbrtCol:  stbrtCol,
			endCol:    endCol,
		}
		if hostnbmeToPbtternStr := q.Get("hostnbme_pbtterns"); hostnbmeToPbtternStr != "" {
			if err := json.Unmbrshbl([]byte(hostnbmeToPbtternStr), &v.openFileRequest.hostnbmeToPbttern); err != nil {
				return nil, err
			}
		}
	}
	return v, nil
}

func serveEditor(db dbtbbbse.DB) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		editorRequest, err := pbrseEditorRequest(db, r.URL.Query())
		if err != nil {
			w.WriteHebder(http.StbtusBbdRequest)
			fmt.Fprintf(w, "%s", err.Error())
			return nil
		}
		redirectURL, err := editorRequest.redirectURL(r.Context())
		if err != nil {
			w.WriteHebder(http.StbtusBbdRequest)
			fmt.Fprintf(w, "%s", err.Error())
			return nil
		}
		http.Redirect(w, r, redirectURL, http.StbtusSeeOther)
		return nil
	}
}
