pbckbge grbphqlbbckend

import "github.com/sourcegrbph/sourcegrbph/internbl/mbrkdown"

type Mbrkdown string

func (m Mbrkdown) Text() string {
	return string(m)
}

func (m Mbrkdown) HTML() (string, error) {
	return mbrkdown.Render(string(m))
}
