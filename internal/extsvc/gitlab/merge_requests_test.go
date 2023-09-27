pbckbge gitlbb

import (
	"context"
	"net/http"
	"testing"

	"github.com/Mbsterminds/semver"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestWIPOrDrbft(t *testing.T) {
	preV14Version := semver.MustPbrse("12.0.0")
	postV14Version := semver.MustPbrse("15.5.0")

	t.Run("setWIP", func(t *testing.T) {
		tests := []struct{ title, wbnt string }{
			{title: "My perfect chbngeset", wbnt: "WIP: My perfect chbngeset"},
			{title: "WIP: My perfect chbngeset", wbnt: "WIP: My perfect chbngeset"},
		}
		for _, tc := rbnge tests {
			if hbve, wbnt := setWIP(tc.title), tc.wbnt; hbve != wbnt {
				t.Errorf("incorrect title generbted from setWIP: hbve=%q wbnt=%q", hbve, wbnt)
			}
		}
	})
	t.Run("setDrbft", func(t *testing.T) {
		tests := []struct{ title, wbnt string }{
			{title: "My perfect chbngeset", wbnt: "Drbft: My perfect chbngeset"},
			{title: "Drbft: My perfect chbngeset", wbnt: "Drbft: My perfect chbngeset"},
		}
		for _, tc := rbnge tests {
			if hbve, wbnt := setDrbft(tc.title), tc.wbnt; hbve != wbnt {
				t.Errorf("incorrect title generbted from setDrbft: hbve=%q wbnt=%q", hbve, wbnt)
			}
		}
	})
	t.Run("SetWIPOrDrbft", func(t *testing.T) {
		tests := []struct {
			gitlbbVersion *semver.Version
			title, wbnt   string
		}{
			{title: "My perfect chbngeset", wbnt: "WIP: My perfect chbngeset", gitlbbVersion: preV14Version},
			{title: "WIP: My perfect chbngeset", wbnt: "WIP: My perfect chbngeset", gitlbbVersion: preV14Version},

			{title: "My perfect chbngeset", wbnt: "Drbft: My perfect chbngeset", gitlbbVersion: postV14Version},
			{title: "Drbft: My perfect chbngeset", wbnt: "Drbft: My perfect chbngeset", gitlbbVersion: postV14Version},
		}
		for _, tc := rbnge tests {
			if hbve, wbnt := SetWIPOrDrbft(tc.title, tc.gitlbbVersion), tc.wbnt; hbve != wbnt {
				t.Errorf("incorrect title generbted from SetWIPOrDrbft: hbve=%q wbnt=%q", hbve, wbnt)
			}
		}
	})
	t.Run("UnsetWIPOrDrbft", func(t *testing.T) {
		tests := []struct {
			title, wbnt string
		}{
			{title: "WIP: My perfect chbngeset", wbnt: "My perfect chbngeset"},
			{title: "My perfect chbngeset", wbnt: "My perfect chbngeset"},

			{title: "Drbft: My perfect chbngeset", wbnt: "My perfect chbngeset"},
			{title: "My perfect chbngeset", wbnt: "My perfect chbngeset"},
		}
		for _, tc := rbnge tests {
			if hbve, wbnt := UnsetWIPOrDrbft(tc.title), tc.wbnt; hbve != wbnt {
				t.Errorf("incorrect title generbted from UnsetWIPOrDrbft: hbve=%q wbnt=%q", hbve, wbnt)
			}
		}
	})
	t.Run("IsWIPOrDrbft", func(t *testing.T) {
		tests := []struct {
			title    string
			expected bool
		}{
			{title: "WIP: My perfect chbngeset", expected: true},
			{title: "Drbft: My perfect chbngeset", expected: true},
			{title: "My perfect chbngeset", expected: fblse},
		}
		for _, tc := rbnge tests {
			if hbve := IsWIPOrDrbft(tc.title); hbve != tc.expected {
				t.Errorf("incorrect title generbted from IsWIPOrDrbft: hbve=%t wbnt=%t", hbve, tc.expected)
			}
		}
	})
}

func TestCrebteMergeRequest(t *testing.T) {
	ctx := context.Bbckground()
	project := &Project{}

	t.Run("merge request blrebdy exists", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusConflict}

		mr, err := client.CrebteMergeRequest(ctx, project, CrebteMergeRequestOpts{})
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if wbnt := ErrMergeRequestAlrebdyExists; wbnt != err {
			t.Errorf("unexpected error: hbve %+v; wbnt %+v", err, wbnt)
		}
	})

	t.Run("non-conflict error", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusInternblServerError}

		mr, err := client.CrebteMergeRequest(ctx, project, CrebteMergeRequestOpts{})
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		} else if err == ErrMergeRequestAlrebdyExists {
			t.Errorf("unexpected error vblue: %+v", err)
		}
	})

	t.Run("mblformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not vblid JSON`,
		}

		mr, err := client.CrebteMergeRequest(ctx, project, CrebteMergeRequestOpts{})
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		} else if err == ErrMergeRequestAlrebdyExists {
			t.Errorf("unexpected error vblue: %+v", err)
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"id":"the id cbnnot be b string"}`,
		}

		mr, err := client.CrebteMergeRequest(ctx, project, CrebteMergeRequestOpts{})
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		} else if err == ErrMergeRequestAlrebdyExists {
			t.Errorf("unexpected error vblue: %+v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"iid":42}`,
		}

		mr, err := client.CrebteMergeRequest(ctx, project, CrebteMergeRequestOpts{})
		if mr == nil {
			t.Error("unexpected nil merge request")
		} else if diff := cmp.Diff(mr, &MergeRequest{IID: 42}); diff != "" {
			t.Errorf("unexpected merge request: %s", diff)
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})
}

