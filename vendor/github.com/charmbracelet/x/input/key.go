package input

// KeySym is a keyboard symbol.
type KeySym int

// Symbol constants.
const (
	KeyNone KeySym = iota

	// Special names in C0

	KeyBackspace
	KeyTab
	KeyEnter
	KeyEscape

	// Special names in G0

	KeySpace
	KeyDelete

	// Special keys

	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyBegin
	KeyFind
	KeyInsert
	KeySelect
	KeyPgUp
	KeyPgDown
	KeyHome
	KeyEnd

	// Keypad keys

	KeyKpEnter
	KeyKpEqual
	KeyKpMultiply
	KeyKpPlus
	KeyKpComma
	KeyKpMinus
	KeyKpDecimal
	KeyKpDivide
	KeyKp0
	KeyKp1
	KeyKp2
	KeyKp3
	KeyKp4
	KeyKp5
	KeyKp6
	KeyKp7
	KeyKp8
	KeyKp9

	// The following are keys defined in the Kitty keyboard protocol.
	// TODO: Investigate the names of these keys
	KeyKpSep
	KeyKpUp
	KeyKpDown
	KeyKpLeft
	KeyKpRight
	KeyKpPgUp
	KeyKpPgDown
	KeyKpHome
	KeyKpEnd
	KeyKpInsert
	KeyKpDelete
	KeyKpBegin

	// Function keys

	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25
	KeyF26
	KeyF27
	KeyF28
	KeyF29
	KeyF30
	KeyF31
	KeyF32
	KeyF33
	KeyF34
	KeyF35
	KeyF36
	KeyF37
	KeyF38
	KeyF39
	KeyF40
	KeyF41
	KeyF42
	KeyF43
	KeyF44
	KeyF45
	KeyF46
	KeyF47
	KeyF48
	KeyF49
	KeyF50
	KeyF51
	KeyF52
	KeyF53
	KeyF54
	KeyF55
	KeyF56
	KeyF57
	KeyF58
	KeyF59
	KeyF60
	KeyF61
	KeyF62
	KeyF63

	// The following are keys defined in the Kitty keyboard protocol.
	// TODO: Investigate the names of these keys

	KeyCapsLock
	KeyScrollLock
	KeyNumLock
	KeyPrintScreen
	KeyPause
	KeyMenu

	KeyMediaPlay
	KeyMediaPause
	KeyMediaPlayPause
	KeyMediaReverse
	KeyMediaStop
	KeyMediaFastForward
	KeyMediaRewind
	KeyMediaNext
	KeyMediaPrev
	KeyMediaRecord

	KeyLowerVol
	KeyRaiseVol
	KeyMute

	KeyLeftShift
	KeyLeftAlt
	KeyLeftCtrl
	KeyLeftSuper
	KeyLeftHyper
	KeyLeftMeta
	KeyRightShift
	KeyRightAlt
	KeyRightCtrl
	KeyRightSuper
	KeyRightHyper
	KeyRightMeta
	KeyIsoLevel3Shift
	KeyIsoLevel5Shift
)

// Key represents a key event.
type Key struct {
	Sym      KeySym
	Rune     rune
	AltRune  rune
	BaseRune rune
	IsRepeat bool
	Mod      KeyMod
}

// KeyDownEvent represents a key down event.
type KeyDownEvent Key

// String implements fmt.Stringer.
func (k KeyDownEvent) String() string {
	return Key(k).String()
}

// KeyUpEvent represents a key up event.
type KeyUpEvent Key

// String implements fmt.Stringer.
func (k KeyUpEvent) String() string {
	return Key(k).String()
}

// String implements fmt.Stringer.
func (k Key) String() string {
	var s string
	if k.Mod.IsCtrl() && k.Sym != KeyLeftCtrl && k.Sym != KeyRightCtrl {
		s += "ctrl+"
	}
	if k.Mod.IsAlt() && k.Sym != KeyLeftAlt && k.Sym != KeyRightAlt {
		s += "alt+"
	}
	if k.Mod.IsShift() && k.Sym != KeyLeftShift && k.Sym != KeyRightShift {
		s += "shift+"
	}
	if k.Mod.IsMeta() && k.Sym != KeyLeftMeta && k.Sym != KeyRightMeta {
		s += "meta+"
	}
	if k.Mod.IsHyper() && k.Sym != KeyLeftHyper && k.Sym != KeyRightHyper {
		s += "hyper+"
	}
	if k.Mod.IsSuper() && k.Sym != KeyLeftSuper && k.Sym != KeyRightSuper {
		s += "super+"
	}

	runeStr := func(r rune) string {
		// Space is the only invisible printable character.
		if r == ' ' {
			return "space"
		}
		return string(r)
	}
	if k.BaseRune != 0 {
		// If a BaseRune is present, use it to represent a key using the standard
		// PC-101 key layout.
		s += runeStr(k.BaseRune)
	} else if k.AltRune != 0 {
		// Otherwise, use the AltRune aka the non-shifted one if present.
		s += runeStr(k.AltRune)
	} else if k.Rune != 0 {
		// Else, just print the rune.
		s += runeStr(k.Rune)
	} else {
		s += k.Sym.String()
	}
	return s
}

// String implements fmt.Stringer.
func (k KeySym) String() string {
	s, ok := keySymString[k]
	if !ok {
		return "unknown"
	}
	return s
}

