package main

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestModeAvailability(t *testing.T) {
	t.Skip("for now")
	t.Parallel()

	t.Run("invalid query returns unavailable", func(t *testing.T) {
		availabilities, err := client.ModeAvailability("fork:insights test", "literal")
		if err != nil {
			t.Fatal(err)
		}
		for _, response := range availabilities {
			if response.Available == true {
				t.Errorf("expected mode %v to be unavailable", response.Mode)
			}
			if response.ReasonUnavailable == nil {
				t.Errorf("expected to receive an unavailable reason, got nil")
			}
		}
	})

	t.Run("returns repo path capture group", func(t *testing.T) {
		query := `(\w)\s\*testing.T`
		availabilities, err := client.ModeAvailability(query, "regexp")
		if err != nil {
			t.Fatal(err)
		}
		for mode, response := range availabilities {
			if mode == "REPO" || mode == "PATH" || mode == "CAPTURE_GROUP" {
				if response.Available != true {
					t.Errorf("expected mode %v to be available for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable != nil {
					t.Errorf("expected to be available, got %q", *response.ReasonUnavailable)
				}
			} else {
				if response.Available == true {
					t.Errorf("expected mode %v to be unavailable for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable == nil {
					t.Errorf("expected to receive an unavailable reason, got nil")
				}
			}
		}
	})

	t.Run("returns repo author", func(t *testing.T) {
		query := "type:commit insights"
		availabilities, err := client.ModeAvailability(query, "literal")
		if err != nil {
			t.Fatal(err)
		}
		for mode, response := range availabilities {
			if mode == "REPO" || mode == "AUTHOR" {
				if response.Available != true {
					t.Errorf("expected mode %v to be available for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable != nil {
					t.Errorf("expected to be available, got %q", *response.ReasonUnavailable)
				}
			} else {
				if response.Available == true {
					t.Errorf("expected mode %v to be unavailable for query %q", response.Mode, query)
				}
				if response.ReasonUnavailable == nil {
					t.Errorf("expected to receive an unavailable reason, got nil")
				}
			}
		}
	})
}

func TestAggregations(t *testing.T) {
	t.Skip("for now")
	if len(*githubToken) == 0 {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	esID, err := client.AddExternalService(gqltestutil.AddExternalServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplayName: "gqltest-aggregation-search",
		Config: mustMarshalJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPathPattern string   `json:"repositoryPathPattern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPathPattern: "github.com/{nameWithOwner}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	removeExternalServiceAfterTest(t, esID)

	err = client.WaitForReposToBeCloned(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fatal(err)
	}

	err = client.WaitForReposToBeIndexed(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("finds default mode if not specified", func(t *testing.T) {
		args := gqltestutil.AggregationArgs{
			Query:       `(\w+) query`,
			PatternType: "regexp",
		}
		resp, err := client.Aggregations(args)
		if err != nil {
			t.Fatal(err)
		}
		if resp.Reason != "" {
			t.Errorf("Expected to work, got %q", resp.Reason)
		}
		if resp.Mode != "CAPTURE_GROUP" {
			t.Errorf("Expected to default to CAPTURE_GROUP, got %v", resp.Mode)
		}
	})

	t.Run("returns unavailable for unavailable mode for query", func(t *testing.T) {
		mode := "CAPTURE_GROUP"
		args := gqltestutil.AggregationArgs{
			Query:       `(\w+) query`,
			PatternType: "literal",
			Mode:        &mode,
		}
		resp, err := client.Aggregations(args)
		if err != nil {
			t.Fatal(err)
		}
		if resp.Reason == "" {
			t.Error("Expected reason unavailable, got empty")
		}
	})

	t.Run("returns results", func(t *testing.T) {
		mode := "CAPTURE_GROUP"
		args := gqltestutil.AggregationArgs{
			Query:       `(\w+) main lang:go`,
			PatternType: "regexp",
			Mode:        &mode,
		}
		var resp gqltestutil.AggregationResponse
		var err error
		// We'll retry with timeout max twice.
		err = gqltestutil.Retry(2*time.Minute, func() error {
			resp, err = client.Aggregations(args)
			if err != nil {
				t.Fatal(err)
			}
			if resp.ReasonType == "TIMEOUT_EXTENSION_AVAILABLE" {
				args.ExtendedTimeout = true
				return gqltestutil.ErrContinueRetry
			}
			if resp.Reason != "" {
				t.Fatalf("Got unexpected unavailable reason: %v", resp.Reason)
			}
			// We don't assert on the results because these could change, but we want to get some.
			// However, the query is for `main`, given the go repo we should always get at least *one* result.
			if len(resp.Groups) == 0 {
				t.Error("Did not get any results")
			}
			return nil
		})
		if err != nil {
			t.Errorf("got error after retrying: %v", err)
		}
	})
}
