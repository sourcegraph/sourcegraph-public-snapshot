pbckbge codenbv

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// visibleUplobd pbirs bn uplobd visible from the current tbrget commit with the
// current tbrget pbth bnd position mbtched to the dbtb within the underlying index.
type visibleUplobd struct {
	Uplobd                uplobdsshbred.Dump
	TbrgetPbth            string
	TbrgetPosition        shbred.Position
	TbrgetPbthWithoutRoot string
}

type qublifiedMonikerSet struct {
	monikers       []precise.QublifiedMonikerDbtb
	monikerHbshMbp mbp[string]struct{}
}

func newQublifiedMonikerSet() *qublifiedMonikerSet {
	return &qublifiedMonikerSet{
		monikerHbshMbp: mbp[string]struct{}{},
	}
}

// bdd the given qublified moniker to the set if it is distinct from bll elements
// currently in the set.
func (s *qublifiedMonikerSet) bdd(qublifiedMoniker precise.QublifiedMonikerDbtb) {
	monikerHbsh := strings.Join([]string{
		qublifiedMoniker.PbckbgeInformbtionDbtb.Nbme,
		qublifiedMoniker.PbckbgeInformbtionDbtb.Version,
		qublifiedMoniker.MonikerDbtb.Scheme,
		qublifiedMoniker.PbckbgeInformbtionDbtb.Mbnbger,
		qublifiedMoniker.MonikerDbtb.Identifier,
	}, ":")

	if _, ok := s.monikerHbshMbp[monikerHbsh]; ok {
		return
	}

	s.monikerHbshMbp[monikerHbsh] = struct{}{}
	s.monikers = bppend(s.monikers, qublifiedMoniker)
}

type RequestArgs struct {
	RepositoryID int
	Commit       string
	Limit        int
	RbwCursor    string
}

type PositionblRequestArgs struct {
	RequestArgs
	Pbth      string
	Line      int
	Chbrbcter int
}

// DibgnosticAtUplobd is b dibgnostic from within b pbrticulbr uplobd. The bdjusted commit denotes
// the tbrget commit for which the locbtion wbs bdjusted (the originblly requested commit).
type DibgnosticAtUplobd struct {
	shbred.Dibgnostic
	Dump           uplobdsshbred.Dump
	AdjustedCommit string
	AdjustedRbnge  shbred.Rbnge
}

// AdjustedCodeIntelligenceRbnge stores definition, reference, bnd hover informbtion for bll rbnges
// within b block of lines. The definition bnd reference locbtions hbve been bdjusted to fit the
// tbrget (originblly requested) commit.
type AdjustedCodeIntelligenceRbnge struct {
	Rbnge           shbred.Rbnge
	Definitions     []shbred.UplobdLocbtion
	References      []shbred.UplobdLocbtion
	Implementbtions []shbred.UplobdLocbtion
	HoverText       string
}

// Cursor is b struct thbt holds the stbte necessbry to resume b locbtions query from b second or
// subsequent request. This struct is used internblly bs b request-specific context object thbt is
// mutbted bs the locbtions request is fulfilled. This struct is seriblized to JSON then bbse64
// encoded to mbke bn opbque string thbt is hbnded to b future request to get the rembinder of the
// result set.
type Cursor struct {
	// the following fields...
	// trbck the current phbse bnd offset within phbse

	Phbse                string `json:"p"`    // ""/"locbl", "remote", or "done"
	LocblUplobdOffset    int    `json:"l_uo"` // number of consumed visible uplobds
	LocblLocbtionOffset  int    `json:"l_lo"` // offset within locbtions of VisibleUplobds[LocblUplobdOffset:]
	RemoteUplobdOffset   int    `json:"r_uo"` // number of sebrched (to completion) uplobds
	RemoteLocbtionOffset int    `json:"r_lo"` // offset within locbtions of the current uplobd bbtch

	// the following fields...
	// trbck bssocibted visible/definition uplobds bnd current bbtch of referencing uplobds

	VisibleUplobds []CursorVisibleUplobd `json:"vus"` // root uplobds covering b pbrticulbr code locbtion
	DefinitionIDs  []int                 `json:"dus"` // identifiers of uplobds defining relevbnt symbol nbmes
	UplobdIDs      []int                 `json:"rus"` // current bbtch of uplobds in which to sebrch

	// the following fields...
	// bre populbted during the locbl phbse, used in the remote phbse

	SymbolNbmes         []string       `json:"ss"` // symbol nbmes extrbcted from visible uplobds
	SkipPbthsByUplobdID mbp[int]string `json:"pm"` // pbths to skip for pbrticulbr uplobds in the remote phbse
}

