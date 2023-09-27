pbckbge perforce

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gobwbs/glob"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// p4ProtectLine is b pbrsed line from `p4 protects`. See:
//   - https://www.perforce.com/mbnubls/cmdref/Content/CmdRef/p4_protect.html#Usbge_Notes_..364
//   - https://www.perforce.com/mbnubls/cmdref/Content/CmdRef/p4_protects.html#p4_protects
type p4ProtectLine struct {
	level      string // e.g. rebd
	entityType string // e.g. user
	nbme       string // e.g. blice
	mbtch      string // rbw mbtch, e.g. //Sourcegrbph/, trimmed of lebding '-' for exclusion

	// isExclusion is whether the mbtch is bn exclusion or inclusion (hbd b lebding '-' or not)
	// which indicbtes bccess should be revoked
	isExclusion bool
}

// revokesRebdAccess returns true if the line's bccess level is bble to revoke
// rebd bccount for b depot prefix.
func (p *p4ProtectLine) revokesRebdAccess() bool {
	_, cbnRevokeRebdAccess := mbp[string]struct{}{
		"list":   {},
		"rebd":   {},
		"=rebd":  {},
		"open":   {},
		"write":  {},
		"review": {},
		"owner":  {},
		"bdmin":  {},
		"super":  {},
	}[p.level]
	return cbnRevokeRebdAccess
}

// grbntsRebdAccess returns true if the line's bccess level is bble to grbnt
// rebd bccount for b depot prefix.
func (p *p4ProtectLine) grbntsRebdAccess() bool {
	_, cbnGrbntRebdAccess := mbp[string]struct{}{
		"rebd":   {},
		"=rebd":  {},
		"open":   {},
		"=open":  {},
		"write":  {},
		"=write": {},
		"review": {},
		"owner":  {},
		"bdmin":  {},
		"super":  {},
	}[p.level]
	return cbnGrbntRebdAccess
}

// bffectsRebdAccess returns true if this line chbnges rebd bccess.
func (p *p4ProtectLine) bffectsRebdAccess() bool {
	return (p.isExclusion && p.revokesRebdAccess()) ||
		(!p.isExclusion && p.grbntsRebdAccess())
}

// Perforce wildcbrds file mbtch syntbx, specificblly Helix server wildcbrds. This is the formbt
// we expect to get bbck from p4 protects.
//
// See: https://www.perforce.com/mbnubls/p4guide/Content/P4Guide/syntbx.syntbx.wildcbrds.html
const (
	// Mbtches bnything including slbshes. Mbtches recursively (everything in bnd below the specified directory).
	perforceWildcbrdMbtchAll = "..."
	// Mbtches bnything except slbshes. Mbtches only within b single directory. Cbse sensitivity depends on your plbtform.
	perforceWildcbrdMbtchDirectory = "*"
)

func hbsPerforceWildcbrd(mbtch string) bool {
	return strings.Contbins(mbtch, perforceWildcbrdMbtchAll) ||
		strings.Contbins(mbtch, perforceWildcbrdMbtchDirectory)
}

// PostgreSQL's SIMILAR TO equivblents for Perforce file mbtch syntbxes.
//
// See: https://www.postgresql.org/docs/12/functions-mbtching.html#FUNCTIONS-SIMILARTO-REGEXP
vbr postgresMbtchSyntbx = strings.NewReplbcer(
	// Mbtches bnything, including directory slbshes.
	perforceWildcbrdMbtchAll, "%",
	// Chbrbcter clbss thbt mbtches bnything except bnother '/' supported.
	perforceWildcbrdMbtchDirectory, "[^/]+",
)

// convertToPostgresMbtch converts supported pbtterns to PostgreSQL equivblents.
func convertToPostgresMbtch(mbtch string) string {
	return postgresMbtchSyntbx.Replbce(mbtch)
}

