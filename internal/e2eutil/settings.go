package e2eutil

import (
	"github.com/pkg/errors"
)

// OverwriteSettings overwrites settings for given subject ID with contents.
func (c *Client) OverwriteSettings(subjectID, contents string) error {
	const query = `
mutation OverwriteSettings($subject: ID!, $contents: String!) {
	settingsMutation(input: { subject: $subject }) {
		overwriteSettings(contents: $contents) {
			empty {
				alwaysNil
			}
		}
	}
}
`
	variables := map[string]interface{}{
		"subject":  subjectID,
		"contents": contents,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// ViewerSettings returns the latest cascaded settings of authenticated user.
func (c *Client) ViewerSettings() (string, error) {
	const query = `
query ViewerSettings {
	viewerSettings {
		final
	}
}
`
	var resp struct {
		Data struct {
			ViewerSettings struct {
				Final string `json:"final"`
			} `json:"viewerSettings"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}
	return resp.Data.ViewerSettings.Final, nil
}
