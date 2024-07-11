package gqltestutil

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SettingsSubject contains contents of a setting.
type SettingsSubject struct {
	ID       int64  `json:"id"`
	Contents string `json:"contents"`
}

// SettingsCascade returns settings of given subject ID with contents.
func (c *Client) SettingsCascade(subjectID string) ([]*SettingsSubject, error) {
	const query = `
query SettingsCascade($subject: ID!) {
	settingsSubject(id: $subject) {
		settingsCascade {
			subjects {
				latestSettings {
					id
					contents
				}
			}
		}
	}
}
`
	variables := map[string]any{
		"subject": subjectID,
	}
	var resp struct {
		Data struct {
			SettingsSubject struct {
				SettingsCascade struct {
					Subjects []*SettingsSubject `json:"subjects"`
				} `json:"settingsCascade"`
			} `json:"settingsSubject"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}
	return resp.Data.SettingsSubject.SettingsCascade.Subjects, nil
}

// OverwriteSettings overwrites settings for given subject ID with contents.
func (c *Client) OverwriteSettings(subjectID, contents string) error {
	lastID, err := c.lastSettingsID(subjectID)
	if err != nil {
		return errors.Wrap(err, "get last settings ID")
	}

	const query = `
mutation OverwriteSettings($subject: ID!, $lastID: Int, $contents: String!) {
	settingsMutation(input: { subject: $subject, lastID: $lastID }) {
		overwriteSettings(contents: $contents) {
			empty {
				alwaysNil
			}
		}
	}
}
`
	variables := map[string]any{
		"subject":  subjectID,
		"lastID":   lastID,
		"contents": contents,
	}
	err = c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// lastSettingsID returns the ID of last settings of given subject.
// It is required to be used to update corresponding settings.
func (c *Client) lastSettingsID(subjectID string) (int, error) {
	const query = `
query ViewerSettings {
	viewerSettings {
		subjects {
			id
			latestSettings {
				id
			}
		}
	}
}
`
	var resp struct {
		Data struct {
			ViewerSettings struct {
				Subjects []struct {
					ID             string `json:"id"`
					LatestSettings *struct {
						ID int
					} `json:"latestSettings"`
				} `json:"subjects"`
			} `json:"viewerSettings"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return 0, errors.Wrap(err, "request GraphQL")
	}

	lastID := 0
	for _, s := range resp.Data.ViewerSettings.Subjects {
		if s.ID != subjectID {
			continue
		}

		// It is nil in the initial state, which effectively makes lastID as 0.
		if s.LatestSettings != nil {
			lastID = s.LatestSettings.ID
		}
		break
	}
	return lastID, nil
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
func (c *Client) SiteConfiguration() (*schema.SiteConfiguration, int32, error) {
	const query = `
query Site {
	site {
		configuration {
            id
			effectiveContents
		}
	}
}
`

	var resp struct {
		Data struct {
			Site struct {
				Configuration struct {
					ID                int32  `json:"id"`
					EffectiveContents string `json:"effectiveContents"`
				} `json:"configuration"`
			} `json:"site"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return nil, 0, errors.Wrap(err, "request GraphQL")
	}

	config := new(schema.SiteConfiguration)
	err = jsonc.Unmarshal(resp.Data.Site.Configuration.EffectiveContents, config)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unmarshal configuration")
	}

	return config, resp.Data.Site.Configuration.ID, nil
}

// UpdateSiteConfiguration updates site configuration.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) UpdateSiteConfiguration(config *schema.SiteConfiguration, lastID int32) error {
	input, err := jsoniter.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshal configuration")
	}

	const query = `
mutation UpdateSiteConfiguration($lastID: Int!, $input: String!) {
	updateSiteConfiguration(lastID: $lastID, input: $input)
}
`
	variables := map[string]any{
		"lastID": lastID,
		"input":  string(input),
	}
	err = c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// ModifySiteConfiguration allows temporarily modifying the site configuration.
// The returned closure can be used to reset the site configuration to its original state.
//
// The returned function may be nil if there was an error.
func (c *Client) ModifySiteConfiguration(modify func(*schema.SiteConfiguration)) (reset func() error, _ error) {
	oldConfig, lastID, err := c.SiteConfiguration()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current site config")
	}
	newConfig, err := deepCopyViaJSON(*oldConfig)
	if err != nil {
		return nil, errors.Wrap(err, "deepcopy failed")
	}
	modify(&newConfig)
	if err = c.UpdateSiteConfiguration(&newConfig, lastID); err != nil {
		return nil, errors.Wrap(err, "update site config failed")
	}
	return func() error {
		_, currentID, err := c.SiteConfiguration()
		if err != nil {
			return err
		}
		return c.UpdateSiteConfiguration(oldConfig, currentID)
	}, nil
}

func deepCopyViaJSON[T any](t T) (T, error) {
	bytes, err := json.Marshal(t)
	var zero T
	if err != nil {
		return zero, errors.Wrap(err, "marshal old value")
	}
	var tcopy T
	if err = json.Unmarshal(bytes, &tcopy); err != nil {
		return zero, errors.Wrap(err, "unmarshal old value")
	}
	return tcopy, nil
}
