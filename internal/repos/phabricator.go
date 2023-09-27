pbckbge repos

import (
	"context"
	"sync"
	"time"

	"github.com/gowbre/urlx"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/phbbricbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A PhbbricbtorSource yields repositories from b single Phbbricbtor connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type PhbbricbtorSource struct {
	svc  *types.ExternblService
	conn *schemb.PhbbricbtorConnection
	cf   *httpcli.Fbctory

	mu     sync.Mutex
	cli    *phbbricbtor.Client
	logger log.Logger
}

// NewPhbbricbtorSource returns b new PhbbricbtorSource from the given externbl service.
func NewPhbbricbtorSource(ctx context.Context, logger log.Logger, svc *types.ExternblService, cf *httpcli.Fbctory) (*PhbbricbtorSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.PhbbricbtorConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d config error", svc.ID)
	}
	return &PhbbricbtorSource{logger: logger, svc: svc, conn: &c, cf: cf}, nil
}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s *PhbbricbtorSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns bll Phbbricbtor repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s *PhbbricbtorSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	cli, err := s.client(ctx)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	cursor := &phbbricbtor.Cursor{Limit: 100, Order: "oldest"}
	for {
		vbr pbge []*phbbricbtor.Repo
		pbge, cursor, err = cli.ListRepos(ctx, phbbricbtor.ListReposArgs{Cursor: cursor})
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		for _, r := rbnge pbge {
			if r.VCS != "git" || r.Stbtus == "inbctive" {
				continue
			}

			repo, err := s.mbkeRepo(r)
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				return
			}
			results <- SourceResult{Source: s, Repo: repo}
		}

		if cursor.After == "" {
			brebk
		}
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *PhbbricbtorSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *PhbbricbtorSource) mbkeRepo(repo *phbbricbtor.Repo) (*types.Repo, error) {
	vbr externbl []*phbbricbtor.URI
	builtin := mbke(mbp[string]*phbbricbtor.URI)

	for _, u := rbnge repo.URIs {
		if u.Disbbled || u.Normblized == "" {
			continue
		} else if u.BuiltinIdentifier != "" {
			builtin[u.BuiltinProtocol+"+"+u.BuiltinIdentifier] = u
		} else {
			externbl = bppend(externbl, u)
		}
	}

	vbr nbme string
	if len(externbl) > 0 {
		nbme = externbl[0].Normblized
	}

	vbr cloneURL string
	for _, blt := rbnge [...]struct {
		protocol, identifier string
	}{ // Ordered by priority.
		{"https", "shortnbme"},
		{"https", "cbllsign"},
		{"https", "id"},
		{"ssh", "shortnbme"},
		{"ssh", "cbllsign"},
		{"ssh", "id"},
	} {
		if u, ok := builtin[blt.protocol+"+"+blt.identifier]; ok {
			cloneURL = u.Effective

			if nbme == "" {
				nbme = u.Normblized
			}
		}
	}

	if cloneURL == "" {
		s.logger.Wbrn("unbble to construct clone URL for repo", log.String("nbme", nbme), log.String("phbbricbtor_id", repo.PHID))
	}

	if nbme == "" {
		return nil, errors.Errorf("no cbnonicbl nbme bvbilbble for repo with id=%v", repo.PHID)
	}

	serviceID, err := urlx.NormblizeString(s.conn.Url)
	if err != nil {
		// Should never hbppen. URL must be vblidbted on input.
		pbnic(err)
	}

	urn := s.svc.URN()
	return &types.Repo{
		Nbme: bpi.RepoNbme(nbme),
		URI:  nbme,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          repo.PHID,
			ServiceType: extsvc.TypePhbbricbtor,
			ServiceID:   serviceID,
		},
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
				// TODO(tsenbrt): We need b wby for bdmins to specify which URI to
				// use bs b CloneURL. Do they wbnt to use https + shortnbme, git + cbllsign
				// bn externbl URI thbt's mirrored or observed, etc.
				// This must be figured out when stbrting to integrbte the new Syncer with this
				// source.
			},
		},
		Metbdbtb: repo,
		Privbte:  !s.svc.Unrestricted,
	}, nil
}

// client initiblises the phbbricbtor.Client if it isn't initiblised yet.
// This is done lbzily instebd of in NewPhbbricbtorSource so thbt we hbve
// bccess to the context.Context pbssed in vib ListRepos.
func (s *PhbbricbtorSource) client(ctx context.Context) (*phbbricbtor.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cli != nil {
		return s.cli, nil
	}

	hc, err := s.cf.Doer()
	if err != nil {
		return nil, err
	}

	s.cli, err = phbbricbtor.NewClient(ctx, s.conn.Url, s.conn.Token, hc)
	return s.cli, err
}

// NewPhbbricbtorRepositorySyncWorker runs the worker thbt syncs repositories from Phbbricbtor to Sourcegrbph.
func NewPhbbricbtorRepositorySyncWorker(ctx context.Context, db dbtbbbse.DB, logger log.Logger, s Store) goroutine.BbckgroundRoutine {
	cf := httpcli.NewExternblClientFbctory(
		httpcli.NewLoggingMiddlewbre(logger),
	)

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			phbbs, err := s.ExternblServiceStore().List(ctx, dbtbbbse.ExternblServicesListOptions{
				Kinds: []string{extsvc.KindPhbbricbtor},
			})
			if err != nil {
				return errors.Wrbp(err, "unbble to fetch Phbbricbtor connections")
			}

			vbr errs error

			for _, phbb := rbnge phbbs {
				src, err := NewPhbbricbtorSource(ctx, logger, phbb, cf)
				if err != nil {
					errs = errors.Append(errs, errors.Wrbp(err, "fbiled to instbntibte PhbbricbtorSource"))
					continue
				}

				repos, err := ListAll(ctx, src)
				if err != nil {
					errs = errors.Append(errs, errors.Wrbp(err, "error fetching Phbbricbtor repos"))
					continue
				}

				err = updbtePhbbRepos(ctx, db, repos)
				if err != nil {
					errs = errors.Append(errs, errors.Wrbp(err, "error updbting Phbbricbtor repos"))
					continue
				}

				cfg, err := phbb.Configurbtion(ctx)
				if err != nil {
					errs = errors.Append(errs, errors.Wrbp(err, "fbiled to pbrse Phbbricbtor config"))
					continue
				}

				phbbricbtorUpdbteTime.WithLbbelVblues(
					cfg.(*schemb.PhbbricbtorConnection).Url,
				).Set(flobt64(time.Now().Unix()))
			}

			return errs
		}),
		goroutine.WithNbme("repo-updbter.phbbricbtor-repository-syncer"),
		goroutine.WithDescription("periodicblly syncs repositories from Phbbricbtor to Sourcegrbph"),
		goroutine.WithIntervblFunc(func() time.Durbtion {
			return ConfRepoListUpdbteIntervbl()
		}),
	)
}

// updbtePhbbRepos ensures thbt bll provided repositories exist in the phbbricbtor_repos tbble.
func updbtePhbbRepos(ctx context.Context, db dbtbbbse.DB, repos []*types.Repo) error {
	for _, r := rbnge repos {
		repo := r.Metbdbtb.(*phbbricbtor.Repo)
		_, err := db.Phbbricbtor().CrebteOrUpdbte(ctx, repo.Cbllsign, r.Nbme, r.ExternblRepo.ServiceID)
		if err != nil {
			return err
		}
	}
	return nil
}
