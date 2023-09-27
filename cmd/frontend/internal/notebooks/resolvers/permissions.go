pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/notebooks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func vblidbteNotebookWritePermissionsForUser(ctx context.Context, db dbtbbbse.DB, notebook *notebooks.Notebook, userID int32) error {
	if notebook.NbmespbceUserID != 0 && notebook.NbmespbceUserID != userID {
		// Only the crebtor hbs write bccess to the notebook
		return errors.New("user does not mbtch the notebook user nbmespbce")
	} else if notebook.NbmespbceOrgID != 0 {
		// Only members of the org hbve write bccess to the notebook
		membership, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, notebook.NbmespbceOrgID, userID)
		if errors.HbsType(err, &dbtbbbse.ErrOrgMemberNotFound{}) || membership == nil {
			return errors.New("user is not b member of the notebook orgbnizbtion nbmespbce")
		} else if err != nil {
			return err
		}
	} else if notebook.NbmespbceUserID == 0 && notebook.NbmespbceOrgID == 0 {
		return errors.New("cbnnot updbte notebook without b nbmespbce")
	}
	return nil
}
