package main

const gqlDocReferencesQuery = `
	query DocReferences(
		$repoName: String!
		$pathID: String!
		$first: Int
		$after: String
	) {
		repository(name: $repoName) {
			commit(rev: "HEAD") {
				tree(path: "/") {
					lsif {
						documentationReferences(pathID: $pathID, first: $first, after: $after) {
							nodes {
								resource {
									repository {
										name
										url
									}
									commit {
										oid
									}
									path
									name
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
								url
							}
							pageInfo {
								endCursor
								hasNextPage
							}
						}
					}
				}
			}
		}
	}
`

type gqlDocReferencesVars struct {
	RepoName string  `json:"repoName"`
	PathID   string  `json:"pathID"`
	First    *int    `json:"first,omitempty"`
	After    *string `json:"after,omitempty"`
}

type gqlDocReferencesResponse struct {
	Data struct {
		Repository struct {
			Commit struct {
				Tree struct {
					LSIF struct {
						DocumentationReferences struct {
							Nodes    []DocumentationReference
							PageInfo struct {
								EndCursor   *string
								HasNextPage bool
							}
						}
						DocumentationPage struct {
							Tree string
						}
					}
				}
			}
		}
	}
	Errors []any
}

type DocumentationReference struct {
	Resource struct {
		Repository struct {
			Name string
			URL  string
		}
		Commit struct {
			OID string
		}
		Path string
		Name string
	}
	Range struct {
		Start struct {
			Line      int
			Character int
		}
		End struct {
			Line      int
			Character int
		}
	}
	URL string
}
