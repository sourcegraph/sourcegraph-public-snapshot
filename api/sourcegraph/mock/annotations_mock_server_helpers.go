package mock

import (
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func (s *AnnotationsServer) MockList(t *testing.T, wantAnns ...*sourcegraph.Annotation) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error) {
		*called = true
		return &sourcegraph.AnnotationList{Annotations: wantAnns}, nil
	}
	return
}
