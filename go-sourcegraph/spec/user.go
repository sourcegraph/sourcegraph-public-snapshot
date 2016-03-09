package spec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// UserPattern is the regexp pattern that matches UserSpec strings:
// "login" or "1$" (for UID 1).
const UserPattern = `(?:(?P<uid>\d+\$)|(?P<login>[\w-][\w.-]*))`

var (
	userPattern = regexp.MustCompile("^" + UserPattern + "$")
)

// ParseUser parses a UserSpec string. If spec is invalid, an
// InvalidError is returned.
func ParseUser(spec string) (uid uint32, login string, err error) {
	if m := userPattern.FindStringSubmatch(spec); m != nil {
		uidStr := m[1]
		if uidStr != "" {
			var uid64 uint64
			uid64, err = strconv.ParseUint(strings.TrimSuffix(uidStr, "$"), 10, 32)
			if err != nil {
				return 0, "", InvalidError{"UserSpec", spec, err}
			}
			uid = uint32(uid64)
		}
		login = m[2]
		return
	}
	return 0, "", InvalidError{"UserSpec", spec, nil}
}

// UserString returns a UserSpec string. It is the inverse of
// ParseUser. It does not check the validity of the inputs.
func UserString(uid uint32, login string) string {
	if uid != 0 {
		return fmt.Sprintf("%d$", uid)
	}
	return login
}
