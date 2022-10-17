package validation

import (
	"net/url"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/reader"
)

// ValidationContext holds shared state about the current validation.
type ValidationContext struct {
	ProjectRoot *url.URL
	Stasher     *reader.Stasher

	Errors     []*reader.ValidationError
	ErrorsLock sync.RWMutex

	NumVertices uint64
	NumEdges    uint64

	ownershipMap map[int]OwnershipContext
	once         sync.Once
}

// NewValidationContext create a new ValidationContext.
func NewValidationContext() *ValidationContext {
	return &ValidationContext{
		Stasher: reader.NewStasher(),
	}
}

// AddError creates a new validaton error and saves it in the validation context.
func (ctx *ValidationContext) AddError(message string, args ...any) *reader.ValidationError {
	err := reader.NewValidationError(message, args...)

	ctx.ErrorsLock.Lock()
	ctx.Errors = append(ctx.Errors, err)
	ctx.ErrorsLock.Unlock()

	return err
}

// OwnershipMap returns the context's ownership map. One will be created from the
// current state of the context's Stasher if one does not yet exist.
func (ctx *ValidationContext) OwnershipMap() map[int]OwnershipContext {
	ctx.once.Do(func() {
		ctx.ownershipMap = ownershipMap(ctx)
	})

	return ctx.ownershipMap
}
