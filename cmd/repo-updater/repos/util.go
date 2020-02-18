package repos

import (
	"net/url"
)

// setUserinfoBestEffort adds the username and password to rawurl. If user is
// not set in rawurl, username is used. If password is not set and there is a
// user, password is used. If anything fails, the original rawurl is returned.
func setUserinfoBestEffort(rawurl, username, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	passwordSet := password != ""

	// Update username and password if specified in rawurl
	if u.User != nil {
		if u.User.Username() != "" {
			username = u.User.Username()
		}
		if p, ok := u.User.Password(); ok {
			password = p
			passwordSet = true
		}
	}

	if username == "" {
		return rawurl
	}

	if passwordSet {
		u.User = url.UserPassword(username, password)
	} else {
		u.User = url.User(username)
	}

	return u.String()
}
