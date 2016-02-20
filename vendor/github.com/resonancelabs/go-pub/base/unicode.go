package base

// Returns if the rune is an Emoji, for some definition of Emoji.
// Data from http://apps.timwhitlock.info/emoji/tables/unicode, which in turn
// was generated from http://www.unicode.org/Public/UNIDATA/EmojiSources.txt
//
// Note this does _NOT_ include various other symbol-like characters (eg,
// Alchemical Symbols, Ancient Symbols, Mathematics Symbols, etc)
func IsEmoji(c rune) bool {
	// Emoticons
	if c >= 0x1f601 && c <= 0x1f64f {
		return true
	}

	// Dingbats
	if c >= 0x2702 && c <= 0x27B0 {
		return true
	}

	// Transport and map symbols
	if c >= 0x1F680 && c <= 0x1F6C0 {
		return true
	}

	// Enclosed characters
	if c >= 0x24C2 && c <= 0x1F251 {
		return true
	}

	// "Missing Emoticons"
	if c >= 0x1F600 && c <= 0x1F636 {
		return true
	}

	// Missing transport and map symbols
	if c >= 0x1F681 && c <= 0x1F6C5 {
		return true
	}

	// Other missing symbols
	if c >= 0x1F30D && c <= 0x1F567 {
		return true
	}

	// Uncategorized
	if c == 0xa9 || c == 0xae {
		return true
	}
	if c == 0x203c {
		return true
	}
	if c == 0x2049 {
		return true
	}
	if c == 0x2122 {
		return true
	}
	if c == 0x2139 {
		return true
	}
	if c >= 0x2194 && c <= 0x2199 {
		return true
	}
	if c >= 0x21a9 && c <= 0x21aa {
		return true
	}
	if c >= 0x231a && c <= 0x231b {
		return true
	}
	if c >= 0x23e9 && c <= 0x23ec {
		return true
	}
	if c == 0x23f0 || c == 0x23f3 {
		return true
	}
	if c >= 0x25aa && c <= 0x25ab {
		return true
	}
	if c == 0x25b6 {
		return true
	}
	if c == 0x25c0 {
		return true
	}
	if c >= 0x25fb && c <= 0x25fe {
		return true
	}
	if c >= 0x2600 && c <= 0x2601 {
		return true
	}
	if c == 0x260e {
		return true
	}
	if c == 0x2611 {
		return true
	}
	if c >= 0x2614 && c <= 0x2615 {
		return true
	}
	if c == 0x261d {
		return true
	}
	if c == 0x263a {
		return true
	}
	if c >= 0x2648 && c <= 0x2653 {
		return true
	}
	if c == 0x2660 {
		return true
	}
	if c == 0x2663 {
		return true
	}
	if c >= 0x2665 && c <= 0x2666 {
		return true
	}
	if c == 0x2668 {
		return true
	}
	if c == 0x267b {
		return true
	}
	if c == 0x267f {
		return true
	}
	if c == 0x2693 {
		return true
	}
	if c >= 0x26a0 && c <= 0x26a1 {
		return true
	}
	if c >= 0x26aa && c <= 0x26ab {
		return true
	}
	if c >= 0x26bd && c <= 0x26be {
		return true
	}
	if c >= 0x26c4 && c <= 0x26c5 {
		return true
	}
	if c == 0x26ce {
		return true
	}
	if c == 0x26d4 {
		return true
	}
	if c == 0x26ea {
		return true
	}
	if c >= 0x26f2 && c <= 0x26f3 {
		return true
	}
	if c == 0x26f5 {
		return true
	}
	if c == 0x26fa {
		return true
	}
	if c == 0x26fd {
		return true
	}
	if c >= 0x2934 && c <= 0x2935 {
		return true
	}
	if c >= 0x2b05 && c <= 0x2b07 {
		return true
	}
	if c >= 0x2b1b && c <= 0x2b1c {
		return true
	}
	if c == 0x2b50 {
		return true
	}
	if c == 0x2b55 {
		return true
	}
	if c == 0x3030 {
		return true
	}
	if c == 0x303d {
		return true
	}
	if c == 0x3297 {
		return true
	}
	if c == 0x3299 {
		return true
	}
	if c == 0x1f004 {
		return true
	}
	if c == 0x1f0cf {
		return true
	}
	if c >= 0x1f300 && c <= 0x1f30c {
		return true
	}
	if c == 0x1f30f {
		return true
	}
	if c == 0x1f311 {
		return true
	}
	if c >= 0x1f313 && c <= 0x1f315 {
		return true
	}
	if c == 0x1f319 {
		return true
	}
	if c == 0x1f31b {
		return true
	}
	if c >= 0x1f31f && c <= 0x1f320 {
		return true
	}
	if c >= 0x1f330 && c <= 0x1f331 {
		return true
	}
	if c >= 0x1f334 && c <= 0x1f335 {
		return true
	}
	if c >= 0x1f337 && c <= 0x1f34a {
		return true
	}
	if c >= 0x1f34c && c <= 0x1f34f {
		return true
	}
	if c >= 0x1f351 && c <= 0x1f37b {
		return true
	}
	if c >= 0x1f380 && c <= 0x1f393 {
		return true
	}
	if c >= 0x1f3a0 && c <= 0x1f3c4 {
		return true
	}
	if c == 0x1f3c6 {
		return true
	}
	if c == 0x1f3c8 {
		return true
	}
	if c == 0x1f3ca {
		return true
	}
	if c >= 0x1f3e0 && c <= 0x1f3e3 {
		return true
	}
	if c >= 0x1f3e5 && c <= 0x1f3f0 {
		return true
	}
	if c >= 0x1f40c && c <= 0x1f40e {
		return true
	}
	if c >= 0x1f411 && c <= 0x1f412 {
		return true
	}
	if c == 0x1f414 {
		return true
	}
	if c >= 0x1f417 && c <= 0x1f429 {
		return true
	}
	if c >= 0x1f42b && c <= 0x1f43e {
		return true
	}
	if c == 0x1f440 {
		return true
	}
	if c >= 0x1f442 && c <= 0x1f464 {
		return true
	}
	if c >= 0x1f466 && c <= 0x1f46b {
		return true
	}
	if c >= 0x1f46e && c <= 0x1f4ac {
		return true
	}
	if c >= 0x1f4ae && c <= 0x1f4b5 {
		return true
	}
	if c >= 0x1f4b8 && c <= 0x1f4eb {
		return true
	}
	if c == 0x1f4ee {
		return true
	}
	if c == 0x1f4f0 {
		return true
	}
	if c >= 0x1f524 && c <= 0x1f52b {
		return true
	}
	if c >= 0x1f52e && c <= 0x1f53d {
		return true
	}
	if c >= 0x1f550 && c <= 0x1f55b {
		return true
	}

	//
	// Symbols added that are not in the page above!
	// See http://www.unicode.org/charts/

	// "Miscellaneous symbols"
	if c >= 0x2600 && c <= 0x26ff {
		return true
	}
	// Miscellaneous Symbols and Pictographs
	if c >= 0x1F300 && c <= 0x1F5FF {
		return true
	}
	// Miscellaneous Symbols and Arrows
	if c >= 0x2B00 && c <= 0x2BFF {
		return true
	}

	return false
}
