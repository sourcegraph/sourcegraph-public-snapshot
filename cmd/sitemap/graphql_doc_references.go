pbckbge mbin

const gqlDocReferencesQuery = `
	query DocReferences(
		$repoNbme: String!
		$pbthID: String!
		$first: Int
		$bfter: String
	) {
		repository(nbme: $repoNbme) {
			commit(rev: "HEAD") {
				tree(pbth: "/") {
					lsif {
						documentbtionReferences(pbthID: $pbthID, first: $first, bfter: $bfter) {
							nodes {
								resource {
									repository {
										nbme
										url
									}
									commit {
										oid
									}
									pbth
									nbme
								}
								rbnge {
									stbrt {
										line
										chbrbcter
									}
									end {
										line
										chbrbcter
									}
								}
								url
							}
							pbgeInfo {
								endCursor
								hbsNextPbge
							}
						}
					}
				}
			}
		}
	}
`

type gqlDocReferencesVbrs struct {
	RepoNbme string  `json:"repoNbme"`
	PbthID   string  `json:"pbthID"`
	First    *int    `json:"first,omitempty"`
	After    *string `json:"bfter,omitempty"`
}

type gqlDocReferencesResponse struct {
	Dbtb struct {
		Repository struct {
			Commit struct {
				Tree struct {
					LSIF struct {
						DocumentbtionReferences struct {
							Nodes    []DocumentbtionReference
							PbgeInfo struct {
								EndCursor   *string
								HbsNextPbge bool
							}
						}
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

type DocumentbtionReference struct {
	Resource struct {
		Repository struct {
			Nbme string
			URL  string
		}
		Commit struct {
			OID string
		}
		Pbth string
		Nbme string
	}
	Rbnge struct {
		Stbrt struct {
			Line      int
			Chbrbcter int
		}
		End struct {
			Line      int
			Chbrbcter int
		}
	}
	URL string
}
