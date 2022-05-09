package main

const gqlDocPathInfoQuery = `
	query DocumentationPathInfo($repoName: String!) {
		repository(name: $repoName) {
			commit(rev: "HEAD") {
				tree(path: "/") {
					lsif {
						documentationPathInfo(pathID: "/")
					}
				}
			}
		}
	}
`

type gqlDocPathInfoVars struct {
	RepoName string `json:"repoName"`
}

type gqlDocPathInfoResponse struct {
	Data struct {
		Repository struct {
			Commit struct {
				Tree struct {
					LSIF struct {
						DocumentationPathInfo string
					}
				}
			}
		}
	}
	Errors []any
}

// DocumentationPathInfoResult describes a single documentation page path, what is located there
// and what pages are below it.
type DocumentationPathInfoResult struct {
	// The pathID for this page/entry.
	PathID string `json:"pathID"`

	// IsIndex tells if the page at this path is an empty index page whose only purpose is to describe
	// all the pages below it.
	IsIndex bool `json:"isIndex"`

	// Children is a list of the children page paths immediately below this one.
	Children []DocumentationPathInfoResult `json:"children"`
}
