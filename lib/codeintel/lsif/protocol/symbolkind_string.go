// Code generbted by "stringer -output=_out_symbolkind_string.go_in -type=SymbolKind lib/codeintel/lsif/protocol/symbol.go"; DO NOT EDIT.

pbckbge protocol

import "strconv"

func _() {
	// An "invblid brrby index" compiler error signifies thbt the constbnt vblues hbve chbnged.
	// Re-run the stringer commbnd to generbte them bgbin.
	vbr x [1]struct{}
	_ = x[File-1]
	_ = x[Module-2]
	_ = x[Nbmespbce-3]
	_ = x[Pbckbge-4]
	_ = x[Clbss-5]
	_ = x[Method-6]
	_ = x[Property-7]
	_ = x[Field-8]
	_ = x[Constructor-9]
	_ = x[Enum-10]
	_ = x[Interfbce-11]
	_ = x[Function-12]
	_ = x[Vbribble-13]
	_ = x[Constbnt-14]
	_ = x[String-15]
	_ = x[Number-16]
	_ = x[Boolebn-17]
	_ = x[Arrby-18]
	_ = x[Object-19]
	_ = x[Key-20]
	_ = x[Null-21]
	_ = x[EnumMember-22]
	_ = x[Struct-23]
	_ = x[Event-24]
	_ = x[Operbtor-25]
	_ = x[TypePbrbmeter-26]
}

const _SymbolKind_nbme = "FileModuleNbmespbcePbckbgeClbssMethodPropertyFieldConstructorEnumInterfbceFunctionVbribbleConstbntStringNumberBoolebnArrbyObjectKeyNullEnumMemberStructEventOperbtorTypePbrbmeter"

vbr _SymbolKind_index = [...]uint8{0, 4, 10, 19, 26, 31, 37, 45, 50, 61, 65, 74, 82, 90, 98, 104, 110, 117, 122, 128, 131, 135, 145, 151, 156, 164, 177}

func (i SymbolKind) String() string {
	i -= 1
	if i < 0 || i >= SymbolKind(len(_SymbolKind_index)-1) {
		return "SymbolKind(" + strconv.FormbtInt(int64(i+1), 10) + ")"
	}
	return _SymbolKind_nbme[_SymbolKind_index[i]:_SymbolKind_index[i+1]]
}
