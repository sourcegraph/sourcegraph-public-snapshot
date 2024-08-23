package input

import (
	"strings"

	"github.com/xo/terminfo"
)

func buildTerminfoKeys(flags int, term string) map[string]Key {
	table := make(map[string]Key)
	ti, _ := terminfo.Load(term)
	if ti == nil {
		return table
	}

	tiTable := defaultTerminfoKeys(flags)

	// Default keys
	for name, seq := range ti.StringCapsShort() {
		if !strings.HasPrefix(name, "k") || len(seq) == 0 {
			continue
		}

		if k, ok := tiTable[name]; ok {
			table[string(seq)] = k
		}
	}

	// Extended keys
	for name, seq := range ti.ExtStringCapsShort() {
		if !strings.HasPrefix(name, "k") || len(seq) == 0 {
			continue
		}

		if k, ok := tiTable[name]; ok {
			table[string(seq)] = k
		}
	}

	return table
}

// This returns a map of terminfo keys to key events. It's a mix of ncurses
// terminfo default and user-defined key capabilities.
// Upper-case caps that are defined in the default terminfo database are
//   - kNXT
//   - kPRV
//   - kHOM
//   - kEND
//   - kDC
//   - kIC
//   - kLFT
//   - kRIT
//
// See https://man7.org/linux/man-pages/man5/terminfo.5.html
// See https://github.com/mirror/ncurses/blob/master/include/Caps-ncurses
func defaultTerminfoKeys(flags int) map[string]Key {
	keys := map[string]Key{
		"kcuu1": {Sym: KeyUp},
		"kUP":   {Sym: KeyUp, Mod: Shift},
		"kUP3":  {Sym: KeyUp, Mod: Alt},
		"kUP4":  {Sym: KeyUp, Mod: Shift | Alt},
		"kUP5":  {Sym: KeyUp, Mod: Ctrl},
		"kUP6":  {Sym: KeyUp, Mod: Shift | Ctrl},
		"kUP7":  {Sym: KeyUp, Mod: Alt | Ctrl},
		"kUP8":  {Sym: KeyUp, Mod: Shift | Alt | Ctrl},
		"kcud1": {Sym: KeyDown},
		"kDN":   {Sym: KeyDown, Mod: Shift},
		"kDN3":  {Sym: KeyDown, Mod: Alt},
		"kDN4":  {Sym: KeyDown, Mod: Shift | Alt},
		"kDN5":  {Sym: KeyDown, Mod: Ctrl},
		"kDN7":  {Sym: KeyDown, Mod: Alt | Ctrl},
		"kDN6":  {Sym: KeyDown, Mod: Shift | Ctrl},
		"kDN8":  {Sym: KeyDown, Mod: Shift | Alt | Ctrl},
		"kcub1": {Sym: KeyLeft},
		"kLFT":  {Sym: KeyLeft, Mod: Shift},
		"kLFT3": {Sym: KeyLeft, Mod: Alt},
		"kLFT4": {Sym: KeyLeft, Mod: Shift | Alt},
		"kLFT5": {Sym: KeyLeft, Mod: Ctrl},
		"kLFT6": {Sym: KeyLeft, Mod: Shift | Ctrl},
		"kLFT7": {Sym: KeyLeft, Mod: Alt | Ctrl},
		"kLFT8": {Sym: KeyLeft, Mod: Shift | Alt | Ctrl},
		"kcuf1": {Sym: KeyRight},
		"kRIT":  {Sym: KeyRight, Mod: Shift},
		"kRIT3": {Sym: KeyRight, Mod: Alt},
		"kRIT4": {Sym: KeyRight, Mod: Shift | Alt},
		"kRIT5": {Sym: KeyRight, Mod: Ctrl},
		"kRIT6": {Sym: KeyRight, Mod: Shift | Ctrl},
		"kRIT7": {Sym: KeyRight, Mod: Alt | Ctrl},
		"kRIT8": {Sym: KeyRight, Mod: Shift | Alt | Ctrl},
		"kich1": {Sym: KeyInsert},
		"kIC":   {Sym: KeyInsert, Mod: Shift},
		"kIC3":  {Sym: KeyInsert, Mod: Alt},
		"kIC4":  {Sym: KeyInsert, Mod: Shift | Alt},
		"kIC5":  {Sym: KeyInsert, Mod: Ctrl},
		"kIC6":  {Sym: KeyInsert, Mod: Shift | Ctrl},
		"kIC7":  {Sym: KeyInsert, Mod: Alt | Ctrl},
		"kIC8":  {Sym: KeyInsert, Mod: Shift | Alt | Ctrl},
		"kdch1": {Sym: KeyDelete},
		"kDC":   {Sym: KeyDelete, Mod: Shift},
		"kDC3":  {Sym: KeyDelete, Mod: Alt},
		"kDC4":  {Sym: KeyDelete, Mod: Shift | Alt},
		"kDC5":  {Sym: KeyDelete, Mod: Ctrl},
		"kDC6":  {Sym: KeyDelete, Mod: Shift | Ctrl},
		"kDC7":  {Sym: KeyDelete, Mod: Alt | Ctrl},
		"kDC8":  {Sym: KeyDelete, Mod: Shift | Alt | Ctrl},
		"khome": {Sym: KeyHome},
		"kHOM":  {Sym: KeyHome, Mod: Shift},
		"kHOM3": {Sym: KeyHome, Mod: Alt},
		"kHOM4": {Sym: KeyHome, Mod: Shift | Alt},
		"kHOM5": {Sym: KeyHome, Mod: Ctrl},
		"kHOM6": {Sym: KeyHome, Mod: Shift | Ctrl},
		"kHOM7": {Sym: KeyHome, Mod: Alt | Ctrl},
		"kHOM8": {Sym: KeyHome, Mod: Shift | Alt | Ctrl},
		"kend":  {Sym: KeyEnd},
		"kEND":  {Sym: KeyEnd, Mod: Shift},
		"kEND3": {Sym: KeyEnd, Mod: Alt},
		"kEND4": {Sym: KeyEnd, Mod: Shift | Alt},
		"kEND5": {Sym: KeyEnd, Mod: Ctrl},
		"kEND6": {Sym: KeyEnd, Mod: Shift | Ctrl},
		"kEND7": {Sym: KeyEnd, Mod: Alt | Ctrl},
		"kEND8": {Sym: KeyEnd, Mod: Shift | Alt | Ctrl},
		"kpp":   {Sym: KeyPgUp},
		"kprv":  {Sym: KeyPgUp},
		"kPRV":  {Sym: KeyPgUp, Mod: Shift},
		"kPRV3": {Sym: KeyPgUp, Mod: Alt},
		"kPRV4": {Sym: KeyPgUp, Mod: Shift | Alt},
		"kPRV5": {Sym: KeyPgUp, Mod: Ctrl},
		"kPRV6": {Sym: KeyPgUp, Mod: Shift | Ctrl},
		"kPRV7": {Sym: KeyPgUp, Mod: Alt | Ctrl},
		"kPRV8": {Sym: KeyPgUp, Mod: Shift | Alt | Ctrl},
		"knp":   {Sym: KeyPgDown},
		"knxt":  {Sym: KeyPgDown},
		"kNXT":  {Sym: KeyPgDown, Mod: Shift},
		"kNXT3": {Sym: KeyPgDown, Mod: Alt},
		"kNXT4": {Sym: KeyPgDown, Mod: Shift | Alt},
		"kNXT5": {Sym: KeyPgDown, Mod: Ctrl},
		"kNXT6": {Sym: KeyPgDown, Mod: Shift | Ctrl},
		"kNXT7": {Sym: KeyPgDown, Mod: Alt | Ctrl},
		"kNXT8": {Sym: KeyPgDown, Mod: Shift | Alt | Ctrl},

		"kbs":  {Sym: KeyBackspace},
		"kcbt": {Sym: KeyTab, Mod: Shift},

		// Function keys
		// This only includes the first 12 function keys. The rest are treated
		// as modifiers of the first 12.
		// Take a look at XTerm modifyFunctionKeys
		//
		// XXX: To use unambiguous function keys, use fixterms or kitty clipboard.
		//
		// See https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyFunctionKeys
		// See https://invisible-island.net/xterm/terminfo.html

		"kf1":  {Sym: KeyF1},
		"kf2":  {Sym: KeyF2},
		"kf3":  {Sym: KeyF3},
		"kf4":  {Sym: KeyF4},
		"kf5":  {Sym: KeyF5},
		"kf6":  {Sym: KeyF6},
		"kf7":  {Sym: KeyF7},
		"kf8":  {Sym: KeyF8},
		"kf9":  {Sym: KeyF9},
		"kf10": {Sym: KeyF10},
		"kf11": {Sym: KeyF11},
		"kf12": {Sym: KeyF12},
		"kf13": {Sym: KeyF1, Mod: Shift},
		"kf14": {Sym: KeyF2, Mod: Shift},
		"kf15": {Sym: KeyF3, Mod: Shift},
		"kf16": {Sym: KeyF4, Mod: Shift},
		"kf17": {Sym: KeyF5, Mod: Shift},
		"kf18": {Sym: KeyF6, Mod: Shift},
		"kf19": {Sym: KeyF7, Mod: Shift},
		"kf20": {Sym: KeyF8, Mod: Shift},
		"kf21": {Sym: KeyF9, Mod: Shift},
		"kf22": {Sym: KeyF10, Mod: Shift},
		"kf23": {Sym: KeyF11, Mod: Shift},
		"kf24": {Sym: KeyF12, Mod: Shift},
		"kf25": {Sym: KeyF1, Mod: Ctrl},
		"kf26": {Sym: KeyF2, Mod: Ctrl},
		"kf27": {Sym: KeyF3, Mod: Ctrl},
		"kf28": {Sym: KeyF4, Mod: Ctrl},
		"kf29": {Sym: KeyF5, Mod: Ctrl},
		"kf30": {Sym: KeyF6, Mod: Ctrl},
		"kf31": {Sym: KeyF7, Mod: Ctrl},
		"kf32": {Sym: KeyF8, Mod: Ctrl},
		"kf33": {Sym: KeyF9, Mod: Ctrl},
		"kf34": {Sym: KeyF10, Mod: Ctrl},
		"kf35": {Sym: KeyF11, Mod: Ctrl},
		"kf36": {Sym: KeyF12, Mod: Ctrl},
		"kf37": {Sym: KeyF1, Mod: Shift | Ctrl},
		"kf38": {Sym: KeyF2, Mod: Shift | Ctrl},
		"kf39": {Sym: KeyF3, Mod: Shift | Ctrl},
		"kf40": {Sym: KeyF4, Mod: Shift | Ctrl},
		"kf41": {Sym: KeyF5, Mod: Shift | Ctrl},
		"kf42": {Sym: KeyF6, Mod: Shift | Ctrl},
		"kf43": {Sym: KeyF7, Mod: Shift | Ctrl},
		"kf44": {Sym: KeyF8, Mod: Shift | Ctrl},
		"kf45": {Sym: KeyF9, Mod: Shift | Ctrl},
		"kf46": {Sym: KeyF10, Mod: Shift | Ctrl},
		"kf47": {Sym: KeyF11, Mod: Shift | Ctrl},
		"kf48": {Sym: KeyF12, Mod: Shift | Ctrl},
		"kf49": {Sym: KeyF1, Mod: Alt},
		"kf50": {Sym: KeyF2, Mod: Alt},
		"kf51": {Sym: KeyF3, Mod: Alt},
		"kf52": {Sym: KeyF4, Mod: Alt},
		"kf53": {Sym: KeyF5, Mod: Alt},
		"kf54": {Sym: KeyF6, Mod: Alt},
		"kf55": {Sym: KeyF7, Mod: Alt},
		"kf56": {Sym: KeyF8, Mod: Alt},
		"kf57": {Sym: KeyF9, Mod: Alt},
		"kf58": {Sym: KeyF10, Mod: Alt},
		"kf59": {Sym: KeyF11, Mod: Alt},
		"kf60": {Sym: KeyF12, Mod: Alt},
		"kf61": {Sym: KeyF1, Mod: Shift | Alt},
		"kf62": {Sym: KeyF2, Mod: Shift | Alt},
		"kf63": {Sym: KeyF3, Mod: Shift | Alt},
	}

	// Preserve F keys from F13 to F63 instead of using them for F-keys
	// modifiers.
	if flags&FlagFKeys != 0 {
		keys["kf13"] = Key{Sym: KeyF13}
		keys["kf14"] = Key{Sym: KeyF14}
		keys["kf15"] = Key{Sym: KeyF15}
		keys["kf16"] = Key{Sym: KeyF16}
		keys["kf17"] = Key{Sym: KeyF17}
		keys["kf18"] = Key{Sym: KeyF18}
		keys["kf19"] = Key{Sym: KeyF19}
		keys["kf20"] = Key{Sym: KeyF20}
		keys["kf21"] = Key{Sym: KeyF21}
		keys["kf22"] = Key{Sym: KeyF22}
		keys["kf23"] = Key{Sym: KeyF23}
		keys["kf24"] = Key{Sym: KeyF24}
		keys["kf25"] = Key{Sym: KeyF25}
		keys["kf26"] = Key{Sym: KeyF26}
		keys["kf27"] = Key{Sym: KeyF27}
		keys["kf28"] = Key{Sym: KeyF28}
		keys["kf29"] = Key{Sym: KeyF29}
		keys["kf30"] = Key{Sym: KeyF30}
		keys["kf31"] = Key{Sym: KeyF31}
		keys["kf32"] = Key{Sym: KeyF32}
		keys["kf33"] = Key{Sym: KeyF33}
		keys["kf34"] = Key{Sym: KeyF34}
		keys["kf35"] = Key{Sym: KeyF35}
		keys["kf36"] = Key{Sym: KeyF36}
		keys["kf37"] = Key{Sym: KeyF37}
		keys["kf38"] = Key{Sym: KeyF38}
		keys["kf39"] = Key{Sym: KeyF39}
		keys["kf40"] = Key{Sym: KeyF40}
		keys["kf41"] = Key{Sym: KeyF41}
		keys["kf42"] = Key{Sym: KeyF42}
		keys["kf43"] = Key{Sym: KeyF43}
		keys["kf44"] = Key{Sym: KeyF44}
		keys["kf45"] = Key{Sym: KeyF45}
		keys["kf46"] = Key{Sym: KeyF46}
		keys["kf47"] = Key{Sym: KeyF47}
		keys["kf48"] = Key{Sym: KeyF48}
		keys["kf49"] = Key{Sym: KeyF49}
		keys["kf50"] = Key{Sym: KeyF50}
		keys["kf51"] = Key{Sym: KeyF51}
		keys["kf52"] = Key{Sym: KeyF52}
		keys["kf53"] = Key{Sym: KeyF53}
		keys["kf54"] = Key{Sym: KeyF54}
		keys["kf55"] = Key{Sym: KeyF55}
		keys["kf56"] = Key{Sym: KeyF56}
		keys["kf57"] = Key{Sym: KeyF57}
		keys["kf58"] = Key{Sym: KeyF58}
		keys["kf59"] = Key{Sym: KeyF59}
		keys["kf60"] = Key{Sym: KeyF60}
		keys["kf61"] = Key{Sym: KeyF61}
		keys["kf62"] = Key{Sym: KeyF62}
		keys["kf63"] = Key{Sym: KeyF63}
	}

	return keys
}
