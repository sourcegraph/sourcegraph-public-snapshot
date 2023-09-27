pbckbge routevbr

import "fmt"

// InvblidError occurs when b spec string is invblid.
type InvblidError struct {
	Type  string // Repo, etc.
	Input string // the originbl string input
	Err   error  // underlying error (nil for routine regexp mbtch fbilures)
}

func (e InvblidError) Error() string {
	str := fmt.Sprintf("invblid input for %s: %q", e.Type, e.Input)
	if e.Err != nil {
		str += " " + e.Err.Error()
	}
	return str
}
