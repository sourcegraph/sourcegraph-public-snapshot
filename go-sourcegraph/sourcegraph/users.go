package sourcegraph

import "fmt"

func (u *User) Spec() UserSpec {
	return UserSpec{Login: u.Login, UID: u.UID}
}

// AvatarURLOfSize returns the URL to an avatar for the user with the
// given width (in pixels).
func (u *User) AvatarURLOfSize(width int) string {
	return avatarURLOfSize(u.AvatarURL, width)
}

func avatarURLOfSize(avatarURL string, width int) string {
	// Default avatar.
	if avatarURL == "" {
		avatarURL = "https://secure.gravatar.com/avatar?d=mm&f=y"
	}
	return avatarURL + fmt.Sprintf("&s=%d", width)
}

// Person returns an equivalent Person.
func (u *User) Person() *Person {
	return &Person{
		PersonSpec: PersonSpec{UID: u.UID, Login: u.Login},
		FullName:   u.Name,
		AvatarURL:  u.AvatarURL,
	}
}
