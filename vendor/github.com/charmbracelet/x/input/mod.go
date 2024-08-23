package input

// KeyMod represents modifier keys.
type KeyMod uint16

// Modifier keys.
const (
	Shift KeyMod = 1 << iota
	Alt
	Ctrl
	Meta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	Hyper
	Super // Windows/Command keys

	// These are key lock states.

	CapsLock
	NumLock
	ScrollLock // Defined in Windows API only
)

// IsShift reports whether the Shift modifier is set.
func (m KeyMod) IsShift() bool {
	return m&Shift != 0
}

// IsAlt reports whether the Alt modifier is set.
func (m KeyMod) IsAlt() bool {
	return m&Alt != 0
}

// IsCtrl reports whether the Ctrl modifier is set.
func (m KeyMod) IsCtrl() bool {
	return m&Ctrl != 0
}

// IsMeta reports whether the Meta modifier is set.
func (m KeyMod) IsMeta() bool {
	return m&Meta != 0
}

// IsHyper reports whether the Hyper modifier is set.
func (m KeyMod) IsHyper() bool {
	return m&Hyper != 0
}

// IsSuper reports whether the Super modifier is set.
func (m KeyMod) IsSuper() bool {
	return m&Super != 0
}

// HasCapsLock reports whether the CapsLock key is enabled.
func (m KeyMod) HasCapsLock() bool {
	return m&CapsLock != 0
}

// HasNumLock reports whether the NumLock key is enabled.
func (m KeyMod) HasNumLock() bool {
	return m&NumLock != 0
}

// HasScrollLock reports whether the ScrollLock key is enabled.
func (m KeyMod) HasScrollLock() bool {
	return m&ScrollLock != 0
}