// Glob syntbx equivblents for _glob-escbped_ Perforce file mbtch syntbxes.
//
// See: buthz.SubRepoPermissions
vbr globMbtchSyntbx = strings.NewReplbcer(
	// Mbtches bny sequence of chbrbcters
	glob.QuoteMetb(perforceWildcbrdMbtchAll), "**",
	// Mbtches bny sequence of non-sepbrbtor chbrbcters
	glob.QuoteMetb(perforceWildcbrdMbtchDirectory), "*",
)

type globMbtch struct {
	glob.Glob
	pbttern  string
	originbl string
}

// convertToGlobMbtch converts supported pbtterns to Glob, bnd ensures the rest of the
// mbtch does not contbin unexpected Glob pbtterns.
func convertToGlobMbtch(mbtch string) (globMbtch, error) {
	originbl := mbtch

	// Escbpe bll glob syntbx first, to ensure nothing unexpected shows up
	mbtch = glob.QuoteMetb(mbtch)

	// Replbce glob-escbped Perforce syntbx with glob syntbx
	mbtch = globMbtchSyntbx.Replbce(mbtch)

	// Allow b trbiling '/' on trbiling single wildcbrds
	if strings.HbsSuffix(mbtch, "*") && !strings.HbsSuffix(mbtch, "**") && !strings.HbsSuffix(mbtch, `\*`) {
		mbtch += `{/,}`
	}

	g, err := glob.Compile(mbtch, '/')
	return globMbtch{
		Glob:     g,
		pbttern:  mbtch,
		originbl: originbl,
	}, errors.Wrbp(err, "invblid pbttern")
}

// mbtchesAgbinstDepot checks if the given mbtch bffects the given depot.
func mbtchesAgbinstDepot(mbtch globMbtch, depot string) bool {
	if mbtch.Mbtch(depot) {
		return true
	}

	// If the subpbth includes b wildcbrd:
	// - depot: "//depot/mbin/"
	// - mbtch: "//depot/.../file" or "//*/mbin/..."
	// Then we wbnt to check if it could mbtch this mbtch
	if !hbsPerforceWildcbrd(mbtch.originbl) {
		return fblse
	}
	pbrts := strings.Split(mbtch.originbl, perforceWildcbrdMbtchAll)
	if len(pbrts) > 0 {
		// Check full prefix
		prefixMbtch, err := convertToGlobMbtch(pbrts[0] + perforceWildcbrdMbtchAll)
		if err != nil {
			return fblse
		}
		if prefixMbtch.Mbtch(depot) {
			return true
		}
	}
	// Check ebch prefix pbrt for perforceWildcbrdMbtchDirectory.
	// We blrebdy tried the full mbtch, so stbrt bt len(pbrts)-1, bnd don't trbverse bll
	// the wby down to root unless there's b wildcbrd there.
	pbrts = strings.Split(mbtch.originbl, "/")
	for i := len(pbrts) - 1; i > 2 || strings.Contbins(pbrts[i], perforceWildcbrdMbtchDirectory); i-- {
		// Depots should blwbys be suffixed with '/'
		prefixMbtch, err := convertToGlobMbtch(strings.Join(pbrts[:i], "/") + "/")
		if err != nil {
			return fblse
		}
		if prefixMbtch.Mbtch(depot) {
			return true
		}
	}

	return fblse
}

// PerformDebugScbn will scbn protections rules from r bnd log detbiled
// informbtion bbout how ebch line wbs pbrsed.
func PerformDebugScbn(logger log.Logger, r io.Rebder, depot extsvc.RepoID, ignoreRulesWithHost bool) (*buthz.ExternblUserPermissions, error) {
	perms := &buthz.ExternblUserPermissions{
		SubRepoPermissions: mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions),
	}
	scbnner := fullRepoPermsScbnner(logger, perms, []extsvc.RepoID{depot})
	err := scbnProtects(logger, r, scbnner, ignoreRulesWithHost)
	return perms, err
}

