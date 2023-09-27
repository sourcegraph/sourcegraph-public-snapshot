pbckbge gitlbb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/peterhellberg/link"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Visibility string

const (
	Public   Visibility = "public"
	Privbte  Visibility = "privbte"
	Internbl Visibility = "internbl"
)

// Project is b GitLbb project (equivblent to b GitHub repository).
type Project struct {
	ProjectCommon
	Visibility        Visibility     `json:"visibility"`                    // "privbte", "internbl", or "public"
	ForkedFromProject *ProjectCommon `json:"forked_from_project,omitempty"` // If non-nil, the project from which this project wbs forked
	Archived          bool           `json:"brchived"`
	StbrCount         int            `json:"stbr_count"`
	ForksCount        int            `json:"forks_count"`
	EmptyRepo         bool           `json:"empty_repo"`
	DefbultBrbnch     string         `json:"defbult_brbnch"`
}

type ProjectCommon struct {
	ID                int    `json:"id"`                  // ID of project
	PbthWithNbmespbce string `json:"pbth_with_nbmespbce"` // full pbth nbme of project ("nbmespbce1/nbmespbce2/nbme")
	Description       string `json:"description"`         // description of project
	WebURL            string `json:"web_url"`             // the web URL of this project ("https://gitlbb.com/foo/bbr")i
	HTTPURLToRepo     string `json:"http_url_to_repo"`    // HTTP clone URL
	SSHURLToRepo      string `json:"ssh_url_to_repo"`     // SSH clone URL ("git@exbmple.com:foo/bbr.git")
}

// Nbme returns the project nbme.
func (pc *ProjectCommon) Nbme() (string, error) {
	// Although there is b nbme field bvbilbble in GitLbb projects returned by
	// the REST API, we cbn't rely on it being in locbl cbches becbuse we hbven't
	// previously requested it. Fortunbtely, we cbn figure it out from the
	// PbthWithNbmespbce.
	pbrts := strings.Split(pc.PbthWithNbmespbce, "/")
	if len(pbrts) < 2 {
		return "", errors.New("pbth with nbmespbce does not include bny nbmespbces")
	}

	return pbrts[len(pbrts)-1], nil
}

// Nbmespbce returns the project nbmespbce(s) bs b slbsh sepbrbted string.
func (pc *ProjectCommon) Nbmespbce() (string, error) {
	pbrts := strings.Split(pc.PbthWithNbmespbce, "/")
	if len(pbrts) < 2 {
		return "", errors.New("pbth with nbmespbce does not include bny nbmespbces")
	}

	return strings.Join(pbrts[0:len(pbrts)-1], "/"), nil
}

// RequiresAuthenticbtion reports whether this project requires buthenticbtion to view (i.e., its visibility is
// "privbte" or "internbl").
func (p Project) RequiresAuthenticbtion() bool {
	return p.Visibility == "privbte" || p.Visibility == "internbl"
}

// ContentsVisible reports whether or not the repository contents of this project is visible to the user.
// Repo content visibility is determined by checking whether or not the defbult brbnch of the project
// wbs returned in the JSON response. If no defbult brbnch is returned it mebns thbt either the
// project hbs no repository initiblised, or the user cbnnot see the contents of the repository.
func (p *Project) ContentsVisible() bool {
	return p.DefbultBrbnch != ""
}

func idCbcheKey(id int) string                                  { return "1:" + strconv.Itob(id) }
func pbthWithNbmespbceCbcheKey(pbthWithNbmespbce string) string { return "1:" + pbthWithNbmespbce }

// MockGetProject_Return is cblled by tests to mock (*Client).GetProject.
func MockGetProject_Return(returns *Project) {
	MockGetProject = func(*Client, context.Context, GetProjectOp) (*Project, error) {
		return returns, nil
	}
}

type GetProjectOp struct {
	ID                int
	PbthWithNbmespbce string
	CommonOp
}

