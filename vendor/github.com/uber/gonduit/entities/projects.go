package entities

import "github.com/uber/gonduit/util"

// Project represents a single Phabricator Project.
type Project struct {
	ID               string             `json:"id"`
	PHID             string             `json:"phid"`
	Name             string             `json:"name"`
	ProfileImagePHID string             `json:"profileImagePHID"`
	Icon             string             `json:"icon"`
	Color            string             `json:"color"`
	Members          []string           `json:"members"`
	Slugs            []string           `json:"slugs"`
	DateCreated      util.UnixTimestamp `json:"dateCreated"`
	DateModified     util.UnixTimestamp `json:"dateModified"`
}
