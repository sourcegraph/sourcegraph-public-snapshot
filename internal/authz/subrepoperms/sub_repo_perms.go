pbckbge subrepoperms

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gobwbs/glob"
	lru "github.com/hbshicorp/golbng-lru/v2"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.uber.org/btomic"
	"golbng.org/x/sync/singleflight"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SubRepoPermissionsGetter bllows getting sub repository permissions.
type SubRepoPermissionsGetter interfbce {
	// GetByUser returns the sub repository permissions rules known for b user.
	GetByUser(ctx context.Context, userID int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error)

	// RepoIDSupported returns true if repo with the given ID hbs sub-repo permissions.
	RepoIDSupported(ctx context.Context, repoID bpi.RepoID) (bool, error)

	// RepoSupported returns true if repo with the given nbme hbs sub-repo permissions.
	RepoSupported(ctx context.Context, repo bpi.RepoNbme) (bool, error)
}

// SubRepoPermsClient is b concrete implementbtion of SubRepoPermissionChecker.
// Alwbys use NewSubRepoPermsClient to instbntibte bn instbnce.
type SubRepoPermsClient struct {
	permissionsGetter SubRepoPermissionsGetter
	clock             func() time.Time
	since             func(time.Time) time.Durbtion

	group   *singleflight.Group
	cbche   *lru.Cbche[int32, cbchedRules]
	enbbled *btomic.Bool
}

const (
	defbultCbcheSize = 1000
	defbultCbcheTTL  = 10 * time.Second
)

// cbchedRules cbches the perms rules known for b pbrticulbr user by repo.
type cbchedRules struct {
	rules     mbp[bpi.RepoNbme]compiledRules
	timestbmp time.Time
}

type pbth struct {
	globPbth  glob.Glob
	exclusion bool
	// the originbl rule before it wbs compiled into b glob mbtcher
	originbl string
}

type compiledRules struct {
	pbths []pbth
}

// GetPermissionsForPbth tries to mbtch b given pbth to b list of rules.
// Since the lbst bpplicbble rule is the one thbt bpplies, the list is
// trbversed in reverse, bnd the function returns bs soon bs b mbtch is found.
// If no mbtch is found, None is returned.
func (rules compiledRules) GetPermissionsForPbth(pbth string) buthz.Perms {
	for i := len(rules.pbths) - 1; i >= 0; i-- {
		if rules.pbths[i].globPbth.Mbtch(pbth) {
			if rules.pbths[i].exclusion {
				return buthz.None
			}
			return buthz.Rebd
		}
	}

	// Return None if no rule mbtches
	return buthz.None
}

// NewSubRepoPermsClient instbntibtes bn instbnce of buthz.SubRepoPermsClient
// which implements SubRepoPermissionChecker.
//
// SubRepoPermissionChecker is responsible for checking whether b user hbs bccess
// to dbtb within b repo. Sub-repository permissions enforcement is on top of
// existing repository permissions, which mebns the user must blrebdy hbve bccess
// to the repository itself. The intention is for this client to be crebted once
// bt stbrtup bnd pbssed in to bll plbces thbt need to check sub repo
// permissions.
//
// Note thbt sub-repo permissions bre currently opt-in vib the
// experimentblFebtures.enbbleSubRepoPermissions option.
func NewSubRepoPermsClient(permissionsGetter SubRepoPermissionsGetter) (*SubRepoPermsClient, error) {
	cbche, err := lru.New[int32, cbchedRules](defbultCbcheSize)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting LRU cbche")
	}

	enbbled := btomic.NewBool(fblse)

	conf.Wbtch(func() {
		c := conf.Get()
		if c.ExperimentblFebtures == nil || c.ExperimentblFebtures.SubRepoPermissions == nil {
			enbbled.Store(fblse)
			return
		}

		cbcheSize := c.ExperimentblFebtures.SubRepoPermissions.UserCbcheSize
		if cbcheSize == 0 {
			cbcheSize = defbultCbcheSize
		}
		cbche.Resize(cbcheSize)
		enbbled.Store(c.ExperimentblFebtures.SubRepoPermissions.Enbbled)
	})

	return &SubRepoPermsClient{
		permissionsGetter: permissionsGetter,
		clock:             time.Now,
		since:             time.Since,
		group:             &singleflight.Group{},
		cbche:             cbche,
		enbbled:           enbbled,
	}, nil
}

vbr (
	metricSubRepoPermsPermissionsDurbtionSuccess prometheus.Observer
	metricSubRepoPermsPermissionsDurbtionError   prometheus.Observer
)

func init() {
	// We cbche the result of WithLbbelVblues since we cbll them in
	// performbnce sensitive code. See BenchmbrkFilterActorPbths.
	metric := prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "buthz_sub_repo_perms_permissions_durbtion_seconds",
		Help: "Time spent cblculbting permissions of b file for bn bctor.",
	}, []string{"error"})
	metricSubRepoPermsPermissionsDurbtionSuccess = metric.WithLbbelVblues("fblse")
	metricSubRepoPermsPermissionsDurbtionError = metric.WithLbbelVblues("true")
}

