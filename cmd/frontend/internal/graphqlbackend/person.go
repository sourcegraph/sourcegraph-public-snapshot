package graphqlbackend

import (
	"crypto/md5"
	"fmt"
	"strings"
)

type personResolver struct {
	name  string
	email string
}

func (r *personResolver) Name() string {
	return r.name
}

func (r *personResolver) Email() string {
	return r.email
}

func (r *personResolver) DisplayName() string {
	// Trim space in case of strings like " <a@example.com>" where the name
	// part is " ". (We wouldn't want to show an all-whitespace name in the UI.)
	if name := strings.TrimSpace(r.name); name != "" {
		return name
	}
	if r.email != "" {
		return r.email
	}
	return "unknown"
}

func (r *personResolver) GravatarHash() string {
	return ConstructGravatarHash(r.email)
}

func (r *personResolver) AvatarURL() string {
	return "https://www.gravatar.com/avatar/" + ConstructGravatarHash(r.email) + "?d=identicon"
}

// ConstructGravatarHash hashes the email into a gravatar hash
func ConstructGravatarHash(email string) string {
	if email != "" {
		h := md5.New()
		h.Write([]byte(strings.ToLower(email)))
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	return ""
}
