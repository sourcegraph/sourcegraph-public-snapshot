package handlerutil

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/cvg"
)

// GetCoverage retrieves the coverage data for the given repository
func GetCoverage(cl *sourcegraph.Client, ctx context.Context, repo string) (map[string]*cvg.Coverage, *sourcegraph.SrclibDataVersion, error) {
	var repoRevSpec sourcegraph.RepoRevSpec
	repoRevSpec.URI = repo

	var rootEntrySpec sourcegraph.TreeEntrySpec
	rootEntrySpec.RepoRev.URI = repo
	srclibDataVer, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &rootEntrySpec)
	if err != nil {
		return nil, nil, err
	}
	repoRevSpec.CommitID = srclibDataVer.CommitID

	cstatus, err := cl.RepoStatuses.GetCombined(ctx, &repoRevSpec)
	if err != nil {
		return nil, nil, err
	}

	var c map[string]*cvg.Coverage
	for _, status := range cstatus.Statuses {
		if status.Context == "coverage" {
			err := json.Unmarshal([]byte(status.Description), &c)
			if err != nil {
				return nil, nil, err
			}
			return c, srclibDataVer, nil
		}
	}
	return nil, nil, fmt.Errorf("coverage data not found for repo %s", repo)
}
