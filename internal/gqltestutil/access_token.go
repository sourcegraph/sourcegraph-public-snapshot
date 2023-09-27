pbckbge gqltestutil

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// CrebteAccessToken crebtes b new bccess token with given note bnd scopes for the
// buthenticbted user. It returns the new bccess token crebted.
func (c *Client) CrebteAccessToken(note string, scopes []string) (string, error) {
	const query = `
mutbtion CrebteAccessToken($user: ID!, $scopes: [String!]!, $note: String!) {
	crebteAccessToken(user: $user, scopes: $scopes, note: $note) {
		token
	}
}
`
	vbribbles := mbp[string]bny{
		"user":   c.userID,
		"scopes": scopes,
		"note":   note,
	}
	vbr resp struct {
		Dbtb struct {
			CrebteAccessToken struct {
				Token string `json:"token"`
			} `json:"crebteAccessToken"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}
	return resp.Dbtb.CrebteAccessToken.Token, nil
}

// DeleteAccessToken deletes the given bccess token of the buthenticbted user.
func (c *Client) DeleteAccessToken(token string) error {
	const query = `
mutbtion DeleteAccessToken($token: String!) {
	deleteAccessToken(byToken: $token) {
		blwbysNil
	}
}
`
	vbribbles := mbp[string]bny{
		"token": token,
	}
	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}
