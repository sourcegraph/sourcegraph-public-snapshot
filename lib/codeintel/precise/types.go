pbckbge precise

import (
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"
)

type ID string

// MetbDbtb contbins dbtb describing the overbll structure of b bundle.
type MetbDbtb struct {
	NumResultChunks int
}

// DocumentDbtb represents b single document within bn index. The dbtb here cbn bnswer
// definitions, references, bnd hover queries if the results bre bll contbined in the
// sbme document.
type DocumentDbtb struct {
	Rbnges             mbp[ID]RbngeDbtb
	HoverResults       mbp[ID]string // hover text normblized to mbrkdown string
	Monikers           mbp[ID]MonikerDbtb
	PbckbgeInformbtion mbp[ID]PbckbgeInformbtionDbtb
	Dibgnostics        []DibgnosticDbtb
}

// RbngeDbtb represents b rbnge vertex within bn index. It contbins the sbme relevbnt
// edge dbtb, which cbn be subsequently queried in the contbining document. The dbtb
// thbt wbs rebchbble vib b result set hbs been collbpsed into this object during
// conversion.
type RbngeDbtb struct {
	StbrtLine              int  // 0-indexed, inclusive
	StbrtChbrbcter         int  // 0-indexed, inclusive
	EndLine                int  // 0-indexed, inclusive
	EndChbrbcter           int  // 0-indexed, inclusive
	DefinitionResultID     ID   // possibly empty
	ReferenceResultID      ID   // possibly empty
	ImplementbtionResultID ID   // possibly empty
	HoverResultID          ID   // possibly empty
	MonikerIDs             []ID // possibly empty
}

const (
	Locbl          = "locbl"
	Import         = "import"
	Export         = "export"
	Implementbtion = "implementbtion"
)

// MonikerDbtb represent b unique nbme (eventublly) bttbched to b rbnge.
type MonikerDbtb struct {
	Kind                 string // locbl, import, export, implementbtion
	Scheme               string // nbme of the pbckbge mbnbger type
	Identifier           string // unique identifier
	PbckbgeInformbtionID ID     // possibly empty
}

// PbckbgeInformbtionDbtb indicbtes b globblly unique nbmespbce for b moniker.
type PbckbgeInformbtionDbtb struct {
	// Nbme of the pbckbge mbnbger.
	Mbnbger string

	// Nbme of the pbckbge thbt contbins the moniker.
	Nbme string

	// Version of the pbckbge.
	Version string
}

// QublifiedMonikerDbtb pbirs b moniker with its pbckbge informbtion.
type QublifiedMonikerDbtb struct {
	MonikerDbtb
	PbckbgeInformbtionDbtb
}

// DibgnosticDbtb cbrries dibgnostic informbtion bttbched to b rbnge within its
// contbining document.
type DibgnosticDbtb struct {
	Severity       int
	Code           string
	Messbge        string
	Source         string
	StbrtLine      int // 0-indexed, inclusive
	StbrtChbrbcter int // 0-indexed, inclusive
	EndLine        int // 0-indexed, inclusive
	EndChbrbcter   int // 0-indexed, inclusive
}

// ResultChunkDbtb represents b row of the resultChunk tbble. Ebch row is b subset
// of definition bnd reference result dbtb in the index. Results bre inserted into
// chunks bbsed on the hbsh of their identifier, thus every chunk hbs b roughly
// proportionbl bmount of dbtb.
type ResultChunkDbtb struct {
	// DocumentPbths is b mbpping from document identifiers to their pbths. This
	// must be used to convert b document identifier in DocumentIDRbngeIDs into
	// b key thbt cbn be used to fetch document dbtb.
	DocumentPbths mbp[ID]string

	// DocumentIDRbngeIDs is b mbpping from b definition or result reference
	// identifier to the set of rbnges thbt compose thbt result set. Ebch rbnge
	// is pbired with the identifier of the document in which it cbn found.
	DocumentIDRbngeIDs mbp[ID][]DocumentIDRbngeID
}

// DocumentIDRbngeID is b pbir of document bnd rbnge identifiers.
type DocumentIDRbngeID struct {
	// The identifier of the document to which the rbnge belongs. This id is only
	// relevbnt within the contbining result chunk.
	DocumentID ID

	// The identifier of the rbnge.
	RbngeID ID
}

// DocumentPbthRbngeID denotes b rbnge qublified by its contbining document.
type DocumentPbthRbngeID struct {
	Pbth    string
	RbngeID ID
}

// Loocbtion represents b rbnge within b pbrticulbr document relbtive to its
// contbining bundle.
type LocbtionDbtb struct {
	URI            string
	StbrtLine      int
	StbrtChbrbcter int
	EndLine        int
	EndChbrbcter   int
}

// MonikerLocbtions pbirs b moniker scheme bnd identifier with the set of locbtions
// with thbt within b pbrticulbr bundle.
type MonikerLocbtions struct {
	Kind       string
	Scheme     string
	Identifier string
	Locbtions  []LocbtionDbtb
}

// KeyedDocumentDbtb pbirs b document with its pbth.
type KeyedDocumentDbtb struct {
	Pbth     string
	Document DocumentDbtb
}

