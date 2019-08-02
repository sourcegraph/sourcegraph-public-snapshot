package graphqlbackend

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (srs *searchResultsStats) Owners(ctx context.Context) ([]*OwnerStatistics, error) {
	srr, err := srs.getResults(ctx)
	if err != nil {
		return nil, err
	}

	getFiles := func(ctx context.Context, repo gitserver.Repo, commitID api.CommitID, path string) ([]os.FileInfo, error) {
		return git.ReadDir(ctx, repo, commitID, "", true)
	}

	byOwner := map[string]int{}
	recordOwner := func(r *GitCommitResolver, lines int) {
		if author := r.Author(); author != nil {
			byOwner[author.person.email] += lines
		}
	}

	for _, res := range srr.Results() {
		if fileMatch, ok := res.ToFileMatch(); ok {
			if !fileMatch.File().IsDirectory() {
				var lines int
				if len(fileMatch.LineMatches()) > 0 {
					lines = len(fileMatch.LineMatches())
				} else {
					content, err := fileMatch.File().Content(ctx)
					if err != nil {
						return nil, err
					}
					lines = strings.Count(content, "\n")
				}
				recordOwner(fileMatch.File().Commit(), lines)
			}
		} else if repo, ok := res.ToRepository(); ok {
			branchRef, err := repo.DefaultBranch(ctx)
			if err != nil {
				return nil, err
			}
			target, err := branchRef.Target().OID(ctx)
			if err != nil {
				return nil, err
			}
			inv, err := backend.Repos.GetInventory(ctx, repo.repo, api.CommitID(target))
			if err != nil {
				return nil, err
			}
			var sum uint64
			for _, l := range inv.Languages {
				sum += l.TotalBytes / 31 // TODO!(sqs): hack adjust for lines
			}

			commit, err := branchRef.Target().Commit(ctx)
			if err != nil {
				return nil, err
			}
			recordOwner(commit, int(sum))
		} else if commit, ok := res.ToCommitSearchResult(); ok {
			if commit.raw.Diff == nil {
				continue
			}
			fileDiffs, err := diff.ParseMultiFileDiff([]byte(commit.raw.Diff.Raw))
			if err != nil {
				return nil, err
			}
			var lines int64
			for _, fileDiff := range fileDiffs {
				for _, hunk := range fileDiff.Hunks {
					c := int64(hunk.NewLines - hunk.OrigLines)
					if c < 0 {
						c = c * -1
					}
					lines += c
				}
			}
			recordOwner(commit.Commit(), int(lines))
		}
	}

	ownerStats := make([]*OwnerStatistics, 0, len(byOwner))
	for owner, totalBytes := range byOwner {
		ownerStats = append(ownerStats, &OwnerStatistics{owner: owner, totalBytes: totalBytes})
	}
	sort.Slice(ownerStats, func(i, j int) bool {
		return ownerStats[i].totalBytes > ownerStats[j].totalBytes
	})
	return ownerStats, nil
}

type OwnerStatistics struct {
	owner      string
	totalBytes int
}

func (v *OwnerStatistics) Owner() string     { return v.owner }
func (v *OwnerStatistics) TotalBytes() int32 { return int32(v.totalBytes) }
