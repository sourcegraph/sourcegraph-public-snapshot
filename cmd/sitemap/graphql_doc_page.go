package main

import "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"

const gqlDocPageQuery = `
	query DocumentationPage($repoName: String!, $pathID: String!) {
		repository(name: $repoName) {
			commit(rev: "HEAD") {
				tree(path: "/") {
					lsif {
						documentationPage(pathID: $pathID) {
							tree
						}
					}
				}
			}
		}
	}
`

type gqlDocPageVars struct {
	RepoName string `json:"repoName"`
	PathID   string `json:"pathID"`
}

type gqlDocPageResponse struct {
	Data struct {
		Repository struct {
			Commit struct {
				Tree struct {
					LSIF struct {
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

// DocumentationNodeChild represents a child of a node.
type DocumentationNodeChild struct {
	// Node is non-nil if this child is another (non-new-page) node.
	Node *DocumentationNode `json:"node,omitempty"`

	// PathID is a non-empty string if this child is itself a new page.
	PathID string `json:"pathID,omitempty"`
}

// DocumentationNode describes one node in a tree of hierarchial documentation.
type DocumentationNode struct {
	// PathID is the path ID of this node itself.
	PathID        string                   `json:"pathID"`
	Documentation protocol.Documentation   `json:"documentation"`
	Label         protocol.MarkupContent   `json:"label"`
	Detail        protocol.MarkupContent   `json:"detail"`
	Children      []DocumentationNodeChild `json:"children"`
}
