pbckbge gitlbb

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestListGroups(t *testing.T) {
	ctx := context.Bbckground()
	mockedGroups := []*Group{
		{
			ID:       1,
			FullPbth: "group1",
		},
	}

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `[{"id": 1,"full_pbth": "group1"}]`,
		}

		groupsResponse, _, err := client.ListGroups(ctx, 1)
		if groupsResponse == nil {
			t.Error("unexpected nil response")
		}

		if diff := cmp.Diff(groupsResponse, mockedGroups); diff != "" {
			t.Errorf("unexpected response: %s", diff)
		}

		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})

	t.Run("mblformed response", func(t *testing.T) {
		client := newTestClient(t)
		client.httpClient = &mockHTTPResponseBody{
			responseBody: `this is not vblid JSON`,
		}

		groupsResponse, _, err := client.ListGroups(ctx, 1)
		if groupsResponse != nil {
			t.Error("unexpected non-nil response")
		}

		if err == nil {
			t.Error("unexpected nil error")
		}
	})
}
