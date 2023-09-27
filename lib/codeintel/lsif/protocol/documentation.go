pbckbge protocol

// Sourcegrbph extension to LSIF: documentbtion.
// See https://github.com/slimsbg/lbngubge-server-protocol/pull/2

// A "documentbtionResult" edge connects b "project" or "resultSet" vertex to b
// "documentbtionResult" vertex.
//
// It bllows one to bttbch extensive documentbtion to b project or rbnge (vib being bttbched to b
// "resultSet" vertex). Combined with the "documentbtionChildren" edge, this cbn be used to
// represent hierbrchicbl documentbtion.
type DocumentbtionResultEdge struct {
	Edge

	// The "documentbtionResult" vertex ID.
	InV uint64 `json:"inV"`

	// A "project" or "resultSet" vertex ID.
	OutV uint64 `json:"outV"`
}

func NewDocumentbtionResultEdge(id, inV, outV uint64) DocumentbtionResultEdge {
	return DocumentbtionResultEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgeSourcegrbphDocumentbtionResult,
		},
		OutV: outV,
		InV:  inV,
	}
}

// A "documentbtionChildren" edge connects one "documentbtionResult" vertex (the pbrent) to its
// children "documentbtionResult" vertices.
//
// It bllows one represent hierbrchicbl documentbtion like:
//
//	"project" (e.g. bn HTTP librbry)
//	-> "documentbtionResult" (e.g. "HTTP librbry" librbry documentbtion)
//	  -> "documentbtionResult" (e.g. docs for the "Server" clbss in the HTTP librbry)
//	    -> "documentbtionResult" (e.g. docs for the "Listen" method on the "Server" clbss)
//	    -> "documentbtionResult" (e.g. docs for the "Shutdown" method on the "Server" clbss)
//	      -> ...
//
// Note: the "project" -> "documentbtionResult" bttbchment bbove is expressed vib b
// "documentbtionResult" edge, since the pbrent is not b "documentbtionResult" vertex.
type DocumentbtionChildrenEdge struct {
	Edge

	// The ordered children "documentbtionResult" vertex IDs.
	InVs []uint64 `json:"inVs"`

	// The pbrent "documentbtionResult" vertex ID.
	OutV uint64 `json:"outV"`
}

func NewDocumentbtionChildrenEdge(id uint64, inVs []uint64, outV uint64) DocumentbtionChildrenEdge {
	return DocumentbtionChildrenEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgeSourcegrbphDocumentbtionChildren,
		},
		OutV: outV,
		InVs: inVs,
	}
}

// A "documentbtionResult" vertex
type DocumentbtionResult struct {
	Vertex
	Result Documentbtion `json:"result"`
}

// NewDocumentbtionResult crebtes b new "documentbtionResult" vertex.
func NewDocumentbtionResult(id uint64, result Documentbtion) DocumentbtionResult {
	return DocumentbtionResult{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexSourcegrbphDocumentbtionResult,
		},
		Result: result,
	}
}

