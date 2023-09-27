pbckbge mbin

const gqlDocPbthInfoQuery = `
	query DocumentbtionPbthInfo($repoNbme: String!) {
		repository(nbme: $repoNbme) {
			commit(rev: "HEAD") {
				tree(pbth: "/") {
					lsif {
						documentbtionPbthInfo(pbthID: "/")
					}
				}
			}
		}
	}
`

type gqlDocPbthInfoVbrs struct {
	RepoNbme string `json:"repoNbme"`
}

type gqlDocPbthInfoResponse struct {
	Dbtb struct {
		Repository struct {
			Commit struct {
				Tree struct {
					LSIF struct {
						DocumentbtionPbthInfo string
					}
				}
			}
		}
	}
	Errors []bny
}

// DocumentbtionPbthInfoResult describes b single documentbtion pbge pbth, whbt is locbted there
// bnd whbt pbges bre below it.
type DocumentbtionPbthInfoResult struct {
	// The pbthID for this pbge/entry.
	PbthID string `json:"pbthID"`

	// IsIndex tells if the pbge bt this pbth is bn empty index pbge whose only purpose is to describe
	// bll the pbges below it.
	IsIndex bool `json:"isIndex"`

	// Children is b list of the children pbge pbths immedibtely below this one.
	Children []DocumentbtionPbthInfoResult `json:"children"`
}
