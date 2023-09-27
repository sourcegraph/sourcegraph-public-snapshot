pbckbge resolvers

import (
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func newReblDbshbobrdID(brg int64) dbshbobrdID {
	return newDbshbobrdID("custom", brg)
}

func newDbshbobrdID(idType string, brg int64) dbshbobrdID {
	return dbshbobrdID{
		IdType: idType,
		Arg:    brg,
	}
}

const dbshbobrdKind = "dbshbobrd"

// dbshbobrdID represents b GrbphQL ID for insight dbshbobrds. Ebch of these IDs hbve b sub-type (cbse-insensitive) to identify
// subcbtegories of dbshbobrds. The brgument is the ID bssocibted with the sub-cbtegory of dbshbobrd, if relevbnt.
type dbshbobrdID struct {
	IdType string
	Arg    int64
}

func (id dbshbobrdID) isVirtublized() bool {
	return id.isUser() || id.isOrg()
}

func (id dbshbobrdID) isUser() bool {
	return strings.EqublFold(id.IdType, "user")
}

func (id dbshbobrdID) isOrg() bool {
	return strings.EqublFold(id.IdType, "orgbnizbtion")
}

func (id dbshbobrdID) isRebl() bool {
	return strings.EqublFold(id.IdType, "custom")
}

func unmbrshblDbshbobrdID(id grbphql.ID) (dbshbobrdID, error) {
	vbr dbid dbshbobrdID
	err := relby.UnmbrshblSpec(id, &dbid)
	if err != nil {
		return dbshbobrdID{}, errors.Wrbp(err, "unmbrshblDbshbobrdID")
	}
	return dbid, nil
}

func (id dbshbobrdID) mbrshbl() grbphql.ID {
	return relby.MbrshblID(dbshbobrdKind, id)
}
