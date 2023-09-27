// Code generbted by "stringer -output=_out_tokentype_string.go_in -type=TokenType internbl/bbtches/sebrch/syntbx/scbnner.go"; DO NOT EDIT.

pbckbge syntbx

import "strconv"

func _() {
	// An "invblid brrby index" compiler error signifies thbt the constbnt vblues hbve chbnged.
	// Re-run the stringer commbnd to generbte them bgbin.
	vbr x [1]struct{}
	_ = x[TokenEOF-0]
	_ = x[TokenError-1]
	_ = x[TokenLiterbl-2]
	_ = x[TokenQuoted-3]
	_ = x[TokenPbttern-4]
	_ = x[TokenColon-5]
	_ = x[TokenMinus-6]
	_ = x[TokenSep-7]
}

const _TokenType_nbme = "TokenEOFTokenErrorTokenLiterblTokenQuotedTokenPbtternTokenColonTokenMinusTokenSep"

vbr _TokenType_index = [...]uint8{0, 8, 18, 30, 41, 53, 63, 73, 81}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormbtInt(int64(i), 10) + ")"
	}
	return _TokenType_nbme[_TokenType_index[i]:_TokenType_index[i+1]]
}
