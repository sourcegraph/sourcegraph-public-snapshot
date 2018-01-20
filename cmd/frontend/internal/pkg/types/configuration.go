package types

import (
	"fmt"
	"time"
)

// A ConfigurationSubject is something that can have settings. A subject with no
// fields set represents the global site settings subject.
type ConfigurationSubject struct {
	Site *string // the site's ID
	Org  *int32  // the org's ID
	User *int32  // the user's ID
}

func (s ConfigurationSubject) String() string {
	switch {
	case s.Site != nil:
		return fmt.Sprintf("site %q", *s.Site)
	case s.Org != nil:
		return fmt.Sprintf("org %d", *s.Org)
	case s.User != nil:
		return fmt.Sprintf("user %d", *s.User)
	default:
		return "unknown configuration subject"
	}
}

// Settings contains configuration settings for a subject.
type Settings struct {
	ID           int32
	Subject      ConfigurationSubject
	AuthorUserID int32
	Contents     string
	CreatedAt    time.Time
}

type SiteConfig struct {
	SiteID           string
	Email            string
	TelemetryEnabled bool
	UpdatedAt        string
}
