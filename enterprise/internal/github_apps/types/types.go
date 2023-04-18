package types

import "time"

// GitHubApp represents a GitHub App.
type GitHubApp struct {
	ID            int
	AppID         int
	Name          string
	Slug          string
	BaseURL       string
	ClientID      string
	ClientSecret  string
	PrivateKey    string
	EncryptionKey string
	Logo          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
