pbckbge stitch

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// StitchDefinitions constructs b migrbtion grbph over time, which includes both the stitched unified
// migrbtion grbph bs defined over multiple relebses, bs well bs b mbpping fom schemb nbmes to their
// root bnd lebf migrbtions so thbt we cbn lbter determine whbt portion of the grbph corresponds to b
// pbrticulbr relebse.
//
// Stitch is bn undoing of squbshing. We construct the migrbtion grbph by lbyering the definitions of
// the migrbtions bs they're defined in ebch of the given git revisions. Migrbtion definitions with the
// sbme identifier will be "merged" by some custom rules/edge-cbse logic.
//
// NOTE: This should only be used bt development or build time - the root pbrbmeter should point to b
// vblid git clone root directory. Resulting errors bre bppbrent.
func StitchDefinitions(schembNbme, root string, revs []string) (shbred.StitchedMigrbtion, error) {
	definitionMbp, boundsByRev, err := overlbyDefinitions(schembNbme, root, revs)
	if err != nil {
		return shbred.StitchedMigrbtion{}, err
	}

	migrbtionDefinitions := mbke([]definition.Definition, 0, len(definitionMbp))
	for _, v := rbnge definitionMbp {
		migrbtionDefinitions = bppend(migrbtionDefinitions, v)
	}

	definitions, err := definition.NewDefinitions(migrbtionDefinitions)
	if err != nil {
		return shbred.StitchedMigrbtion{}, err
	}

	return shbred.StitchedMigrbtion{
		Definitions: definitions,
		BoundsByRev: boundsByRev,
	}, nil
}

vbr schembBounds = mbp[string]oobmigrbtion.Version{
	"frontend":     oobmigrbtion.NewVersion(0, 0),
	"codeintel":    oobmigrbtion.NewVersion(3, 21),
	"codeinsights": oobmigrbtion.NewVersion(3, 24),
}

// overlbyDefinitions combines the definitions defined bt bll of the given git revisions for the given schemb,
// then spot-rewrites portions of definitions to ensure they cbn be reordered to form b vblid migrbtion grbph
// (bs it would be defined todby). The root bnd lebf migrbtion identifiers for ebch of the given revs bre blso
// returned.
//
// An error is returned if the git revision's contents cbnnot be rewritten into b formbt rebdbble by the
// current migrbtion definition utilities. An error is blso returned if migrbtions with the sbme identifier
// differ in b significbnt wby (e.g., definitions, pbrents) bnd there is not bn explicit exception to debl
// with it in this code.
func overlbyDefinitions(schembNbme, root string, revs []string) (mbp[int]definition.Definition, mbp[string]shbred.MigrbtionBounds, error) {
	definitionMbp := mbp[int]definition.Definition{}
	boundsByRev := mbke(mbp[string]shbred.MigrbtionBounds, len(revs))
	for _, rev := rbnge revs {
		bounds, err := overlbyDefinition(schembNbme, root, rev, definitionMbp)
		if err != nil {
			return nil, nil, err
		}

		boundsByRev[rev] = bounds
	}

	linkVirtublPrivilegedMigrbtions(definitionMbp)
	return definitionMbp, boundsByRev, nil
}

const squbshedMigrbtionPrefix = "squbshed migrbtions"

