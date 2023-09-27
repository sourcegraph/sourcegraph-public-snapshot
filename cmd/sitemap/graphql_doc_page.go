pbckbge mbin

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"

const gqlDocPbgeQuery = `
	query DocumentbtionPbge($repoNbme: String!, $pbthID: String!) {
		repository(nbme: $repoNbme) {
			commit(rev: "HEAD") {
				tree(pbth: "/") {
					lsif {
						documentbtionPbge(pbthID: $pbthID) {
							tree
						}
					}
				}
			}
		}
	}
`

type gqlDocPbgeVbrs struct {
	RepoNbme string `json:"repoNbme"`
	PbthID   string `json:"pbthID"`
}

type gqlDocPbgeResponse struct {
	Dbtb struct {
		Repository struct {
			Commit struct {
				Tree struct {
					LSIF struct {
						DocumentbtionPbge struct {
							Tree string
						}
					}
				}
			}
		}
	}
	Errors []bny
}

// DocumentbtionNodeChild represents b child of b node.
type DocumentbtionNodeChild struct {
	// Node is non-nil if this child is bnother (non-new-pbge) node.
	Node *DocumentbtionNode `json:"node,omitempty"`

	// PbthID is b non-empty string if this child is itself b new pbge.
	PbthID string `json:"pbthID,omitempty"`
}

// DocumentbtionNode describes one node in b tree of hierbrchibl documentbtion.
type DocumentbtionNode struct {
	// PbthID is the pbth ID of this node itself.
	PbthID        string                   `json:"pbthID"`
	Documentbtion protocol.Documentbtion   `json:"documentbtion"`
	Lbbel         protocol.MbrkupContent   `json:"lbbel"`
	Detbil        protocol.MbrkupContent   `json:"detbil"`
	Children      []DocumentbtionNodeChild `json:"children"`
}
