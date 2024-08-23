package input

import (
	"fmt"
)

// UnknownCsiEvent represents an unknown CSI sequence event.
type UnknownCsiEvent string

// String implements fmt.Stringer.
func (e UnknownCsiEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// UnknownOscEvent represents an unknown OSC sequence event.
type UnknownOscEvent string

// String implements fmt.Stringer.
func (e UnknownOscEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// UnknownDcsEvent represents an unknown DCS sequence event.
type UnknownDcsEvent string

// String implements fmt.Stringer.
func (e UnknownDcsEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// UnknownApcEvent represents an unknown APC sequence event.
type UnknownApcEvent string

// String implements fmt.Stringer.
func (e UnknownApcEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// UnknownSs3Event represents an unknown SS3 sequence event.
type UnknownSs3Event string

// String implements fmt.Stringer.
func (e UnknownSs3Event) String() string {
	return fmt.Sprintf("%q", string(e))
}
