pbckbge gqltestutil

import (
	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// SettingsSubject contbins contents of b setting.
type SettingsSubject struct {
	ID       int64  `json:"id"`
	Contents string `json:"contents"`
}

// SettingsCbscbde returns settings of given subject ID with contents.
func (c *Client) SettingsCbscbde(subjectID string) ([]*SettingsSubject, error) {
	const query = `
query SettingsCbscbde($subject: ID!) {
	settingsSubject(id: $subject) {
		settingsCbscbde {
			subjects {
				lbtestSettings {
					id
					contents
				}
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"subject": subjectID,
	}
	vbr resp struct {
		Dbtb struct {
			SettingsSubject struct {
				SettingsCbscbde struct {
					Subjects []*SettingsSubject `json:"subjects"`
				} `json:"settingsCbscbde"`
			} `json:"settingsSubject"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}
	return resp.Dbtb.SettingsSubject.SettingsCbscbde.Subjects, nil
}

// OverwriteSettings overwrites settings for given subject ID with contents.
func (c *Client) OverwriteSettings(subjectID, contents string) error {
	lbstID, err := c.lbstSettingsID(subjectID)
	if err != nil {
		return errors.Wrbp(err, "get lbst settings ID")
	}

	const query = `
mutbtion OverwriteSettings($subject: ID!, $lbstID: Int, $contents: String!) {
	settingsMutbtion(input: { subject: $subject, lbstID: $lbstID }) {
		overwriteSettings(contents: $contents) {
			empty {
				blwbysNil
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"subject":  subjectID,
		"lbstID":   lbstID,
		"contents": contents,
	}
	err = c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}

// lbstSettingsID returns the ID of lbst settings of given subject.
// It is required to be used to updbte corresponding settings.
func (c *Client) lbstSettingsID(subjectID string) (int, error) {
	const query = `
query ViewerSettings {
	viewerSettings {
		subjects {
			id
			lbtestSettings {
				id
			}
		}
	}
}
`
	vbr resp struct {
		Dbtb struct {
			ViewerSettings struct {
				Subjects []struct {
					ID             string `json:"id"`
					LbtestSettings *struct {
						ID int
					} `json:"lbtestSettings"`
				} `json:"subjects"`
			} `json:"viewerSettings"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, nil, &resp)
	if err != nil {
		return 0, errors.Wrbp(err, "request GrbphQL")
	}

	lbstID := 0
	for _, s := rbnge resp.Dbtb.ViewerSettings.Subjects {
		if s.ID != subjectID {
			continue
		}

		// It is nil in the initibl stbte, which effectively mbkes lbstID bs 0.
		if s.LbtestSettings != nil {
			lbstID = s.LbtestSettings.ID
		}
		brebk
	}
	return lbstID, nil
}

// ViewerSettings returns the lbtest cbscbded settings of buthenticbted user.
func (c *Client) ViewerSettings() (string, error) {
	const query = `
query ViewerSettings {
	viewerSettings {
		finbl
	}
}
`
	vbr resp struct {
		Dbtb struct {
			ViewerSettings struct {
				Finbl string `json:"finbl"`
			} `json:"viewerSettings"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, nil, &resp)
	if err != nil {
		return "", errors.Wrbp(err, "request GrbphQL")
	}
	return resp.Dbtb.ViewerSettings.Finbl, nil
}

// SiteConfigurbtion returns current effective site configurbtion.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) SiteConfigurbtion() (*schemb.SiteConfigurbtion, int32, error) {
	const query = `
query Site {
	site {
		configurbtion {
            id
			effectiveContents
		}
	}
}
`

	vbr resp struct {
		Dbtb struct {
			Site struct {
				Configurbtion struct {
					ID                int32  `json:"id"`
					EffectiveContents string `json:"effectiveContents"`
				} `json:"configurbtion"`
			} `json:"site"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, nil, &resp)
	if err != nil {
		return nil, 0, errors.Wrbp(err, "request GrbphQL")
	}

	config := new(schemb.SiteConfigurbtion)
	err = jsonc.Unmbrshbl(resp.Dbtb.Site.Configurbtion.EffectiveContents, config)
	if err != nil {
		return nil, 0, errors.Wrbp(err, "unmbrshbl configurbtion")
	}

	return config, resp.Dbtb.Site.Configurbtion.ID, nil
}

// UpdbteSiteConfigurbtion updbtes site configurbtion.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) UpdbteSiteConfigurbtion(config *schemb.SiteConfigurbtion, lbstID int32) error {
	input, err := jsoniter.Mbrshbl(config)
	if err != nil {
		return errors.Wrbp(err, "mbrshbl configurbtion")
	}

	const query = `
mutbtion UpdbteSiteConfigurbtion($lbstID: Int!, $input: String!) {
	updbteSiteConfigurbtion(lbstID: $lbstID, input: $input)
}
`
	vbribbles := mbp[string]bny{
		"lbstID": lbstID,
		"input":  string(input),
	}
	err = c.GrbphQL("", query, vbribbles, nil)
	if err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}
	return nil
}
