package main

import (
	"context"
	"sort"
	"strings"
)

const preciseIndexesQuery = `
	query PreciseIndexes {
		preciseIndexes(states: [COMPLETED], first: 1000) {
			nodes {
				inputRoot
				projectRoot {
					repository {
						name
					}
					commit {
						oid
					}
				}
			}
		}
	}
`

type CommitAndRoot struct {
	Commit string
	Root   string
}

func queryPreciseIndexes(ctx context.Context) (_ map[string][]CommitAndRoot, err error) {
	var payload struct {
		Data struct {
			PreciseIndexes struct {
				Nodes []struct {
					InputRoot   string `json:"inputRoot"`
					ProjectRoot struct {
						Repository struct {
							Name string `json:"name"`
						} `json:"repository"`
						Commit struct {
							OID string `json:"oid"`
						} `json:"commit"`
					} `json:"projectRoot"`
				} `json:"nodes"`
			} `json:"preciseIndexes"`
		} `json:"data"`
	}
	if err := queryGraphQL(ctx, "CodeIntelQA_Query_PreciseIndexes", preciseIndexesQuery, map[string]any{}, &payload); err != nil {
		return nil, err
	}

	rootsByCommitsByRepo := map[string][]CommitAndRoot{}
	for _, node := range payload.Data.PreciseIndexes.Nodes {
		root := node.InputRoot
		projectRoot := node.ProjectRoot
		name := projectRoot.Repository.Name
		commit := projectRoot.Commit.OID
		rootsByCommitsByRepo[name] = append(rootsByCommitsByRepo[name], CommitAndRoot{commit, root})
	}

	return rootsByCommitsByRepo, nil
}

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

func queryDefinitions(ctx context.Context, location Location) (locations []Location, err error) {
	variables := map[string]any{
		"repository": location.Repo,
		"commit":     location.Rev,
		"path":       location.Path,
		"line":       location.Line,
		"character":  location.Character,
	}

	var payload QueryResponse
	if err := queryGraphQL(ctx, "CodeIntelQA_Query_Definitions", definitionsQuery, variables, &payload); err != nil {
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

func queryReferences(ctx context.Context, location Location) (locations []Location, err error) {
	endCursor := ""
	for {
		variables := map[string]any{
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
		if err := queryGraphQL(ctx, "CodeIntelQA_Query_References", referencesQuery, variables, &payload); err != nil {
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

const implementationsQuery = `
	query Implementations($repository: String!, $commit: String!, $path: String!, $line: Int!, $character: Int!, $after: String) {
		repository(name: $repository) {
			commit(rev: $commit) {
				blob(path: $path) {
					lsif {
						implementations(line: $line, character: $character, after: $after) {
							` + locationsFragment + `
						}
					}
				}
			}
		}
	}
`

func queryImplementations(ctx context.Context, location Location) (locations []Location, err error) {
	endCursor := ""
	for {
		variables := map[string]any{
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
		if err := queryGraphQL(ctx, "CodeIntelQA_Query_Implementations", implementationsQuery, variables, &payload); err != nil {
			return nil, err
		}

		for _, node := range payload.Data.Repository.Commit.Blob.LSIF.Implementations.Nodes {
			locations = append(locations, Location{
				Repo:      node.Resource.Repository.Name,
				Rev:       node.Resource.Commit.Oid,
				Path:      node.Resource.Path,
				Line:      node.Range.Start.Line,
				Character: node.Range.Start.Character,
			})
		}

		if endCursor = payload.Data.Repository.Commit.Blob.LSIF.Implementations.PageInfo.EndCursor; endCursor == "" {
			break
		}
	}

	return locations, nil
}

const prototypesQuery = `
	query Prototypes($repository: String!, $commit: String!, $path: String!, $line: Int!, $character: Int!, $after: String) {
		repository(name: $repository) {
			commit(rev: $commit) {
				blob(path: $path) {
					lsif {
						prototypes(line: $line, character: $character, after: $after) {
							` + locationsFragment + `
						}
					}
				}
			}
		}
	}
`

func queryPrototypes(ctx context.Context, location Location) (locations []Location, err error) {
	endCursor := ""
	for {
		variables := map[string]any{
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
		if err := queryGraphQL(ctx, "CodeIntelQA_Query_Prototypes", prototypesQuery, variables, &payload); err != nil {
			return nil, err
		}

		for _, node := range payload.Data.Repository.Commit.Blob.LSIF.Prototypes.Nodes {
			locations = append(locations, Location{
				Repo:      node.Resource.Repository.Name,
				Rev:       node.Resource.Commit.Oid,
				Path:      node.Resource.Path,
				Line:      node.Range.Start.Line,
				Character: node.Range.Start.Character,
			})
		}

		if endCursor = payload.Data.Repository.Commit.Blob.LSIF.Prototypes.PageInfo.EndCursor; endCursor == "" {
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
