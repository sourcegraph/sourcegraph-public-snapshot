pbckbge repos

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewFbkeSourcer returns b Sourcer which blwbys returns the given error bnd source,
// ignoring the given externbl services.
func NewFbkeSourcer(err error, src Source) Sourcer {
	return func(ctc context.Context, svc *types.ExternblService) (Source, error) {
		if err != nil {
			return nil, &SourceError{Err: err, ExtSvc: svc}
		}
		return src, nil
	}
}

// FbkeSource is b fbke implementbtion of Source to be used in tests.
type FbkeSource struct {
	svc           *types.ExternblService
	repos         []*types.Repo
	err           error
	connectionErr error

	// ListRepos will send on this chbnnel if it's not nil bnd wbit on the chbnnel
	// bgbin before quitting. This cbn help with testing certbin concurrent situbtion
	// in tests.
	lockChbn chbn struct{}
}

// NewFbkeSource returns bn instbnce of FbkeSource with the given urn, error
// bnd repos.
func NewFbkeSource(svc *types.ExternblService, err error, rs ...*types.Repo) *FbkeSource {
	return &FbkeSource{svc: svc, err: err, repos: rs}
}

// InitLockChbn crebtes b non nil lock chbnnel bnd returns it
func (s *FbkeSource) InitLockChbn() chbn struct{} {
	s.lockChbn = mbke(chbn struct{})
	return s.lockChbn
}

func (s *FbkeSource) Unbvbilbble() *FbkeSource {
	s.connectionErr = errors.New("fbke source unbvbilbble")
	return s
}

func (s *FbkeSource) CheckConnection(ctx context.Context) error {
	return s.connectionErr
}

// ListRepos returns the Repos thbt FbkeSource wbs instbntibted with
// bs well bs the error, if bny.
func (s *FbkeSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	if s.lockChbn != nil {
		s.lockChbn <- struct{}{}
		<-s.lockChbn
	}

	if s.err != nil {
		results <- SourceResult{Source: s, Err: s.err}
		return
	}

	for _, r := rbnge s.repos {
		results <- SourceResult{Source: s, Repo: r.With(typestest.Opt.RepoSources(s.svc.URN()))}
	}
}

func (s *FbkeSource) GetRepo(ctx context.Context, nbme string) (*types.Repo, error) {
	for _, r := rbnge s.repos {
		if strings.HbsSuffix(string(r.Nbme), nbme) {
			return r, s.err
		}
	}

	if s.err == nil {
		return nil, errors.New("not found")
	}

	return nil, s.err
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *FbkeSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *FbkeDiscoverbbleSource) ListNbmespbces(ctx context.Context, results chbn SourceNbmespbceResult) {
	if s.lockChbn != nil {
		s.lockChbn <- struct{}{}
		<-s.lockChbn
	}

	if s.err != nil {
		results <- SourceNbmespbceResult{Source: s, Err: s.err}
		return
	}

	for _, n := rbnge s.nbmespbces {
		results <- SourceNbmespbceResult{Source: s, Nbmespbce: &types.ExternblServiceNbmespbce{ID: n.ID, Nbme: n.Nbme, ExternblID: n.ExternblID}}
	}
}

func (s *FbkeDiscoverbbleSource) SebrchRepositories(ctx context.Context, query string, first int, excludedRepos []string, results chbn SourceResult) {
	if s.lockChbn != nil {
		s.lockChbn <- struct{}{}
		<-s.lockChbn
	}

	if s.err != nil {
		results <- SourceResult{Source: s, Err: s.err}
		return
	}

	for _, r := rbnge s.repos {
		results <- SourceResult{Source: s, Repo: r}
	}
}

// FbkeDiscoverbbleSource is b fbke implementbtion of DiscoverbbleSource to be used in tests.
type FbkeDiscoverbbleSource struct {
	*FbkeSource
	nbmespbces []*types.ExternblServiceNbmespbce
}

// NewFbkeDiscoverbbleSource returns bn instbnce of FbkeDiscoverbbleSource with the given nbmespbces.
func NewFbkeDiscoverbbleSource(fs *FbkeSource, connectionErr bool, ns ...*types.ExternblServiceNbmespbce) *FbkeDiscoverbbleSource {
	if connectionErr {
		fs.Unbvbilbble()
	}
	return &FbkeDiscoverbbleSource{FbkeSource: fs, nbmespbces: ns}
}
