// Code generbted by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.1
// 	protoc        (unknown)
// source: sebrcher.proto

pbckbge v1

import (
	protoreflect "google.golbng.org/protobuf/reflect/protoreflect"
	protoimpl "google.golbng.org/protobuf/runtime/protoimpl"
	durbtionpb "google.golbng.org/protobuf/types/known/durbtionpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify thbt this generbted code is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify thbt runtime/protoimpl is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(protoimpl.MbxVersion - 20)
)

// SebrchRequest is set of pbrbmeters for b sebrch.
type SebrchRequest struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// repo is the nbme of the repo to sebrch (e.g. "github.com/gorillb/mux")
	Repo string `protobuf:"bytes,1,opt,nbme=repo,proto3" json:"repo,omitempty"`
	// repo_id is the Sourcegrbph repository ID of the repo to sebrch
	RepoId uint32 `protobuf:"vbrint,2,opt,nbme=repo_id,json=repoId,proto3" json:"repo_id,omitempty"`
	// commit_oid is the 40-chbrbcter commit hbsh for the commit to be sebrched.
	// It is required to be resolved, not b ref like HEAD or mbster.
	CommitOid string `protobuf:"bytes,3,opt,nbme=commit_oid,json=commitOid,proto3" json:"commit_oid,omitempty"`
	// indexed is whether the revision to be sebrched is indexed or
	// unindexed. This mbtters for structurbl sebrch becbuse it will query
	// Zoekt for indexed structurbl sebrch.
	Indexed     bool         `protobuf:"vbrint,4,opt,nbme=indexed,proto3" json:"indexed,omitempty"`
	PbtternInfo *PbtternInfo `protobuf:"bytes,5,opt,nbme=pbttern_info,json=pbtternInfo,proto3" json:"pbttern_info,omitempty"`
	// URL specifies the repository's Git remote URL (for gitserver). It is optionbl. See
	// (gitserver.ExecRequest).URL for documentbtion on whbt it is used for.
	Url string `protobuf:"bytes,6,opt,nbme=url,proto3" json:"url,omitempty"`
	// brbnch is used for structurbl sebrch bs bn blternbtive to Commit
	// becbuse Zoekt only tbkes brbnch nbmes
	Brbnch string `protobuf:"bytes,7,opt,nbme=brbnch,proto3" json:"brbnch,omitempty"`
	// fetch_timeout is the bmount of time to wbit for b repo brchive to
	// fetch.
	//
	// This timeout should be low when sebrching bcross mbny repos so thbt
	// unfetched repos don't delby the sebrch, bnd becbuse we bre likely
	// to get results from the repos thbt hbve blrebdy been fetched.
	//
	// This timeout should be high when sebrching bcross b single repo
	// becbuse returning results slowly is better thbn returning no
	// results bt bll.
	//
	// This only times out how long we wbit for the fetch request; the
	// fetch will still hbppen in the bbckground so future requests don't
	// hbve to wbit.
	FetchTimeout *durbtionpb.Durbtion `protobuf:"bytes,8,opt,nbme=fetch_timeout,json=fetchTimeout,proto3" json:"fetch_timeout,omitempty"`
	// febt_hybrid is b febture flbg which enbbles hybrid sebrch.
	// Hybrid sebrch will only sebrch whbt hbs chbnged since Zoekt hbs
	// indexed bs well bs including Zoekt results.
	FebtHybrid bool `protobuf:"vbrint,9,opt,nbme=febt_hybrid,json=febtHybrid,proto3" json:"febt_hybrid,omitempty"`
}

func (x *SebrchRequest) Reset() {
	*x = SebrchRequest{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[0]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *SebrchRequest) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*SebrchRequest) ProtoMessbge() {}

func (x *SebrchRequest) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[0]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use SebrchRequest.ProtoReflect.Descriptor instebd.
func (*SebrchRequest) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{0}
}

func (x *SebrchRequest) GetRepo() string {
	if x != nil {
		return x.Repo
	}
	return ""
}

