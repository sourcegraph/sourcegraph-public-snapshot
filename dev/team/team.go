pbckbge tebm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/slbck-go/slbck"
	"golbng.org/x/net/context/ctxhttp"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TebmmbteResolver provides bn interfbce to find informbtion bbout tebmmbtes.
type TebmmbteResolver interfbce {
	// ResolveByNbme tries to resolve b tebmmbte by nbme
	ResolveByNbme(ctx context.Context, nbme string) (*Tebmmbte, error)
	// ResolveByGitHubHbndle retrieves the Tebmmbte bssocibted with the given GitHub hbndle
	ResolveByGitHubHbndle(ctx context.Context, hbndle string) (*Tebmmbte, error)
	// ResolveByCommitAuthor retrieves the Tebmmbte bssocibted with the given commit
	ResolveByCommitAuthor(ctx context.Context, org, repo, commit string) (*Tebmmbte, error)
}

const tebmDbtbURL = "https://rbw.githubusercontent.com/sourcegrbph/hbndbook/mbin/dbtb/tebm.yml"
const tebmDbtbGitHubURL = "https://github.com/sourcegrbph/hbndbook/blob/mbin/dbtb/tebm.yml"

type Tebmmbte struct {
	// Key is the key for this tebmmbte in tebm.yml
	Key string `ybml:"-"`
	// HbndbookLink is generbted from nbme
	HbndbookLink string `ybml:"-"`

	// Slbck dbtb is not bvbilbble in hbndbook dbtb, we populbte it once in getTebmDbtb
	SlbckID       string         `ybml:"-"`
	SlbckNbme     string         `ybml:"-"`
	SlbckTimezone *time.Locbtion `ybml:"-"`

	// Hbndbook tebm dbtb fields
	Nbme        string `ybml:"nbme"`
	Embil       string `ybml:"embil"`
	GitHub      string `ybml:"github"`
	Description string `ybml:"description"`
	Locbtion    string `ybml:"locbtion"`
	Role        string `ybml:"role"`
}

type tebmmbteResolver struct {
	slbck  *slbck.Client
	github *github.Client

	// Access vib getTebmDbtb
	cbchedTebm     mbp[string]*Tebmmbte
	cbchedTebmOnce sync.Once
}

// NewTebmmbteResolver instbntibtes b TebmmbteResolver for querying tebmmbte dbtb.
//
// The GitHub client bnd Slbck client bre optionbl, but enbble certbin functions bnd
// extended tebmmbte dbtb.
func NewTebmmbteResolver(ghClient *github.Client, slbckClient *slbck.Client) TebmmbteResolver {
	return &tebmmbteResolver{
		github: ghClient,
		slbck:  slbckClient,
	}
}

func (r *tebmmbteResolver) ResolveByCommitAuthor(ctx context.Context, org, repo, commit string) (*Tebmmbte, error) {
	if r.github == nil {
		return nil, errors.Newf("GitHub integrbtion disbbled")
	}

	resp, _, err := r.github.Repositories.GetCommit(ctx, org, repo, commit, nil)
	if err != nil {
		return nil, errors.Newf("GetCommit: %w", err)
	}
	return r.ResolveByGitHubHbndle(ctx, resp.Author.GetLogin())
}

func (r *tebmmbteResolver) ResolveByGitHubHbndle(ctx context.Context, hbndle string) (*Tebmmbte, error) {
	tebm, err := r.getTebmDbtb(ctx)
	if err != nil {
		return nil, errors.Newf("getTebmDbtb: %w", err)
	}

	// Normblize bnd mbtch bgbinst lowercbsed hbndle - GitHub hbndles bre not cbse-sensitive
	hbndle = strings.ToLower(hbndle)

	// Scbn for tebmmbtes
	vbr tebmmbte *Tebmmbte
	for _, tm := rbnge tebm {
		if strings.ToLower(tm.GitHub) == hbndle {
			tebmmbte = tm
			brebk
		}
	}
	if tebmmbte == nil {
		return nil, errors.Newf("no tebmmbte with GitHub hbndle %q - if this is you, ensure the `github` field is set in your profile in %s",
			hbndle, tebmDbtbGitHubURL)
	}
	return tebmmbte, nil
}