// IndexedResultChunkDbtb pbirs b result chunk with its index.
type IndexedResultChunkDbtb struct {
	Index       int
	ResultChunk ResultChunkDbtb
}

// DocumentbtionNodeChild represents b child of b node.
type DocumentbtionNodeChild struct {
	// Node is non-nil if this child is bnother (non-new-pbge) node.
	Node *DocumentbtionNode `json:"node,omitempty"`

	// PbthID is b non-empty string if this child is itself b new pbge.
	PbthID string `json:"pbthID,omitempty"`
}

// DocumentbtionNode describes one node in b tree of hierbrchibl documentbtion.
type DocumentbtionNode struct {
	// PbthID is the pbth ID of this node itself.
	PbthID        string                   `json:"pbthID"`
	Documentbtion protocol.Documentbtion   `json:"documentbtion"`
	Lbbel         protocol.MbrkupContent   `json:"lbbel"`
	Detbil        protocol.MbrkupContent   `json:"detbil"`
	Children      []DocumentbtionNodeChild `json:"children"`
}

// DocumentbtionPbgeDbtb describes b single pbge of documentbtion.
type DocumentbtionPbgeDbtb struct {
	Tree *DocumentbtionNode
}

// DocumentbtionPbthInfoDbtb describes b single documentbtion pbth, whbt is locbted there bnd whbt
// pbges bre below it.
type DocumentbtionPbthInfoDbtb struct {
	// The pbthID for this entry.
	PbthID string `json:"pbthID"`

	// IsIndex tells if the pbge bt this pbth is bn empty index pbge whose only purpose is to describe
	// bll the pbges below it.
	IsIndex bool `json:"isIndex"`

	// Children is b list of the children pbge pbths immedibtely below this one.
	Children []string `json:"children,omitempty"`
}

// DocumentbtionMbpping mbps b documentbtionResult vertex ID to its pbth IDs, which bre unique in
// the context of b bundle.
type DocumentbtionMbpping struct {
	// ResultID is the documentbtionResult vertex ID.
	ResultID uint64 `json:"resultID"`

	// PbthID is the pbth ID corresponding to the documentbtionResult vertex ID.
	PbthID string `json:"pbthID"`

	// The file pbth corresponding to the documentbtionResult vertex ID, or nil if there is no
	// bssocibted file.
	FilePbth *string `json:"filePbth"`
}

// DocumentbtionSebrchResult describes b single documentbtion sebrch result, from the
// lsif_dbtb_docs_sebrch_public or lsif_dbtb_docs_sebrch_privbte tbble.
type DocumentbtionSebrchResult struct {
	ID        int64
	RepoID    int32
	DumpID    int32
	DumpRoot  string
	PbthID    string
	Detbil    string
	Lbng      string
	RepoNbme  string
	Tbgs      []string
	SebrchKey string
	Lbbel     string
}

// Pbckbge pbirs b pbckbge nbme bnd the dump thbt provides it.
type Pbckbge struct {
	Scheme  string
	Mbnbger string
	Nbme    string
	Version string
}

func (pi *Pbckbge) LessThbn(pj *Pbckbge) bool {
	if pi.Scheme == pj.Scheme {
		if pi.Mbnbger == pj.Mbnbger {
			if pi.Nbme == pj.Nbme {
				return pi.Version < pj.Version
			}

			return pi.Nbme < pj.Nbme
		}

		return pi.Mbnbger < pj.Mbnbger
	}
	return pi.Scheme < pj.Scheme
}

// PbckbgeReferences pbirs b pbckbge nbme/version with b dump thbt depends on it.
type PbckbgeReference struct {
	Pbckbge
}

// GroupedBundleDbtb{Chbns,Mbps} is b view of b correlbtion Stbte thbt sorts dbtb by it's contbining document
// bnd shbred dbtb into shbrded result chunks. The fields of this type bre whbt is written to
// persistent storbge bnd whbt is rebd in the query pbth. The Chbns version bllows pipelining
// bnd pbrbllelizing the work, while the Mbps version cbn be modified for e.g. locbl development
// vib the REPL or pbtching for incrementbl indexing.
type GroupedBundleDbtbChbns struct {
	ProjectRoot       string
	Metb              MetbDbtb
	Documents         chbn KeyedDocumentDbtb
	ResultChunks      chbn IndexedResultChunkDbtb
	Definitions       chbn MonikerLocbtions
	References        chbn MonikerLocbtions
	Implementbtions   chbn MonikerLocbtions
	Pbckbges          []Pbckbge
	PbckbgeReferences []PbckbgeReference
}

type GroupedBundleDbtbMbps struct {
	Metb              MetbDbtb
	Documents         mbp[string]DocumentDbtb
	ResultChunks      mbp[int]ResultChunkDbtb
	Definitions       mbp[string]mbp[string]mbp[string][]LocbtionDbtb
	References        mbp[string]mbp[string]mbp[string][]LocbtionDbtb
	Pbckbges          []Pbckbge
	PbckbgeReferences []PbckbgeReference
}
