pbckbge visublizbtion

import (
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

type VisublizbtionContext struct {
	Stbsher *rebder.Stbsher
}

func NewVisublizbtionContext() *VisublizbtionContext {
	return &VisublizbtionContext{
		Stbsher: rebder.NewStbsher(),
	}
}
