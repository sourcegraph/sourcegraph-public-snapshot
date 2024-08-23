package input

import (
	"strconv"

	"github.com/charmbracelet/x/ansi"
)

func buildKeysTable(flags int, term string) map[string]Key {
	nul := Key{Rune: ' ', Sym: KeySpace, Mod: Ctrl} // ctrl+@ or ctrl+space
	if flags&FlagCtrlAt != 0 {
		nul = Key{Rune: '@', Mod: Ctrl}
	}

	tab := Key{Sym: KeyTab} // ctrl+i or tab
	if flags&FlagCtrlI != 0 {
		tab = Key{Rune: 'i', Mod: Ctrl}
	}

	enter := Key{Sym: KeyEnter} // ctrl+m or enter
	if flags&FlagCtrlM != 0 {
		enter = Key{Rune: 'm', Mod: Ctrl}
	}

	esc := Key{Sym: KeyEscape} // ctrl+[ or escape
	if flags&FlagCtrlOpenBracket != 0 {
		esc = Key{Rune: '[', Mod: Ctrl} // ctrl+[ or escape
	}

	del := Key{Sym: KeyBackspace}
	if flags&FlagBackspace != 0 {
		del.Sym = KeyDelete
	}

	find := Key{Sym: KeyHome}
	if flags&FlagFind != 0 {
		find.Sym = KeyFind
	}

	sel := Key{Sym: KeyEnd}
	if flags&FlagSelect != 0 {
		sel.Sym = KeySelect
	}

	// The following is a table of key sequences and their corresponding key
	// events based on the VT100/VT200 terminal specs.
	//
	// See: https://vt100.net/docs/vt100-ug/chapter3.html#S3.2
	// See: https://vt100.net/docs/vt220-rm/chapter3.html
	//
	// XXX: These keys may be overwritten by other options like XTerm or
	// Terminfo.
	table := map[string]Key{
		// C0 control characters
		string(byte(ansi.NUL)): nul,
		string(byte(ansi.SOH)): {Rune: 'a', Mod: Ctrl},
		string(byte(ansi.STX)): {Rune: 'b', Mod: Ctrl},
		string(byte(ansi.ETX)): {Rune: 'c', Mod: Ctrl},
		string(byte(ansi.EOT)): {Rune: 'd', Mod: Ctrl},
		string(byte(ansi.ENQ)): {Rune: 'e', Mod: Ctrl},
		string(byte(ansi.ACK)): {Rune: 'f', Mod: Ctrl},
		string(byte(ansi.BEL)): {Rune: 'g', Mod: Ctrl},
		string(byte(ansi.BS)):  {Rune: 'h', Mod: Ctrl},
		string(byte(ansi.HT)):  tab,
		string(byte(ansi.LF)):  {Rune: 'j', Mod: Ctrl},
		string(byte(ansi.VT)):  {Rune: 'k', Mod: Ctrl},
		string(byte(ansi.FF)):  {Rune: 'l', Mod: Ctrl},
		string(byte(ansi.CR)):  enter,
		string(byte(ansi.SO)):  {Rune: 'n', Mod: Ctrl},
		string(byte(ansi.SI)):  {Rune: 'o', Mod: Ctrl},
		string(byte(ansi.DLE)): {Rune: 'p', Mod: Ctrl},
		string(byte(ansi.DC1)): {Rune: 'q', Mod: Ctrl},
		string(byte(ansi.DC2)): {Rune: 'r', Mod: Ctrl},
		string(byte(ansi.DC3)): {Rune: 's', Mod: Ctrl},
		string(byte(ansi.DC4)): {Rune: 't', Mod: Ctrl},
		string(byte(ansi.NAK)): {Rune: 'u', Mod: Ctrl},
		string(byte(ansi.SYN)): {Rune: 'v', Mod: Ctrl},
		string(byte(ansi.ETB)): {Rune: 'w', Mod: Ctrl},
		string(byte(ansi.CAN)): {Rune: 'x', Mod: Ctrl},
		string(byte(ansi.EM)):  {Rune: 'y', Mod: Ctrl},
		string(byte(ansi.SUB)): {Rune: 'z', Mod: Ctrl},
		string(byte(ansi.ESC)): esc,
		string(byte(ansi.FS)):  {Rune: '\\', Mod: Ctrl},
		string(byte(ansi.GS)):  {Rune: ']', Mod: Ctrl},
		string(byte(ansi.RS)):  {Rune: '^', Mod: Ctrl},
		string(byte(ansi.US)):  {Rune: '_', Mod: Ctrl},

		// Special keys in G0
		string(byte(ansi.SP)):  {Sym: KeySpace, Rune: ' '},
		string(byte(ansi.DEL)): del,

		// Special keys

		"\x1b[Z": {Sym: KeyTab, Mod: Shift},

		"\x1b[1~": find,
		"\x1b[2~": {Sym: KeyInsert},
		"\x1b[3~": {Sym: KeyDelete},
		"\x1b[4~": sel,
		"\x1b[5~": {Sym: KeyPgUp},
		"\x1b[6~": {Sym: KeyPgDown},
		"\x1b[7~": {Sym: KeyHome},
		"\x1b[8~": {Sym: KeyEnd},

		// Normal mode
		"\x1b[A": {Sym: KeyUp},
		"\x1b[B": {Sym: KeyDown},
		"\x1b[C": {Sym: KeyRight},
		"\x1b[D": {Sym: KeyLeft},
		"\x1b[E": {Sym: KeyBegin},
		"\x1b[F": {Sym: KeyEnd},
		"\x1b[H": {Sym: KeyHome},
		"\x1b[P": {Sym: KeyF1},
		"\x1b[Q": {Sym: KeyF2},
		"\x1b[R": {Sym: KeyF3},
		"\x1b[S": {Sym: KeyF4},

		// Application Cursor Key Mode (DECCKM)
		"\x1bOA": {Sym: KeyUp},
		"\x1bOB": {Sym: KeyDown},
		"\x1bOC": {Sym: KeyRight},
		"\x1bOD": {Sym: KeyLeft},
		"\x1bOE": {Sym: KeyBegin},
		"\x1bOF": {Sym: KeyEnd},
		"\x1bOH": {Sym: KeyHome},
		"\x1bOP": {Sym: KeyF1},
		"\x1bOQ": {Sym: KeyF2},
		"\x1bOR": {Sym: KeyF3},
		"\x1bOS": {Sym: KeyF4},

		// Keypad Application Mode (DECKPAM)

		"\x1bOM": {Sym: KeyKpEnter},
		"\x1bOX": {Sym: KeyKpEqual},
		"\x1bOj": {Sym: KeyKpMultiply},
		"\x1bOk": {Sym: KeyKpPlus},
		"\x1bOl": {Sym: KeyKpComma},
		"\x1bOm": {Sym: KeyKpMinus},
		"\x1bOn": {Sym: KeyKpDecimal},
		"\x1bOo": {Sym: KeyKpDivide},
		"\x1bOp": {Sym: KeyKp0},
		"\x1bOq": {Sym: KeyKp1},
		"\x1bOr": {Sym: KeyKp2},
		"\x1bOs": {Sym: KeyKp3},
		"\x1bOt": {Sym: KeyKp4},
		"\x1bOu": {Sym: KeyKp5},
		"\x1bOv": {Sym: KeyKp6},
		"\x1bOw": {Sym: KeyKp7},
		"\x1bOx": {Sym: KeyKp8},
		"\x1bOy": {Sym: KeyKp9},

		// Function keys

		"\x1b[11~": {Sym: KeyF1},
		"\x1b[12~": {Sym: KeyF2},
		"\x1b[13~": {Sym: KeyF3},
		"\x1b[14~": {Sym: KeyF4},
		"\x1b[15~": {Sym: KeyF5},
		"\x1b[17~": {Sym: KeyF6},
		"\x1b[18~": {Sym: KeyF7},
		"\x1b[19~": {Sym: KeyF8},
		"\x1b[20~": {Sym: KeyF9},
		"\x1b[21~": {Sym: KeyF10},
		"\x1b[23~": {Sym: KeyF11},
		"\x1b[24~": {Sym: KeyF12},
		"\x1b[25~": {Sym: KeyF13},
		"\x1b[26~": {Sym: KeyF14},
		"\x1b[28~": {Sym: KeyF15},
		"\x1b[29~": {Sym: KeyF16},
		"\x1b[31~": {Sym: KeyF17},
		"\x1b[32~": {Sym: KeyF18},
		"\x1b[33~": {Sym: KeyF19},
		"\x1b[34~": {Sym: KeyF20},
	}

	// CSI ~ sequence keys
	csiTildeKeys := map[string]Key{
		"1": find, "2": {Sym: KeyInsert},
		"3": {Sym: KeyDelete}, "4": sel,
		"5": {Sym: KeyPgUp}, "6": {Sym: KeyPgDown},
		"7": {Sym: KeyHome}, "8": {Sym: KeyEnd},
		// There are no 9 and 10 keys
		"11": {Sym: KeyF1}, "12": {Sym: KeyF2},
		"13": {Sym: KeyF3}, "14": {Sym: KeyF4},
		"15": {Sym: KeyF5}, "17": {Sym: KeyF6},
		"18": {Sym: KeyF7}, "19": {Sym: KeyF8},
		"20": {Sym: KeyF9}, "21": {Sym: KeyF10},
		"23": {Sym: KeyF11}, "24": {Sym: KeyF12},
		"25": {Sym: KeyF13}, "26": {Sym: KeyF14},
		"28": {Sym: KeyF15}, "29": {Sym: KeyF16},
		"31": {Sym: KeyF17}, "32": {Sym: KeyF18},
		"33": {Sym: KeyF19}, "34": {Sym: KeyF20},
	}

	// URxvt keys
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	table["\x1b[a"] = Key{Sym: KeyUp, Mod: Shift}
	table["\x1b[b"] = Key{Sym: KeyDown, Mod: Shift}
	table["\x1b[c"] = Key{Sym: KeyRight, Mod: Shift}
	table["\x1b[d"] = Key{Sym: KeyLeft, Mod: Shift}
	table["\x1bOa"] = Key{Sym: KeyUp, Mod: Ctrl}
	table["\x1bOb"] = Key{Sym: KeyDown, Mod: Ctrl}
	table["\x1bOc"] = Key{Sym: KeyRight, Mod: Ctrl}
	table["\x1bOd"] = Key{Sym: KeyLeft, Mod: Ctrl}
	// TODO: invistigate if shift-ctrl arrow keys collide with DECCKM keys i.e.
	// "\x1bOA", "\x1bOB", "\x1bOC", "\x1bOD"

	// URxvt modifier CSI ~ keys
	for k, v := range csiTildeKeys {
		key := v
		// Normal (no modifier) already defined part of VT100/VT200
		// Shift modifier
		key.Mod = Shift
		table["\x1b["+k+"$"] = key
		// Ctrl modifier
		key.Mod = Ctrl
		table["\x1b["+k+"^"] = key
		// Shift-Ctrl modifier
		key.Mod = Shift | Ctrl
		table["\x1b["+k+"@"] = key
	}

	// URxvt F keys
	// Note: Shift + F1-F10 generates F11-F20.
	// This means Shift + F1 and Shift + F2 will generate F11 and F12, the same
	// applies to Ctrl + Shift F1 & F2.
	//
	// P.S. Don't like this? Blame URxvt, configure your terminal to use
	// different escapes like XTerm, or switch to a better terminal ¯\_(ツ)_/¯
	//
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	table["\x1b[23$"] = Key{Sym: KeyF11, Mod: Shift}
	table["\x1b[24$"] = Key{Sym: KeyF12, Mod: Shift}
	table["\x1b[25$"] = Key{Sym: KeyF13, Mod: Shift}
	table["\x1b[26$"] = Key{Sym: KeyF14, Mod: Shift}
	table["\x1b[28$"] = Key{Sym: KeyF15, Mod: Shift}
	table["\x1b[29$"] = Key{Sym: KeyF16, Mod: Shift}
	table["\x1b[31$"] = Key{Sym: KeyF17, Mod: Shift}
	table["\x1b[32$"] = Key{Sym: KeyF18, Mod: Shift}
	table["\x1b[33$"] = Key{Sym: KeyF19, Mod: Shift}
	table["\x1b[34$"] = Key{Sym: KeyF20, Mod: Shift}
	table["\x1b[11^"] = Key{Sym: KeyF1, Mod: Ctrl}
	table["\x1b[12^"] = Key{Sym: KeyF2, Mod: Ctrl}
	table["\x1b[13^"] = Key{Sym: KeyF3, Mod: Ctrl}
	table["\x1b[14^"] = Key{Sym: KeyF4, Mod: Ctrl}
	table["\x1b[15^"] = Key{Sym: KeyF5, Mod: Ctrl}
	table["\x1b[17^"] = Key{Sym: KeyF6, Mod: Ctrl}
	table["\x1b[18^"] = Key{Sym: KeyF7, Mod: Ctrl}
	table["\x1b[19^"] = Key{Sym: KeyF8, Mod: Ctrl}
	table["\x1b[20^"] = Key{Sym: KeyF9, Mod: Ctrl}
	table["\x1b[21^"] = Key{Sym: KeyF10, Mod: Ctrl}
	table["\x1b[23^"] = Key{Sym: KeyF11, Mod: Ctrl}
	table["\x1b[24^"] = Key{Sym: KeyF12, Mod: Ctrl}
	table["\x1b[25^"] = Key{Sym: KeyF13, Mod: Ctrl}
	table["\x1b[26^"] = Key{Sym: KeyF14, Mod: Ctrl}
	table["\x1b[28^"] = Key{Sym: KeyF15, Mod: Ctrl}
	table["\x1b[29^"] = Key{Sym: KeyF16, Mod: Ctrl}
	table["\x1b[31^"] = Key{Sym: KeyF17, Mod: Ctrl}
	table["\x1b[32^"] = Key{Sym: KeyF18, Mod: Ctrl}
	table["\x1b[33^"] = Key{Sym: KeyF19, Mod: Ctrl}
	table["\x1b[34^"] = Key{Sym: KeyF20, Mod: Ctrl}
	table["\x1b[23@"] = Key{Sym: KeyF11, Mod: Shift | Ctrl}
	table["\x1b[24@"] = Key{Sym: KeyF12, Mod: Shift | Ctrl}
	table["\x1b[25@"] = Key{Sym: KeyF13, Mod: Shift | Ctrl}
	table["\x1b[26@"] = Key{Sym: KeyF14, Mod: Shift | Ctrl}
	table["\x1b[28@"] = Key{Sym: KeyF15, Mod: Shift | Ctrl}
	table["\x1b[29@"] = Key{Sym: KeyF16, Mod: Shift | Ctrl}
	table["\x1b[31@"] = Key{Sym: KeyF17, Mod: Shift | Ctrl}
	table["\x1b[32@"] = Key{Sym: KeyF18, Mod: Shift | Ctrl}
	table["\x1b[33@"] = Key{Sym: KeyF19, Mod: Shift | Ctrl}
	table["\x1b[34@"] = Key{Sym: KeyF20, Mod: Shift | Ctrl}

	// Register Alt + <key> combinations
	// XXX: this must come after URxvt but before XTerm keys to register URxvt
	// keys with alt modifier
	tmap := map[string]Key{}
	for seq, key := range table {
		key := key
		key.Mod |= Alt
		tmap["\x1b"+seq] = key
	}
	for seq, key := range tmap {
		table[seq] = key
	}

	// XTerm modifiers
	// These are offset by 1 to be compatible with our Mod type.
	// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-PC-Style-Function-Keys
	modifiers := []KeyMod{
		Shift,                     // 1
		Alt,                       // 2
		Shift | Alt,               // 3
		Ctrl,                      // 4
		Shift | Ctrl,              // 5
		Alt | Ctrl,                // 6
		Shift | Alt | Ctrl,        // 7
		Meta,                      // 8
		Meta | Shift,              // 9
		Meta | Alt,                // 10
		Meta | Shift | Alt,        // 11
		Meta | Ctrl,               // 12
		Meta | Shift | Ctrl,       // 13
		Meta | Alt | Ctrl,         // 14
		Meta | Shift | Alt | Ctrl, // 15
	}

	// SS3 keypad function keys
	ss3FuncKeys := map[string]Key{
		// These are defined in XTerm
		// Taken from Foot keymap.h and XTerm modifyOtherKeys
		// https://codeberg.org/dnkl/foot/src/branch/master/keymap.h
		"M": {Sym: KeyKpEnter}, "X": {Sym: KeyKpEqual},
		"j": {Sym: KeyKpMultiply}, "k": {Sym: KeyKpPlus},
		"l": {Sym: KeyKpComma}, "m": {Sym: KeyKpMinus},
		"n": {Sym: KeyKpDecimal}, "o": {Sym: KeyKpDivide},
		"p": {Sym: KeyKp0}, "q": {Sym: KeyKp1},
		"r": {Sym: KeyKp2}, "s": {Sym: KeyKp3},
		"t": {Sym: KeyKp4}, "u": {Sym: KeyKp5},
		"v": {Sym: KeyKp6}, "w": {Sym: KeyKp7},
		"x": {Sym: KeyKp8}, "y": {Sym: KeyKp9},
	}

	// XTerm keys
	csiFuncKeys := map[string]Key{
		"A": {Sym: KeyUp}, "B": {Sym: KeyDown},
		"C": {Sym: KeyRight}, "D": {Sym: KeyLeft},
		"E": {Sym: KeyBegin}, "F": {Sym: KeyEnd},
		"H": {Sym: KeyHome}, "P": {Sym: KeyF1},
		"Q": {Sym: KeyF2}, "R": {Sym: KeyF3},
		"S": {Sym: KeyF4},
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	modifyOtherKeys := map[int]Key{
		ansi.BS:  {Sym: KeyBackspace},
		ansi.HT:  {Sym: KeyTab},
		ansi.CR:  {Sym: KeyEnter},
		ansi.ESC: {Sym: KeyEscape},
		ansi.DEL: {Sym: KeyBackspace},
	}

	for _, m := range modifiers {
		// XTerm modifier offset +1
		xtermMod := strconv.Itoa(int(m) + 1)

		//  CSI 1 ; <modifier> <func>
		for k, v := range csiFuncKeys {
			// Functions always have a leading 1 param
			seq := "\x1b[1;" + xtermMod + k
			key := v
			key.Mod = m
			table[seq] = key
		}
		// SS3 <modifier> <func>
		for k, v := range ss3FuncKeys {
			seq := "\x1bO" + xtermMod + k
			key := v
			key.Mod = m
			table[seq] = key
		}
		//  CSI <number> ; <modifier> ~
		for k, v := range csiTildeKeys {
			seq := "\x1b[" + k + ";" + xtermMod + "~"
			key := v
			key.Mod = m
			table[seq] = key
		}
		// CSI 27 ; <modifier> ; <code> ~
		for k, v := range modifyOtherKeys {
			code := strconv.Itoa(k)
			seq := "\x1b[27;" + xtermMod + ";" + code + "~"
			key := v
			key.Mod = m
			table[seq] = key
		}
	}

	// Register terminfo keys
	// XXX: this might override keys already registered in table
	if flags&FlagTerminfo != 0 {
		titable := buildTerminfoKeys(flags, term)
		for seq, key := range titable {
			table[seq] = key
		}
	}

	return table
}