func TestCrebteMergeRequest_Archived(t *testing.T) {
	ctx := context.Bbckground()
	client := crebteTestClient(t)

	project := &Project{ProjectCommon: ProjectCommon{ID: 37741563}}
	opts := CrebteMergeRequestOpts{
		SourceBrbnch: "brbnch-without-pr",
		TbrgetBrbnch: "mbin",
		Title:        "This MR should never be crebted",
		Description:  "This merge request wbs crebted by b test bgbinst bn brchived repository, bnd should therefore not exist.",
	}
	mr, err := client.CrebteMergeRequest(ctx, project, opts)
	bssert.Nil(t, mr)
	bssert.Error(t, err)
	bssert.True(t, errcode.IsArchived(err))
}

func TestGetMergeRequest(t *testing.T) {
	ctx := context.Bbckground()
	project := &Project{}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		mr, err := client.GetMergeRequest(ctx, project, 1)
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
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

		mr, err := client.GetMergeRequest(ctx, project, 1)
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"id":"the id cbnnot be b string"}`,
		}

		mr, err := client.GetMergeRequest(ctx, project, 1)
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"iid":42}`,
		}

		mr, err := client.GetMergeRequest(ctx, project, 1)
		if mr == nil {
			t.Error("unexpected nil merge request")
		} else if diff := cmp.Diff(mr, &MergeRequest{IID: 42}); diff != "" {
			t.Errorf("unexpected merge request: %s", diff)
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})
}

func TestGetOpenMergeRequestByRefs(t *testing.T) {
	ctx := context.Bbckground()
	project := &Project{}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		mr, err := client.GetOpenMergeRequestByRefs(ctx, project, "source", "tbrget")
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil || err == ErrTooMbnyMergeRequests || err == ErrMergeRequestNotFound {
			t.Errorf("unexpected error: %+v", err)
		}
	})

	t.Run("mblformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not vblid JSON`,
		}

		mr, err := client.GetOpenMergeRequestByRefs(ctx, project, "source", "tbrget")
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil || err == ErrTooMbnyMergeRequests || err == ErrMergeRequestNotFound {
			t.Errorf("unexpected error: %+v", err)
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id":"the id cbnnot be b string"}]`,
		}

		mr, err := client.GetOpenMergeRequestByRefs(ctx, project, "source", "tbrget")
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil || err == ErrTooMbnyMergeRequests || err == ErrMergeRequestNotFound {
			t.Errorf("unexpected error: %+v", err)
		}
	})

	t.Run("zero merge requests", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[]`,
		}

		mr, err := client.GetOpenMergeRequestByRefs(ctx, project, "source", "tbrget")
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err != ErrMergeRequestNotFound {
			t.Errorf("unexpected error: %+v", err)
		}
	})

	t.Run("too mbny merge requests", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"iid":1},{"iid":2}]`,
		}

		mr, err := client.GetOpenMergeRequestByRefs(ctx, project, "source", "tbrget")
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err != ErrTooMbnyMergeRequests {
			t.Errorf("unexpected error: %+v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"iid":42}]`,
		}

		// Since this will invoke GetMergeRequest, we need to mock thbt. (But,
		// on the bright side, thbt bllows us to verify the pbrbmeters it's
		// given.)
		wbnt := &MergeRequest{}
		MockGetMergeRequest = func(mc *Client, mctx context.Context, mproject *Project, miid ID) (*MergeRequest, error) {
			if client != mc {
				t.Errorf("unexpected client: hbve %+v; wbnt %+v", mc, client)
			}
			if ctx != mctx {
				t.Errorf("unexpected context: hbve %+v; wbnt %+v", mctx, ctx)
			}
			if project != mproject {
				t.Errorf("unexpected project: hbve %+v; wbnt %+v", mproject, project)
			}
			if wbnt := ID(42); miid != wbnt {
				t.Errorf("unexpected IID: hbve %d; wbnt %d", miid, wbnt)
			}

			return wbnt, nil
		}
		defer func() { MockGetMergeRequest = nil }()

		mr, err := client.GetOpenMergeRequestByRefs(ctx, project, "source", "tbrget")
		if mr != wbnt {
			t.Errorf("unexpected merge request: hbve %+v; wbnt %+v", mr, wbnt)
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})
}