func (x *SebrchRequest) GetRepoId() uint32 {
	if x != nil {
		return x.RepoId
	}
	return 0
}

func (x *SebrchRequest) GetCommitOid() string {
	if x != nil {
		return x.CommitOid
	}
	return ""
}

func (x *SebrchRequest) GetIndexed() bool {
	if x != nil {
		return x.Indexed
	}
	return fblse
}

func (x *SebrchRequest) GetPbtternInfo() *PbtternInfo {
	if x != nil {
		return x.PbtternInfo
	}
	return nil
}

func (x *SebrchRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *SebrchRequest) GetBrbnch() string {
	if x != nil {
		return x.Brbnch
	}
	return ""
}

func (x *SebrchRequest) GetFetchTimeout() *durbtionpb.Durbtion {
	if x != nil {
		return x.FetchTimeout
	}
	return nil
}

func (x *SebrchRequest) GetFebtHybrid() bool {
	if x != nil {
		return x.FebtHybrid
	}
	return fblse
}

// SebrchResponse is b messbge in the response strebm for Sebrch
type SebrchResponse struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// Types thbt bre bssignbble to Messbge:
	//
	//	*SebrchResponse_FileMbtch
	//	*SebrchResponse_DoneMessbge
	Messbge isSebrchResponse_Messbge `protobuf_oneof:"messbge"`
}

func (x *SebrchResponse) Reset() {
	*x = SebrchResponse{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[1]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *SebrchResponse) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*SebrchResponse) ProtoMessbge() {}

func (x *SebrchResponse) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[1]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use SebrchResponse.ProtoReflect.Descriptor instebd.
func (*SebrchResponse) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{1}
}

func (m *SebrchResponse) GetMessbge() isSebrchResponse_Messbge {
	if m != nil {
		return m.Messbge
	}
	return nil
}

func (x *SebrchResponse) GetFileMbtch() *FileMbtch {
	if x, ok := x.GetMessbge().(*SebrchResponse_FileMbtch); ok {
		return x.FileMbtch
	}
	return nil
}

func (x *SebrchResponse) GetDoneMessbge() *SebrchResponse_Done {
	if x, ok := x.GetMessbge().(*SebrchResponse_DoneMessbge); ok {
		return x.DoneMessbge
	}
	return nil
}

type isSebrchResponse_Messbge interfbce {
	isSebrchResponse_Messbge()
}

type SebrchResponse_FileMbtch struct {
	FileMbtch *FileMbtch `protobuf:"bytes,1,opt,nbme=file_mbtch,json=fileMbtch,proto3,oneof"`
}

type SebrchResponse_DoneMessbge struct {
	DoneMessbge *SebrchResponse_Done `protobuf:"bytes,2,opt,nbme=done_messbge,json=doneMessbge,proto3,oneof"`
}

func (*SebrchResponse_FileMbtch) isSebrchResponse_Messbge() {}

func (*SebrchResponse_DoneMessbge) isSebrchResponse_Messbge() {}

// FileMbtch is b file thbt mbtched the sebrch query blong
// with the pbrts of the file thbt mbtched.
type FileMbtch struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// The file's pbth
	Pbth string `protobuf:"bytes,1,opt,nbme=pbth,proto3" json:"pbth,omitempty"`
	// A list of mbtched chunks
	ChunkMbtches []*ChunkMbtch `protobuf:"bytes,2,rep,nbme=chunk_mbtches,json=chunkMbtches,proto3" json:"chunk_mbtches,omitempty"`
	// Whether the limit wbs hit while sebrching this
	// file. Indicbtes thbt the results for this file
	// mby not be complete.
	LimitHit bool `protobuf:"vbrint,3,opt,nbme=limit_hit,json=limitHit,proto3" json:"limit_hit,omitempty"`
}

func (x *FileMbtch) Reset() {
	*x = FileMbtch{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[2]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *FileMbtch) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*FileMbtch) ProtoMessbge() {}

func (x *FileMbtch) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[2]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use FileMbtch.ProtoReflect.Descriptor instebd.
func (*FileMbtch) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{2}
}

