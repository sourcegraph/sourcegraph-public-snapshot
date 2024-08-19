package input

import (
	"regexp"

	"github.com/charmbracelet/x/ansi"
)

// MouseButton represents the button that was pressed during a mouse event.
type MouseButton byte

// Mouse event buttons
//
// This is based on X11 mouse button codes.
//
//	1 = left button
//	2 = middle button (pressing the scroll wheel)
//	3 = right button
//	4 = turn scroll wheel up
//	5 = turn scroll wheel down
//	6 = push scroll wheel left
//	7 = push scroll wheel right
//	8 = 4th button (aka browser backward button)
//	9 = 5th button (aka browser forward button)
//	10
//	11
//
// Other buttons are not supported.
const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
	MouseWheelLeft
	MouseWheelRight
	MouseBackward
	MouseForward
	MouseExtra1
	MouseExtra2
)

var mouseButtons = map[MouseButton]string{
	MouseNone:       "none",
	MouseLeft:       "left",
	MouseMiddle:     "middle",
	MouseRight:      "right",
	MouseWheelUp:    "wheelup",
	MouseWheelDown:  "wheeldown",
	MouseWheelLeft:  "wheelleft",
	MouseWheelRight: "wheelright",
	MouseBackward:   "backward",
	MouseForward:    "forward",
	MouseExtra1:     "button10",
	MouseExtra2:     "button11",
}

// Mouse represents a Mouse event.
type Mouse struct {
	X, Y   int
	Button MouseButton
	Mod    KeyMod
}

// String implements fmt.Stringer.
func (m Mouse) String() (s string) {
	if m.Mod.IsCtrl() {
		s += "ctrl+"
	}
	if m.Mod.IsAlt() {
		s += "alt+"
	}
	if m.Mod.IsShift() {
		s += "shift+"
	}

	str, ok := mouseButtons[m.Button]
	if !ok {
		s += "unknown"
	} else if str != "none" { // motion events don't have a button
		s += str
	}

	return s
}

// MouseDownEvent represents a mouse button down event.
type MouseDownEvent Mouse

// String implements fmt.Stringer.
func (e MouseDownEvent) String() string {
	return Mouse(e).String()
}

// MouseUpEvent represents a mouse button up event.
type MouseUpEvent Mouse

// String implements fmt.Stringer.
func (e MouseUpEvent) String() string {
	return Mouse(e).String()
}

// MouseWheelEvent represents a mouse wheel event.
type MouseWheelEvent Mouse

// String implements fmt.Stringer.
func (e MouseWheelEvent) String() string {
	return Mouse(e).String()
}

// MouseMotionEvent represents a mouse motion event.
type MouseMotionEvent Mouse

// String implements fmt.Stringer.
func (e MouseMotionEvent) String() string {
	m := Mouse(e)
	if m.Button != 0 {
		return m.String() + "+motion"
	}
	return m.String() + "motion"
}

var mouseSGRRegex = regexp.MustCompile(`(\d+);(\d+);(\d+)([Mm])`)

// Parse SGR-encoded mouse events; SGR extended mouse events. SGR mouse events
// look like:
//
//	ESC [ < Cb ; Cx ; Cy (M or m)
//
// where:
//
//	Cb is the encoded button code
//	Cx is the x-coordinate of the mouse
//	Cy is the y-coordinate of the mouse
//	M is for button press, m is for button release
//
// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseSGRMouseEvent(csi *ansi.CsiSequence) Event {
	x := csi.Param(1)
	y := csi.Param(2)
	release := csi.Command() == 'm'
	mod, btn, _, isMotion := parseMouseButton(csi.Param(0))

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	x--
	y--

	m := Mouse{X: x, Y: y, Button: btn, Mod: mod}

	// Wheel buttons don't have release events
	// Motion can be reported as a release event in some terminals (Windows Terminal)
	if isWheel(m.Button) {
		return MouseWheelEvent(m)
	} else if !isMotion && release {
		return MouseUpEvent(m)
	} else if isMotion {
		return MouseMotionEvent(m)
	}
	return MouseDownEvent(m)
}

const x10MouseByteOffset = 32

// Parse X10-encoded mouse events; the simplest kind. The last release of X10
// was December 1986, by the way. The original X10 mouse protocol limits the Cx
// and Cy coordinates to 223 (=255-032).
//
// X10 mouse events look like:
//
//	ESC [M Cb Cx Cy
//
// See: http://www.xfree86.org/current/ctlseqs.html#Mouse%20Tracking
func parseX10MouseEvent(buf []byte) Event {
	v := buf[3:6]
	b := int(v[0])
	if b >= x10MouseByteOffset {
		// XXX: b < 32 should be impossible, but we're being defensive.
		b -= x10MouseByteOffset
	}

	mod, btn, isRelease, isMotion := parseMouseButton(b)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	x := int(v[1]) - x10MouseByteOffset - 1
	y := int(v[2]) - x10MouseByteOffset - 1

	m := Mouse{X: x, Y: y, Button: btn, Mod: mod}
	if isWheel(m.Button) {
		return MouseWheelEvent(m)
	} else if isMotion {
		return MouseMotionEvent(m)
	} else if isRelease {
		return MouseUpEvent(m)
	}
	return MouseDownEvent(m)
}

// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseMouseButton(b int) (mod KeyMod, btn MouseButton, isRelease bool, isMotion bool) {
	// mouse bit shifts
	const (
		bitShift  = 0b0000_0100
		bitAlt    = 0b0000_1000
		bitCtrl   = 0b0001_0000
		bitMotion = 0b0010_0000
		bitWheel  = 0b0100_0000
		bitAdd    = 0b1000_0000 // additional buttons 8-11

		bitsMask = 0b0000_0011
	)

	// Modifiers
	if b&bitAlt != 0 {
		mod |= Alt
	}
	if b&bitCtrl != 0 {
		mod |= Ctrl
	}
	if b&bitShift != 0 {
		mod |= Shift
	}

	if b&bitAdd != 0 {
		btn = MouseBackward + MouseButton(b&bitsMask)
	} else if b&bitWheel != 0 {
		btn = MouseWheelUp + MouseButton(b&bitsMask)
	} else {
		btn = MouseLeft + MouseButton(b&bitsMask)
		// X10 reports a button release as 0b0000_0011 (3)
		if b&bitsMask == bitsMask {
			btn = MouseNone
			isRelease = true
		}
	}

	// Motion bit doesn't get reported for wheel events.
	if b&bitMotion != 0 && !isWheel(btn) {
		isMotion = true
	}

	return
}

// isWheel returns true if the mouse event is a wheel event.
func isWheel(btn MouseButton) bool {
	return btn >= MouseWheelUp && btn <= MouseWheelRight
}