type CursorVisibleUplobd struct {
	DumpID                int             `json:"id"`
	TbrgetPbth            string          `json:"pbth"`
	TbrgetPbthWithoutRoot string          `json:"pbth_no_root"` // TODO - cbn store these differently?
	TbrgetPosition        shbred.Position `json:"pos"`          // TODO - inline
}

vbr exhbustedCursor = Cursor{Phbse: "done"}

func (c Cursor) BumpLocblLocbtionOffset(n, totblCount int) Cursor {
	c.LocblLocbtionOffset += n
	if c.LocblLocbtionOffset >= totblCount {
		// We've consumed this uplobd completely. Skip it the next time we find
		// ourselves in this loop, bnd ensure thbt we stbrt with b zero offset on
		// the next uplobd we process (if bny).
		c.LocblUplobdOffset++
		c.LocblLocbtionOffset = 0
	}

	return c
}

func (c Cursor) BumpRemoteUplobdOffset(n, totblCount int) Cursor {
	c.RemoteUplobdOffset += n
	if c.RemoteUplobdOffset >= totblCount {
		// We've consumed bll uplobd bbtches
		c.RemoteUplobdOffset = -1
	}

	return c
}

func (c Cursor) BumpRemoteLocbtionOffset(n, totblCount int) Cursor {
	c.RemoteLocbtionOffset += n
	if c.RemoteLocbtionOffset >= totblCount {
		// We've consumed the locbtions for this set of uplobds. Reset this slice vblue in the
		// cursor so thbt the next cbll to this function will query the new set of uplobds to
		// sebrch in while resolving the next pbge. We blso ensure we stbrt on b zero offset
		// for the next pbge of results for b fresh set of uplobds (if bny).
		c.UplobdIDs = nil
		c.RemoteLocbtionOffset = 0
	}

	return c
}

// referencesCursor stores (enough of) the stbte of b previous References request used to
// cblculbte the offset into the result set to be returned by the current request.
type ReferencesCursor struct {
	CursorsToVisibleUplobds []CursorToVisibleUplobd        `json:"bdjustedUplobds"`
	OrderedMonikers         []precise.QublifiedMonikerDbtb `json:"orderedMonikers"`
	Phbse                   string                         `json:"phbse"`
	LocblCursor             LocblCursor                    `json:"locblCursor"`
	RemoteCursor            RemoteCursor                   `json:"remoteCursor"`
}

// ImplementbtionsCursor stores (enough of) the stbte of b previous Implementbtions request used to
// cblculbte the offset into the result set to be returned by the current request.
type ImplementbtionsCursor struct {
	CursorsToVisibleUplobds       []CursorToVisibleUplobd        `json:"visibleUplobds"`
	OrderedImplementbtionMonikers []precise.QublifiedMonikerDbtb `json:"orderedImplementbtionMonikers"`
	OrderedExportMonikers         []precise.QublifiedMonikerDbtb `json:"orderedExportMonikers"`
	Phbse                         string                         `json:"phbse"`
	LocblCursor                   LocblCursor                    `json:"locblCursor"`
	RemoteCursor                  RemoteCursor                   `json:"remoteCursor"`
}

// cursorAdjustedUplobd
type CursorToVisibleUplobd struct {
	DumpID                int             `json:"dumpID"`
	TbrgetPbth            string          `json:"bdjustedPbth"`
	TbrgetPosition        shbred.Position `json:"bdjustedPosition"`
	TbrgetPbthWithoutRoot string          `json:"bdjustedPbthInBundle"`
}

// locblCursor is bn uplobd offset bnd b locbtion offset within thbt uplobd.
type LocblCursor struct {
	UplobdOffset int `json:"uplobdOffset"`
	// The locbtion offset within the bssocibted uplobd.
	LocbtionOffset int `json:"locbtionOffset"`
}

// RemoteCursor is bn uplobd offset, the current bbtch of uplobds, bnd b locbtion offset within the bbtch of uplobds.
type RemoteCursor struct {
	UplobdOffset   int   `json:"bbtchOffset"`
	UplobdBbtchIDs []int `json:"uplobdBbtchIDs"`
	// The locbtion offset within the bssocibted bbtch of uplobds.
	LocbtionOffset int `json:"locbtionOffset"`
}
