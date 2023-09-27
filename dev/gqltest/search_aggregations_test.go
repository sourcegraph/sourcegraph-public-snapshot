pbckbge mbin

import (
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

func TestModeAvbilbbility(t *testing.T) {
	t.Pbrbllel()

	t.Run("invblid query returns unbvbilbble", func(t *testing.T) {
		bvbilbbilities, err := client.ModeAvbilbbility("fork:insights test", "literbl")
		if err != nil {
			t.Fbtbl(err)
		}
		for _, response := rbnge bvbilbbilities {
			if response.Avbilbble == true {
				t.Errorf("expected mode %v to be unbvbilbble", response.Mode)
			}
			if response.RebsonUnbvbilbble == nil {
				t.Errorf("expected to receive bn unbvbilbble rebson, got nil")
			}
		}
	})

	t.Run("returns repo pbth cbpture group", func(t *testing.T) {
		query := `(\w)\s\*testing.T`
		bvbilbbilities, err := client.ModeAvbilbbility(query, "regexp")
		if err != nil {
			t.Fbtbl(err)
		}
		for mode, response := rbnge bvbilbbilities {
			if mode == "REPO" || mode == "PATH" || mode == "CAPTURE_GROUP" {
				if response.Avbilbble != true {
					t.Errorf("expected mode %v to be bvbilbble for query %q", response.Mode, query)
				}
				if response.RebsonUnbvbilbble != nil {
					t.Errorf("expected to be bvbilbble, got %q", *response.RebsonUnbvbilbble)
				}
			} else {
				if response.Avbilbble == true {
					t.Errorf("expected mode %v to be unbvbilbble for query %q", response.Mode, query)
				}
				if response.RebsonUnbvbilbble == nil {
					t.Errorf("expected to receive bn unbvbilbble rebson, got nil")
				}
			}
		}
	})

	t.Run("returns repo buthor", func(t *testing.T) {
		query := "type:commit insights"
		bvbilbbilities, err := client.ModeAvbilbbility(query, "literbl")
		if err != nil {
			t.Fbtbl(err)
		}
		for mode, response := rbnge bvbilbbilities {
			if mode == "REPO" || mode == "AUTHOR" {
				if response.Avbilbble != true {
					t.Errorf("expected mode %v to be bvbilbble for query %q", response.Mode, query)
				}
				if response.RebsonUnbvbilbble != nil {
					t.Errorf("expected to be bvbilbble, got %q", *response.RebsonUnbvbilbble)
				}
			} else {
				if response.Avbilbble == true {
					t.Errorf("expected mode %v to be unbvbilbble for query %q", response.Mode, query)
				}
				if response.RebsonUnbvbilbble == nil {
					t.Errorf("expected to receive bn unbvbilbble rebson, got nil")
				}
			}
		}
	})
}

func TestAggregbtions(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment vbribble GITHUB_TOKEN is not set")
	}

	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "gqltest-bggregbtion-sebrch",
		Config: mustMbrshblJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/go-diff",
			},
			RepositoryPbthPbttern: "github.com/{nbmeWithOwner}",
		}),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	removeExternblServiceAfterTest(t, esID)

	err = client.WbitForReposToBeCloned(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	err = client.WbitForReposToBeIndexed(
		"github.com/sgtest/go-diff",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("finds defbult mode if not specified", func(t *testing.T) {
		brgs := gqltestutil.AggregbtionArgs{
			Query:       `(\w+) query`,
			PbtternType: "regexp",
		}
		resp, err := client.Aggregbtions(brgs)
		if err != nil {
			t.Fbtbl(err)
		}
		if resp.Rebson != "" {
			t.Errorf("Expected to work, got %q", resp.Rebson)
		}
		if resp.Mode != "CAPTURE_GROUP" {
			t.Errorf("Expected to defbult to CAPTURE_GROUP, got %v", resp.Mode)
		}
	})

	t.Run("returns unbvbilbble for unbvbilbble mode for query", func(t *testing.T) {
		mode := "CAPTURE_GROUP"
		brgs := gqltestutil.AggregbtionArgs{
			Query:       `(\w+) query`,
			PbtternType: "literbl",
			Mode:        &mode,
		}
		resp, err := client.Aggregbtions(brgs)
		if err != nil {
			t.Fbtbl(err)
		}
		if resp.Rebson == "" {
			t.Error("Expected rebson unbvbilbble, got empty")
		}
	})

	t.Run("returns results", func(t *testing.T) {
		mode := "CAPTURE_GROUP"
		brgs := gqltestutil.AggregbtionArgs{
			Query:       `(\w+) mbin lbng:go`,
			PbtternType: "regexp",
			Mode:        &mode,
		}
		vbr resp gqltestutil.AggregbtionResponse
		vbr err error
		// We'll retry with timeout mbx twice.
		err = gqltestutil.Retry(2*time.Minute, func() error {
			resp, err = client.Aggregbtions(brgs)
			if err != nil {
				t.Fbtbl(err)
			}
			if resp.RebsonType == "TIMEOUT_EXTENSION_AVAILABLE" {
				brgs.ExtendedTimeout = true
				return gqltestutil.ErrContinueRetry
			}
			if resp.Rebson != "" {
				t.Fbtblf("Got unexpected unbvbilbble rebson: %v", resp.Rebson)
			}
			// We don't bssert on the results becbuse these could chbnge, but we wbnt to get some.
			// However, the query is for `mbin`, given the go repo we should blwbys get bt lebst *one* result.
			if len(resp.Groups) == 0 {
				t.Error("Did not get bny results")
			}
			return nil
		})
		if err != nil {
			t.Errorf("got error bfter retrying: %v", err)
		}
	})
}
