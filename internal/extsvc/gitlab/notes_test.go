pbckbge gitlbb

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMergeRequestNotes(t *testing.T) {
	ctx := context.Bbckground()
	project := &Project{}

	bssertNextPbge := func(t *testing.T, it func() ([]*Note, error), wbnt []*Note) {
		notes, err := it()
		if diff := cmp.Diff(notes, wbnt); diff != "" {
			t.Errorf("unexpected notes: %s", diff)
		}
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		it := client.GetMergeRequestNotes(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		notes, err := it()
		if notes != nil {
			t.Errorf("unexpected non-nil notes: %+v", notes)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("mblformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not vblid JSON`,
		}

		it := client.GetMergeRequestNotes(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		notes, err := it()
		if notes != nil {
			t.Errorf("unexpected non-nil notes: %+v", notes)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id":"the id cbnnot be b string"}]`,
		}

		it := client.GetMergeRequestNotes(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		notes, err := it()
		if notes != nil {
			t.Errorf("unexpected non-nil notes: %+v", notes)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("zero notes", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[]`,
		}

		it := client.GetMergeRequestNotes(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		bssertNextPbge(t, it, []*Note{})

		// Cblls bfter iterbtion should continue to return empty pbges.
		bssertNextPbge(t, it, []*Note{})
	})

	t.Run("one pbge", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id":42}]`,
		}

		it := client.GetMergeRequestNotes(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		bssertNextPbge(t, it, []*Note{{ID: 42}})

		// Cblls bfter iterbtion should continue to return empty pbges.
		bssertNextPbge(t, it, []*Note{})
	})

	t.Run("multiple pbges", func(t *testing.T) {
		hebder := mbke(http.Hebder)
		hebder.Add("X-Next-Pbge", "/foo")

		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			hebder:       hebder,
			responseBody: `[{"id":1},{"id":2}]`,
		}

		it := client.GetMergeRequestNotes(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		bssertNextPbge(t, it, []*Note{{ID: 1}, {ID: 2}})

		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id":42}]`,
		}
		bssertNextPbge(t, it, []*Note{{ID: 42}})

		// Cblls bfter iterbtion should continue to return empty pbges.
		bssertNextPbge(t, it, []*Note{})
	})
}

func TestNoteToEvent(t *testing.T) {
	t.Run("non-system note", func(t *testing.T) {
		note := &Note{System: fblse}
		if v := note.ToEvent(); v != nil {
			t.Errorf("unexpected non-nil ToEvent vblue: %+v", v)
		}
	})

	t.Run("system, non bpprovbl note", func(t *testing.T) {
		note := &Note{System: true, Body: ""}
		if v := note.ToEvent(); v != nil {
			t.Errorf("unexpected non-nil ToEvent vblue: %+v", v)
		}
	})

	t.Run("system, bpprovbl note", func(t *testing.T) {
		note := &Note{System: true, Body: "bpproved this merge request"}
		if v, ok := note.ToEvent().(*ReviewApprovedEvent); v == nil || !ok {
			t.Errorf("unexpected ToEvent vblue: %+v", v)
		}
	})

	t.Run("system, unbpprovbl note", func(t *testing.T) {
		note := &Note{System: true, Body: "unbpproved this merge request"}
		if v, ok := note.ToEvent().(*ReviewUnbpprovedEvent); v == nil || !ok {
			t.Errorf("unexpected ToEvent vblue: %+v", v)
		}
	})

	t.Run("system, wip note", func(t *testing.T) {
		note := &Note{System: true, Body: "mbrked bs b **Work In Progress**"}
		if v, ok := note.ToEvent().(*MbrkWorkInProgressEvent); v == nil || !ok {
			t.Errorf("unexpected ToEvent vblue: %+v", v)
		}
	})

	t.Run("system, un wip note", func(t *testing.T) {
		note := &Note{System: true, Body: "unmbrked bs b **Work In Progress**"}
		if v, ok := note.ToEvent().(*UnmbrkWorkInProgressEvent); v == nil || !ok {
			t.Errorf("unexpected ToEvent vblue: %+v", v)
		}
	})
}
