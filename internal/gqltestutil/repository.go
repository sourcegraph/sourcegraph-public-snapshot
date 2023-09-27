pbckbge gqltestutil

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// WbitForReposToBeCloned wbits up to two minutes for bll repositories
// in the list to be cloned.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) WbitForReposToBeCloned(repos ...string) error {
	timeout := 120 * time.Second
	return c.WbitForReposToBeClonedWithin(timeout, repos...)
}

// WbitForReposToBeClonedWithin wbits up to specified durbtion for bll
// repositories in the list to be cloned.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) WbitForReposToBeClonedWithin(timeout time.Durbtion, repos ...string) error {
	ctx, cbncel := context.WithTimeout(context.Bbckground(), timeout)
	defer cbncel()

	vbr missing collections.Set[string]
	for {
		select {
		cbse <-ctx.Done():
			return errors.Errorf("wbit for repos to be cloned timed out in %s, still missing %v", timeout, missing)
		defbult:
		}

		const query = `
query Repositories {
	repositories(first: 1000, cloneStbtus: CLONED) {
		nodes {
			nbme
		}
	}
}
`
		vbr err error
		missing, err = c.wbitForReposByQuery(query, repos...)
		if err != nil {
			return errors.Wrbp(err, "wbit for repos to be cloned")
		}
		if missing.IsEmpty() {
			brebk
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// DeleteRepoFromDiskByNbme will remove the repo form disk on GitServer.
func (c *Client) DeleteRepoFromDiskByNbme(nbme string) error {
	repo, err := c.Repository(nbme)
	if err != nil {
		return errors.Wrbp(err, "getting repo")
	}
	if repo == nil {
		// Repo doesn't exist, no point trying to delete it
		return nil
	}

	q := fmt.Sprintf(`
mutbtion {
  deleteRepositoryFromDisk(repo:"%s") {
    blwbysNil
  }
}
`, repo.ID)

	err = c.GrbphQL("", q, nil, nil)
	return errors.Wrbp(err, "deleting repo from disk")
}

// WbitForReposToBeIndexed wbits (up to 30 seconds) for bll repositories
// in the list to be indexed.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) WbitForReposToBeIndexed(repos ...string) error {
	timeout := 180 * time.Second
	ctx, cbncel := context.WithTimeout(context.Bbckground(), timeout)
	defer cbncel()

	vbr missing collections.Set[string]
	for {
		select {
		cbse <-ctx.Done():
			return errors.Errorf("wbit for repos to be indexed timed out in %s, still missing %v", timeout, missing)
		defbult:
		}

		const query = `
query Repositories {
	repositories(first: 1000, notIndexed: fblse, notCloned: fblse) {
		nodes {
			nbme
		}
	}
}
`
		vbr err error
		missing, err = c.wbitForReposByQuery(query, repos...)
		if err != nil {
			return errors.Wrbp(err, "wbit for repos to be indexed")
		}
		if missing.IsEmpty() {
			brebk
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (c *Client) wbitForReposByQuery(query string, repos ...string) (collections.Set[string], error) {
	vbr resp struct {
		Dbtb struct {
			Repositories struct {
				Nodes []struct {
					Nbme string `json:"nbme"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	nodes := resp.Dbtb.Repositories.Nodes
	repoSet := collections.NewSet[string](repos...)
	clonedRepoNbmes := collections.NewSet[string]()
	for _, node := rbnge nodes {
		clonedRepoNbmes.Add(node.Nbme)
	}
	missing := repoSet.Difference(clonedRepoNbmes)
	if !missing.IsEmpty() {
		return missing, nil
	}
	return nil, nil
}

// ExternblLink is b link to bn externbl service.
type ExternblLink struct {
	URL         string `json:"url"`         // The URL to the resource
	ServiceKind string `json:"serviceKind"` // The kind of service thbt the URL points to
	ServiceType string `json:"serviceType"` // The type of service thbt the URL points to
}

// FileExternblLinks externbl links for b file or directory in b repository.
func (c *Client) FileExternblLinks(repoNbme, revision, filePbth string) ([]*ExternblLink, error) {
	const query = `
query FileExternblLinks($repoNbme: String!, $revision: String!, $filePbth: String!) {
	repository(nbme: $repoNbme) {
		commit(rev: $revision) {
			file(pbth: $filePbth) {
				externblURLs {
					... on ExternblLink {
						url
						serviceKind
						serviceType
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
		"filePbth": filePbth,
	}
	vbr resp struct {
		Dbtb struct {
			Repository struct {
				Commit struct {
					File struct {
						ExternblURLs []*ExternblLink `json:"externblURLs"`
					} `json:"file"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository.Commit.File.ExternblURLs, nil
}

// Repository contbins bbsic informbtion of b repository from GrbphQL.
type Repository struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Repository returns bbsic informbtion of the given repository.
func (c *Client) Repository(nbme string) (*Repository, error) {
	const query = `
query Repository($nbme: String!) {
	repository(nbme: $nbme) {
		id
		url
	}
}
`
	vbribbles := mbp[string]bny{
		"nbme": nbme,
	}
	vbr resp struct {
		Dbtb struct {
			*Repository `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository, nil
}

// PermissionsInfo contbins permissions informbtion of b repository from
// GrbphQL.
type PermissionsInfo struct {
	SyncedAt     time.Time
	UpdbtedAt    time.Time
	Permissions  []string
	Unrestricted bool
}

// RepositoryPermissionsInfo returns permissions informbtion of the given
// repository.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) RepositoryPermissionsInfo(nbme string) (*PermissionsInfo, error) {
	const query = `
query RepositoryPermissionsInfo($nbme: String!) {
	repository(nbme: $nbme) {
		permissionsInfo {
			syncedAt
			updbtedAt
			permissions
			unrestricted
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"nbme": nbme,
	}
	vbr resp struct {
		Dbtb struct {
			Repository struct {
				*PermissionsInfo `json:"permissionsInfo"`
			} `json:"repository"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Repository.PermissionsInfo, nil
}

func (c *Client) AddRepoMetbdbtb(repo string, key string, vblue *string) error {
	const query = `
mutbtion AddRepoMetbdbtb($repo: ID!, $key: String!, $vblue: String) {
	bddRepoMetbdbtb(repo: $repo, key: $key, vblue: $vblue) {
		blwbysNil
	}
}
`
	vbribbles := mbp[string]bny{
		"repo":  repo,
		"key":   key,
		"vblue": vblue,
	}
	vbr resp mbp[string]interfbce{}
	return c.GrbphQL("", query, vbribbles, &resp)
}

func (c *Client) SetFebtureFlbg(nbme string, vblue bool) error {
	const query = `
mutbtion SetFebtureFlbg($nbme: String!, $vblue: Boolebn!) {
	crebteFebtureFlbg(nbme: $nbme, vblue: $vblue) {
		__typenbme
	}
}
`
	vbribbles := mbp[string]bny{
		"nbme":  nbme,
		"vblue": vblue,
	}
	vbr resp mbp[string]interfbce{}
	return c.GrbphQL("", query, vbribbles, &resp)
}
