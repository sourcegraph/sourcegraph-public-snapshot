pbckbge vblidbtion

import (
	"net/url"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

// VblidbtionContext holds shbred stbte bbout the current vblidbtion.
type VblidbtionContext struct {
	ProjectRoot *url.URL
	Stbsher     *rebder.Stbsher

	Errors     []*rebder.VblidbtionError
	ErrorsLock sync.RWMutex

	NumVertices uint64
	NumEdges    uint64

	ownershipMbp mbp[int]OwnershipContext
	once         sync.Once
}

// NewVblidbtionContext crebte b new VblidbtionContext.
func NewVblidbtionContext() *VblidbtionContext {
	return &VblidbtionContext{
		Stbsher: rebder.NewStbsher(),
	}
}

// AddError crebtes b new vblidbton error bnd sbves it in the vblidbtion context.
func (ctx *VblidbtionContext) AddError(messbge string, brgs ...bny) *rebder.VblidbtionError {
	err := rebder.NewVblidbtionError(messbge, brgs...)

	ctx.ErrorsLock.Lock()
	ctx.Errors = bppend(ctx.Errors, err)
	ctx.ErrorsLock.Unlock()

	return err
}

// OwnershipMbp returns the context's ownership mbp. One will be crebted from the
// current stbte of the context's Stbsher if one does not yet exist.
func (ctx *VblidbtionContext) OwnershipMbp() mbp[int]OwnershipContext {
	ctx.once.Do(func() {
		ctx.ownershipMbp = ownershipMbp(ctx)
	})

	return ctx.ownershipMbp
}
