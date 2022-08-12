package repos

import (
	"net/url"
)

// setUserinfoBestEffort adds the username and password to rawurl. If username param
// is not set, user from url is used. If password is not set and there is a
// user, password is used. If anything fails, the original rawurl is returned.
func setUserinfoBestEffort(rawurl, username, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	// Fallback to get username and password from URL if not specified already
	if username == "" && u.User != nil && u.User.Username() != "" {
		username = u.User.Username()
		password = ""
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	if username == "" {
		return rawurl
	}

	if password != "" {
		u.User = url.UserPassword(username, password)
	} else {
		u.User = url.User(username)
	}

	return u.String()
}
