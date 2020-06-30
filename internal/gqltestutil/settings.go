package gqltestutil

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
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

// SiteConfiguration returns current effective site configuration.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) SiteConfiguration() (*schema.SiteConfiguration, error) {
	const query = `
query Site {
	site {
		configuration {
			effectiveContents
		}
	}
}
`

	var resp struct {
		Data struct {
			Site struct {
				Configuration struct {
					EffectiveContents string `json:"effectiveContents"`
				} `json:"configuration"`
			} `json:"site"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	config := new(schema.SiteConfiguration)
	err = jsonc.Unmarshal(resp.Data.Site.Configuration.EffectiveContents, config)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal configuration")
	}
	return config, nil
}

// UpdateSiteConfiguration updates site configuration.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) UpdateSiteConfiguration(config *schema.SiteConfiguration) error {
	input, err := jsoniter.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshal configuration")
	}

	const query = `
mutation UpdateSiteConfiguration($input: String!) {
	updateSiteConfiguration(lastID: 0, input: $input)
}
`
	variables := map[string]interface{}{
		"input": string(input),
	}
	err = c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}