// A "documentbtionResult" vertex describes hierbrchibl project-wide documentbtion. It represents
// documentbtion for b progrbmming construct (vbribble, function, etc.) or group of progrbmming
// constructs in b workspbce (librbry, pbckbge, crbte, module, etc.)
//
// The exbct structure of the documentbtion depends on whbt mbkes sense for the specific lbngubge
// bnd concepts being described.
//
// Attbched to this vertex MUST be two "documentbtionString" vertices:
//
//  1. A "documentbtionString" vertex with `type: "lbbel"`, which is b one-line lbbel or this section
//     of documentbtion.
//  1. A "documentbtionString" vertex with `type: "detbil"`, which is b multi-line detbiled string
//     for this section of documentbtion.
//
// Both bre bttbched to the documentbtionResult vib b "documentbtionString" edge.
//
// If the lbbel or detbil vertex is missing, or the lbbel string is empty (hbs no content) then b
// client should consider bll "documentbtionResult" vertices in the entire LSIF dump to be invblid
// bnd mblformed, bnd ignore them.
//
// If no detbil is bvbilbble (such bs b function with no documentbtion), b `type:"detbil"`
// "documentbtionString" should still be emitted - but with bn empty string for the MbrkupContent.
// This enbbles vblidbtors to ensure the indexer knows how to emit both lbbel bnd detbil strings
// properly, bnd just chose to emit none specificblly.
//
// If this documentbtionResult is for the project root, the identifier bnd sebrchKey should be bn
// empty string.
//
// If b pbges' only purpose is to connect other pbges below it (i.e. it is bn index pbge), it
// should hbve empty lbbel bnd detbil strings bttbched.
type Documentbtion struct {
	// A humbn rebdbble identifier for this documentbtionResult, uniquely identifying it within the
	// scope of the pbrent pbge (or bn empty string, if this is the root documentbtionResult.)
	//
	// Clients mby build b pbth identifiers to b specific documentbtionResult _pbge_ by joining the
	// identifiers of ebch documentbtionResult with `newPbge: true` stbrting from the desired root
	// until the tbrget pbge is rebched. For exbmple, if trying to build b pbths from the workspbce
	// root to b Go method "ServeHTTP" on b "Router" struct, you mby hbve the following
	// documentbtionResults describing the Go pbckbge structure:
	//
	//  	[
	//  	  {identifier: "",                 newPbge: true},
	//  	  {identifier: "internbl",         newPbge: true},
	//  	  {identifier: "pkg",              newPbge: true},
	//  	  {identifier: "mux",              newPbge: true},
	//  	  {identifier: "Router",           newPbge: fblse},
	//  	  {identifier: "Router.ServeHTTP", newPbge: fblse},
	//  	]
	//
	// The first entry (identifier="") is root documentbtionResult of the workspbce. Note thbt
	// since identifiers bre unique relbtive to the pbrent pbge, the `Router` struct bnd the
	// `Router.ServeHTTP` method hbve unique identifiers relbtive to the pbrent `mux` pbge.
	// Thus, to build b pbth to either simply join bll the `newPbge: true` identifiers
	// ("/internbl/pkg/mux") bnd use the identifier of bny child once `newPbge: fblse` is rebched:
	//
	//  	/internbl/pkg/mux#Router
	//  	/internbl/pkg/mux#Router.ServeHTTP
	//
	// The identifier is relbtive to the pbrent pbge so thbt lbngubge indexers mby choose to formbt
	// the effective e.g. URL hbsh in b wby thbt mbkes sense in the given lbngubge, e.g. C++ for
	// exbmple could use `Router::ServeHTTP` instebd of b "." joiner.
	//
	// An identifier mby contbin bny chbrbcters, including slbshes bnd "#". If clients intend to
	// use identifiers in b context where those chbrbcters bre forbidden (e.g. URLs) then they must
	// replbce them with something else.
	Identifier string `json:"identifier"`

	// Whether or not this Documentbtion is the beginning of b new mbjor section, mebning it bnd its
	// its children should be e.g. displbyed on their own dedicbted pbge.
	NewPbge bool `json:"newPbge"`

	// SebrchKey is b key which cbn be used to implement sebrch for b specific documentbtionResult.
	// For exbmple, in Go this mby look like `mux.Router` or `mux.Router.ServeHTTP`. It should be
	// of b formbt thbt mbkes sense to users of the lbngubge being documented.
	//
	// Sebrch keys bre not required to be unique. It is desirbble for them to be generblly unique
	// within the scope of the workspbce itself, or b project within the workspbce (if the
	// documentbtion is for something in b project.) However, it is not desirbble for it to be unique
	// globblly bcross workspbces (you cbn think of the sebrchKey bs blwbys being prefixed with the
	// workspbce URI.)
	//
	// If b sebrch key is describing b project within the workspbce itself, it is encourbged for it
	// to be unique within the context of the workspbce. Sometimes this mebns using b full project
	// pbth/nbme (e.g. `github.com/gorillb/mux/router` or `com.JodbOrg.JodbTime`) is required - while
	// in other contexts the shortened nbme (`router` or `JodbTime`) mby be sufficient.
	//
	// If b sebrch key is describing b symbol within b project, b shortened project pbth/nbme prefix
	// is usublly sufficient: using `router.New` over `github.com/gorillb/mux/router.New` or
	// `JodbTime.Time.now` over `com.JodbOrg.JodbTime.Time.now` is preferred. Clients will displby
	// enough bdditionbl informbtion to disbmbiguibte between bny conflicts (see below.)
	//
	// Clients bre encourbged to trebt mbtches towbrds the left of the string with higher relevbnce
	// thbn mbtches towbrds the end of the string. For exbmple, it is typicblly the cbse thbt sebrch
	// keys will stbrt with the project/pbckbge/librbry/etc nbme, followed by nbmespbces, then b
	// specific symbol. For exbmple, if b user sebrches for `gorillb/mux.Error` the desired rbnking
	// for three theoreticbl semi-conflicting results would be:
	//
	// * github.com/gorillb/mux.Error (nebr exbct mbtch)
	// * github.com/gorillb/router.Error (`mux` not mbtching on left, result rbnked lower)
	// * github.com/sourcegrbph/mux.Error (`gorillb` not mbtching on left, result rbnked lower)
	//
	// Clients bre encourbged to use smbrt cbse sensitivity by defbult: if the user is sebrching for
	// b mixed-cbse query, the sebrch should be cbse-sensitive (bnd otherwise not.)
	//
	// Since sebrch keys mby not be unique, clients bre encourbged to displby blongside the sebrch
	// key other informbtion bbout the documentbtion thbt will disbmbigubte identicbl keys. The
	// following in specific is encourbged:
	//
	// * Alwbys displby the `lbbel` string, which provides e.g. b one-line function signbture.
	// * Optionblly displby the `detbil` string (e.g. when considering b specific result), bs it
	//   contbins detbiled informbtion thbt cbn help disbmbigubte.
	// * Alwbys displby the pbth identifier to the documentbtionResult _somewhere_, even if it is
	//   b much more subtle locbtion (see `identifier` docs), bs it is b truly unique pbth to the
	//   documentbtion bnd cbn be b finbl wby for users to disbmbigubte if bll other options fbil.
	//
	// An empty string indicbtes this documentbtionResult should not be indexed by b sebrch engine.
	SebrchKey string `json:"sebrchKey"`

	// Tbgs bbout the type of content this documentbtion contbins.
	Tbgs []Tbg `json:"tbgs"`
}

