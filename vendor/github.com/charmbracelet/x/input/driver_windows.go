//go:build windows
// +build windows

package input

import (
	"errors"
	"fmt"
	"unicode/utf16"

	"github.com/charmbracelet/x/ansi"
	termwindows "github.com/charmbracelet/x/windows"
	"github.com/erikgeiser/coninput"
	"golang.org/x/sys/windows"
)

// ReadEvents reads input events from the terminal.
//
// It reads the events available in the input buffer and returns them.
func (d *Driver) ReadEvents() ([]Event, error) {
	events, err := d.handleConInput(coninput.ReadConsoleInput)
	if errors.Is(err, errNotConInputReader) {
		return d.readEvents()
	}
	return events, err
}

var errNotConInputReader = fmt.Errorf("handleConInput: not a conInputReader")

func (d *Driver) handleConInput(
	finput func(windows.Handle, []coninput.InputRecord) (uint32, error),
) ([]Event, error) {
	cc, ok := d.rd.(*conInputReader)
	if !ok {
		return nil, errNotConInputReader
	}

	// read up to 256 events, this is to allow for sequences events reported as
	// key events.
	var events [256]coninput.InputRecord
	_, err := finput(cc.conin, events[:])
	if err != nil {
		return nil, fmt.Errorf("read coninput events: %w", err)
	}

	var evs []Event
	for _, event := range events {
		if e := parseConInputEvent(event, &d.prevMouseState); e != nil {
			evs = append(evs, e)
		}
	}

	return d.detectConInputQuerySequences(evs), nil
}

// Using ConInput API, Windows Terminal responds to sequence query events with
// KEY_EVENT_RECORDs so we need to collect them and parse them as a single
// sequence.
// Is this a hack?
func (d *Driver) detectConInputQuerySequences(events []Event) []Event {
	var newEvents []Event
	start, end := -1, -1

loop:
	for i, e := range events {
		switch e := e.(type) {
		case KeyDownEvent:
			switch e.Rune {
			case ansi.ESC, ansi.CSI, ansi.OSC, ansi.DCS, ansi.APC:
				// start of a sequence
				if start == -1 {
					start = i
				}
			}
		default:
			break loop
		}
		end = i
	}

	if start == -1 || end <= start {
		return events
	}

	var seq []byte
	for i := start; i <= end; i++ {
		switch e := events[i].(type) {
		case KeyDownEvent:
			seq = append(seq, byte(e.Rune))
		}
	}

	n, seqevent := ParseSequence(seq)
	switch seqevent.(type) {
	case UnknownEvent:
		// We're not interested in unknown events
	default:
		if start+n > len(events) {
			return events
		}
		newEvents = events[:start]
		newEvents = append(newEvents, seqevent)
		newEvents = append(newEvents, events[start+n:]...)
		return d.detectConInputQuerySequences(newEvents)
	}

	return events
}

