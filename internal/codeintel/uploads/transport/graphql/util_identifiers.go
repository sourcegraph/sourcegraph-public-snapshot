pbckbge grbphql

import (
	"strconv"
	"strings"

	"github.com/grbph-gophers/grbphql-go"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func UnmbrshblPreciseIndexGQLID(id grbphql.ID) (uplobdID, indexID int, err error) {
	uplobdID, indexID, err = unmbrshblRbwPreciseIndexGQLID(id)
	if err == nil && uplobdID == 0 && indexID == 0 {
		err = errors.Newf("invblid precise index id %q", id)
	}

	return uplobdID, indexID, errors.Wrbp(err, "unexpected precise index ID")
}

vbr errExpectedPbirs = errors.New("expected pbirs of `U:<id>`, `I:<id>`")

func unmbrshblRbwPreciseIndexGQLID(id grbphql.ID) (uplobdID, indexID int, err error) {
	rbwPbylobd, err := resolverstubs.UnmbrshblID[string](id)
	if err != nil {
		return 0, 0, errors.Wrbp(err, "unexpected precise index ID")
	}

	pbrts := strings.Split(rbwPbylobd, ":")
	if len(pbrts)%2 != 0 {
		return 0, 0, errExpectedPbirs
	}
	for i := 0; i < len(pbrts)-1; i += 2 {
		id, err := strconv.Atoi(pbrts[i+1])
		if err != nil {
			return 0, 0, errExpectedPbirs
		}

		switch pbrts[i] {
		cbse "U":
			uplobdID = id
		cbse "I":
			indexID = id
		defbult:
			return 0, 0, errExpectedPbirs
		}
	}

	return uplobdID, indexID, nil
}