var keySymString = map[KeySym]string{
	KeyEnter:      "enter",
	KeyTab:        "tab",
	KeyBackspace:  "backspace",
	KeyEscape:     "esc",
	KeySpace:      "space",
	KeyUp:         "up",
	KeyDown:       "down",
	KeyLeft:       "left",
	KeyRight:      "right",
	KeyBegin:      "begin",
	KeyFind:       "find",
	KeyInsert:     "insert",
	KeyDelete:     "delete",
	KeySelect:     "select",
	KeyPgUp:       "pgup",
	KeyPgDown:     "pgdown",
	KeyHome:       "home",
	KeyEnd:        "end",
	KeyKpEnter:    "kpenter",
	KeyKpEqual:    "kpequal",
	KeyKpMultiply: "kpmul",
	KeyKpPlus:     "kpplus",
	KeyKpComma:    "kpcomma",
	KeyKpMinus:    "kpminus",
	KeyKpDecimal:  "kpperiod",
	KeyKpDivide:   "kpdiv",
	KeyKp0:        "kp0",
	KeyKp1:        "kp1",
	KeyKp2:        "kp2",
	KeyKp3:        "kp3",
	KeyKp4:        "kp4",
	KeyKp5:        "kp5",
	KeyKp6:        "kp6",
	KeyKp7:        "kp7",
	KeyKp8:        "kp8",
	KeyKp9:        "kp9",

	// Kitty keyboard extension
	KeyKpSep:    "kpsep",
	KeyKpUp:     "kpup",
	KeyKpDown:   "kpdown",
	KeyKpLeft:   "kpleft",
	KeyKpRight:  "kpright",
	KeyKpPgUp:   "kppgup",
	KeyKpPgDown: "kppgdown",
	KeyKpHome:   "kphome",
	KeyKpEnd:    "kpend",
	KeyKpInsert: "kpinsert",
	KeyKpDelete: "kpdelete",
	KeyKpBegin:  "kpbegin",

	KeyF1:  "f1",
	KeyF2:  "f2",
	KeyF3:  "f3",
	KeyF4:  "f4",
	KeyF5:  "f5",
	KeyF6:  "f6",
	KeyF7:  "f7",
	KeyF8:  "f8",
	KeyF9:  "f9",
	KeyF10: "f10",
	KeyF11: "f11",
	KeyF12: "f12",
	KeyF13: "f13",
	KeyF14: "f14",
	KeyF15: "f15",
	KeyF16: "f16",
	KeyF17: "f17",
	KeyF18: "f18",
	KeyF19: "f19",
	KeyF20: "f20",
	KeyF21: "f21",
	KeyF22: "f22",
	KeyF23: "f23",
	KeyF24: "f24",
	KeyF25: "f25",
	KeyF26: "f26",
	KeyF27: "f27",
	KeyF28: "f28",
	KeyF29: "f29",
	KeyF30: "f30",
	KeyF31: "f31",
	KeyF32: "f32",
	KeyF33: "f33",
	KeyF34: "f34",
	KeyF35: "f35",
	KeyF36: "f36",
	KeyF37: "f37",
	KeyF38: "f38",
	KeyF39: "f39",
	KeyF40: "f40",
	KeyF41: "f41",
	KeyF42: "f42",
	KeyF43: "f43",
	KeyF44: "f44",
	KeyF45: "f45",
	KeyF46: "f46",
	KeyF47: "f47",
	KeyF48: "f48",
	KeyF49: "f49",
	KeyF50: "f50",
	KeyF51: "f51",
	KeyF52: "f52",
	KeyF53: "f53",
	KeyF54: "f54",
	KeyF55: "f55",
	KeyF56: "f56",
	KeyF57: "f57",
	KeyF58: "f58",
	KeyF59: "f59",
	KeyF60: "f60",
	KeyF61: "f61",
	KeyF62: "f62",
	KeyF63: "f63",

	// Kitty keyboard extension
	KeyCapsLock:         "capslock",
	KeyScrollLock:       "scrolllock",
	KeyNumLock:          "numlock",
	KeyPrintScreen:      "printscreen",
	KeyPause:            "pause",
	KeyMenu:             "menu",
	KeyMediaPlay:        "mediaplay",
	KeyMediaPause:       "mediapause",
	KeyMediaPlayPause:   "mediaplaypause",
	KeyMediaReverse:     "mediareverse",
	KeyMediaStop:        "mediastop",
	KeyMediaFastForward: "mediafastforward",
	KeyMediaRewind:      "mediarewind",
	KeyMediaNext:        "medianext",
	KeyMediaPrev:        "mediaprev",
	KeyMediaRecord:      "mediarecord",
	KeyLowerVol:         "lowervol",
	KeyRaiseVol:         "raisevol",
	KeyMute:             "mute",
	KeyLeftShift:        "leftshift",
	KeyLeftAlt:          "leftalt",
	KeyLeftCtrl:         "leftctrl",
	KeyLeftSuper:        "leftsuper",
	KeyLeftHyper:        "lefthyper",
	KeyLeftMeta:         "leftmeta",
	KeyRightShift:       "rightshift",
	KeyRightAlt:         "rightalt",
	KeyRightCtrl:        "rightctrl",
	KeyRightSuper:       "rightsuper",
	KeyRightHyper:       "righthyper",
	KeyRightMeta:        "rightmeta",
	KeyIsoLevel3Shift:   "isolevel3shift",
	KeyIsoLevel5Shift:   "isolevel5shift",
}