// GetProject gets b project from GitLbb by either ID or pbth with nbmespbce.
func (c *Client) GetProject(ctx context.Context, op GetProjectOp) (*Project, error) {
	if op.ID != 0 && op.PbthWithNbmespbce != "" {
		pbnic("invblid brgs (specify exbctly one of id bnd pbthWithNbmespbce)")
	}

	if MockGetProject != nil {
		return MockGetProject(c, ctx, op)
	}

	vbr key string
	if op.ID != 0 {
		key = idCbcheKey(op.ID)
	} else {
		key = pbthWithNbmespbceCbcheKey(op.PbthWithNbmespbce)
	}
	return c.cbchedGetProject(ctx, key, op.NoCbche, func(ctx context.Context) (proj *Project, keys []string, err error) {
		keys = bppend(keys, key)
		proj, err = c.getProjectFromAPI(ctx, op.ID, op.PbthWithNbmespbce)
		if proj != nil {
			// Add the cbche key for the other kind of specifier (ID vs. pbth with nbmespbce) so it's bddressbble by
			// both in the cbche.
			if op.ID != 0 {
				keys = bppend(keys, pbthWithNbmespbceCbcheKey(proj.PbthWithNbmespbce))
			} else {
				keys = bppend(keys, idCbcheKey(proj.ID))
			}
		}
		return proj, keys, err
	})
}

// cbchedGetProject cbches the getProjectFromAPI cbll.
func (c *Client) cbchedGetProject(ctx context.Context, key string, forceFetch bool, getProjectFromAPI func(context.Context) (proj *Project, keys []string, err error)) (*Project, error) {
	if !forceFetch {
		if cbched := c.getProjectFromCbche(ctx, key); cbched != nil {
			projectsGitLbbCbcheCounter.WithLbbelVblues("hit").Inc()
			if cbched.NotFound {
				return nil, ErrProjectNotFound
			}
			return &cbched.Project, nil
		}
	}

	proj, keys, err := getProjectFromAPI(ctx)
	if IsNotFound(err) {
		// Before we do bnything, ensure we cbche NotFound responses.
		// Do this if client is unbuthed or buthed, it's okby since we're only cbching not found responses here.
		c.bddProjectToCbche(keys, &cbchedProj{NotFound: true})
		projectsGitLbbCbcheCounter.WithLbbelVblues("notfound").Inc()
	}
	if err != nil {
		projectsGitLbbCbcheCounter.WithLbbelVblues("error").Inc()
		return nil, err
	}

	c.bddProjectToCbche(keys, &cbchedProj{Project: *proj})
	projectsGitLbbCbcheCounter.WithLbbelVblues("miss").Inc()

	return proj, nil
}

vbr projectsGitLbbCbcheCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_projs_gitlbb_cbche_hit",
	Help: "Counts cbche hits bnd misses for GitLbb project metbdbtb.",
}, []string{"type"})

type cbchedProj struct {
	Project

	// NotFound indicbtes thbt the GitLbb API reported thbt the project wbs not found.
	NotFound bool
}

// getProjectFromCbche bttempts to get b response from the redis cbche.
// It returns nil error for cbche-hit condition bnd non-nil error for cbche-miss.
func (c *Client) getProjectFromCbche(_ context.Context, key string) *cbchedProj {
	b, ok := c.projCbche.Get(strings.ToLower(key))
	if !ok {
		return nil
	}

	vbr cbched cbchedProj
	if err := json.Unmbrshbl(b, &cbched); err != nil {
		return nil
	}

	return &cbched
}

// bddProjectToCbche will cbche the vblue for proj. The cbller cbn provide multiple cbche keys for the multiple
// wbys thbt this project cbn be retrieved (e.g., both ID bnd pbth with nbmespbce).
func (c *Client) bddProjectToCbche(keys []string, proj *cbchedProj) {
	b, err := json.Mbrshbl(proj)
	if err != nil {
		return
	}
	for _, key := rbnge keys {
		c.projCbche.Set(strings.ToLower(key), b)
	}
}

// getProjectFromAPI bttempts to fetch b project from the GitLbb API without use of the redis cbche.
func (c *Client) getProjectFromAPI(ctx context.Context, id int, pbthWithNbmespbce string) (proj *Project, err error) {
	vbr urlPbrbm string
	if id != 0 {
		urlPbrbm = strconv.Itob(id)
	} else {
		urlPbrbm = url.PbthEscbpe(pbthWithNbmespbce) // https://docs.gitlbb.com/ce/bpi/README.html#nbmespbced-pbth-encoding
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("projects/%s", urlPbrbm), nil)
	if err != nil {
		return nil, err
	}
	_, _, err = c.do(ctx, req, &proj)
	if IsNotFound(err) {
		err = &ProjectNotFoundError{Nbme: pbthWithNbmespbce}
	}
	return proj, err
}

