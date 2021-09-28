package main

import (
	"context"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
)

type QueryResponse struct {
	Data struct {
		Repository struct {
			Commit struct {
				Blob struct {
					LSIF struct {
						Definitions Definitions `json:"definitions"`
						References  References  `json:"references"`
					} `json:"lsif"`
				} `json:"blob"`
			} `json:"commit"`
		} `json:"repository"`
	} `json:"data"`
}

type Definitions struct {
	Nodes []Node `json:"nodes"`
}

type References struct {
	Nodes    []Node   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Node struct {
	Resource `json:"resource"`
	Range    `json:"range"`
}

type Resource struct {
	Path       string     `json:"path"`
	Repository Repository `json:"repository"`
	Commit     Commit     `json:"commit"`
}

type Repository struct {
	Name string `json:"name"`
}

type Commit struct {
	Oid string `json:"oid"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type PageInfo struct {
	EndCursor string `json:"endCursor"`
}

const locationsFragment = `
nodes {
	resource {
		path
		repository {
			name
		}
		commit {
			oid
		}
	}
	range {
		start {
			line
			character
		}
		end {
			line
			character
		}
	}
}

pageInfo {
	endCursor
}
`

const definitionsQuery = `
	query Definitions($repository: String!, $commit: String!, $path: String!, $line: Int!, $character: Int!) {
		repository(name: $repository) {
			commit(rev: $commit) {
				blob(path: $path) {
					lsif {
						definitions(line: $line, character: $character) {
							` + locationsFragment + `
						}
					}
				}
			}
		}
	}
`

// queryDefinitions returns all of the LSIF definitions for the given location.
func queryDefinitions(ctx context.Context, location Location) (locations []Location, err error) {
	variables := map[string]interface{}{
		"repository": location.Repo,
		"commit":     location.Rev,
		"path":       location.Path,
		"line":       location.Line,
		"character":  location.Character,
	}

	var payload QueryResponse
	if err := internal.QueryGraphQL(ctx, "CodeIntelTesterDefinitions", definitionsQuery, variables, &payload); err != nil {
		return nil, err
	}

	for _, node := range payload.Data.Repository.Commit.Blob.LSIF.Definitions.Nodes {
		locations = append(locations, Location{
			Repo:      node.Resource.Repository.Name,
			Rev:       node.Resource.Commit.Oid,
			Path:      node.Resource.Path,
			Line:      node.Range.Start.Line,
			Character: node.Range.Start.Character,
		})
	}

	return locations, nil
}

const referencesQuery = `
	query References($repository: String!, $commit: String!, $path: String!, $line: Int!, $character: Int!, $after: String) {
		repository(name: $repository) {
			commit(rev: $commit) {
				blob(path: $path) {
					lsif {
						references(line: $line, character: $character, after: $after) {
							` + locationsFragment + `
						}
					}
				}
			}
		}
	}
`

// queryReferences returns all of the LSIF references for the given location.
func queryReferences(ctx context.Context, location Location) (locations []Location, err error) {
	endCursor := ""
	for {
		variables := map[string]interface{}{
			"repository": location.Repo,
			"commit":     location.Rev,
			"path":       location.Path,
			"line":       location.Line,
			"character":  location.Character,
		}
		if endCursor != "" {
			variables["after"] = endCursor
		}

		var payload QueryResponse
		if err := internal.QueryGraphQL(ctx, "CodeIntelTesterReferences", referencesQuery, variables, &payload); err != nil {
			return nil, err
		}

		for _, node := range payload.Data.Repository.Commit.Blob.LSIF.References.Nodes {
			locations = append(locations, Location{
				Repo:      node.Resource.Repository.Name,
				Rev:       node.Resource.Commit.Oid,
				Path:      node.Resource.Path,
				Line:      node.Range.Start.Line,
				Character: node.Range.Start.Character,
			})
		}

		if endCursor = payload.Data.Repository.Commit.Blob.LSIF.References.PageInfo.EndCursor; endCursor == "" {
			break
		}
	}

	return locations, nil
}

// sortLocations sorts a slice of Locations by repo, rev, path, line, then character.
func sortLocations(locations []Location) {
	sort.Slice(locations, func(i, j int) bool {
		return compareLocations(locations[i], locations[j]) < 0
	})
}

// Compare returns an integer comparing two locations. The result will be 0 if a == b,
// -1 if a < b, and +1 if a > b.
func compareLocations(a, b Location) int {
	fieldComparison := []int{
		strings.Compare(a.Repo, b.Repo),
		strings.Compare(a.Rev, b.Rev),
		strings.Compare(a.Path, b.Path),
		a.Line - b.Line,
		a.Character - b.Character,
	}

	for _, cmp := range fieldComparison {
		if cmp != 0 {
			return cmp
		}
	}
	return 0
}