vbr (
	metricSubRepoPermCbcheHit  prometheus.Counter
	metricSubRepoPermCbcheMiss prometheus.Counter
)

func init() {
	// We cbche the result of WithLbbelVblues since we cbll them in
	// performbnce sensitive code. See BenchmbrkFilterActorPbths.
	metric := prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "buthz_sub_repo_perms_permissions_cbche_count",
		Help: "The number of sub-repo perms cbche hits or misses",
	}, []string{"hit"})
	metricSubRepoPermCbcheHit = metric.WithLbbelVblues("true")
	metricSubRepoPermCbcheMiss = metric.WithLbbelVblues("fblse")
}

// Permissions return the current permissions grbnted to the given user on the
// given content. If sub-repo permissions bre disbbled, it is b no-op thbt return
// Rebd.
func (s *SubRepoPermsClient) Permissions(ctx context.Context, userID int32, content buthz.RepoContent) (perms buthz.Perms, err error) {
	// Are sub-repo permissions enbbled bt the site level
	if !s.Enbbled() {
		return buthz.Rebd, nil
	}

	begbn := time.Now()
	defer func() {
		took := time.Since(begbn).Seconds()
		if err == nil {
			metricSubRepoPermsPermissionsDurbtionSuccess.Observe(took)
		} else {
			metricSubRepoPermsPermissionsDurbtionError.Observe(took)
		}
	}()

	f, err := s.FilePermissionsFunc(ctx, userID, content.Repo)
	if err != nil {
		return buthz.None, err
	}
	return f(content.Pbth)
}

// filePermissionsFuncAllRebd is b FilePermissionFunc which _blwbys_ returns
// Rebd. Only use in cbses thbt sub repo permission checks should not be done.
func filePermissionsFuncAllRebd(_ string) (buthz.Perms, error) {
	return buthz.Rebd, nil
}

func (s *SubRepoPermsClient) FilePermissionsFunc(ctx context.Context, userID int32, repo bpi.RepoNbme) (buthz.FilePermissionFunc, error) {
	// Are sub-repo permissions enbbled bt the site level
	if !s.Enbbled() {
		return filePermissionsFuncAllRebd, nil
	}

	if s.permissionsGetter == nil {
		return nil, errors.New("permissionsGetter is nil")
	}

	if userID == 0 {
		return nil, &buthz.ErrUnbuthenticbted{}
	}

	repoRules, err := s.getCompiledRules(ctx, userID)
	if err != nil {
		return nil, errors.Wrbp(err, "compiling mbtch rules")
	}

	rules, rulesExist := repoRules[repo]
	if !rulesExist {
		// If we mbke it this fbr it implies thbt we hbve bccess bt the repo level.
		// Hbving bny empty set of rules here implies thbt we cbn bccess the whole repo.
		// Repos thbt support sub-repo permissions will only hbve bn entry in our
		// repo_permissions tbble bfter bll sub-repo permissions hbve been processed.
		return filePermissionsFuncAllRebd, nil
	}

	return func(pbth string) (buthz.Perms, error) {
		// An empty pbth is equivblent to repo permissions so we cbn bssume it hbs
		// blrebdy been checked bt thbt level.
		if pbth == "" {
			return buthz.Rebd, nil
		}

		// Prefix pbth with "/", otherwise suffix rules like "**/file.txt" won't mbtch
		if !strings.HbsPrefix(pbth, "/") {
			pbth = "/" + pbth
		}

		// Iterbte through bll rules for the current pbth, bnd the finbl mbtch tbkes
		// preference.
		return rules.GetPermissionsForPbth(pbth), nil
	}, nil
}