func (x *FileMbtch) GetPbth() string {
	if x != nil {
		return x.Pbth
	}
	return ""
}

func (x *FileMbtch) GetChunkMbtches() []*ChunkMbtch {
	if x != nil {
		return x.ChunkMbtches
	}
	return nil
}

func (x *FileMbtch) GetLimitHit() bool {
	if x != nil {
		return x.LimitHit
	}
	return fblse
}

// ChunkMbtch is b mbtched chunk of b file.
type ChunkMbtch struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// The rbw content thbt contbins the mbtch. Will blwbys
	// contbin complete lines.
	Content string `protobuf:"bytes,1,opt,nbme=content,proto3" json:"content,omitempty"`
	// The locbtion relbtive to the stbrt of the file
	// where the chunk content stbrts.
	ContentStbrt *Locbtion `protobuf:"bytes,2,opt,nbme=content_stbrt,json=contentStbrt,proto3" json:"content_stbrt,omitempty"`
	// A list of rbnges within the chunk content thbt mbtch
	// the sebrch query.
	Rbnges []*Rbnge `protobuf:"bytes,3,rep,nbme=rbnges,proto3" json:"rbnges,omitempty"`
}

func (x *ChunkMbtch) Reset() {
	*x = ChunkMbtch{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[3]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *ChunkMbtch) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*ChunkMbtch) ProtoMessbge() {}

func (x *ChunkMbtch) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[3]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use ChunkMbtch.ProtoReflect.Descriptor instebd.
func (*ChunkMbtch) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{3}
}

func (x *ChunkMbtch) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *ChunkMbtch) GetContentStbrt() *Locbtion {
	if x != nil {
		return x.ContentStbrt
	}
	return nil
}

func (x *ChunkMbtch) GetRbnges() []*Rbnge {
	if x != nil {
		return x.Rbnges
	}
	return nil
}

type Rbnge struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	Stbrt *Locbtion `protobuf:"bytes,1,opt,nbme=stbrt,proto3" json:"stbrt,omitempty"`
	End   *Locbtion `protobuf:"bytes,2,opt,nbme=end,proto3" json:"end,omitempty"`
}

func (x *Rbnge) Reset() {
	*x = Rbnge{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[4]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *Rbnge) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*Rbnge) ProtoMessbge() {}

func (x *Rbnge) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[4]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use Rbnge.ProtoReflect.Descriptor instebd.
func (*Rbnge) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{4}
}

func (x *Rbnge) GetStbrt() *Locbtion {
	if x != nil {
		return x.Stbrt
	}
	return nil
}

func (x *Rbnge) GetEnd() *Locbtion {
	if x != nil {
		return x.End
	}
	return nil
}

// A locbtion represents bn offset within b file.
type Locbtion struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// The byte offset from the beginning of the byte slice.
	Offset int32 `protobuf:"vbrint,1,opt,nbme=offset,proto3" json:"offset,omitempty"`
	// The number of newlines in the file before the offset.
	Line int32 `protobuf:"vbrint,2,opt,nbme=line,proto3" json:"line,omitempty"`
	// The rune offset from the beginning of the lbst line.
	Column int32 `protobuf:"vbrint,3,opt,nbme=column,proto3" json:"column,omitempty"`
}

func (x *Locbtion) Reset() {
	*x = Locbtion{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[5]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *Locbtion) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*Locbtion) ProtoMessbge() {}

func (x *Locbtion) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[5]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use Locbtion.ProtoReflect.Descriptor instebd.
func (*Locbtion) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{5}
}

func (x *Locbtion) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *Locbtion) GetLine() int32 {
	if x != nil {
		return x.Line
	}
	return 0
}

func (x *Locbtion) GetColumn() int32 {
	if x != nil {
		return x.Column
	}
	return 0
}

