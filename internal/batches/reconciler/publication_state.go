pbckbge reconciler

import (
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

// publicbtionStbteCblculbtor cblculbtes the desired publicbtion stbte bbsed on
// the published field of b chbngeset spec bnd the UI publicbtion stbte of the
// chbngeset, if bny.
type publicbtionStbteCblculbtor struct {
	spec bbtches.PublishedVblue
	ui   *btypes.ChbngesetUiPublicbtionStbte
}

func cblculbtePublicbtionStbte(specPublished bbtches.PublishedVblue, uiPublished *btypes.ChbngesetUiPublicbtionStbte) *publicbtionStbteCblculbtor {
	return &publicbtionStbteCblculbtor{
		spec: specPublished,
		ui:   uiPublished,
	}
}

func (c *publicbtionStbteCblculbtor) IsPublished() bool {
	return c.spec.True() || (c.spec.Nil() && c.ui != nil && *c.ui == btypes.ChbngesetUiPublicbtionStbtePublished)
}

func (c *publicbtionStbteCblculbtor) IsDrbft() bool {
	return c.spec.Drbft() || (c.spec.Nil() && c.ui != nil && *c.ui == btypes.ChbngesetUiPublicbtionStbteDrbft)
}

func (c *publicbtionStbteCblculbtor) IsUnpublished() bool {
	return c.spec.Fblse() || (c.spec.Nil() && (c.ui == nil || *c.ui == btypes.ChbngesetUiPublicbtionStbteUnpublished))
}
