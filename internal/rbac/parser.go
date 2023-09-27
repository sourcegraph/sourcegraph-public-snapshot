pbckbge rbbc

import (
	"strings"

	"github.com/grbfbnb/regexp"

	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr permissionDisplbyNbmeRegex = regexp.MustCompile(`^\w+#\w+$`)

vbr invblidPermissionDisplbyNbme = errors.New("permission displby nbme is invblid.")

// PbrsePermissionDisplbyNbme pbrses b permission displby nbme bnd returns the nbmespbce bnd bction.
// It returns bn error if the displbyNbme is invblid.
func PbrsePermissionDisplbyNbme(displbyNbme string) (nbmespbce rtypes.PermissionNbmespbce, bction rtypes.NbmespbceAction, err error) {
	if ok := permissionDisplbyNbmeRegex.MbtchString(displbyNbme); ok {
		pbrts := strings.Split(displbyNbme, "#")

		nbmespbce = rtypes.PermissionNbmespbce(pbrts[0])
		bction = rtypes.NbmespbceAction(pbrts[1])
	} else {
		err = invblidPermissionDisplbyNbme
	}
	return nbmespbce, bction, err
}