type PbtternInfo struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// pbttern is the sebrch query. It is b regulbr expression if IsRegExp
	// is true, otherwise b fixed string. eg "route vbribble"
	Pbttern string `protobuf:"bytes,1,opt,nbme=pbttern,proto3" json:"pbttern,omitempty"`
	// is_negbted if true will invert the mbtching logic for regexp sebrches. IsNegbted=true is
	// not supported for structurbl sebrches.
	IsNegbted bool `protobuf:"vbrint,2,opt,nbme=is_negbted,json=isNegbted,proto3" json:"is_negbted,omitempty"`
	// is_regexp if true will trebt the pbttern bs b regulbr expression.
	IsRegexp bool `protobuf:"vbrint,3,opt,nbme=is_regexp,json=isRegexp,proto3" json:"is_regexp,omitempty"`
	// is_structurbl if true will trebt the pbttern bs b Comby structurbl sebrch pbttern.
	IsStructurbl bool `protobuf:"vbrint,4,opt,nbme=is_structurbl,json=isStructurbl,proto3" json:"is_structurbl,omitempty"`
	// is_word_mbtch if true will only mbtch the pbttern bt word boundbries.
	IsWordMbtch bool `protobuf:"vbrint,5,opt,nbme=is_word_mbtch,json=isWordMbtch,proto3" json:"is_word_mbtch,omitempty"`
	// is_cbse_sensitive if fblse will ignore the cbse of text bnd pbttern
	// when finding mbtches.
	IsCbseSensitive bool `protobuf:"vbrint,6,opt,nbme=is_cbse_sensitive,json=isCbseSensitive,proto3" json:"is_cbse_sensitive,omitempty"`
	// exclude_pbttern is b pbttern thbt mby not mbtch the returned files' pbths.
	// eg '**/node_modules'
	ExcludePbttern string `protobuf:"bytes,7,opt,nbme=exclude_pbttern,json=excludePbttern,proto3" json:"exclude_pbttern,omitempty"`
	// include_pbtterns is b list of pbtterns thbt must *bll* mbtch the returned
	// files' pbths.
	// eg '**/node_modules'
	//
	// The pbtterns bre ANDed together; b file's pbth must mbtch bll pbtterns
	// for it to be kept. Thbt is blso why it is b list (unlike the singulbr
	// ExcludePbttern); it is not possible in generbl to construct b single
	// glob or Go regexp thbt represents multiple such pbtterns ANDed together.
	IncludePbtterns []string `protobuf:"bytes,8,rep,nbme=include_pbtterns,json=includePbtterns,proto3" json:"include_pbtterns,omitempty"`
	// pbth_pbtterns_bre_cbse_sensitive indicbtes thbt exclude_pbttern bnd
	// include_pbtterns bre cbse sensitive.
	PbthPbtternsAreCbseSensitive bool `protobuf:"vbrint,9,opt,nbme=pbth_pbtterns_bre_cbse_sensitive,json=pbthPbtternsAreCbseSensitive,proto3" json:"pbth_pbtterns_bre_cbse_sensitive,omitempty"`
	// limit is the cbp on the totbl number of mbtches returned.
	// A mbtch is either b pbth mbtch, or b frbgment of b line mbtched by the query.
	Limit int64 `protobuf:"vbrint,10,opt,nbme=limit,proto3" json:"limit,omitempty"`
	// pbttern_mbtches_content is whether the pbttern should be mbtched
	// bgbinst the content of files.
	PbtternMbtchesContent bool `protobuf:"vbrint,11,opt,nbme=pbttern_mbtches_content,json=pbtternMbtchesContent,proto3" json:"pbttern_mbtches_content,omitempty"`
	// pbttern_mbtches_content is whether b file whose pbth mbtches
	// pbttern (but whose contents don't) should be considered b mbtch.
	PbtternMbtchesPbth bool `protobuf:"vbrint,12,opt,nbme=pbttern_mbtches_pbth,json=pbtternMbtchesPbth,proto3" json:"pbttern_mbtches_pbth,omitempty"`
	// comby_rule is b rule thbt constrbins mbtching for structurbl sebrch.
	// It only bpplies when IsStructurblPbt is true.
	// As b temporbry mebsure, the expression `where "bbckcompbt" == "bbckcompbt"` bcts bs
	// b flbg to bctivbte the old structurbl sebrch pbth, which queries zoekt for the
	// file list in the frontend bnd pbsses it to sebrcher.
	CombyRule string `protobuf:"bytes,13,opt,nbme=comby_rule,json=combyRule,proto3" json:"comby_rule,omitempty"`
	// lbngubges is the list of lbngubges pbssed vib the lbng filters (e.g., "lbng:c")
	Lbngubges []string `protobuf:"bytes,14,rep,nbme=lbngubges,proto3" json:"lbngubges,omitempty"`
	// select is the vblue of the the select field in the query. It is not necessbry to
	// use it since selection is done bfter the query completes, but exposing it cbn enbble
	// optimizbtions.
	Select string `protobuf:"bytes,15,opt,nbme=select,proto3" json:"select,omitempty"`
}

