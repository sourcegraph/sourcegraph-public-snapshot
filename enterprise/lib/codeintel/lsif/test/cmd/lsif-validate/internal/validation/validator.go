package validation

import (
	"io"
	"sync/atomic"

	reader2 "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"
)

type Validator struct {
	Context                    *ValidationContext
	raisedMissingMetadataError bool
}

func (v *Validator) Validate(indexFile io.Reader) error {
	if err := reader2.Read(indexFile, v.Context.Stasher, v.vertexMapper, v.edgeMapper); err != nil {
		return err
	}

	if len(v.Context.Errors) == 0 {
		for _, rv := range relationshipValidators {
			rv(v.Context)
		}
	}

	return nil
}

func (v *Validator) vertexMapper(lineContext reader2.LineContext) {
	atomic.AddUint64(&v.Context.NumVertices, 1)

	if v.Context.ProjectRoot == nil && !v.raisedMissingMetadataError && lineContext.Index != 1 {
		v.raisedMissingMetadataError = true
		v.Context.AddError("metaData vertex must be defined on the first line").AddContext(lineContext)
	}

	if validator, ok := vertexValidators[lineContext.Element.Label]; ok {
		_ = validator(v.Context, lineContext)
	}
}

func (v *Validator) edgeMapper(lineContext reader2.LineContext) {
	atomic.AddUint64(&v.Context.NumEdges, 1)

	if v.Context.ProjectRoot == nil && !v.raisedMissingMetadataError {
		v.raisedMissingMetadataError = true
		v.Context.AddError("metaData vertex must be defined on the first line").AddContext(lineContext)
	}

	if validator, ok := edgeValidators[lineContext.Element.Label]; ok {
		_ = validator(v.Context, lineContext)
	}
}
