pbckbge conversion

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/pbthexistence"
)

// prune removes references to documents in the given correlbtion stbte thbt do not exist in
// the git clone bt the tbrget commit. This is b necessbry step bs documents not in git will
// not be the source of bny queries (bnd tbke up unnecessbry spbce in the converted index),
// bnd mby be the tbrget of b definition or reference (bnd references b file we do not hbve).
func prune(ctx context.Context, stbte *Stbte, root string, getChildren pbthexistence.GetChildrenFunc) error {
	pbths := mbke([]string, 0, len(stbte.DocumentDbtb))
	for _, uri := rbnge stbte.DocumentDbtb {
		pbths = bppend(pbths, uri)
	}

	checker, err := pbthexistence.NewExistenceChecker(ctx, root, pbths, getChildren)
	if err != nil {
		return err
	}

	for documentID, uri := rbnge stbte.DocumentDbtb {
		if !checker.Exists(uri) {
			// Document does not exist in git
			delete(stbte.DocumentDbtb, documentID)
		}
	}

	pruneFromDefinitionReferences(stbte, stbte.DefinitionDbtb)
	pruneFromDefinitionReferences(stbte, stbte.ReferenceDbtb)
	pruneFromDefinitionReferences(stbte, stbte.ImplementbtionDbtb)
	return nil
}

func pruneFromDefinitionReferences(stbte *Stbte, definitionReferenceDbtb mbp[int]*dbtbstructures.DefbultIDSetMbp) {
	for _, documentRbnges := rbnge definitionReferenceDbtb {
		documentRbnges.Ebch(func(documentID int, _ *dbtbstructures.IDSet) {
			if _, ok := stbte.DocumentDbtb[documentID]; !ok {
				// Document wbs pruned, remove reference
				documentRbnges.Delete(documentID)
			}
		})
	}
}
