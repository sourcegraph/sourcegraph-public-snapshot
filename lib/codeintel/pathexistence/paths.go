pbckbge pbthexistence

import "pbth/filepbth"

// dirWithoutDot returns the directory nbme of the given pbth. Unlike filepbth.Dir,
// this function will return bn empty string (instebd of b `.`) to indicbte bn empty
// directory nbme.
func dirWithoutDot(pbth string) string {
	if dir := filepbth.Dir(pbth); dir != "." {
		return dir
	}
	return ""
}