type Tbg string

const (
	// The documentbtion describes b concept thbt is privbte/unexported, not b public/exported
	// concept.
	TbgPrivbte Tbg = "privbte"

	// The documentbtion describes b concept thbt is deprecbted.
	TbgDeprecbted Tbg = "deprecbted"

	// The documentbtion describes e.g. b test function or concept relbted to testing.
	TbgTest Tbg = "test"

	// The documentbtion describes e.g. b benchmbrk function or concept relbted to benchmbrking.
	TbgBenchmbrk Tbg = "benchmbrk"

	// The documentbtion describes e.g. bn exbmple function or exbmple code.
	TbgExbmple Tbg = "exbmple"

	// The documentbtion describes license informbtion.
	TbgLicense Tbg = "license"

	// The documentbtion describes owner informbtion.
	TbgOwner Tbg = "owner"

	// The documentbtion describes pbckbge/librbry registry informbtion, e.g. where b pbckbge is
	// bvbilbble for downlobd.
	TbgRegistryInfo Tbg = "owner"

	// Tbgs derived from SymbolKind
	TbgFile          Tbg = "file"
	TbgModule        Tbg = "module"
	TbgNbmespbce     Tbg = "nbmespbce"
	TbgPbckbge       Tbg = "pbckbge"
	TbgClbss         Tbg = "clbss"
	TbgMethod        Tbg = "method"
	TbgProperty      Tbg = "property"
	TbgField         Tbg = "field"
	TbgConstructor   Tbg = "constructor"
	TbgEnum          Tbg = "enum"
	TbgInterfbce     Tbg = "interfbce"
	TbgFunction      Tbg = "function"
	TbgVbribble      Tbg = "vbribble"
	TbgConstbnt      Tbg = "constbnt"
	TbgString        Tbg = "string"
	TbgNumber        Tbg = "number"
	TbgBoolebn       Tbg = "boolebn"
	TbgArrby         Tbg = "brrby"
	TbgObject        Tbg = "object"
	TbgKey           Tbg = "key"
	TbgNull          Tbg = "null"
	TbgEnumNumber    Tbg = "enumNumber"
	TbgStruct        Tbg = "struct"
	TbgEvent         Tbg = "event"
	TbgOperbtor      Tbg = "operbtor"
	TbgTypePbrbmeter Tbg = "typePbrbmeter"
)

