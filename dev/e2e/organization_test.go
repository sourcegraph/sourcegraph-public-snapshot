// +build e2e

package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestOrganization(t *testing.T) {
	orgID, err := client.CreateOrganization("e2e-test-org", "e2e-test-org")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteOrganization(orgID)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("settings cascade", func(t *testing.T) {
		err := client.OverwriteSettings(orgID, `{"quicklinks":[{"name":"Test quicklink","url":"http://test-quicklink.local"}]}`)
		if err != nil {
			t.Fatal(err)
		}

		{
			contents, err := client.ViewerSettings()
			if err != nil {
				t.Fatal(err)
			}

			var got struct {
				QuickLinks []schema.QuickLink `json:"quicklinks"`
			}
			err = jsoniter.UnmarshalFromString(contents, &got)
			if err != nil {
				t.Fatal(err)
			}

			wantQuickLinks := []schema.QuickLink{
				{
					Name: "Test quicklink",
					Url:  "http://test-quicklink.local",
				},
			}
			if diff := cmp.Diff(wantQuickLinks, got.QuickLinks); diff != "" {
				t.Fatalf("QuickLinks mismatch (-want +got):\n%s", diff)
			}
		}

		// Remove authenticate user (e2e-admin) from organization (e2e-test-org) should
		// no longer get cascaded settings from this organization.
		err = client.RemoveUserFromOrganization(client.AuthenticatedUserID(), orgID)
		if err != nil {
			t.Fatal(err)
		}

		{
			contents, err := client.ViewerSettings()
			if err != nil {
				t.Fatal(err)
			}

			var got struct {
				QuickLinks []schema.QuickLink `json:"quicklinks"`
			}
			err = jsoniter.UnmarshalFromString(contents, &got)
			if err != nil {
				t.Fatal(err)
			}

			var wantQuickLinks []schema.QuickLink
			if diff := cmp.Diff(wantQuickLinks, got.QuickLinks); diff != "" {
				t.Fatalf("QuickLinks mismatch (-want +got):\n%s", diff)
			}
		}
	})
}