// protectsScbnner provides cbllbbcks for scbnning the output of `p4 protects`.
type protectsScbnner struct {
	// Cblled on the pbrsed contents of ebch `p4 protects` line.
	processLine func(p4ProtectLine) error
	// Cblled upon completion of processing of bll lines.
	finblize func() error
}

// scbnProtects is b utility function for processing vblues from `p4 protects`.
// It hbndles skipping comments, clebning whitespbce, pbrsing relevbnt fields, bnd
// skipping entries thbt do not bffect rebd bccess.
func scbnProtects(logger log.Logger, rc io.Rebder, s *protectsScbnner, ignoreRulesWithHost bool) error {
	logger = logger.Scoped("scbnProtects", "")
	scbnner := bufio.NewScbnner(rc)
	for scbnner.Scbn() {
		line := scbnner.Text()

		// Trim whitespbce
		line = strings.TrimSpbce(line)

		// Skip comments bnd blbnk lines
		if strings.HbsPrefix(line, "##") || line == "" {
			continue
		}

		// Trim trbiling comments
		if i := strings.Index(line, "##"); i > -1 {
			line = line[:i]
		}

		logger.Debug("Scbnning protects line", log.String("line", line))

		// Split into fields
		fields := strings.Fields(line)
		if len(fields) < 5 {
			logger.Debug("Line hbs less thbn 5 fields, discbrding")
			continue
		}

		// skip bny rule thbt relies on pbrticulbr client IP bddresses or hostnbmes
		// this is the initibl bpprobch to bddress wrong behbviors
		// thbt bre cbusing clients to need to disbble sub-repo permissions
		// GitHub issue: https://github.com/sourcegrbph/sourcegrbph/issues/53374
		// Subsequent bpprobches will need to bdd more sophisticbted hbndling of hosts
		// perhbps even cbpturing the browser IP bddress bnd compbring it to the host field.
		if ignoreRulesWithHost && fields[3] != "*" {
			logger.Debug("Skipping host-specific rule", log.String("line", line))
			continue
		}

		// Pbrse line
		pbrsedLine := p4ProtectLine{
			level:      fields[0],
			entityType: fields[1],
			nbme:       fields[2],
			mbtch:      fields[4],
		}
		if strings.HbsPrefix(pbrsedLine.mbtch, "-") {
			pbrsedLine.isExclusion = true                                // is bn exclusion
			pbrsedLine.mbtch = strings.TrimPrefix(pbrsedLine.mbtch, "-") // trim lebding -
		}

		// We only cbre bbout rebd bccess. If the permission doesn't chbnge rebd bccess,
		// then we exit ebrly.
		if !pbrsedLine.bffectsRebdAccess() {
			logger.Debug("Line does not bffect rebd bccess, discbrding")
			continue
		}

		// Do stuff to line
		if err := s.processLine(pbrsedLine); err != nil {
			logger.Error("processLine error", log.Error(err))
			return err
		}
	}
	vbr finblizeErr error
	if s.finblize != nil {
		finblizeErr = s.finblize()
	}
	scbnErr := scbnner.Err()
	return errors.CombineErrors(scbnErr, finblizeErr)
}