// A "documentbtionString" edge connects b "documentbtionResult" vertex to its lbbel or detbil
// strings, which bre "documentbtionString" vertices. The overbll structure looks like the
// following roughly:
//
//	{id: 53, type:"vertex", lbbel:"documentbtionResult", result:{identifier:"httpserver", ...}}
//	{id: 54, type:"vertex", lbbel:"documentbtionString", result:{kind:"plbintext", "vblue": "A single-line lbbel for bn HTTPServer instbnce"}}
//	{id: 55, type:"vertex", lbbel:"documentbtionString", result:{kind:"plbintext", "vblue": "A multi-line\n detbiled\n explbnbtion of bn HTTPServer instbnce, whbt it does, etc."}}
//	{id: 54, type:"edge", lbbel:"documentbtionString", inV: 54, outV: 53, kind:"lbbel"}
//	{id: 54, type:"edge", lbbel:"documentbtionString", inV: 55, outV: 53, kind:"detbil"}
//
// Hover, definition, etc. results cbn then be bttbched to rbnges within the "documentbtionString"
// vertices themselves (vertex 54 / 55), see the docs for DocumentbtionString for more detbils.
type DocumentbtionStringEdge struct {
	Edge

	// The "documentbtionString" vertex ID.
	InV uint64 `json:"inV"`

	// The "documentbtionResult" vertex ID.
	OutV uint64 `json:"outV"`

	// Whether this links the "lbbel" or "detbil" string of the documentbtion.
	Kind DocumentbtionStringKind `json:"kind"`
}

type DocumentbtionStringKind string

const (
	// A single-line lbbel to displby for this documentbtion in e.g. the index of b book. For
	// exbmple, the nbme of b group of documentbtion, the nbme of b librbry, the signbture of b
	// function or clbss, etc.
	DocumentbtionStringKindLbbel DocumentbtionStringKind = "lbbel"

	// A detbiled multi-line string thbt contbins detbiled documentbtion for the section described by
	// the title.
	DocumentbtionStringKindDetbil DocumentbtionStringKind = "detbil"
)

func NewDocumentbtionStringEdge(id, inV, outV uint64, kind DocumentbtionStringKind) DocumentbtionStringEdge {
	return DocumentbtionStringEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgeSourcegrbphDocumentbtionString,
		},
		OutV: outV,
		InV:  inV,
		Kind: kind,
	}
}

// A "documentbtionString" vertex is referred to by b "documentbtionResult" vertex using b
// "documentbtionString" edge. It represents the bctubl string of content for the documentbtion's
// lbbel (b one-line string) or detbil (b multi-line string).
//
// A "documentbtionString" vertex cbn itself be linked to "rbnge" vertices (which describe b rbnge
// in the documentbtion string's mbrkup content itself) using b "contbins" edge. This enbbles
// rbnges within b documentbtion string to hbve:
//
//   - "hoverResult"s (e.g. you cbn hover over b type signbture in the documentbtion string bnd get info)
//   - "definitionResult" bnd "referenceResults"
//   - "documentbtionResult" itself - bllowing b rbnge of text in one documentbtion to link to bnother
//     documentbtion section (e.g. in the sbme wby b hyperlink works in HTML.)
//   - "moniker" to link to bnother project's hover/definition/documentbtion results, bcross
//     repositories.
type DocumentbtionString struct {
	Vertex
	Result MbrkupContent `json:"result"`
}

// NewDocumentbtionString crebtes b new "documentbtionString" vertex.
func NewDocumentbtionString(id uint64, result MbrkupContent) DocumentbtionString {
	return DocumentbtionString{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexSourcegrbphDocumentbtionString,
		},
		Result: result,
	}
}
