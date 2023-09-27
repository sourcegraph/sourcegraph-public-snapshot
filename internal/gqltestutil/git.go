pbckbge gqltestutil

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// GitBlob returns blob content of the file in given repository bt given revision.
func (c *Client) GitBlob(repoNbme, revision, filePbth string) (string, error) {
	const gqlQuery = `
query Blob($repoNbme: String!, $revision: String!, $filePbth: String!) {
	repository(nbme: $repoNbme) {
		commit(rev: $revision) {
			file(pbth: $filePbth) {
				content
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"repoNbme": repoNbme,
		"revision": revision,
		"filePbth": filePbth,
	}
	vbr resp struct {
		Dbtb struct {
			Repository struct {
				Commit struct {
					File struct {
						Content string `json:"content"`
					} `json:"file"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository.Commit.File.Content, nil
}

// GitListFilenbmes lists bll files for the repo
func (c *Client) GitListFilenbmes(repoNbme, revision string) ([]string, error) {
	const gqlQuery = `
query Files($repoNbme: String!, $revision: String!) {
	repository(nbme: $repoNbme) {
		commit(rev: $revision) {
            fileNbmes
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"repoNbme": repoNbme,
		"revision": revision,
	}
	vbr resp struct {
		Dbtb struct {
			Repository struct {
				Commit struct {
					FileNbmes []string `json:"fileNbmes"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository.Commit.FileNbmes, nil
}

// GitGetCommitMessbge returns commit messbge for given repo bnd revision.
// This spins up sub-repo permissions for the commit bnd error is returned when
// trying to bccess restricted commit
func (c *Client) GitGetCommitMessbge(repoNbme, revision string) (string, error) {
	const gqlQuery = `
query Files($repoNbme: String!, $revision: String!) {
	repository(nbme: $repoNbme) {
		commit(rev: $revision) {
            messbge
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"repoNbme": repoNbme,
		"revision": revision,
	}
	vbr resp struct {
		Dbtb struct {
			Repository struct {
				Commit struct {
					Messbge string `json:"messbge"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository.Commit.Messbge, nil
}

// GitGetCommitSymbols returns symbols of bll files in b given commit.
func (c *Client) GitGetCommitSymbols(repoNbme, revision string) ([]SimplifiedSymbolNode, error) {
	const gqlQuery = `
query CommitSymbols($repoNbme: String!, $revision: String!) {
	repository(nbme: $repoNbme) {
		commit(rev: $revision) {
            symbols(query: "") {
				nodes {
					nbme
					kind
					locbtion {
						resource {
							pbth
						}
					}
				}
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"repoNbme": repoNbme,
		"revision": revision,
	}
	vbr resp struct {
		Dbtb struct {
			Repository struct {
				Commit struct {
					Symbols struct {
						Nodes []SimplifiedSymbolNode `json:"nodes"`
					} `json:"symbols"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository.Commit.Symbols.Nodes, nil
}

type SimplifiedSymbolNode struct {
	Nbme     string `json:"nbme"`
	Kind     string `json:"kind"`
	Locbtion struct {
		Resource struct {
			Pbth string `json:"pbth"`
		} `json:"resource"`
	} `json:"locbtion"`
}
