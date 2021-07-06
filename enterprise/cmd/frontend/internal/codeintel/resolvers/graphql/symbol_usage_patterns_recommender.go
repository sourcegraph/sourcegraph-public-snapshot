package graphql

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/xeonx/timeago"
)

func sortAndRankExampleLocations(ctx context.Context, locationResolver *CachedLocationResolver, symbol resolvers.AdjustedSymbol, exampleLocations []symbolUsagePatternExampleLocation) ([]symbolUsagePatternExampleLocation, error) {
	infos := make([]*exampleLocationInfo, len(exampleLocations))
	for i, exampleLocation := range exampleLocations {
		info, err := lookupExampleLocationInfo(ctx, exampleLocation.location)
		if err != nil {
			return nil, err
		}
		infos[i] = info
	}

	const (
		currentUserEmail      = "quinn@slack.org"  // TODO(sqs): un-hardcode
		teamMemberEmailDomain = "@sourcegraph.com" // TODO(sqs): un-hardcode
		maintainerEmailDomain = "@hashicorp.com"   // TODO(sqs): un-hardcode; for go-multierror
	)

	// Sort most recent first.
	sort.Slice(infos, func(i, j int) bool { return infos[i].author.Date.After(infos[j].author.Date) })

	var (
		seenCurrentUserMostRecent = false
		seenTeamMemberMostRecent  = false
		seenMaintainerMostRecent  = false
	)
	for i, info := range infos {
		var parts []string

		timeAgo := timeago.NoMax(timeago.English).Format(info.author.Date)
		const (
			recent = 45 * 24 * time.Hour  // 45 days
			old    = 365 * 24 * time.Hour // 1 year
		)
		if i == 0 {
			parts = append(parts, fmt.Sprintf("Most recent (%s)", timeAgo))
		} else if time.Since(info.author.Date) < recent {
			parts = append(parts, fmt.Sprintf("Recent (%s)", timeAgo))
		} else if time.Since(info.author.Date) < old {
			parts = append(parts, timeAgo)
		} else {
			parts = append(parts, fmt.Sprintf("Old (%s)", timeAgo))
		}

		isByCurrentUser := info.author.Email == currentUserEmail
		if isByCurrentUser {
			if !seenCurrentUserMostRecent && i != 0 {
				parts = append(parts, "most recent usage by you")
			} else {
				parts = append(parts, "by you")
			}
			seenCurrentUserMostRecent = true
		}

		isByTeamMember := strings.HasSuffix(info.author.Email, teamMemberEmailDomain)
		if isByTeamMember && !isByCurrentUser {
			if !seenTeamMemberMostRecent && i != 0 {
				parts = append(parts, fmt.Sprintf("most recent usage by your team (%s)", info.author.Email))
			} else {
				parts = append(parts, fmt.Sprintf("by a team member (%s)", info.author.Email))
			}
			seenTeamMemberMostRecent = true
		}

		isByMaintainer := strings.HasSuffix(info.author.Email, maintainerEmailDomain)
		if !isByCurrentUser && !isByTeamMember && isByMaintainer {
			if !seenMaintainerMostRecent && i != 0 {
				parts = append(parts, fmt.Sprintf("most recent usage by maintainer (%s)", info.author.Email))
			} else {
				parts = append(parts, fmt.Sprintf("by maintainer (%s)", info.author.Email))
			}
			seenMaintainerMostRecent = true
		}

		if !isByCurrentUser && !isByTeamMember && !isByMaintainer {
			parts = append(parts, "by a community member")
		}

		symbolDefinitionRepo := symbol.AdjustedLocation.Dump.RepositoryID
		if isInExternalRepo := info.location.Dump.RepositoryID != symbolDefinitionRepo; isInExternalRepo {
			parts = append(parts, "in a separate project")
		} else if isInTestCode := strings.Contains(info.location.Path, "test"); isInTestCode {
			parts = append(parts, "in test code")
		}

		info.description = strings.Join(parts, ", ")
	}

	// Skip an example if the previous one was from the same author.
	keep := infos[:0]
	for i, info := range infos {
		if i >= 1 && infos[i-1].author.Email == info.author.Email {
			continue
		}
		keep = append(keep, info)
	}
	infos = keep

	// Re-extract the exampleLocations.
	exampleLocations = exampleLocations[:0]
	for _, info := range infos {
		exampleLocations = append(exampleLocations, symbolUsagePatternExampleLocation{
			symbol:      symbol,
			description: info.description,
			location:    info.location,
		})
	}

	return exampleLocations, nil
}

type exampleLocationInfo struct {
	description string
	location    resolvers.AdjustedLocation
	author      git.Signature
}

func lookupExampleLocationInfo(ctx context.Context, loc resolvers.AdjustedLocation) (*exampleLocationInfo, error) {
	info := exampleLocationInfo{location: loc}

	// TODO(sqs): only takes 1st author (which is usually ok because most calls only span 1 line)
	authors, err := getLocationBlameAuthors(ctx, loc)
	if err != nil {
		return nil, err
	}
	if len(authors) != 0 {
		info.author = authors[0]
	}

	return &info, nil
}
