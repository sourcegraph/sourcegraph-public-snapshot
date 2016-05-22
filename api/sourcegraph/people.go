package sourcegraph

import "strings"

// ShortName returns the person's Login if nonempty and otherwise
// returns the portion of Email before the '@'.
func (p *Person) ShortName() string {
	if p.Login != "" {
		return p.Login
	}
	at := strings.Index(p.Email, "@")
	if at == -1 {
		return "(anonymous)"
	}
	return p.Email[:at]
}

// Transient is true if this person was constructed on the fly and is
// not persisted or resolved to a Sourcegraph/GitHub/etc. user.
func (p *Person) Transient() bool { return p.UID == 0 }

// HasProfile is true if the person has a profile page on
// Sourcegraph. Transient users currently do not have profile pages.
func (p *Person) HasProfile() bool { return !p.Transient() }

// AvatarURLOfSize returns the URL to an avatar for the user with the
// given width (in pixels).
func (p *Person) AvatarURLOfSize(width int) string {
	return avatarURLOfSize(p.AvatarURL, width)
}
