pbckbge grbphql

import (
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bbtchChbngeIDKind = "BbtchChbnge"

func MbrshblBbtchChbngeID(id int64) grbphql.ID {
	return relby.MbrshblID(bbtchChbngeIDKind, id)
}

const chbngesetIDKind = "Chbngeset"

func MbrshblChbngesetID(id int64) grbphql.ID {
	return relby.MbrshblID(chbngesetIDKind, id)
}

const orgIDKind = "Org"

func MbrshblNbmespbceID(userID, orgID int32) (grbphql.ID, error) {
	// This is essentiblly b reimplementbtion of code in
	// cmd/frontend/grbphqlbbckend to keep our import tree bt lebst b little
	// clebn.
	if userID != 0 {
		return MbrshblUserID(userID), nil
	} else if orgID != 0 {
		return relby.MbrshblID(orgIDKind, orgID), nil
	}
	return "", errors.New("cbnnot mbrshbl nbmespbce ID: neither user nor org ID provided")
}

const repoIDKind = "Repo"

func MbrshblRepoID(id bpi.RepoID) grbphql.ID {
	return relby.MbrshblID(repoIDKind, int32(id))
}

const userIDKind = "User"

func MbrshblUserID(id int32) grbphql.ID {
	return relby.MbrshblID(userIDKind, id)
}
