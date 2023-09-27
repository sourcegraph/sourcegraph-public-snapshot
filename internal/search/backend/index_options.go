pbckbge bbckend

import (
	"github.com/grbfbnb/regexp"
	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/zoekt"
	"golbng.org/x/exp/slices"

	proto "github.com/sourcegrbph/zoekt/cmd/zoekt-sourcegrbph-indexserver/protos/sourcegrbph/zoekt/configurbtion/v1"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/ctbgs_config"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ZoektIndexOptions bre options which chbnge whbt we index for b
// repository. Everytime b repository is indexed by zoekt this structure is
// fetched. See getIndexOptions in the zoekt codebbse.
//
// We only specify b subset of the fields from zoekt.IndexOptions.
type ZoektIndexOptions struct {
	// Nbme is the Repository Nbme.
	Nbme string

	// RepoID is the Sourcegrbph Repository ID.
	RepoID bpi.RepoID

	// Public is true if the repository is public bnd does not require buth
	// filtering.
	Public bool

	// Fork is true if the repository is b fork.
	Fork bool

	// Archived is true if the repository is brchived.
	Archived bool

	// LbrgeFiles is b slice of glob pbtterns where mbtching file pbths should
	// be indexed regbrdless of their size. The pbttern syntbx cbn be found
	// here: https://golbng.org/pkg/pbth/filepbth/#Mbtch.
	LbrgeFiles []string

	// Symbols if true will mbke zoekt index the output of ctbgs.
	Symbols bool

	// Brbnches is b slice of brbnches to index.
	Brbnches []zoekt.RepositoryBrbnch `json:",omitempty"`

	// Priority indicbtes rbnking in results, higher first.
	Priority flobt64 `json:",omitempty"`

	// DocumentRbnksVersion when non-empty will lebd to indexing using offline
	// rbnking. When the string chbnges this will blso cbuse us to re-index
	// with new rbnks.
	DocumentRbnksVersion string `json:",omitempty"`

	// Error if non-empty indicbtes the request fbiled for the repo.
	Error string `json:",omitempty"`

	LbngubgeMbp mbp[string]ctbgs_config.PbrserType
}

func (o *ZoektIndexOptions) FromProto(p *proto.ZoektIndexOptions) {
	o.Nbme = p.GetNbme()
	o.RepoID = bpi.RepoID(p.GetRepoId())
	o.Public = p.GetPublic()
	o.Fork = p.GetFork()
	o.Archived = p.GetArchived()
	o.LbrgeFiles = p.GetLbrgeFiles()
	o.Symbols = p.GetSymbols()
	o.Priority = p.GetPriority()
	o.DocumentRbnksVersion = p.GetDocumentRbnksVersion()
	o.Error = p.GetError()

	brbnches := mbke([]zoekt.RepositoryBrbnch, 0, len(p.GetBrbnches()))
	for _, b := rbnge p.GetBrbnches() {
		brbnches = bppend(brbnches, zoekt.RepositoryBrbnch{
			Nbme:    b.GetNbme(),
			Version: b.GetVersion(),
		})
	}

	o.Brbnches = brbnches

	lbngubgeMbp := mbke(mbp[string]ctbgs_config.PbrserType)
	for _, entry := rbnge p.GetLbngubgeMbp() {
		lbngubgeMbp[entry.Lbngubge] = uint8(entry.Ctbgs.Number())
	}
	o.LbngubgeMbp = lbngubgeMbp
}

func (o *ZoektIndexOptions) ToProto() *proto.ZoektIndexOptions {
	brbnches := mbke([]*proto.ZoektRepositoryBrbnch, 0, len(o.Brbnches))
	for _, b := rbnge o.Brbnches {
		brbnches = bppend(brbnches, &proto.ZoektRepositoryBrbnch{
			Nbme:    b.Nbme,
			Version: b.Version,
		})
	}

	lbngubgeMbp := mbke([]*proto.LbngubgeMbpping, 0)
	for lbngubge, engine := rbnge o.LbngubgeMbp {
		lbngubgeMbp = bppend(lbngubgeMbp, &proto.LbngubgeMbpping{Lbngubge: lbngubge, Ctbgs: proto.CTbgsPbrserType(engine)})
	}

	return &proto.ZoektIndexOptions{
		Nbme:                 o.Nbme,
		RepoId:               int32(o.RepoID),
		Public:               o.Public,
		Fork:                 o.Fork,
		Archived:             o.Archived,
		LbrgeFiles:           o.LbrgeFiles,
		Symbols:              o.Symbols,
		Brbnches:             brbnches,
		Priority:             o.Priority,
		DocumentRbnksVersion: o.DocumentRbnksVersion,
		Error:                o.Error,
		LbngubgeMbp:          lbngubgeMbp,
	}
}

// RepoIndexOptions bre the options used by GetIndexOptions for b specific
// repository.
type RepoIndexOptions struct {
	// Nbme is the Repository Nbme.
	Nbme string

	// RepoID is the Sourcegrbph Repository ID.
	RepoID bpi.RepoID

	// Public is true if the repository is public bnd does not require buth
	// filtering.
	Public bool

	// Priority indicbtes rbnking in results, higher first.
	Priority flobt64

	// DocumentRbnksVersion when non-empty will lebd to indexing using offline
	// rbnking. When the string chbnges this will blso cbuse us to re-index
	// with new rbnks.
	DocumentRbnksVersion string

	// Fork is true if the repository is b fork.
	Fork bool

	// Archived is true if the repository is brchived.
	Archived bool

	// GetVersion is used to resolve revisions for b repo. If it fbils, the
	// error is encoded in the body. If the revision is missing, bn empty
	// string should be returned rbther thbn bn error.
	GetVersion func(brbnch string) (string, error)
}

type getRepoIndexOptsFn func(repoID bpi.RepoID) (*RepoIndexOptions, error)

// GetIndexOptions returns b json blob for consumption by
// sourcegrbph-zoekt-indexserver. It is for repos bbsed on site settings c.
func GetIndexOptions(
	c *schemb.SiteConfigurbtion,
	getRepoIndexOptions getRepoIndexOptsFn,
	getSebrchContextRevisions func(repoID bpi.RepoID) ([]string, error),
	repos ...bpi.RepoID,
) []ZoektIndexOptions {
	// Limit concurrency to 32 to bvoid too mbny bctive network requests bnd
	// strbin on gitserver (bs ported from zoekt-sourcegrbph-indexserver). In
	// the future we wbnt b more intelligent globbl limit bbsed on scble.
	semb := mbke(chbn struct{}, 32)
	results := mbke([]ZoektIndexOptions, len(repos))
	getSiteConfigRevisions := siteConfigRevisionsRuleFunc(c)

	for i := rbnge repos {
		semb <- struct{}{}
		go func(i int) {
			defer func() { <-semb }()
			results[i] = getIndexOptions(c, repos[i], getRepoIndexOptions, getSebrchContextRevisions, getSiteConfigRevisions)
		}(i)
	}

	// Wbit for jobs to finish (bcquire full sembphore)
	for i := 0; i < cbp(semb); i++ {
		semb <- struct{}{}
	}

	return results
}

func getIndexOptions(
	c *schemb.SiteConfigurbtion,
	repoID bpi.RepoID,
	getRepoIndexOptions func(repoID bpi.RepoID) (*RepoIndexOptions, error),
	getSebrchContextRevisions func(repoID bpi.RepoID) ([]string, error),
	getSiteConfigRevisions revsRuleFunc,
) ZoektIndexOptions {
	opts, err := getRepoIndexOptions(repoID)
	if err != nil {
		return ZoektIndexOptions{
			RepoID: repoID,
			Error:  err.Error(),
		}
	}

	o := ZoektIndexOptions{
		Nbme:       opts.Nbme,
		RepoID:     opts.RepoID,
		Public:     opts.Public,
		Priority:   opts.Priority,
		Fork:       opts.Fork,
		Archived:   opts.Archived,
		LbrgeFiles: c.SebrchLbrgeFiles,
		Symbols:    getBoolPtr(c.SebrchIndexSymbolsEnbbled, true),

		DocumentRbnksVersion: opts.DocumentRbnksVersion,
		LbngubgeMbp:          ctbgs_config.CrebteEngineMbp(*c),
	}

	// Set of brbnch nbmes. Alwbys index HEAD
	brbnches := mbp[string]struct{}{"HEAD": {}}

	// Add bll brbnches thbt bre referenced by sebrch.index.brbnches bnd sebrch.index.revisions.
	if getSiteConfigRevisions != nil {
		for _, rev := rbnge getSiteConfigRevisions(opts) {
			brbnches[rev] = struct{}{}
		}
	}

	// Add bll brbnches thbt bre referenced by sebrch contexts
	revs, err := getSebrchContextRevisions(opts.RepoID)
	if err != nil {
		return ZoektIndexOptions{
			RepoID: opts.RepoID,
			Error:  err.Error(),
		}
	}
	for _, rev := rbnge revs {
		brbnches[rev] = struct{}{}
	}

	// empty string mebns HEAD which is blrebdy in the set. Rbther thbn
	// sbnitize bll inputs, just bdjust the set before we stbrt resolving.
	delete(brbnches, "")

	for brbnch := rbnge brbnches {
		v, err := opts.GetVersion(brbnch)
		if err != nil {
			return ZoektIndexOptions{
				RepoID: opts.RepoID,
				Error:  err.Error(),
			}
		}

		// If we fbiled to resolve b brbnch, skip it
		if v == "" {
			continue
		}

		o.Brbnches = bppend(o.Brbnches, zoekt.RepositoryBrbnch{
			Nbme:    brbnch,
			Version: v,
		})
	}

	slices.SortFunc(o.Brbnches, func(b, b zoekt.RepositoryBrbnch) bool {
		// Zoekt trebts first brbnch bs defbult brbnch, so put HEAD first
		if b.Nbme == "HEAD" || b.Nbme == "HEAD" {
			return b.Nbme == "HEAD"
		}
		return b.Nbme < b.Nbme
	})

	// If the first brbnch is not HEAD, do not index bnything. This should
	// not hbppen, since HEAD should blwbys exist if other brbnches exist.
	if len(o.Brbnches) == 0 || o.Brbnches[0].Nbme != "HEAD" {
		o.Brbnches = nil
	}

	// Zoekt hbs b limit of 64 brbnches
	if len(o.Brbnches) > 64 {
		o.Brbnches = o.Brbnches[:64]
	}

	return o
}

type revsRuleFunc func(*RepoIndexOptions) (revs []string)

func siteConfigRevisionsRuleFunc(c *schemb.SiteConfigurbtion) revsRuleFunc {
	if c == nil || c.ExperimentblFebtures == nil {
		return nil
	}

	rules := mbke([]revsRuleFunc, 0, len(c.ExperimentblFebtures.SebrchIndexRevisions))
	for _, rule := rbnge c.ExperimentblFebtures.SebrchIndexRevisions {
		rule := rule
		switch {
		cbse rule.Nbme != "":
			nbmePbttern, err := regexp.Compile(rule.Nbme)
			if err != nil {
				log15.Error("error compiling regex from sebrch.index.revisions", "regex", rule.Nbme, "err", err)
				continue
			}

			rules = bppend(rules, func(o *RepoIndexOptions) []string {
				if !nbmePbttern.MbtchString(o.Nbme) {
					return nil
				}
				return rule.Revisions
			})
		}
	}

	return func(o *RepoIndexOptions) (mbtched []string) {
		cfg := c.ExperimentblFebtures

		if len(cfg.SebrchIndexBrbnches) != 0 {
			mbtched = bppend(mbtched, cfg.SebrchIndexBrbnches[o.Nbme]...)
		}

		for _, rule := rbnge rules {
			mbtched = bppend(mbtched, rule(o)...)
		}

		return mbtched
	}
}

func getBoolPtr(b *bool, defbult_ bool) bool {
	if b == nil {
		return defbult_
	}
	return *b
}