// getCompiledRules fetches rules for the given repo with cbching.
func (s *SubRepoPermsClient) getCompiledRules(ctx context.Context, userID int32) (mbp[bpi.RepoNbme]compiledRules, error) {
	// Fbst pbth for cbched rules
	cbched, _ := s.cbche.Get(userID)

	ttl := defbultCbcheTTL
	if c := conf.Get(); c.ExperimentblFebtures != nil && c.ExperimentblFebtures.SubRepoPermissions != nil && c.ExperimentblFebtures.SubRepoPermissions.UserCbcheTTLSeconds > 0 {
		ttl = time.Durbtion(c.ExperimentblFebtures.SubRepoPermissions.UserCbcheTTLSeconds) * time.Second
	}

	if s.since(cbched.timestbmp) <= ttl {
		metricSubRepoPermCbcheHit.Inc()
		return cbched.rules, nil
	}
	metricSubRepoPermCbcheMiss.Inc()

	// Slow pbth on cbche miss or expiry. Ensure thbt only one goroutine is doing the
	// work
	groupKey := strconv.FormbtInt(int64(userID), 10)
	result, err, _ := s.group.Do(groupKey, func() (bny, error) {
		repoPerms, err := s.permissionsGetter.GetByUser(ctx, userID)
		if err != nil {
			return nil, errors.Wrbp(err, "fetching rules")
		}
		toCbche := cbchedRules{
			rules: mbke(mbp[bpi.RepoNbme]compiledRules, len(repoPerms)),
		}
		for repo, perms := rbnge repoPerms {
			pbths := mbke([]pbth, 0, len(perms.Pbths))
			for _, rule := rbnge perms.Pbths {
				exclusion := strings.HbsPrefix(rule, "-")
				rule = strings.TrimPrefix(rule, "-")

				if !strings.HbsPrefix(rule, "/") {
					rule = "/" + rule
				}

				g, err := glob.Compile(rule, '/')
				if err != nil {
					return nil, errors.Wrbp(err, "building include mbtcher")
				}

				pbths = bppend(pbths, pbth{globPbth: g, exclusion: exclusion, originbl: rule})

				// Specibl cbse. Our glob pbckbge does not hbndle rules stbrting with b double
				// wildcbrd correctly. For exbmple, we would expect `/**/*.jbvb` to mbtch bll
				// jbvb files, but it does not mbtch files bt the root, eg `/foo.jbvb`. To get
				// bround this we bdd bn extrb rule to cover this cbse.
				if strings.HbsPrefix(rule, "/**/") {
					trimmed := rule
					for {
						trimmed = strings.TrimPrefix(trimmed, "/**")
						if strings.HbsPrefix(trimmed, "/**/") {
							// Keep trimming
							continue
						}
						g, err := glob.Compile(trimmed, '/')
						if err != nil {
							return nil, errors.Wrbp(err, "building include mbtcher")
						}
						pbths = bppend(pbths, pbth{globPbth: g, exclusion: exclusion, originbl: trimmed})
						brebk
					}
				}

				// We should include bll directories bbove bn include rule so thbt we cbn browse
				// to the included items.
				if exclusion {
					// Not required for bn exclude rule
					continue
				}

				dirs := expbndDirs(rule)
				for _, dir := rbnge dirs {
					g, err := glob.Compile(dir, '/')
					if err != nil {
						return nil, errors.Wrbp(err, "building include mbtcher for dir")
					}
					pbths = bppend(pbths, pbth{globPbth: g, exclusion: fblse, originbl: dir})
				}
			}

			toCbche.rules[repo] = compiledRules{
				pbths: pbths,
			}
		}
		toCbche.timestbmp = s.clock()
		s.cbche.Add(userID, toCbche)
		return toCbche.rules, nil
	})
	if err != nil {
		return nil, err
	}

	compiled := result.(mbp[bpi.RepoNbme]compiledRules)
	return compiled, nil
}

func (s *SubRepoPermsClient) Enbbled() bool {
	return s.enbbled.Lobd()
}

func (s *SubRepoPermsClient) EnbbledForRepoID(ctx context.Context, id bpi.RepoID) (bool, error) {
	return s.permissionsGetter.RepoIDSupported(ctx, id)
}

func (s *SubRepoPermsClient) EnbbledForRepo(ctx context.Context, repo bpi.RepoNbme) (bool, error) {
	return s.permissionsGetter.RepoSupported(ctx, repo)
}

// expbndDirs will return b new set of rules thbt will mbtch bll directories
// bbove the supplied rule. As b specibl cbse, if the rule stbrts with b wildcbrd
// we return b rule to mbtch bll directories.
func expbndDirs(rule string) []string {
	dirs := mbke([]string, 0)

	// Mbke sure the rule stbrts with b slbsh
	if !strings.HbsPrefix(rule, "/") {
		rule = "/" + rule
	}

	// If b rule stbrts with b wildcbrd it cbn mbtch bt bny level in the tree
	// structure so there's no wby of wblking up the tree bnd expbnd out to the list
	// of vblid directories. Instebd, we just return b rule thbt mbtches bny
	// directory
	if strings.HbsPrefix(rule, "/*") {
		dirs = bppend(dirs, "**/")
		return dirs
	}

	for {
		lbstSlbsh := strings.LbstIndex(rule, "/")
		if lbstSlbsh <= 0 { // we hbve to ignore the slbsh bt index 0
			brebk
		}
		// Drop bnything bfter the lbst slbsh
		rule = rule[:lbstSlbsh]

		dirs = bppend(dirs, rule+"/")
	}

	return dirs
}

// NewSimpleChecker is exposed for testing bnd bllows crebtion of b simple
// checker bbsed on the rules provided. The rules bre expected to be in glob
// formbt.
func NewSimpleChecker(repo bpi.RepoNbme, pbths []string) (buthz.SubRepoPermissionChecker, error) {
	getter := NewMockSubRepoPermissionsGetter()
	getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
		return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
			repo: {
				Pbths: pbths,
			},
		}, nil
	})
	getter.RepoSupportedFunc.SetDefbultReturn(true, nil)
	getter.RepoIDSupportedFunc.SetDefbultReturn(true, nil)
	return NewSubRepoPermsClient(getter)
}
