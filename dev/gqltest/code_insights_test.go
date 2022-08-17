package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestCreateDashboard(t *testing.T) {
	t.Run("can create an insights dashboard", func(t *testing.T) {
		title := "Dashboard Title 1"
		result, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{
			Title:       title,
			GlobalGrant: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		want := gqltestutil.DashboardResponse{
			Title: title,
			Grants: gqltestutil.GrantsResponse{
				Users:         []string{},
				Organizations: []string{},
				Global:        true,
			},
		}
		err = client.DeleteDashboard(result.Id)
		if err != nil {
			t.Fatal(err)
		}

		// Ignore the newly created id
		result.Id = ""
		if diff := cmp.Diff(want, result); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("errors on a grant that the user does not have permission to give", func(t *testing.T) {
		title := "Dashboard Title 1"
		_, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{
			Title:     title,
			UserGrant: string(graphqlbackend.MarshalUserID(9999)),
		})
		if !strings.Contains(err.Error(), "user does not have permission") {
			t.Fatal("Should have thrown an error")
		}
	})
	t.Run("errors on zero grants", func(t *testing.T) {
		title := "Dashboard Title 1"
		_, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{
			Title: title,
		})
		if !strings.Contains(err.Error(), "dashboard must be created with at least one grant") {
			t.Fatal("Should have thrown an error")
		}
	})
}

func TestGetDashboards(t *testing.T) {
	titles := []string{"Title 1", "Title 2", "Title 3", "Title 4", "Title 5"}
	ids := []string{}
	for _, title := range titles {
		response, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{Title: title, GlobalGrant: true})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, response.Id)
	}

	defer func() {
		for _, id := range ids {
			err := client.DeleteDashboard(id)
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	t.Run("can get all dashboards", func(t *testing.T) {
		resultTitles := getTitles(t, gqltestutil.GetDashboardArgs{})
		if diff := cmp.Diff(titles, resultTitles); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("can get the first 2 dashboards", func(t *testing.T) {
		first := 2
		args := gqltestutil.GetDashboardArgs{First: &first}
		resultTitles := getTitles(t, args)
		if diff := cmp.Diff(titles[0:2], resultTitles); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("can get a dashboard by id", func(t *testing.T) {
		args := gqltestutil.GetDashboardArgs{Id: &ids[3]}
		resultTitles := getTitles(t, args)
		if diff := cmp.Diff([]string{titles[3]}, resultTitles); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("can get all dashboards after a cursor", func(t *testing.T) {
		args := gqltestutil.GetDashboardArgs{After: &ids[1]}
		resultTitles := getTitles(t, args)
		if diff := cmp.Diff(titles[2:5], resultTitles); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("can get a single dashboard after a cursor", func(t *testing.T) {
		first := 1
		args := gqltestutil.GetDashboardArgs{First: &first, After: &ids[2]}
		resultTitles := getTitles(t, args)
		if diff := cmp.Diff([]string{titles[3]}, resultTitles); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestUpdateDashboard(t *testing.T) {
	dashboard, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{Title: "Title", GlobalGrant: true})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := client.DeleteDashboard(dashboard.Id)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("can update a dashboard", func(t *testing.T) {
		updatedTitle := "Updated title"
		userGrant := client.AuthenticatedUserID()
		updatedDashboard, err := client.UpdateDashboard(dashboard.Id, gqltestutil.DashboardInputArgs{Title: updatedTitle, UserGrant: userGrant})
		if err != nil {
			t.Fatal(err)
		}

		wantDashboard := gqltestutil.DashboardResponse{
			Id:    dashboard.Id,
			Title: updatedTitle,
			Grants: gqltestutil.GrantsResponse{
				Users:         []string{userGrant},
				Organizations: []string{},
				Global:        false,
			},
		}
		if diff := cmp.Diff(wantDashboard, updatedDashboard); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestDeleteDashboard(t *testing.T) {
	t.Run("can delete an insights dashboard", func(t *testing.T) {
		dashboard, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{Title: "Should be deleted", GlobalGrant: true})
		if err != nil {
			t.Fatal(err)
		}
		err = client.DeleteDashboard(dashboard.Id)
		if err != nil {
			t.Fatal(err)
		}
		responseDashboard, err := client.GetDashboards(gqltestutil.GetDashboardArgs{Id: &dashboard.Id})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(0, len(responseDashboard)); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("cannot delete an insights dashboard without permission", func(t *testing.T) {
		dashboard, err := client.CreateDashboard(gqltestutil.DashboardInputArgs{Title: "Should be deleted", GlobalGrant: true})
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.UpdateDashboard(dashboard.Id, gqltestutil.DashboardInputArgs{})
		if err == nil || !strings.Contains(err.Error(), "got nil for non-null") {
			t.Fatal(err)
		}
		err = client.DeleteDashboard(dashboard.Id)
		if !strings.Contains(err.Error(), "dashboard not found") {
			t.Fatal("Should have thrown an error")
		}
	})
	t.Run("returns an error when a dashboard does not exist", func(t *testing.T) {
		err := client.DeleteDashboard("ZGFzaGJvYXJkOnsiSWRUeXBlIjoiY3VzdG9tIiwiQXJnIjo5OTk5OX0=")
		if !strings.Contains(err.Error(), "dashboard not found") {
			t.Fatal("Should have thrown an error")
		}
	})
}

func getTitles(t *testing.T, args gqltestutil.GetDashboardArgs) []string {
	dashboards, err := client.GetDashboards(args)
	if err != nil {
		t.Fatal(err)
	}

	retry := false
	resultTitles := []string{}
	for _, dashboard := range dashboards {
		// Sometimes the LAM dashboard will be present since the service is running. We do not want to count it in the test,
		// so we hide the LAM dashboard and query the dashboards again.
		if dashboard.Title == "Limited Access Mode Dashboard" {
			_, err = client.UpdateDashboard(dashboard.Id, gqltestutil.DashboardInputArgs{Title: "Limited Access Mode Dashboard"})
			if err == nil || !strings.Contains(err.Error(), "got nil for non-null") {
				t.Fatal(err)
			}
			retry = true
		}
		resultTitles = append(resultTitles, dashboard.Title)
	}

	if retry {
		return getTitles(t, args)
	}
	return resultTitles
}
