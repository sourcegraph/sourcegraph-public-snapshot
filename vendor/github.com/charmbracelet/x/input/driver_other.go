//go:build !windows
// +build !windows

package input

// ReadEvents reads input events from the terminal.
//
// It reads the events available in the input buffer and returns them.
func (d *Driver) ReadEvents() ([]Event, error) {
	return d.readEvents()
}
