package emailaddrs

import (
	"fmt"
	"strings"
)

func Split(email string) (user, domain string, err error) {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("email has no '@': %q", email)
	}
	user, domain = parts[0], parts[1]
	if len(user) == 0 {
		return "", "", fmt.Errorf("email user is empty: %q", email)
	}
	if len(domain) == 0 {
		return "", "", fmt.Errorf("email domain is empty: %q", email)
	}
	return
}
