package term

import (
	"image/color"
	"io"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/input"
)

// File represents a file that has a file descriptor and can be read from,
// written to, and closed.
type File interface {
	io.ReadWriteCloser
	Fd() uintptr
}

// QueryBackgroundColor queries the terminal for the background color.
// If the terminal does not support querying the background color, nil is
// returned.
//
// Note: you will need to set the input to raw mode before calling this
// function.
//
//	state, _ := term.MakeRaw(in.Fd())
//	defer term.Restore(in.Fd(), state)
func QueryBackgroundColor(in io.Reader, out io.Writer) (c color.Color, err error) {
	// nolint: errcheck
	err = QueryTerminal(in, out, defaultQueryTimeout,
		func(events []input.Event) bool {
			for _, e := range events {
				switch e := e.(type) {
				case input.BackgroundColorEvent:
					c = e.Color
					continue // we need to consume the next DA1 event
				case input.PrimaryDeviceAttributesEvent:
					return false
				}
			}
			return true
		}, ansi.RequestBackgroundColor+ansi.RequestPrimaryDeviceAttributes)
	return
}

// QueryKittyKeyboard returns the enabled Kitty keyboard protocol options.
// -1 means the terminal does not support the feature.
//
// Note: you will need to set the input to raw mode before calling this
// function.
//
//	state, _ := term.MakeRaw(in.Fd())
//	defer term.Restore(in.Fd(), state)
func QueryKittyKeyboard(in io.Reader, out io.Writer) (flags int, err error) {
	flags = -1
	// nolint: errcheck
	err = QueryTerminal(in, out, defaultQueryTimeout,
		func(events []input.Event) bool {
			for _, e := range events {
				switch event := e.(type) {
				case input.KittyKeyboardEvent:
					flags = int(event)
					continue // we need to consume the next DA1 event
				case input.PrimaryDeviceAttributesEvent:
					return false
				}
			}
			return true
		}, ansi.RequestKittyKeyboard+ansi.RequestPrimaryDeviceAttributes)
	return
}

const defaultQueryTimeout = time.Second * 2

// QueryTerminalFilter is a function that filters input events using a type
// switch. If false is returned, the QueryTerminal function will stop reading
// input.
type QueryTerminalFilter func(events []input.Event) bool

// QueryTerminal queries the terminal for support of various features and
// returns a list of response events.
// Most of the time, you will need to set stdin to raw mode before calling this
// function.
// Note: This function will block until the terminal responds or the timeout
// is reached.
func QueryTerminal(
	in io.Reader,
	out io.Writer,
	timeout time.Duration,
	filter QueryTerminalFilter,
	query string,
) error {
	rd, err := input.NewDriver(in, "", 0)
	if err != nil {
		return err
	}

	defer rd.Close() // nolint: errcheck

	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-time.After(timeout):
			rd.Cancel()
		}
	}()

	if _, err := io.WriteString(out, query); err != nil {
		return err
	}

	for {
		events, err := rd.ReadEvents()
		if err != nil {
			return err
		}

		if !filter(events) {
			break
		}
	}

	return nil
}