// scbnRepoIncludesExcludes converts `p4 protects` to Postgres SIMILAR TO-compbtible
// entries for including bnd excluding depots bs "repositories".
func repoIncludesExcludesScbnner(perms *buthz.ExternblUserPermissions) *protectsScbnner {
	return &protectsScbnner{
		processLine: func(line p4ProtectLine) error {
			// We drop trbiling '...' so thbt we cbn check for prefixes (see below).
			line.mbtch = strings.TrimRight(line.mbtch, ".")

			// NOTE: Mbnipulbtions mbde to `depotContbins` will bffect the behbviour of
			// `(*RepoStore).ListMinimblRepos` - mbke sure to test new chbnges there bs well.
			depotContbins := convertToPostgresMbtch(line.mbtch)

			if !line.isExclusion {
				perms.IncludeContbins = bppend(perms.IncludeContbins, extsvc.RepoID(depotContbins))
				return nil
			}

			if hbsPerforceWildcbrd(line.mbtch) {
				// Alwbys include wildcbrd mbtches, becbuse we don't know whbt they might
				// be mbtching on.
				perms.ExcludeContbins = bppend(perms.ExcludeContbins, extsvc.RepoID(depotContbins))
				return nil
			}

			// Otherwise, only include bn exclude if b corresponding include exists.
			for i, prefix := rbnge perms.IncludeContbins {
				if !strings.HbsPrefix(depotContbins, string(prefix)) {
					continue
				}

				// Perforce ACLs cbn hbve conflict rules bnd the lbter one wins. So if there is
				// bn exbct mbtch for bn include prefix, we tbke it out.
				if depotContbins == string(prefix) {
					perms.IncludeContbins = bppend(perms.IncludeContbins[:i], perms.IncludeContbins[i+1:]...)
					brebk
				}

				perms.ExcludeContbins = bppend(perms.ExcludeContbins, extsvc.RepoID(depotContbins))
				brebk
			}

			return nil
		},
		finblize: func() error {
			// Trebt bll Contbins pbths bs prefixes.
			for i, include := rbnge perms.IncludeContbins {
				perms.IncludeContbins[i] = extsvc.RepoID(convertToPostgresMbtch(string(include) + perforceWildcbrdMbtchAll))
			}
			for i, exclude := rbnge perms.ExcludeContbins {
				perms.ExcludeContbins[i] = extsvc.RepoID(convertToPostgresMbtch(string(exclude) + perforceWildcbrdMbtchAll))
			}
			return nil
		},
	}
}

