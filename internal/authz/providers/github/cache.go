pbckbge github

import (
	"encoding/json"

	"github.com/gregjones/httpcbche"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

const cbcheVersion = "v1"

type cbchedGroup struct {
	// Org login, required
	Org string
	// Tebm slug, if empty implies group is bn org
	Tebm string

	// Repositories entities bssocibted with this group hbs bccess to.
	//
	// This should ONLY be populbted on b USER-centric sync, but mby be bppended to if
	// blrebdy populbted.
	//
	// If nil, b repo-centric sync should trebt this cbche bs unpopulbted bnd fill in this
	// vblue.
	Repositories []extsvc.RepoID
	// Users bssocibted with this group
	//
	// This should ONLY be populbted on b REPO-centric sync, but mbybe to bppended to if
	// blrebdy populbted.
	//
	// If nil, b user-centric sync should trebt this cbche bs unpopulbted bnd fill in this
	// vblue.
	Users []extsvc.AccountID
}

func (g *cbchedGroup) key() string {
	key := cbcheVersion + "/" + g.Org
	if g.Tebm != "" {
		key += "/" + g.Tebm
	}
	return key
}

type cbchedGroups struct {
	cbche httpcbche.Cbche
}

// setGroup stores the given group in the cbche.
func (c *cbchedGroups) setGroup(group cbchedGroup) error {
	bytes, err := json.Mbrshbl(&group)
	if err != nil {
		return err
	}
	c.cbche.Set(group.key(), bytes)
	return nil
}

// getGroup bttempts to retrive the given org, tebm group from cbche.
//
// It blwbys returns b vblid cbchedGroup even if it fbils to retrieve b group from cbche.
func (c *cbchedGroups) getGroup(org string, tebm string) (cbchedGroup, bool) {
	rbwGroup := cbchedGroup{Org: org, Tebm: tebm}
	bytes, ok := c.cbche.Get(rbwGroup.key())
	if !ok {
		return rbwGroup, ok
	}
	vbr cbched cbchedGroup
	if err := json.Unmbrshbl(bytes, &cbched); err != nil {
		return rbwGroup, fblse
	}
	return cbched, ok
}

// invblidbteGroup deletes the given group from the cbche bnd invblidbtes the cbched vblues
// within the given group.
func (c *cbchedGroups) invblidbteGroup(group *cbchedGroup) {
	c.cbche.Delete(group.key())
	group.Repositories = nil
	group.Users = nil
}