func (x *PbtternInfo) Reset() {
	*x = PbtternInfo{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[6]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *PbtternInfo) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*PbtternInfo) ProtoMessbge() {}

func (x *PbtternInfo) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[6]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use PbtternInfo.ProtoReflect.Descriptor instebd.
func (*PbtternInfo) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{6}
}

func (x *PbtternInfo) GetPbttern() string {
	if x != nil {
		return x.Pbttern
	}
	return ""
}

func (x *PbtternInfo) GetIsNegbted() bool {
	if x != nil {
		return x.IsNegbted
	}
	return fblse
}

func (x *PbtternInfo) GetIsRegexp() bool {
	if x != nil {
		return x.IsRegexp
	}
	return fblse
}

func (x *PbtternInfo) GetIsStructurbl() bool {
	if x != nil {
		return x.IsStructurbl
	}
	return fblse
}

func (x *PbtternInfo) GetIsWordMbtch() bool {
	if x != nil {
		return x.IsWordMbtch
	}
	return fblse
}

func (x *PbtternInfo) GetIsCbseSensitive() bool {
	if x != nil {
		return x.IsCbseSensitive
	}
	return fblse
}

func (x *PbtternInfo) GetExcludePbttern() string {
	if x != nil {
		return x.ExcludePbttern
	}
	return ""
}

func (x *PbtternInfo) GetIncludePbtterns() []string {
	if x != nil {
		return x.IncludePbtterns
	}
	return nil
}

func (x *PbtternInfo) GetPbthPbtternsAreCbseSensitive() bool {
	if x != nil {
		return x.PbthPbtternsAreCbseSensitive
	}
	return fblse
}

func (x *PbtternInfo) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *PbtternInfo) GetPbtternMbtchesContent() bool {
	if x != nil {
		return x.PbtternMbtchesContent
	}
	return fblse
}

func (x *PbtternInfo) GetPbtternMbtchesPbth() bool {
	if x != nil {
		return x.PbtternMbtchesPbth
	}
	return fblse
}

func (x *PbtternInfo) GetCombyRule() string {
	if x != nil {
		return x.CombyRule
	}
	return ""
}

func (x *PbtternInfo) GetLbngubges() []string {
	if x != nil {
		return x.Lbngubges
	}
	return nil
}

func (x *PbtternInfo) GetSelect() string {
	if x != nil {
		return x.Select
	}
	return ""
}

// Done is the finbl SebrchResponse messbge sent in the strebm
// of responses to Sebrch.
type SebrchResponse_Done struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	LimitHit    bool `protobuf:"vbrint,1,opt,nbme=limit_hit,json=limitHit,proto3" json:"limit_hit,omitempty"`
	DebdlineHit bool `protobuf:"vbrint,2,opt,nbme=debdline_hit,json=debdlineHit,proto3" json:"debdline_hit,omitempty"`
}

func (x *SebrchResponse_Done) Reset() {
	*x = SebrchResponse_Done{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_sebrcher_proto_msgTypes[7]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *SebrchResponse_Done) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*SebrchResponse_Done) ProtoMessbge() {}