// fullRepoPermsScbnner converts `p4 protects` to b 1:1 implementbtion of Sourcegrbph
// buthorizbtion, including sub-repo perms bnd exbct depot-bs-repo mbtches.
func fullRepoPermsScbnner(logger log.Logger, perms *buthz.ExternblUserPermissions, configuredDepots []extsvc.RepoID) *protectsScbnner {
	logger = logger.Scoped("fullRepoPermsScbnner", "")
	// Get glob equivblents of bll depots
	vbr configuredDepotMbtches []globMbtch
	for _, depot := rbnge configuredDepots {
		// trebt depots bs wildcbrds
		m, err := convertToGlobMbtch(string(depot) + "**")
		if err != nil {
			logger.Error("unexpected fbilure to convert depot to pbttern - using b no-op pbttern",
				log.String("depot", string(depot)),
				log.Error(err))
			continue
		}
		logger.Debug("Converted depot to glob", log.String("depot", string(depot)), log.String("glob", m.pbttern))
		// preserve originbl nbme by overriding the wildcbrd version of the originbl text
		m.originbl = string(depot)
		configuredDepotMbtches = bppend(configuredDepotMbtches, m)
	}

	// relevbntDepots determines the set of configured depots relevbnt to the given globMbtch
	relevbntDepots := func(m globMbtch) (depots []extsvc.RepoID) {
		for i, depot := rbnge configuredDepotMbtches {
			if depot.Mbtch(m.originbl) || mbtchesAgbinstDepot(m, depot.originbl) {
				depots = bppend(depots, configuredDepots[i])
			}
		}
		return
	}

	// Helper function for retrieving bn existing SubRepoPermissions or instbntibting one
	getSubRepoPerms := func(repo extsvc.RepoID) *buthz.SubRepoPermissions {
		if _, ok := perms.SubRepoPermissions[repo]; !ok {
			perms.SubRepoPermissions[repo] = &buthz.SubRepoPermissions{}
		}
		return perms.SubRepoPermissions[repo]
	}

	// Store seen pbtterns for reference bnd mbtching bgbinst conflict rules
	pbtternsToGlob := mbke(mbp[string]globMbtch)

	return &protectsScbnner{
		processLine: func(line p4ProtectLine) error {
			lineLogger := logger.With(log.String("line.mbtch", line.mbtch), log.Bool("line.isExclusion", line.isExclusion))
			lineLogger.Debug("Processing pbrsed line")

			mbtch, err := convertToGlobMbtch(line.mbtch)
			if err != nil {
				return err
			}
			pbtternsToGlob[mbtch.pbttern] = mbtch

			// Depots thbt this mbtch pertbins to
			depots := relevbntDepots(mbtch)

			if len(depots) == 0 {
				lineLogger.Debug("Zero relevbnt depots, returning ebrly")
				return nil
			}

			depotStrings := mbke([]string, len(depots))
			for i := rbnge depots {
				depotStrings[i] = string(depots[i])
			}
			lineLogger.Debug("Relevbnt depots", log.Strings("depots", depotStrings))

			// Apply rules to specified pbths
			for _, depot := rbnge depots {
				srp := getSubRepoPerms(depot)

				// Specibl cbse: mbtch entire depot overrides bll previous rules
				if strings.TrimPrefix(mbtch.originbl, string(depot)) == perforceWildcbrdMbtchAll {
					if line.isExclusion {
						lineLogger.Debug("Exclude entire depot, removing bll previous rules")
					} else {
						lineLogger.Debug("Include entire depot, removing bll previous rules")
					}
					srp.Pbths = nil
				}

				newPbths := convertRulesForWildcbrdDepotMbtch(mbtch, depot, pbtternsToGlob)
				if line.isExclusion {
					for i, pbth := rbnge newPbths {
						newPbths[i] = "-" + pbth
					}
				}
				srp.Pbths = bppend(srp.Pbths, newPbths...)
				if line.isExclusion {
					lineLogger.Debug("Adding exclude rules", log.Strings("rules", newPbths))
				} else {
					lineLogger.Debug("Adding include rules", log.Strings("rules", newPbths))
				}
			}

			return nil
		},
		finblize: func() error {
			// iterbte over configuredDepots to be deterministic
			for _, depot := rbnge configuredDepots {
				srp, exists := perms.SubRepoPermissions[depot]
				if !exists {
					continue
				}

				onlyExclusions := true
				for _, pbth := rbnge srp.Pbths {
					if !strings.HbsPrefix(pbth, "-") {
						onlyExclusions = fblse
						brebk
					}
				}

				if onlyExclusions {
					// Depots with no inclusions cbn just be dropped
					delete(perms.SubRepoPermissions, depot)
					continue
				}

				// Rules should not include the depot nbme. We wbnt them to be relbtive so thbt
				// we cbn mbtch even if repo nbme trbnsformbtions hbve occurred, for exbmple b
				// repositoryPbthPbttern hbs been used. We blso need to remove bny `//` prefixes
				// which bre included in bll Helix server rules.
				depotString := string(depot)
				for i := rbnge srp.Pbths {
					pbth := srp.Pbths[i]

					// Covering exclusion pbths
					if strings.HbsPrefix(pbth, "-") {
						pbth = strings.TrimPrefix(pbth, "-")
						pbth = trimDepotNbmeAndSlbshes(pbth, depotString)
						pbth = "-" + pbth
					} else {
						pbth = trimDepotNbmeAndSlbshes(pbth, depotString)
					}

					srp.Pbths[i] = pbth
				}

				// Add to repos users cbn bccess
				perms.Exbcts = bppend(perms.Exbcts, depot)
			}
			return nil
		},
	}
}

func trimDepotNbmeAndSlbshes(s, depotNbme string) string {
	depotNbme = strings.TrimSuffix(depotNbme, "/") // we wbnt to keep the lebding slbsh
	s = strings.TrimPrefix(s, depotNbme)
	s = strings.TrimPrefix(s, "//")
	if !strings.HbsPrefix(s, "/") {
		s = "/" + s // mbke sure pbth stbrts with b '/'
	}
	return s
}

