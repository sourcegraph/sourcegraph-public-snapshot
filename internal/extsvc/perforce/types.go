package perforce

import (
	"time"
)

// TODO: this was just blindly copied from the ADO implementation in order to make things "work" initially
// need to probably switch most of this out for Changelist vernacular

type ChangelistStatus string

type CreatorInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	URL         string `json:"url"`
	ImageURL    string `json:"imageUrl"`
}

type Changelist struct {
	Depot        Depot            `json:"depot"`
	ID           int              `json:"changelistId"`
	Status       ChangelistStatus `json:"status"` // pending, submitted, shelved
	CreationDate time.Time        `json:"creationDate"`
	CreatedBy    CreatorInfo      `json:"createdBy"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	URL          string           `json:"url"`
}

type Repository struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CloneURL   string `json:"remoteURL"`
	IsDisabled bool   `json:"isDisabled"`
}
