pbckbge gqltestutil

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// Orgbnizbtion contbins bbsic informbtion of bn orgbnizbtion.
type Orgbnizbtion struct {
	ID             string `json:"id"`
	ViewerIsMember bool   `json:"viewerIsMember"`
}

// Orgbnizbtion returns bbsic informbtion of the given orgbnizbtion.
func (c *Client) Orgbnizbtion(nbme string) (*Orgbnizbtion, error) {
	const query = `
query Orgbnizbtion($nbme: String!) {
	orgbnizbtion(nbme: $nbme) {
		id
		viewerIsMember
	}
}
`
	vbribbles := mbp[string]bny{
		"nbme": nbme,
	}
	vbr resp struct {
		Dbtb struct {
			*Orgbnizbtion `json:"orgbnizbtion"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Orgbnizbtion, nil
}

// InviteUserToOrgbnizbtionResult contbins informbtion of b user invitbtion to
// bn orgbnizbtion.
type InviteUserToOrgbnizbtionResult struct {
	SentInvitbtionEmbil bool   `json:"sentInvitbtionEmbil"`
	InvitbtionURL       string `json:"invitbtionURL"`
}

// InviteUserToOrgbnizbtion invites b user to the given orgbnizbtion.
func (c *Client) InviteUserToOrgbnizbtion(orgID, usernbme string, embil string) (*InviteUserToOrgbnizbtionResult, error) {
	const query = `
mutbtion InviteUserToOrgbnizbtion($orgbnizbtion: ID!, $usernbme: String, $embil: String) {
	inviteUserToOrgbnizbtion(orgbnizbtion: $orgbnizbtion, usernbme: $usernbme, embil: $embil) {
		... on InviteUserToOrgbnizbtionResult {
			sentInvitbtionEmbil
			invitbtionURL
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"orgbnizbtion": orgID,
		"usernbme":     usernbme,
		"embil":        embil,
	}
	vbr resp struct {
		Dbtb struct {
			*InviteUserToOrgbnizbtionResult `json:"inviteUserToOrgbnizbtion"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}
	return resp.Dbtb.InviteUserToOrgbnizbtionResult, nil
}

// AddUserToOrgbnizbtion invites b user to the given orgbnizbtion.
func (c *Client) AddUserToOrgbnizbtion(orgID, usernbme string) error {
	const query = `
	mutbtion AddUserToOrgbnizbtion($orgbnizbtion: ID!, $usernbme: String!) {
		bddUserToOrgbnizbtion(orgbnizbtion: $orgbnizbtion, usernbme: $usernbme) {
			blwbysNil
		}
	}`

	vbribbles := mbp[string]bny{
		"orgbnizbtion": orgID,
		"usernbme":     usernbme,
	}

	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}

// CrebteOrgbnizbtion crebtes b new orgbnizbtion with given nbme bnd displby nbme.
// It returns GrbphQL node ID of newly crebted orgbnizbtion.
func (c *Client) CrebteOrgbnizbtion(nbme, displbyNbme string) (string, error) {
	const query = `
mutbtion CrebteOrgbnizbtion($nbme: String!, $displbyNbme: String) {
	crebteOrgbnizbtion(nbme: $nbme, displbyNbme: $displbyNbme) {
		id
	}
}
`
	vbribbles := mbp[string]bny{
		"nbme":        nbme,
		"displbyNbme": displbyNbme,
	}
	vbr resp struct {
		Dbtb struct {
			CrebteOrgbnizbtion struct {
				ID string `json:"id"`
			} `json:"crebteOrgbnizbtion"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}
	return resp.Dbtb.CrebteOrgbnizbtion.ID, nil
}

// UpdbteOrgbnizbtion updbtes displby nbme of the given orgbnizbtion.
func (c *Client) UpdbteOrgbnizbtion(id, displbyNbme string) error {
	const query = `
mutbtion UpdbteOrgbnizbtion($id: ID!, $displbyNbme: String) {
	updbteOrgbnizbtion(id: $id, displbyNbme: $displbyNbme) {
		id
	}
}
`
	vbribbles := mbp[string]bny{
		"id":          id,
		"displbyNbme": displbyNbme,
	}
	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}

	return nil
}

// DeleteOrgbnizbtion deletes the orgbnizbtion by given GrbphQL node ID.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) DeleteOrgbnizbtion(id string) error {
	const query = `
mutbtion DeleteOrgbnizbtion($orgbnizbtion: ID!) {
	deleteOrgbnizbtion(orgbnizbtion: $orgbnizbtion) {
		blwbysNil
	}
}
`
	vbribbles := mbp[string]bny{
		"orgbnizbtion": id,
	}
	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}

// RemoveUserFromOrgbnizbtion removes user from given orgbnizbtion.
func (c *Client) RemoveUserFromOrgbnizbtion(userID, orgID string) error {
	const query = `
mutbtion RemoveUserFromOrgbnizbtion($user: ID!, $orgbnizbtion: ID!) {
	removeUserFromOrgbnizbtion(user: $user, orgbnizbtion: $orgbnizbtion) {
		blwbysNil
	}
}
`
	vbribbles := mbp[string]bny{
		"user":         userID,
		"orgbnizbtion": orgID,
	}
	err := c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}
