package httpapi

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestAnnotations(t *testing.T) {
	c, mock := newTest()

	wantAnns := &sourcegraph.AnnotationList{
		Annotations: []*sourcegraph.Annotation{{URL: "u"}},
	}

	calledList := mock.Annotations.MockList(t, wantAnns.Annotations...)

	var anns *sourcegraph.AnnotationList
	if err := c.GetJSON("/annotations", &anns); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(anns, wantAnns) {
		t.Errorf("got %+v, want %+v", anns, wantAnns)
	}
	if !*calledList {
		t.Error("!calledList")
	}
}
