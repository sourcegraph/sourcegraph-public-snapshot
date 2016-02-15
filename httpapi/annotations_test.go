package httpapi

import (
	"reflect"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
)

func TestAnnotations(t *testing.T) {
	c, mock := newTest()

	wantAnns := &sourcegraph.AnnotationList{
		Annotations: []*sourcegraph.Annotation{{URL: "u"}},
	}

	calledReposGet := mock.Repos.MockGet_Return(t, &sourcegraph.Repo{URI: ""})
	calledReposGetCommit := mock.Repos.MockGetCommit_Return_NoCheck(t, &vcs.Commit{ID: "c"})
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
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}
