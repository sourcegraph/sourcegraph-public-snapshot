package ansi

import "strconv"

// Kitty keyboard protocol progressive enhancement flags.
// See: https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
const (
	KittyDisambiguateEscapeCodes = 1 << iota
	KittyReportEventTypes
	KittyReportAlternateKeys
	KittyReportAllKeys
	KittyReportAssociatedKeys

	KittyAllFlags = KittyDisambiguateEscapeCodes | KittyReportEventTypes |
		KittyReportAlternateKeys | KittyReportAllKeys | KittyReportAssociatedKeys
)

// RequestKittyKeyboard is a sequence to request the terminal Kitty keyboard
// protocol enabled flags.
//
// See: https://sw.kovidgoyal.net/kitty/keyboard-protocol/
const RequestKittyKeyboard = "\x1b[?u"

// PushKittyKeyboard returns a sequence to push the given flags to the terminal
// Kitty Keyboard stack.
//
//	CSI > flags u
//
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
func PushKittyKeyboard(flags int) string {
	var f string
	if flags > 0 {
		f = strconv.Itoa(flags)
	}

	return "\x1b[>" + f + "u"
}

// DisableKittyKeyboard is a sequence to push zero into the terminal Kitty
// Keyboard stack to disable the protocol.
//
// This is equivalent to PushKittyKeyboard(0).
const DisableKittyKeyboard = "\x1b[>0u"

// PopKittyKeyboard returns a sequence to pop n number of flags from the
// terminal Kitty Keyboard stack.
//
//	CSI < flags u
//
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
func PopKittyKeyboard(n int) string {
	var num string
	if n > 0 {
		num = strconv.Itoa(n)
	}

	return "\x1b[<" + num + "u"
}