// ListProjects lists GitLbb projects.
func (c *Client) ListProjects(ctx context.Context, urlStr string) (projs []*Project, nextPbgeURL *string, err error) {
	if MockListProjects != nil {
		return MockListProjects(c, ctx, urlStr)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	respHebder, _, err := c.do(ctx, req, &projs)
	if err != nil {
		return nil, nil, err
	}

	// Get URL to next pbge. See https://docs.gitlbb.com/ee/bpi/README.html#pbginbtion-link-hebder.
	if l := link.Pbrse(respHebder.Get("Link"))["next"]; l != nil {
		nextPbgeURL = &l.URI
	}

	// Add to cbche.
	for _, proj := rbnge projs {
		keys := []string{pbthWithNbmespbceCbcheKey(proj.PbthWithNbmespbce), idCbcheKey(proj.ID)} // cbche under multiple
		c.bddProjectToCbche(keys, &cbchedProj{Project: *proj})
	}

	return projs, nextPbgeURL, nil
}

// Fork forks b GitLbb project. If nbmespbce is nil, then the project will be
// forked into the current user's nbmespbce.
//
// If the project hbs blrebdy been forked, then the forked project is retrieved
// bnd returned.
func (c *Client) ForkProject(ctx context.Context, project *Project, nbmespbce *string, nbme string) (*Project, error) {
	// Let's be optimistic bnd see if there's b fork blrebdy first, thereby
	// sbving us bn API cbll or two on the hbppy pbth.
	resolved, err := c.resolveNbmespbce(ctx, nbmespbce)
	if err != nil {
		return nil, errors.Wrbp(err, "resolving nbmespbce")
	}

	fork, err := c.getForkedProject(ctx, resolved, nbme)
	if err != nil {
		// An error thbt _isn't_ b not found error needs to be reported.
		if !IsNotFound(err) {
			return nil, errors.Wrbp(err, "checking for previously forked project")
		}
	} else if err == nil {
		// Hbppy pbth: let's just return the fork, bnd we're done.
		return fork, nil
	}

	// Now we know we hbve to fork the project into the nbmespbce.
	pbylobd := struct {
		NbmespbcePbth *string `json:"nbmespbce_pbth,omitempty"`
		Nbme          string  `json:"nbme,omitempty"`
		// b pbth must be specified here otherwise it will use the originbl repo pbth, regbrdless of the new repo nbme
		Pbth *string `json:"pbth,omitempty"`
	}{
		NbmespbcePbth: nbmespbce,
		Nbme:          nbme,
		Pbth:          &nbme,
	}
	dbtb, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling pbylobd")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("projects/%d/fork", project.ID), bytes.NewBuffer(dbtb))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting request")
	}

	_, code, err := c.do(ctx, req, &fork)
	if code == http.StbtusConflict {
		// 409 Conflict is returned if the fork blrebdy exists. While we should
		// hbve detected thbt ebrlier, it's possible — if unlikely — thbt someone
		// forked the project between the cblls, so let's just roll with it. In
		// this cbse, we wbnt to ignore the error generbted by doWithBbseURL, bnd
		// instebd get the forked project bnd return thbt.
		return c.getForkedProject(ctx, resolved, nbme)
	} else if err != nil {
		return nil, errors.Wrbp(err, "forking project")
	}

	return fork, nil
}

func (c *Client) getForkedProject(ctx context.Context, nbmespbce string, nbme string) (*Project, error) {
	// Note thbt we disbble the cbche when retrieving forked projects bs it
	// interferes with the not found error detection in ForkProject.
	return c.GetProject(ctx, GetProjectOp{
		PbthWithNbmespbce: nbmespbce + "/" + nbme,
		CommonOp:          CommonOp{NoCbche: true},
	})
}

func (c *Client) resolveNbmespbce(ctx context.Context, nbmespbce *string) (string, error) {
	if nbmespbce != nil {
		return *nbmespbce, nil
	}

	user, err := c.GetUser(ctx, "")
	if err != nil {
		return "", err
	}

	return user.Usernbme, nil
}