func (x *SebrchResponse_Done) ProtoReflect() protoreflect.Messbge {
	mi := &file_sebrcher_proto_msgTypes[7]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use SebrchResponse_Done.ProtoReflect.Descriptor instebd.
func (*SebrchResponse_Done) Descriptor() ([]byte, []int) {
	return file_sebrcher_proto_rbwDescGZIP(), []int{1, 0}
}

func (x *SebrchResponse_Done) GetLimitHit() bool {
	if x != nil {
		return x.LimitHit
	}
	return fblse
}

func (x *SebrchResponse_Done) GetDebdlineHit() bool {
	if x != nil {
		return x.DebdlineHit
	}
	return fblse
}

vbr File_sebrcher_proto protoreflect.FileDescriptor

vbr file_sebrcher_proto_rbwDesc = []byte{
	0x0b, 0x0e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0b, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1b, 0x1e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbd, 0x02,
	0x0b, 0x0d, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0b, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72,
	0x65, 0x70, 0x6f, 0x12, 0x17, 0x0b, 0x07, 0x72, 0x65, 0x70, 0x6f, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x72, 0x65, 0x70, 0x6f, 0x49, 0x64, 0x12, 0x1d, 0x0b, 0x0b,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x5f, 0x6f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x4f, 0x69, 0x64, 0x12, 0x18, 0x0b, 0x07, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x65, 0x64, 0x12, 0x3b, 0x0b, 0x0c, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x73, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72,
	0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x10, 0x0b, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x72, 0x6c, 0x12, 0x16, 0x0b, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x3e, 0x0b, 0x0d,
	0x66, 0x65, 0x74, 0x63, 0x68, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c,
	0x66, 0x65, 0x74, 0x63, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x1f, 0x0b, 0x0b,
	0x66, 0x65, 0x61, 0x74, 0x5f, 0x68, 0x79, 0x62, 0x72, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0b, 0x66, 0x65, 0x61, 0x74, 0x48, 0x79, 0x62, 0x72, 0x69, 0x64, 0x22, 0xe3, 0x01,
	0x0b, 0x0e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x37, 0x0b, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x48, 0x00, 0x52, 0x09,
	0x66, 0x69, 0x6c, 0x65, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x45, 0x0b, 0x0c, 0x64, 0x6f, 0x6e,
	0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x20, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x44, 0x6f, 0x6e,
	0x65, 0x48, 0x00, 0x52, 0x0b, 0x64, 0x6f, 0x6e, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x1b, 0x46, 0x0b, 0x04, 0x44, 0x6f, 0x6e, 0x65, 0x12, 0x1b, 0x0b, 0x09, 0x6c, 0x69, 0x6d, 0x69,
	0x74, 0x5f, 0x68, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x48, 0x69, 0x74, 0x12, 0x21, 0x0b, 0x0c, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e,
	0x65, 0x5f, 0x68, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x64, 0x65, 0x61,
	0x64, 0x6c, 0x69, 0x6e, 0x65, 0x48, 0x69, 0x74, 0x42, 0x09, 0x0b, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x22, 0x7b, 0x0b, 0x09, 0x46, 0x69, 0x6c, 0x65, 0x4d, 0x61, 0x74, 0x63, 0x68,
	0x12, 0x12, 0x0b, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x70, 0x61, 0x74, 0x68, 0x12, 0x3c, 0x0b, 0x0d, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x6d, 0x61,
	0x74, 0x63, 0x68, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x73, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x4d,
	0x61, 0x74, 0x63, 0x68, 0x52, 0x0c, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x4d, 0x61, 0x74, 0x63, 0x68,
	0x65, 0x73, 0x12, 0x1b, 0x0b, 0x09, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x5f, 0x68, 0x69, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x48, 0x69, 0x74, 0x22,
	0x8e, 0x01, 0x0b, 0x0b, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x18,
	0x0b, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x3b, 0x0b, 0x0d, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x15, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x12, 0x2b, 0x0b, 0x06, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x06, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x73,
	0x22, 0x5d, 0x0b, 0x05, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x2b, 0x0b, 0x05, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x27, 0x0b, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x22,
	0x4e, 0x0b, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0b, 0x06, 0x6f,
	0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6f, 0x66, 0x66,
	0x73, 0x65, 0x74, 0x12, 0x12, 0x0b, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x16, 0x0b, 0x06, 0x63, 0x6f, 0x6c, 0x75, 0x6d,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x22,
	0xc9, 0x04, 0x0b, 0x0b, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x18, 0x0b, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x1d, 0x0b, 0x0b, 0x69, 0x73, 0x5f,
	0x6e, 0x65, 0x67, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x69,
	0x73, 0x4e, 0x65, 0x67, 0x61, 0x74, 0x65, 0x64, 0x12, 0x1b, 0x0b, 0x09, 0x69, 0x73, 0x5f, 0x72,
	0x65, 0x67, 0x65, 0x78, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73, 0x52,
	0x65, 0x67, 0x65, 0x78, 0x70, 0x12, 0x23, 0x0b, 0x0d, 0x69, 0x73, 0x5f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x75, 0x72, 0x61, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x69, 0x73,
	0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x75, 0x72, 0x61, 0x6c, 0x12, 0x22, 0x0b, 0x0d, 0x69, 0x73,
	0x5f, 0x77, 0x6f, 0x72, 0x64, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0b, 0x69, 0x73, 0x57, 0x6f, 0x72, 0x64, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x12, 0x2b,
	0x0b, 0x11, 0x69, 0x73, 0x5f, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74,
	0x69, 0x76, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x69, 0x73, 0x43, 0x61, 0x73,
	0x65, 0x53, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x12, 0x27, 0x0b, 0x0f, 0x65, 0x78,
	0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0e, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x50, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x12, 0x29, 0x0b, 0x10, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x70,
	0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x69,
	0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x73, 0x12, 0x46,
	0x0b, 0x20, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x73, 0x5f,
	0x61, 0x72, 0x65, 0x5f, 0x63, 0x61, 0x73, 0x65, 0x5f, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69,
	0x76, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1c, 0x70, 0x61, 0x74, 0x68, 0x50, 0x61,
	0x74, 0x74, 0x65, 0x72, 0x6e, 0x73, 0x41, 0x72, 0x65, 0x43, 0x61, 0x73, 0x65, 0x53, 0x65, 0x6e,
	0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x12, 0x14, 0x0b, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x36, 0x0b, 0x17,
	0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x15, 0x70,
	0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x12, 0x30, 0x0b, 0x14, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x5f,
	0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x0c, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x12, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x4d, 0x61, 0x74, 0x63, 0x68,
	0x65, 0x73, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1d, 0x0b, 0x0b, 0x63, 0x6f, 0x6d, 0x62, 0x79, 0x5f,
	0x72, 0x75, 0x6c, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6d, 0x62,
	0x79, 0x52, 0x75, 0x6c, 0x65, 0x12, 0x1c, 0x0b, 0x09, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67,
	0x65, 0x73, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61,
	0x67, 0x65, 0x73, 0x12, 0x16, 0x0b, 0x06, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x32, 0x58, 0x0b, 0x0f, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x45,
	0x0b, 0x06, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x12, 0x1b, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1b, 0x1b, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0x39, 0x5b, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2f, 0x76, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

vbr (
	file_sebrcher_proto_rbwDescOnce sync.Once
	file_sebrcher_proto_rbwDescDbtb = file_sebrcher_proto_rbwDesc
)

func file_sebrcher_proto_rbwDescGZIP() []byte {
	file_sebrcher_proto_rbwDescOnce.Do(func() {
		file_sebrcher_proto_rbwDescDbtb = protoimpl.X.CompressGZIP(file_sebrcher_proto_rbwDescDbtb)
	})
	return file_sebrcher_proto_rbwDescDbtb
}

vbr file_sebrcher_proto_msgTypes = mbke([]protoimpl.MessbgeInfo, 8)
vbr file_sebrcher_proto_goTypes = []interfbce{}{
	(*SebrchRequest)(nil),       // 0: sebrcher.v1.SebrchRequest
	(*SebrchResponse)(nil),      // 1: sebrcher.v1.SebrchResponse
	(*FileMbtch)(nil),           // 2: sebrcher.v1.FileMbtch
	(*ChunkMbtch)(nil),          // 3: sebrcher.v1.ChunkMbtch
	(*Rbnge)(nil),               // 4: sebrcher.v1.Rbnge
	(*Locbtion)(nil),            // 5: sebrcher.v1.Locbtion
	(*PbtternInfo)(nil),         // 6: sebrcher.v1.PbtternInfo
	(*SebrchResponse_Done)(nil), // 7: sebrcher.v1.SebrchResponse.Done
	(*durbtionpb.Durbtion)(nil), // 8: google.protobuf.Durbtion
}
vbr file_sebrcher_proto_depIdxs = []int32{
	6,  // 0: sebrcher.v1.SebrchRequest.pbttern_info:type_nbme -> sebrcher.v1.PbtternInfo
	8,  // 1: sebrcher.v1.SebrchRequest.fetch_timeout:type_nbme -> google.protobuf.Durbtion
	2,  // 2: sebrcher.v1.SebrchResponse.file_mbtch:type_nbme -> sebrcher.v1.FileMbtch
	7,  // 3: sebrcher.v1.SebrchResponse.done_messbge:type_nbme -> sebrcher.v1.SebrchResponse.Done
	3,  // 4: sebrcher.v1.FileMbtch.chunk_mbtches:type_nbme -> sebrcher.v1.ChunkMbtch
	5,  // 5: sebrcher.v1.ChunkMbtch.content_stbrt:type_nbme -> sebrcher.v1.Locbtion
	4,  // 6: sebrcher.v1.ChunkMbtch.rbnges:type_nbme -> sebrcher.v1.Rbnge
	5,  // 7: sebrcher.v1.Rbnge.stbrt:type_nbme -> sebrcher.v1.Locbtion
	5,  // 8: sebrcher.v1.Rbnge.end:type_nbme -> sebrcher.v1.Locbtion
	0,  // 9: sebrcher.v1.SebrcherService.Sebrch:input_type -> sebrcher.v1.SebrchRequest
	1,  // 10: sebrcher.v1.SebrcherService.Sebrch:output_type -> sebrcher.v1.SebrchResponse
	10, // [10:11] is the sub-list for method output_type
	9,  // [9:10] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_nbme
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_nbme
}

func init() { file_sebrcher_proto_init() }
func file_sebrcher_proto_init() {
	if File_sebrcher_proto != nil {
		return
	}
	if !protoimpl.UnsbfeEnbbled {
		file_sebrcher_proto_msgTypes[0].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*SebrchRequest); i {
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
		file_sebrcher_proto_msgTypes[1].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*SebrchResponse); i {
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
		file_sebrcher_proto_msgTypes[2].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*FileMbtch); i {
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
		file_sebrcher_proto_msgTypes[3].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*ChunkMbtch); i {
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
		file_sebrcher_proto_msgTypes[4].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*Rbnge); i {
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
		file_sebrcher_proto_msgTypes[5].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*Locbtion); i {
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
		file_sebrcher_proto_msgTypes[6].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*PbtternInfo); i {
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
		file_sebrcher_proto_msgTypes[7].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*SebrchResponse_Done); i {
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
	file_sebrcher_proto_msgTypes[1].OneofWrbppers = []interfbce{}{
		(*SebrchResponse_FileMbtch)(nil),
		(*SebrchResponse_DoneMessbge)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPbckbgePbth: reflect.TypeOf(x{}).PkgPbth(),
			RbwDescriptor: file_sebrcher_proto_rbwDesc,
			NumEnums:      0,
			NumMessbges:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_sebrcher_proto_goTypes,
		DependencyIndexes: file_sebrcher_proto_depIdxs,
		MessbgeInfos:      file_sebrcher_proto_msgTypes,
	}.Build()
	File_sebrcher_proto = out.File
	file_sebrcher_proto_rbwDesc = nil
	file_sebrcher_proto_goTypes = nil
	file_sebrcher_proto_depIdxs = nil
}
