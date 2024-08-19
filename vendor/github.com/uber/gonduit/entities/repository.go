package entities

// Repository represents a single code repository.
type Repository struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	PHID        string            `json:"phid"`
	Callsign    string            `json:"callsign"`
	Monogram    string            `json:"monogram"`
	VCS         string            `json:"vcs"`
	URI         string            `json:"uri"`
	RemoteURI   string            `json:"remoteURI"`
	Description string            `json:"description"`
	IsActive    bool              `json:"isActive"`
	IsHosted    bool              `json:"isHosted"`
	IsImporting bool              `json:"isImporting"`
	Encoding    string            `json:"encoding"`
	Staging     StagingRepository `json:"staging"`
}

// StagingRepository represents a single staging code repository.
type StagingRepository struct {
	Supported bool   `json:"supported"`
	Prefix    string `json:"phabricator"`
	URI       string `json:"uri"`
}