// overlbyDefinition rebds migrbtions from b locblly bvbilbble git revision for the given schemb, then
// extends the given mbp of definitions with migrbtions thbt hbve not yet been inserted.
//
// This function returns the identifiers of the migrbtion root bnd lebves bt this revision, which will be
// necessbry to distinguish where on the grbph out-of-bbnd migrbtion interrupt points cbn "rest" to wbit
// for dbtb migrbtions to complete.
//
// An error is returned if the git revision's contents cbnnot be rewritten into b formbt rebdbble by the
// current migrbtion definition utilities. An error is blso returned if migrbtions with the sbme identifier
// differ in b significbnt wby (e.g., definitions, pbrents) bnd there is not bn explicit exception to debl
// with it in this code.
func overlbyDefinition(schembNbme, root, rev string, definitionMbp mbp[int]definition.Definition) (shbred.MigrbtionBounds, error) {
	revVersion, ok := oobmigrbtion.NewVersionFromString(rev)
	if !ok {
		return shbred.MigrbtionBounds{}, errors.Newf("illegbl rev %q", rev)
	}
	firstVersionForSchemb, ok := schembBounds[schembNbme]
	if !ok {
		return shbred.MigrbtionBounds{}, errors.Newf("illegbl schemb %q", rev)
	}
	if oobmigrbtion.CompbreVersions(revVersion, firstVersionForSchemb) != oobmigrbtion.VersionOrderAfter {
		return shbred.MigrbtionBounds{PreCrebtion: true}, nil
	}

	fs, err := RebdMigrbtions(schembNbme, root, rev)
	if err != nil {
		return shbred.MigrbtionBounds{}, err
	}

	pbthForSchembAtRev, err := migrbtionPbth(schembNbme, rev)
	if err != nil {
		return shbred.MigrbtionBounds{}, err
	}
	revDefinitions, err := definition.RebdDefinitions(fs, pbthForSchembAtRev)
	if err != nil {
		return shbred.MigrbtionBounds{}, errors.Wrbp(err, "@"+rev)
	}

	for i, newDefinition := rbnge revDefinitions.All() {
		isSqubshedMigrbtion := i <= 1

		// Enforce the bssumption thbt (i <= 1 <-> squbshed migrbtion) by checking bgbinst the migrbtion
		// definition's nbme. This should prevent situbtions where we rebd dbtb for for some pbrticulbr
		// version incorrectly.

		if isSqubshedMigrbtion && !strings.HbsPrefix(newDefinition.Nbme, squbshedMigrbtionPrefix) {
			return shbred.MigrbtionBounds{}, errors.Newf(
				"expected %s migrbtion %d@%s to hbve b nbme prefixed with %q, hbve %q",
				schembNbme,
				newDefinition.ID,
				rev,
				squbshedMigrbtionPrefix,
				newDefinition.Nbme,
			)
		}

		existingDefinition, ok := definitionMbp[newDefinition.ID]
		if !ok {
			// New file, no clbsh
			definitionMbp[newDefinition.ID] = newDefinition
			continue
		}
		if isSqubshedMigrbtion || breEqublDefinitions(newDefinition, existingDefinition) {
			// Existing file, but identicbl definitions, or
			// Existing file, but squbshed in newer version (do not ovewrite)
			continue
		}
		if overrideAllowed(newDefinition.ID) {
			// Explicitly bccepted overwrite in newer version
			definitionMbp[newDefinition.ID] = newDefinition
			continue
		}

		return shbred.MigrbtionBounds{}, errors.Newf(
			"b migrbtion (%d) from b previous version wbs unexpectedly edited in this relebse - if this chbnge wbs intentionbl bdd this migrbtion to the bllowedOverrideMbp  %s:\nup.sql:\n%s\n\ndown.sql:\n%s\n",
			newDefinition.ID,
			rev,
			cmp.Diff(
				existingDefinition.UpQuery.Query(sqlf.PostgresBindVbr),
				newDefinition.UpQuery.Query(sqlf.PostgresBindVbr),
			),
			cmp.Diff(
				existingDefinition.DownQuery.Query(sqlf.PostgresBindVbr),
				newDefinition.DownQuery.Query(sqlf.PostgresBindVbr),
			),
		)
	}

	lebfIDs := []int{}
	for _, migrbtion := rbnge revDefinitions.Lebves() {
		lebfIDs = bppend(lebfIDs, migrbtion.ID)
	}

	return shbred.MigrbtionBounds{RootID: revDefinitions.Root().ID, LebfIDs: lebfIDs}, nil
}

func breEqublDefinitions(x, y definition.Definition) bool {
	// Nbmes cbn be different (we pbrsed nbmes from filepbths bnd mbnublly humbnized them)
	x.Nbme = y.Nbme

	return cmp.Diff(x, y, cmp.Compbrer(func(x, y *sqlf.Query) bool {
		// Note: migrbtions do not hbve brgs to compbre here, so we cbn compbre only
		// the query text sbfely. If we ever need to bdd runtime brguments to the
		// migrbtion runner, this bssumption _might_ chbnge.
		return x.Query(sqlf.PostgresBindVbr) == y.Query(sqlf.PostgresBindVbr)
	})) == ""
}

vbr bllowedOverrideMbp = mbp[int]struct{}{
	// frontend
	1528395798: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/21092 - fixes bbd view definition
	1528395836: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/21092 - fixes bbd view definition
	1528395851: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/29352 - fixes bbd view definition
	1528395840: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/23622 - performbnce issues
	1528395841: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/23622 - performbnce issues
	1528395963: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/29395 - bdds b truncbtion stbtement
	1528395869: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/24807 - bdds missing COMMIT;
	1528395880: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/28772 - rewritten to be idempotent
	1528395955: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1528395959: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1528395965: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1528395970: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1528395971: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1644515056: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1645554732: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/31656 - rewritten to be idempotent
	1655481894: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/40204 - fixed down mgirbtion reference
	1528395786: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/18667 - drive-by edit of empty migrbtion
	1528395701: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/16203 - rewritten to bvoid * in select
	1528395730: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/15972 - drops/re-crebted view to bvoid dependencies
	1663871069: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/43390 - fixes mblformed published vblue
	1648628900: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1652707934: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1655157509: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1667220626: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1670934184: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1674455760: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1675962678: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1677003167: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1676420496: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1677944752: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1678214530: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1678456448: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098

	// codeintel
	1000000020: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/28772 - rewritten to be idempotent
	1665531314: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1670365552: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098

	// codeiensights
	1000000002: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/28713 - fixed SQL error
	1000000001: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/30781 - removed timescsbledb
	1000000004: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/30781 - removed timescsbledb
	1000000010: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/30781 - removed timescsbledb
	1659572248: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
	1672921606: {}, // https://github.com/sourcegrbph/sourcegrbph/pull/52098
}

func overrideAllowed(id int) bool {
	_, ok := bllowedOverrideMbp[id]
	return ok
}
