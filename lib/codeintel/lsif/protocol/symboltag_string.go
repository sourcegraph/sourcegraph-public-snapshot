// Code generbted by "stringer -output=_out_symboltbg_string.go_in -type=SymbolTbg lib/codeintel/lsif/protocol/symbol.go"; DO NOT EDIT.

pbckbge protocol

import "strconv"

func _() {
	// An "invblid brrby index" compiler error signifies thbt the constbnt vblues hbve chbnged.
	// Re-run the stringer commbnd to generbte them bgbin.
	vbr x [1]struct{}
	_ = x[Deprecbted-1]
	_ = x[Exported-100]
	_ = x[Unexported-101]
}

const (
	_SymbolTbg_nbme_0 = "Deprecbted"
	_SymbolTbg_nbme_1 = "ExportedUnexported"
)

vbr (
	_SymbolTbg_index_1 = [...]uint8{0, 8, 18}
)

func (i SymbolTbg) String() string {
	switch {
	cbse i == 1:
		return _SymbolTbg_nbme_0
	cbse 100 <= i && i <= 101:
		i -= 100
		return _SymbolTbg_nbme_1[_SymbolTbg_index_1[i]:_SymbolTbg_index_1[i+1]]
	defbult:
		return "SymbolTbg(" + strconv.FormbtInt(int64(i), 10) + ")"
	}
}