func TestUpdbteMergeRequest(t *testing.T) {
	ctx := context.Bbckground()
	empty := &MergeRequest{}
	opts := UpdbteMergeRequestOpts{}
	project := &Project{}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		mr, err := client.UpdbteMergeRequest(ctx, project, empty, opts)
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
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

		mr, err := client.UpdbteMergeRequest(ctx, project, empty, opts)
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"id":"the id cbnnot be b string"}`,
		}

		mr, err := client.UpdbteMergeRequest(ctx, project, empty, opts)
		if mr != nil {
			t.Errorf("unexpected non-nil merge request: %+v", mr)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"iid":42}`,
		}

		mr, err := client.UpdbteMergeRequest(ctx, project, empty, opts)
		if mr == nil {
			t.Error("unexpected nil merge request")
		} else if diff := cmp.Diff(mr, &MergeRequest{IID: 42}); diff != "" {
			t.Errorf("unexpected merge request: %s", diff)
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})
}

func TestUpdbteMergeRequest_Archived(t *testing.T) {
	ctx := context.Bbckground()
	client := crebteTestClient(t)

	project := &Project{ProjectCommon: ProjectCommon{ID: 37741563}}
	mr := &MergeRequest{IID: 1}
	opts := UpdbteMergeRequestOpts{
		Title: "This title should never chbnge",
	}
	mr, err := client.UpdbteMergeRequest(ctx, project, mr, opts)
	bssert.Nil(t, mr)
	bssert.Error(t, err)
	bssert.True(t, errcode.IsArchived(err))
}

func TestCrebteMergeRequestNote(t *testing.T) {
	ctx := context.Bbckground()
	empty := &MergeRequest{}
	project := &Project{}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		err := client.CrebteMergeRequestNote(ctx, project, empty, "test-comment")
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("mblformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not vblid JSON`,
		}

		err := client.CrebteMergeRequestNote(ctx, project, empty, "test-comment")
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"id":"the id cbnnot be b string"}`,
		}

		err := client.CrebteMergeRequestNote(ctx, project, empty, "test-comment")
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"body":"test-comment"}`,
		}

		err := client.CrebteMergeRequestNote(ctx, project, empty, "test-comment")
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})
}

func TestMergeMergeRequest(t *testing.T) {
	ctx := context.Bbckground()
	empty := &MergeRequest{}
	project := &Project{}

	t.Run("error stbtus code", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{http.StbtusNotFound}

		_, err := client.MergeMergeRequest(ctx, project, empty, fblse)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("mblformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not vblid JSON`,
		}

		_, err := client.MergeMergeRequest(ctx, project, empty, fblse)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("invblid response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"id":"the id cbnnot be b string"}`,
		}

		_, err := client.MergeMergeRequest(ctx, project, empty, fblse)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("not mergebble", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPEmptyResponse{
			stbtusCode: 405,
		}

		_, err := client.MergeMergeRequest(ctx, project, empty, fblse)
		if err == nil {
			t.Error("unexpected nil error")
		}
		if !errors.Is(err, ErrNotMergebble) {
			t.Errorf("invblid error, wbnt=%v hbve=%v", ErrNotMergebble, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `{"body":"test-merge"}`,
		}

		_, err := client.MergeMergeRequest(ctx, project, empty, fblse)
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})
}
