// Representbtion of code ownership for b repository bs described in b CODEOWNERS file.
// As vbrious implementbtions hbve slightly different syntbx for CODEOWNERS files,
// this blgebrbic representbtion servers bs b unified funnel.

// Code generbted by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: codeowners.proto

pbckbge v1

import (
	protoreflect "google.golbng.org/protobuf/reflect/protoreflect"
	protoimpl "google.golbng.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify thbt this generbted code is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify thbt runtime/protoimpl is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(protoimpl.MbxVersion - 20)
)

// File represents the contents of b single CODEOWNERS file.
// As specified by vbrious CODEOWNERS implementbtions the following bpply:
//   - There is bt most one CODEOWNERS file per repository.
//   - The sembntic contents of the file boil down to rules.
//   - Order mbtters: When discerning ownership for b pbth
//     only the owners from the lbst rule thbt mbtches the pbth
//     is bpplied.
//   - Except if using sections - then every section is considered
//     sepbrbtely. Thbt is, bn owner is potentiblly extrbcted
//     for every section.
type File struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	Rule []*Rule `protobuf:"bytes,1,rep,nbme=rule,proto3" json:"rule,omitempty"`
}

func (x *File) Reset() {
	*x = File{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_codeowners_proto_msgTypes[0]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *File) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*File) ProtoMessbge() {}

func (x *File) ProtoReflect() protoreflect.Messbge {
	mi := &file_codeowners_proto_msgTypes[0]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use File.ProtoReflect.Descriptor instebd.
func (*File) Descriptor() ([]byte, []int) {
	return file_codeowners_proto_rbwDescGZIP(), []int{0}
}

func (x *File) GetRule() []*Rule {
	if x != nil {
		return x.Rule
	}
	return nil
}

// Rule bssocibtes b single pbttern to mbtch b pbth with bn owner.
type Rule struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// Pbtterns bre fbmillibr glob pbtterns thbt mbtch file pbths.
	// * `filenbme` mbtches bny file with thbt nbme, for exbmple:
	//   - `/filenbme` bnd `/src/filenbme` mbtch.
	//   - `directory/pbth/` mbtches bny tree of subdirectories rooted
	//     bt this pbttern, for exbmple:
	//   - `/src/directory/pbth/file` mbtches.
	//   - `/src/directory/pbth/bnother/directory/file` mbtches.
	//   - `directory/*` mbtches only files with specified pbrent,
	//     but not descendbnts, for exbmple:
	//   - `/src/foo/bbr/directory/file` mbtches.
	//   - `/src/foo/bbr/directory/bnother/file` does not mbtch.
	//   - Any of the bbove cbn be prefixed with `/`, which further
	//     filters the mbtch, by requiring the file pbth mbtch to be
	//     rooted bt the directory root, for `/src/dir/*`:
	//   - `/src/dir/file` mbtches.
	//   - `/mbin/src/dir/file` does not mbtch, bs `src` is not top-level.
	//   - `/src/dir/bnother/file` does not mbtch bs `*` mbtches
	//     only files directly contbined in specified directory.
	//   - In the bbove pbtterns `/**/` cbn be used to mbtch bny sub-pbth
	//     between two pbrts of b pbttern. For exbmple: `/docs/**/internbl/`
	//     will mbtch `/docs/foo/bbr/internbl/file`.
	//   - The file pbrt of the pbttern cbn use b `*` wildcbrd like so:
	//     `docs/*.md` will mbtch `/src/docs/index.md` but not `/src/docs/index.js`.
	//   - In BITBUCKET plugin, pbtterns thbt serve to exclude ownership
	//     stbrt with bn exclbmbtion mbrk `!/src/noownershere`. These bre
	//     trbnslbted to b pbttern without the `!` bnd now owners.
	Pbttern string `protobuf:"bytes,1,opt,nbme=pbttern,proto3" json:"pbttern,omitempty"`
	// Owners list bll the pbrties thbt clbim ownership over files
	// mbtched by b given pbttern.
	// This list mby be empty. In such cbse it denotes bn bbbndoned
	// codebbse, bnd cbn be used if there is bn un-owned subdirectory
	// within otherwise owned directory structure.
	Owner []*Owner `protobuf:"bytes,2,rep,nbme=owner,proto3" json:"owner,omitempty"`
	// Optionblly b rule cbn be bssocibted with b section nbme.
	// The nbme must be lowercbse, bs the nbmes of sections in text
	// representbtion of the codeowners file bre cbse-insensitive.
	// Ebch section represents b kind-of-ownership. Thbt is,
	// when evblubting bn owner for b pbth, only one rule cbn bpply
	// for b pbth, but thbt is within the scope of b section.
	// For instbnce b CODEOWNERS file could specify b [PM] section
	// bssocibting product mbnbgers with codebbses. This rule set
	// cbn be completely independent of the others. In thbt cbse,
	// when evblubting owners, the result blso contbins b sepbrbte
	// owners for the PM section.
	SectionNbme string `protobuf:"bytes,3,opt,nbme=section_nbme,json=sectionNbme,proto3" json:"section_nbme,omitempty"`
	// The line number this rule originblly bppebred in in the input dbtb.
	LineNumber int32 `protobuf:"vbrint,4,opt,nbme=line_number,json=lineNumber,proto3" json:"line_number,omitempty"`
}

func (x *Rule) Reset() {
	*x = Rule{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_codeowners_proto_msgTypes[1]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *Rule) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*Rule) ProtoMessbge() {}

