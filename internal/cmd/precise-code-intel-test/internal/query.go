package internal

import (
	"sort"
	"strings"
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

// QueryDefinitions returns all of the LSIF definitions for the given location.
func QueryDefinitions(baseURL, token string, location Location) (locations []Location, err error) {
	var query = `
		query Definitions($repository: String!, $commit: String!, $path: String!, $line: Int!, $character: Int!) {
			repository(name: $repository) {
				commit(rev: $commit) {
					blob(path: $path) {
						lsif {
							definitions(line: $line, character: $character) {
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
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"repository": location.Repo,
		"commit":     location.Rev,
		"path":       location.Path,
		"line":       location.Line,
		"character":  location.Character,
	}

	payload := QueryResponse{}
	if err := graphQL(baseURL, token, query, variables, &payload); err != nil {
		return nil, err
	}

	lsifPayload := payload.Data.Repository.Commit.Blob.LSIF

	for _, node := range lsifPayload.Definitions.Nodes {
		locations = append(locations, Location{
			Repo:      node.Resource.Repository.Name,
			Rev:       node.Resource.Commit.Oid,
			Path:      node.Resource.Path,
			Line:      node.Range.Start.Line,
			Character: node.Range.Start.Character,
		})
	}

	sortLocations(locations)
	return locations, nil
}

// QueryReferences returns all of the LSIF references for the given location.
func QueryReferences(baseURL, token string, location Location) (locations []Location, err error) {
	var query = `
		query References($repository: String!, $commit: String!, $path: String!, $line: Int!, $character: Int!, $after: String) {
			repository(name: $repository) {
				commit(rev: $commit) {
					blob(path: $path) {
						lsif {
							references(line: $line, character: $character, after: $after) {
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
							}
						}
					}
				}
			}
		}
	`

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

		payload := QueryResponse{}
		if err := graphQL(baseURL, token, query, variables, &payload); err != nil {
			return nil, err
		}

		lsifPayload := payload.Data.Repository.Commit.Blob.LSIF

		for _, node := range lsifPayload.References.Nodes {
			locations = append(locations, Location{
				Repo:      node.Resource.Repository.Name,
				Rev:       node.Resource.Commit.Oid,
				Path:      node.Resource.Path,
				Line:      node.Range.Start.Line,
				Character: node.Range.Start.Character,
			})
		}

		if endCursor = lsifPayload.References.PageInfo.EndCursor; endCursor == "" {
			break
		}
	}

	sortLocations(locations)
	return locations, nil
}

// sortLocations sorts a slice of Locations by repo, rev, path, line, then character.
func sortLocations(locations []Location) {
	sort.Slice(locations, func(i, j int) bool {
		cmps := []int{
			strings.Compare(locations[i].Repo, locations[j].Repo),
			strings.Compare(locations[i].Rev, locations[j].Rev),
			strings.Compare(locations[i].Path, locations[j].Path),
			locations[i].Line - locations[j].Line,
			locations[i].Character - locations[j].Character,
		}

		for _, cmp := range cmps {
			if cmp < 0 {
				return true
			}
			if cmp > 0 {
				return false
			}
		}
		return false
	})
}