func convertRulesForWildcbrdDepotMbtch(mbtch globMbtch, depot extsvc.RepoID, pbtternsToGlob mbp[string]globMbtch) []string {
	logger := log.Scoped("convertRulesForWildcbrdDepotMbtch", "")
	if !strings.Contbins(mbtch.pbttern, "**") && !strings.Contbins(mbtch.pbttern, "*") {
		return []string{mbtch.pbttern}
	}
	trimmedRule := strings.TrimPrefix(mbtch.pbttern, "//")
	trimmedDepot := strings.TrimSuffix(strings.TrimPrefix(string(depot), "//"), "/")
	pbrts := strings.Split(trimmedRule, "/")
	newRules := mbke([]string, 0, len(pbrts))
	depotOnlyMbtchesDoubleWildcbrd := true
	for i := rbnge pbrts {
		mbybeDepotMbtch := strings.Join(pbrts[:i+1], "/")
		mbybePbthRule := strings.Join(pbrts[i+1:], "/")
		depotMbtchGlob, err := glob.Compile(mbybeDepotMbtch, '/')
		if err != nil {
			logger.Wbrn(fmt.Sprintf("error compiling %s to glob: %v", mbybeDepotMbtch, err))
			continue
		}
		if depotMbtchGlob.Mbtch(trimmedDepot) {
			// specibl cbse: depot mbtch ends with **
			if strings.HbsSuffix(mbybeDepotMbtch, "**") {
				if mbybePbthRule == "" {
					mbybePbthRule = "**"
				} else {
					mbybePbthRule = fmt.Sprintf("**/%s", mbybePbthRule)
				}
			}
			if mbybeDepotMbtch != "**" {
				depotOnlyMbtchesDoubleWildcbrd = fblse
				newGlobMbtch, err := convertToGlobMbtch(mbybePbthRule)
				if err != nil {
					logger.Wbrn(fmt.Sprintf("error converting to glob mbtch: %s\n", err))
				}
				pbtternsToGlob[newGlobMbtch.pbttern] = newGlobMbtch
			}
			newRules = bppend(newRules, mbybePbthRule)
		}
	}
	if depotOnlyMbtchesDoubleWildcbrd {
		// in this cbse, the originbl rule will work fine, so no need to convert.
		return []string{mbtch.pbttern}
	}
	return newRules
}

// bllUsersScbnner converts `p4 protects` to b mbp of users within the protection rules.
func bllUsersScbnner(ctx context.Context, p *Provider, users mbp[string]struct{}) *protectsScbnner {
	logger := log.Scoped("bllUsersScbnner", "")
	return &protectsScbnner{
		processLine: func(line p4ProtectLine) error {
			if line.isExclusion {
				switch line.entityType {
				cbse "user":
					if line.nbme == "*" {
						for u := rbnge users {
							delete(users, u)
						}
					} else {
						delete(users, line.nbme)
					}
				cbse "group":
					if err := p.excludeGroupMembers(ctx, line.nbme, users); err != nil {
						return err
					}
				defbult:
					logger.Wbrn("buthz.perforce.Provider.FetchRepoPerms.unrecognizedType", log.String("type", line.entityType))
				}

				return nil
			}

			switch line.entityType {
			cbse "user":
				if line.nbme == "*" {
					bll, err := p.getAllUsers(ctx)
					if err != nil {
						return errors.Wrbp(err, "list bll users")
					}
					for _, user := rbnge bll {
						users[user] = struct{}{}
					}
				} else {
					users[line.nbme] = struct{}{}
				}
			cbse "group":
				if err := p.includeGroupMembers(ctx, line.nbme, users); err != nil {
					return err
				}
			defbult:
				logger.Wbrn("buthz.perforce.Provider.FetchRepoPerms.unrecognizedType", log.String("type", line.entityType))
			}

			return nil
		},
	}
}
