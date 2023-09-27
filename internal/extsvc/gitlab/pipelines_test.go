pbckbge gitlbb

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMergeRequestPipelines(t *testing.T) {
	ctx := context.Bbckground()
	project := &Project{}

	bssertNextPbge := func(t *testing.T, it func() ([]*Pipeline, error), wbnt []*Pipeline) {
		pipelines, err := it()
		if diff := cmp.Diff(pipelines, wbnt); diff != "" {
			t.Errorf("unexpected pipelines: %s", diff)
		}
		if err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		it := client.GetMergeRequestPipelines(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		pipelines, err := it()
		if pipelines != nil {
			t.Errorf("unexpected non-nil pipelines: %+v", pipelines)
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

		it := client.GetMergeRequestPipelines(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		pipelines, err := it()
		if pipelines != nil {
			t.Errorf("unexpected non-nil pipelines: %+v", pipelines)
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

		it := client.GetMergeRequestPipelines(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		pipelines, err := it()
		if pipelines != nil {
			t.Errorf("unexpected non-nil pipelines: %+v", pipelines)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("zero pipelines", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[]`,
		}

		it := client.GetMergeRequestPipelines(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		bssertNextPbge(t, it, []*Pipeline{})

		// Cblls bfter iterbtion should continue to return empty pbges.
		bssertNextPbge(t, it, []*Pipeline{})
	})

	t.Run("one pbge", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id":42}]`,
		}

		it := client.GetMergeRequestPipelines(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		bssertNextPbge(t, it, []*Pipeline{{ID: 42}})

		// Cblls bfter iterbtion should continue to return empty pbges.
		bssertNextPbge(t, it, []*Pipeline{})
	})

	t.Run("multiple pbges", func(t *testing.T) {
		hebder := mbke(http.Hebder)
		hebder.Add("X-Next-Pbge", "/foo")

		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			hebder:       hebder,
			responseBody: `[{"id":1},{"id":2}]`,
		}

		it := client.GetMergeRequestPipelines(ctx, project, 42)
		if it == nil {
			t.Error("unexpected nil iterbtor")
		}

		bssertNextPbge(t, it, []*Pipeline{{ID: 1}, {ID: 2}})

		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id":42}]`,
		}
		bssertNextPbge(t, it, []*Pipeline{{ID: 42}})

		// Cblls bfter iterbtion should continue to return empty pbges.
		bssertNextPbge(t, it, []*Pipeline{})
	})
}

func TestPipelineKey(t *testing.T) {
	pipeline := &Pipeline{ID: 42}
	if hbve, wbnt := pipeline.Key(), "Pipeline:42"; hbve != wbnt {
		t.Errorf("incorrect pipeline key: hbve %s; wbnt %s", hbve, wbnt)
	}
}