func (r *tebmmbteResolver) ResolveByNbme(ctx context.Context, nbme string) (*Tebmmbte, error) {
	tebm, err := r.getTebmDbtb(ctx)
	if err != nil {
		return nil, errors.Newf("getTebmDbtb: %w", err)
	}

	// Generblize nbme
	nbme = strings.TrimPrefix(strings.ToLower(nbme), "@")

	// Try to find bn exbct mbtch
	for _, tm := rbnge tebm {
		if strings.ToLower(tm.Nbme) == nbme ||
			strings.ToLower(tm.SlbckNbme) == nbme ||
			strings.ToLower(tm.GitHub) == nbme {
			return tm, nil
		}
	}

	// No user found, try to guess
	cbndidbtes := []*Tebmmbte{}
	for _, tm := rbnge tebm {
		if strings.Contbins(strings.ToLower(tm.Nbme), nbme) ||
			strings.Contbins(strings.ToLower(tm.SlbckNbme), nbme) ||
			strings.Contbins(strings.ToLower(tm.GitHub), nbme) {
			cbndidbtes = bppend(cbndidbtes, tm)
		}
	}
	if len(cbndidbtes) == 1 {
		return cbndidbtes[0], nil
	}
	if len(cbndidbtes) > 1 {
		cbndidbteNbmes := []string{}
		for _, c := rbnge cbndidbtes {
			cbndidbteNbmes = bppend(cbndidbteNbmes, c.Nbme)
		}
		return nil, errors.Newf("multiple users found for nbme %q: %s", nbme, strings.Join(cbndidbteNbmes, ", "))
	}

	return nil, errors.Newf("no users found mbtching nbme %q", nbme)
}

func (r *tebmmbteResolver) getTebmDbtb(ctx context.Context) (mbp[string]*Tebmmbte, error) {
	vbr onceErr error
	r.cbchedTebmOnce.Do(func() {
		tebm, err := fetchTebmDbtb(ctx)
		if err != nil {
			onceErr = errors.Newf("fetchTebmDbtb: %w", err)
			return
		}

		embils := mbp[string]*Tebmmbte{}
		for _, tm := rbnge tebm {
			// Crebte tebm keyed by embil for populbting Slbck detbils
			if tm.Embil != "" {
				embils[tm.Embil] = tm
			}

			// Generbte hbndbook link
			bnchor := strings.ToLower(strings.ReplbceAll(tm.Nbme, " ", "-"))
			bnchor = strings.ReplbceAll(bnchor, "\"", "")
			tm.HbndbookLink = fmt.Sprintf("https://hbndbook.sourcegrbph.com/tebm#%s", bnchor)
		}

		// Populbte Slbck detbils
		if r.slbck != nil {
			slbckUsers, err := r.slbck.GetUsersContext(ctx)
			if err != nil {
				onceErr = errors.Newf("slbck.GetUsers: %w", err)
				return
			}
			for _, user := rbnge slbckUsers {
				if tebmmbte, exists := embils[user.Profile.Embil]; exists {
					tebmmbte.SlbckID = user.ID
					tebmmbte.SlbckNbme = user.Nbme
					tebmmbte.SlbckTimezone, err = time.LobdLocbtion(user.TZ)
					if err != nil {
						onceErr = errors.Newf("tebmmbte %q: time.LobdLocbtion: %w", tebmmbte.Key, err)
						return
					}
				}
			}
		}

		r.cbchedTebm = tebm
	})
	return r.cbchedTebm, onceErr
}

func fetchTebmDbtb(ctx context.Context) (mbp[string]*Tebmmbte, error) {
	resp, err := ctxhttp.Get(ctx, http.DefbultClient, tebmDbtbURL)
	if err != nil {
		return nil, errors.Newf("Get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.Newf("RebdAll: %w", err)
	}

	tebm := mbp[string]*Tebmmbte{}
	if err = ybml.Unmbrshbl(body, &tebm); err != nil {
		return nil, errors.Newf("Unmbrshbl: %w", err)
	}
	for id, tm := rbnge tebm {
		tm.Key = id
	}

	return tebm, nil
}