func parseConInputEvent(event coninput.InputRecord, ps *coninput.ButtonState) Event {
	switch e := event.Unwrap().(type) {
	case coninput.KeyEventRecord:
		event := parseWin32InputKeyEvent(e.VirtualKeyCode, e.VirtualScanCode,
			e.Char, e.KeyDown, e.ControlKeyState, e.RepeatCount)

		var key Key
		switch event := event.(type) {
		case KeyDownEvent:
			key = Key(event)
		case KeyUpEvent:
			key = Key(event)
		default:
			return nil
		}

		// If the key is not printable, return the event as is
		// (e.g. function keys, arrows, etc.)
		// Otherwise, try to translate it to a rune based on the active keyboard
		// layout.
		if key.Rune == 0 {
			return event
		}

		// Get active keyboard layout
		fgWin := windows.GetForegroundWindow()
		fgThread, err := windows.GetWindowThreadProcessId(fgWin, nil)
		if err != nil {
			return event
		}

		layout, err := termwindows.GetKeyboardLayout(fgThread)
		if layout == windows.InvalidHandle || err != nil {
			return event
		}

		// Translate key to rune
		var keyState [256]byte
		var utf16Buf [16]uint16
		const dontChangeKernelKeyboardLayout = 0x4
		ret, err := termwindows.ToUnicodeEx(
			uint32(e.VirtualKeyCode),
			uint32(e.VirtualScanCode),
			&keyState[0],
			&utf16Buf[0],
			int32(len(utf16Buf)),
			dontChangeKernelKeyboardLayout,
			layout,
		)

		// -1 indicates a dead key
		// 0 indicates no translation for this key
		if ret < 1 || err != nil {
			return event
		}

		runes := utf16.Decode(utf16Buf[:ret])
		if len(runes) != 1 {
			// Key doesn't translate to a single rune
			return event
		}

		key.BaseRune = runes[0]
		if e.KeyDown {
			return KeyDownEvent(key)
		}

		return KeyUpEvent(key)

	case coninput.WindowBufferSizeEventRecord:
		return WindowSizeEvent{
			Width:  int(e.Size.X),
			Height: int(e.Size.Y),
		}
	case coninput.MouseEventRecord:
		mevent := mouseEvent(*ps, e)
		*ps = e.ButtonState
		return mevent
	case coninput.FocusEventRecord, coninput.MenuEventRecord:
		// ignore
	}
	return nil
}

func mouseEventButton(p, s coninput.ButtonState) (button MouseButton, isRelease bool) {
	btn := p ^ s
	if btn&s == 0 {
		isRelease = true
	}

	if btn == 0 {
		switch {
		case s&coninput.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			button = MouseLeft
		case s&coninput.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			button = MouseMiddle
		case s&coninput.RIGHTMOST_BUTTON_PRESSED > 0:
			button = MouseRight
		case s&coninput.FROM_LEFT_3RD_BUTTON_PRESSED > 0:
			button = MouseBackward
		case s&coninput.FROM_LEFT_4TH_BUTTON_PRESSED > 0:
			button = MouseForward
		}
		return
	}

	switch {
	case btn == coninput.FROM_LEFT_1ST_BUTTON_PRESSED: // left button
		button = MouseLeft
	case btn == coninput.RIGHTMOST_BUTTON_PRESSED: // right button
		button = MouseRight
	case btn == coninput.FROM_LEFT_2ND_BUTTON_PRESSED: // middle button
		button = MouseMiddle
	case btn == coninput.FROM_LEFT_3RD_BUTTON_PRESSED: // unknown (possibly mouse backward)
		button = MouseBackward
	case btn == coninput.FROM_LEFT_4TH_BUTTON_PRESSED: // unknown (possibly mouse forward)
		button = MouseForward
	}

	return
}

func mouseEvent(p coninput.ButtonState, e coninput.MouseEventRecord) (ev Event) {
	var mod KeyMod
	var isRelease bool
	if e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
		mod |= Alt
	}
	if e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED) {
		mod |= Ctrl
	}
	if e.ControlKeyState.Contains(coninput.SHIFT_PRESSED) {
		mod |= Shift
	}
	m := Mouse{
		X:   int(e.MousePositon.X),
		Y:   int(e.MousePositon.Y),
		Mod: mod,
	}
	switch e.EventFlags {
	case coninput.CLICK, coninput.DOUBLE_CLICK:
		m.Button, isRelease = mouseEventButton(p, e.ButtonState)
	case coninput.MOUSE_WHEELED:
		if e.WheelDirection > 0 {
			m.Button = MouseWheelUp
		} else {
			m.Button = MouseWheelDown
		}
	case coninput.MOUSE_HWHEELED:
		if e.WheelDirection > 0 {
			m.Button = MouseWheelRight
		} else {
			m.Button = MouseWheelLeft
		}
	case coninput.MOUSE_MOVED:
		m.Button, _ = mouseEventButton(p, e.ButtonState)
		return MouseMotionEvent(m)
	}

	if isWheel(m.Button) {
		return MouseWheelEvent(m)
	} else if isRelease {
		return MouseUpEvent(m)
	}

	return MouseDownEvent(m)
}