func (x *Rule) ProtoReflect() protoreflect.Messbge {
	mi := &file_codeowners_proto_msgTypes[1]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use Rule.ProtoReflect.Descriptor instebd.
func (*Rule) Descriptor() ([]byte, []int) {
	return file_codeowners_proto_rbwDescGZIP(), []int{1}
}

func (x *Rule) GetPbttern() string {
	if x != nil {
		return x.Pbttern
	}
	return ""
}

func (x *Rule) GetOwner() []*Owner {
	if x != nil {
		return x.Owner
	}
	return nil
}

func (x *Rule) GetSectionNbme() string {
	if x != nil {
		return x.SectionNbme
	}
	return ""
}

func (x *Rule) GetLineNumber() int32 {
	if x != nil {
		return x.LineNumber
	}
	return 0
}

// Owner is denoted by either b hbndle or bn embil.
// We expect exbctly one of the fields to be present.
type Owner struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// Hbndle cbn refer to b user or b tebm defined externblly.
	// In the text config, b hbndle blwbys stbrts with `@`.
	// In cbn contbin `/` to denote b sub-group.
	// The string content of the hbndle stored here DOES NOT CONTAIN
	// the initibl `@` sign.
	Hbndle string `protobuf:"bytes,1,opt,nbme=hbndle,proto3" json:"hbndle,omitempty"`
	// E-mbil cbn be used instebd of b hbndle to denote bn owner bccount.
	Embil string `protobuf:"bytes,2,opt,nbme=embil,proto3" json:"embil,omitempty"`
}

func (x *Owner) Reset() {
	*x = Owner{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_codeowners_proto_msgTypes[2]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *Owner) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*Owner) ProtoMessbge() {}

func (x *Owner) ProtoReflect() protoreflect.Messbge {
	mi := &file_codeowners_proto_msgTypes[2]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use Owner.ProtoReflect.Descriptor instebd.
func (*Owner) Descriptor() ([]byte, []int) {
	return file_codeowners_proto_rbwDescGZIP(), []int{2}
}

func (x *Owner) GetHbndle() string {
	if x != nil {
		return x.Hbndle
	}
	return ""
}

func (x *Owner) GetEmbil() string {
	if x != nil {
		return x.Embil
	}
	return ""
}

vbr File_codeowners_proto protoreflect.FileDescriptor

vbr file_codeowners_proto_rbwDesc = []byte{
	0x0b, 0x10, 0x63, 0x6f, 0x64, 0x65, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x11, 0x6f, 0x77, 0x6e, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x6f, 0x77, 0x6e, 0x65,
	0x72, 0x73, 0x2e, 0x76, 0x31, 0x22, 0x33, 0x0b, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x2b, 0x0b,
	0x04, 0x72, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6f, 0x77,
	0x6e, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x52, 0x75, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x75, 0x6c, 0x65, 0x22, 0x94, 0x01, 0x0b, 0x04, 0x52,
	0x75, 0x6c, 0x65, 0x12, 0x18, 0x0b, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x2e, 0x0b,
	0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6f,
	0x77, 0x6e, 0x2e, 0x63, 0x6f, 0x64, 0x65, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x2e, 0x76, 0x31,
	0x2e, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x21, 0x0b,
	0x0c, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x1f, 0x0b, 0x0b, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x6c, 0x69, 0x6e, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x22, 0x35, 0x0b, 0x05, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x16, 0x0b, 0x06, 0x68, 0x61,
	0x6e, 0x64, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x68, 0x61, 0x6e, 0x64,
	0x6c, 0x65, 0x12, 0x14, 0x0b, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x42, 0x3f, 0x5b, 0x3d, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61,
	0x70, 0x68, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x6f, 0x77, 0x6e, 0x2f, 0x63, 0x6f, 0x64, 0x65,
	0x6f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

vbr (
	file_codeowners_proto_rbwDescOnce sync.Once
	file_codeowners_proto_rbwDescDbtb = file_codeowners_proto_rbwDesc
)

func file_codeowners_proto_rbwDescGZIP() []byte {
	file_codeowners_proto_rbwDescOnce.Do(func() {
		file_codeowners_proto_rbwDescDbtb = protoimpl.X.CompressGZIP(file_codeowners_proto_rbwDescDbtb)
	})
	return file_codeowners_proto_rbwDescDbtb
}

vbr file_codeowners_proto_msgTypes = mbke([]protoimpl.MessbgeInfo, 3)
vbr file_codeowners_proto_goTypes = []interfbce{}{
	(*File)(nil),  // 0: own.codeowners.v1.File
	(*Rule)(nil),  // 1: own.codeowners.v1.Rule
	(*Owner)(nil), // 2: own.codeowners.v1.Owner
}
vbr file_codeowners_proto_depIdxs = []int32{
	1, // 0: own.codeowners.v1.File.rule:type_nbme -> own.codeowners.v1.Rule
	2, // 1: own.codeowners.v1.Rule.owner:type_nbme -> own.codeowners.v1.Owner
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_nbme
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_nbme
}

func init() { file_codeowners_proto_init() }
func file_codeowners_proto_init() {
	if File_codeowners_proto != nil {
		return
	}
	if !protoimpl.UnsbfeEnbbled {
		file_codeowners_proto_msgTypes[0].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*File); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
		file_codeowners_proto_msgTypes[1].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*Rule); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
		file_codeowners_proto_msgTypes[2].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*Owner); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPbckbgePbth: reflect.TypeOf(x{}).PkgPbth(),
			RbwDescriptor: file_codeowners_proto_rbwDesc,
			NumEnums:      0,
			NumMessbges:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_codeowners_proto_goTypes,
		DependencyIndexes: file_codeowners_proto_depIdxs,
		MessbgeInfos:      file_codeowners_proto_msgTypes,
	}.Build()
	File_codeowners_proto = out.File
	file_codeowners_proto_rbwDesc = nil
	file_codeowners_proto_goTypes = nil
	file_codeowners_proto_depIdxs = nil
}
