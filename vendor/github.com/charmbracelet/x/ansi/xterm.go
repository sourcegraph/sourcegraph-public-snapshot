package ansi

// DisableModifyOtherKeys disables the modifyOtherKeys mode.
//
//	CSI > 4 ; 0 m
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
const DisableModifyOtherKeys = "\x1b[>4;0m"

// EnableModifyOtherKeys1 enables the modifyOtherKeys mode 1.
//
//	CSI > 4 ; 1 m
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
const EnableModifyOtherKeys1 = "\x1b[>4;1m"

// EnableModifyOtherKeys2 enables the modifyOtherKeys mode 2.
//
//	CSI > 4 ; 2 m
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
const EnableModifyOtherKeys2 = "\x1b[>4;2m"

// RequestModifyOtherKeys requests the modifyOtherKeys mode.
//
//	CSI ? 4  m
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
const RequestModifyOtherKeys = "\x1b[?4m"
