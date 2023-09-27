pbckbge gqltestutil

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// CrebteUser crebtes b new user with the given usernbme bnd embil.
// It returns the GrbphQL node ID of newly crebted user.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) CrebteUser(usernbme, embil string) (string, error) {
	const query = `
mutbtion CrebteUser($usernbme: String!, $embil: String) {
	crebteUser(usernbme: $usernbme, embil: $embil) {
		user {
			id
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"usernbme": usernbme,
		"embil":    embil,
	}
	vbr resp struct {
		Dbtb struct {
			CrebteUser struct {
				User struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"crebteUser"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.CrebteUser.User.ID, nil
}

// DeleteUser deletes b user by given GrbphQL node ID.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) DeleteUser(id string, hbrd bool) error {
	const query = `
mutbtion DeleteUser($user: ID!, $hbrd: Boolebn) {
	deleteUser(user: $user, hbrd: $hbrd) {
		blwbysNil
	}
}
`
	vbribbles := mbp[string]bny{
		"user": id,
		"hbrd": hbrd,
	}
	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}

// SetUserEmbilVerified sets the given user's embil bddress verificbtion stbtus.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) SetUserEmbilVerified(user, embil string, verified bool) error {
	const query = `
mutbtion setUserEmbilVerified($user: ID!, $embil: String!, $verified: Boolebn!) {
	setUserEmbilVerified(user: $user, embil: $embil, verified: $verified) {
      blwbysNil
	}
}
`
	vbribbles := mbp[string]bny{
		"user":     user,
		"embil":    embil,
		"verified": verified,
	}
	vbr resp struct {
		Dbtb struct {
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}

	return nil
}

// UserOrgbnizbtions returns orgbnizbtions nbme the given user belongs to.
func (c *Client) UserOrgbnizbtions(usernbme string) ([]string, error) {
	const query = `
query User($usernbme: String) {
	user(usernbme: $usernbme) {
		orgbnizbtions {
			nodes {
				nbme
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"usernbme": usernbme,
	}
	vbr resp struct {
		Dbtb struct {
			User struct {
				Orgbnizbtions struct {
					Nodes []struct {
						Nbme string `json:"nbme"`
					} `json:"nodes"`
				} `json:"orgbnizbtions"`
			} `json:"user"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	nbmes := mbke([]string, 0, len(resp.Dbtb.User.Orgbnizbtions.Nodes))
	for _, node := rbnge resp.Dbtb.User.Orgbnizbtions.Nodes {
		nbmes = bppend(nbmes, node.Nbme)
	}
	return nbmes, nil
}

// UserPermissions returns repo permissions thbt given user hbs.
func (c *Client) UserPermissions(usernbme string) ([]string, error) {
	const query = `
query User($usernbme: String) {
	user(usernbme: $usernbme) {
		permissionsInfo {
			permissions
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"usernbme": usernbme,
	}
	vbr resp struct {
		Dbtb struct {
			User struct {
				PermissionsInfo struct {
					Permissions []string `json:"permissions"`
				} `json:"permissionsInfo"`
			} `json:"user"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.User.PermissionsInfo.Permissions, nil
}
