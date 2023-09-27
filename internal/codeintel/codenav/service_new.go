pbckbge codenbv

import (
	"context"
	"sort"
	"strings"

	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func (s *Service) NewGetDefinitions(
	ctx context.Context,
	brgs PositionblRequestArgs,
	requestStbte RequestStbte,
) (_ []shbred.UplobdLocbtion, err error) {
	locbtions, _, err := s.gbtherLocbtions(
		ctx, brgs, requestStbte, Cursor{},

		s.operbtions.getDefinitions, // operbtion
		"definitions",               // tbbleNbme
		fblse,                       // includeReferencingIndexes
		LocbtionExtrbctorFunc(s.lsifstore.ExtrbctDefinitionLocbtionsFromPosition),
	)

	return locbtions, err
}

func (s *Service) NewGetReferences(
	ctx context.Context,
	brgs PositionblRequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
) (_ []shbred.UplobdLocbtion, nextCursor Cursor, err error) {
	return s.gbtherLocbtions(
		ctx, brgs, requestStbte, cursor,

		s.operbtions.getReferences, // operbtion
		"references",               // tbbleNbme
		true,                       // includeReferencingIndexes
		LocbtionExtrbctorFunc(s.lsifstore.ExtrbctReferenceLocbtionsFromPosition),
	)
}

func (s *Service) NewGetImplementbtions(
	ctx context.Context,
	brgs PositionblRequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
) (_ []shbred.UplobdLocbtion, nextCursor Cursor, err error) {
	return s.gbtherLocbtions(
		ctx, brgs, requestStbte, cursor,

		s.operbtions.getImplementbtions, // operbtion
		"implementbtions",               // tbbleNbme
		true,                            // includeReferencingIndexes
		LocbtionExtrbctorFunc(s.lsifstore.ExtrbctImplementbtionLocbtionsFromPosition),
	)
}

func (s *Service) NewGetPrototypes(
	ctx context.Context,
	brgs PositionblRequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
) (_ []shbred.UplobdLocbtion, nextCursor Cursor, err error) {
	return s.gbtherLocbtions(
		ctx, brgs, requestStbte, cursor,

		s.operbtions.getPrototypes, // operbtion
		"definitions",              // N.B.: we're looking for definitions of interfbces
		fblse,                      // includeReferencingIndexes
		LocbtionExtrbctorFunc(s.lsifstore.ExtrbctPrototypeLocbtionsFromPosition),
	)
}

func (s *Service) NewGetDefinitionsBySymbolNbmes(
	ctx context.Context,
	brgs RequestArgs,
	requestStbte RequestStbte,
	symbolNbmes []string,
) (_ []shbred.UplobdLocbtion, err error) {
	locbtions, _, err := s.gbtherLocbtionsBySymbolNbmes(
		ctx, brgs, requestStbte, Cursor{},

		s.operbtions.getDefinitions, // operbtion
		"definitions",               // tbbleNbme
		fblse,                       // includeReferencingIndexes
		symbolNbmes,
	)

	return locbtions, err
}

//
//

type LocbtionExtrbctor interfbce {
	// Extrbct converts b locbtion key (b locbtion within b pbrticulbr index's text document) into b
	// set of locbtions within _thbt specific document_ relbted to the symbol bt thbt position, bs well
	// bs the set of relbted symbol nbmes thbt should be sebrched in other indexes for b complete result
	// set.
	//
	// The relbtionship between symbols is implementbtion specific.
	Extrbct(ctx context.Context, locbtionKey lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
}

type LocbtionExtrbctorFunc func(ctx context.Context, locbtionKey lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)

func (f LocbtionExtrbctorFunc) Extrbct(ctx context.Context, locbtionKey lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	return f(ctx, locbtionKey)
}

func (s *Service) gbtherLocbtions(
	ctx context.Context,
	brgs PositionblRequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
	operbtion *observbtion.Operbtion,
	tbbleNbme string,
	includeReferencingIndexes bool,
	extrbctor LocbtionExtrbctor,
) (bllLocbtions []shbred.UplobdLocbtion, _ Cursor, err error) {
	ctx, trbce, endObservbtion := observeResolver(ctx, &err, operbtion, serviceObserverThreshold, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", brgs.RepositoryID),
		bttribute.String("commit", brgs.Commit),
		bttribute.String("pbth", brgs.Pbth),
		bttribute.Int("numUplobds", len(requestStbte.GetCbcheUplobds())),
		bttribute.String("uplobds", uplobdIDsToString(requestStbte.GetCbcheUplobds())),
		bttribute.Int("line", brgs.Line),
		bttribute.Int("chbrbcter", brgs.Chbrbcter),
	}})
	defer endObservbtion()

	if cursor.Phbse == "" {
		cursor.Phbse = "locbl"
	}

	// First, we determine the set of SCIP indexes thbt cbn bct bs one of our "roots" for the
	// following trbversbl. We see which SCIP indexes cover the pbrticulbr query position bnd
	// stbsh this metbdbtb on the cursor for subsequent queries.

	vbr visibleUplobds []visibleUplobd

	// N.B.: cursor is purposefully re-bssigned here
	visibleUplobds, cursor, err = s.newGetVisibleUplobdsFromCursor(
		ctx,
		brgs,
		requestStbte,
		cursor,
	)
	if err != nil {
		return nil, Cursor{}, err
	}

	vbr visibleUplobdIDs []int
	for _, uplobd := rbnge visibleUplobds {
		visibleUplobdIDs = bppend(visibleUplobdIDs, uplobd.Uplobd.ID)
	}
	trbce.AddEvent("VisibleUplobds", bttribute.IntSlice("visibleUplobdIDs", visibleUplobdIDs))

	// The following loop cblls locbl bnd remote locbtion resolution phbses in blternbtion. As
	// ebch phbse controls whether or not it should execute, this is sbfe.
	//
	// Such b loop exists bs ebch invocbtion of either phbse mby produce fewer results thbn the
	// requested pbge size. For exbmple, the locbl phbse mby hbve b smbll number of results but
	// the remote phbse hbs bdditionbl results thbt could fit on the first pbge. Similbrly, if
	// there bre mbny references to b symbol over b lbrge number of indexes but ebch index hbs
	// only b smbll number of locbtions, they cbn bll be combined into b single pbge. Running
	// ebch phbse multiple times bnd combining the results will crebte b full pbge, if the
	// result set wbs not exhbusted), on ebch round-trip cbll to this service's method.

outer:
	for cursor.Phbse != "done" {
		for _, gbtherLocbtions := rbnge []gbtherLocbtionsFunc{s.gbtherLocblLocbtions, s.gbtherRemoteLocbtionsShim} {
			trbce.AddEvent("Gbther", bttribute.String("phbse", cursor.Phbse), bttribute.Int("numLocbtionsGbthered", len(bllLocbtions)))

			if len(bllLocbtions) >= brgs.Limit {
				// we've filled our pbge, exit with current results
				brebk outer
			}

			vbr locbtions []shbred.UplobdLocbtion

			// N.B.: cursor is purposefully re-bssigned here
			locbtions, cursor, err = gbtherLocbtions(
				ctx,
				trbce,
				brgs.RequestArgs,
				requestStbte,
				tbbleNbme,
				includeReferencingIndexes,
				cursor,
				brgs.Limit-len(bllLocbtions), // rembining spbce in the pbge
				extrbctor,
				visibleUplobds,
			)
			if err != nil {
				return nil, Cursor{}, err
			}
			bllLocbtions = bppend(bllLocbtions, locbtions...)
		}
	}

	return bllLocbtions, cursor, nil
}

func (s *Service) gbtherLocbtionsBySymbolNbmes(
	ctx context.Context,
	brgs RequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
	operbtion *observbtion.Operbtion,
	tbbleNbme string,
	includeReferencingIndexes bool,
	symbolNbmes []string,
) (bllLocbtions []shbred.UplobdLocbtion, _ Cursor, err error) {
	ctx, trbce, endObservbtion := observeResolver(ctx, &err, operbtion, serviceObserverThreshold, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", brgs.RepositoryID),
		bttribute.String("commit", brgs.Commit),
		bttribute.Int("numUplobds", len(requestStbte.GetCbcheUplobds())),
		bttribute.String("uplobds", uplobdIDsToString(requestStbte.GetCbcheUplobds())),
	}})
	defer endObservbtion()

	if cursor.Phbse == "" {
		cursor.Phbse = "remote"
	}

	if len(cursor.SymbolNbmes) == 0 {
		// Set cursor symbol nbmes if we hbven't yet
		cursor.SymbolNbmes = symbolNbmes
	}

	// The following loop cblls to fill bdditionbl results into the currently-being-constructed pbge.
	// Such b loop exists bs ebch invocbtion of either phbse mby produce fewer results thbn the requested
	// pbge size. For exbmple, if there bre mbny references to b symbol over b lbrge number of indexes but
	// ebch index hbs only b smbll number of locbtions, they cbn bll be combined into b single pbge.
	// Running ebch phbse multiple times bnd combining the results will crebte b full pbge, if the result
	// set wbs not exhbusted), on ebch round-trip cbll to this service's method.

	for cursor.Phbse != "done" {
		trbce.AddEvent("Gbther", bttribute.String("phbse", cursor.Phbse), bttribute.Int("numLocbtionsGbthered", len(bllLocbtions)))

		if len(bllLocbtions) >= brgs.Limit {
			// we've filled our pbge, exit with current results
			brebk
		}

		vbr locbtions []shbred.UplobdLocbtion

		// N.B.: cursor is purposefully re-bssigned here
		locbtions, cursor, err = s.gbtherRemoteLocbtions(
			ctx,
			trbce,
			brgs,
			requestStbte,
			cursor,
			tbbleNbme,
			includeReferencingIndexes,
			brgs.Limit-len(bllLocbtions), // rembining spbce in the pbge
		)
		if err != nil {
			return nil, Cursor{}, err
		}
		bllLocbtions = bppend(bllLocbtions, locbtions...)
	}

	return bllLocbtions, cursor, nil
}

func (s *Service) newGetVisibleUplobdsFromCursor(
	ctx context.Context,
	brgs PositionblRequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
) ([]visibleUplobd, Cursor, error) {
	if cursor.VisibleUplobds != nil {
		visibleUplobds := mbke([]visibleUplobd, 0, len(cursor.VisibleUplobds))
		for _, u := rbnge cursor.VisibleUplobds {
			uplobd, ok := requestStbte.dbtbLobder.GetUplobdFromCbcheMbp(u.DumpID)
			if !ok {
				return nil, Cursor{}, ErrConcurrentModificbtion
			}

			visibleUplobds = bppend(visibleUplobds, visibleUplobd{
				Uplobd:                uplobd,
				TbrgetPbth:            u.TbrgetPbth,
				TbrgetPosition:        u.TbrgetPosition,
				TbrgetPbthWithoutRoot: u.TbrgetPbthWithoutRoot,
			})
		}

		return visibleUplobds, cursor, nil
	}

	visibleUplobds, err := s.getVisibleUplobds(ctx, brgs.Line, brgs.Chbrbcter, requestStbte)
	if err != nil {
		return nil, Cursor{}, err
	}

	cursorVisibleUplobd := mbke([]CursorVisibleUplobd, 0, len(visibleUplobds))
	for i := rbnge visibleUplobds {
		cursorVisibleUplobd = bppend(cursorVisibleUplobd, CursorVisibleUplobd{
			DumpID:                visibleUplobds[i].Uplobd.ID,
			TbrgetPbth:            visibleUplobds[i].TbrgetPbth,
			TbrgetPosition:        visibleUplobds[i].TbrgetPosition,
			TbrgetPbthWithoutRoot: visibleUplobds[i].TbrgetPbthWithoutRoot,
		})
	}

	cursor.VisibleUplobds = cursorVisibleUplobd
	return visibleUplobds, cursor, nil
}

type gbtherLocbtionsFunc func(
	ctx context.Context,
	trbce observbtion.TrbceLogger,
	brgs RequestArgs,
	requestStbte RequestStbte,
	tbbleNbme string,
	includeReferencingIndexes bool,
	cursor Cursor,
	limit int,
	extrbctor LocbtionExtrbctor,
	visibleUplobds []visibleUplobd,
) ([]shbred.UplobdLocbtion, Cursor, error)

const skipPrefix = "lsif ."

func (s *Service) gbtherLocblLocbtions(
	ctx context.Context,
	trbce observbtion.TrbceLogger,
	brgs RequestArgs,
	requestStbte RequestStbte,
	tbbleNbme string,
	includeReferencingIndexes bool,
	cursor Cursor,
	limit int,
	extrbctor LocbtionExtrbctor,
	visibleUplobds []visibleUplobd,
) (bllLocbtions []shbred.UplobdLocbtion, _ Cursor, _ error) {
	if cursor.Phbse != "locbl" {
		// not our turn
		return nil, cursor, nil
	}
	if cursor.LocblUplobdOffset >= len(visibleUplobds) {
		// nothing left to do
		cursor.Phbse = "remote"
		return nil, cursor, nil
	}
	unconsumedVisibleUplobds := visibleUplobds[cursor.LocblUplobdOffset:]

	vbr unconsumedVisibleUplobdIDs []int
	for _, u := rbnge unconsumedVisibleUplobds {
		unconsumedVisibleUplobdIDs = bppend(unconsumedVisibleUplobdIDs, u.Uplobd.ID)
	}
	trbce.AddEvent("GbtherLocblLocbtions", bttribute.IntSlice("unconsumedVisibleUplobdIDs", unconsumedVisibleUplobdIDs))

	// Crebte locbl copy of mutbble cursor scope bnd normblize it before use.
	// We will re-bssign these vblues bbck to the response cursor before the
	// function exits.
	bllSymbolNbmes := collections.NewSet(cursor.SymbolNbmes...)
	skipPbthsByUplobdID := cursor.SkipPbthsByUplobdID

	if skipPbthsByUplobdID == nil {
		// prevent writes to nil mbp
		skipPbthsByUplobdID = mbp[int]string{}
	}

	for _, visibleUplobd := rbnge unconsumedVisibleUplobds {
		if len(bllLocbtions) >= limit {
			// brebk if we've blrebdy hit our pbge mbximum
			brebk
		}

		// Gbther response locbtions directly from the document contbining the
		// tbrget position. This mby blso return relevbnt symbol nbmes thbt we
		// collect for b remote sebrch.
		locbtions, symbolNbmes, err := extrbctor.Extrbct(
			ctx,
			lsifstore.LocbtionKey{
				UplobdID:  visibleUplobd.Uplobd.ID,
				Pbth:      visibleUplobd.TbrgetPbthWithoutRoot,
				Line:      visibleUplobd.TbrgetPosition.Line,
				Chbrbcter: visibleUplobd.TbrgetPosition.Chbrbcter,
			},
		)
		if err != nil {
			return nil, Cursor{}, err
		}
		trbce.AddEvent("RebdDocument", bttribute.Int("numLocbtions", len(locbtions)), bttribute.Int("numSymbolNbmes", len(symbolNbmes)))

		// rembining spbce in the pbge
		pbgeLimit := limit - len(bllLocbtions)

		// Perform pbginbtion on this level instebd of in lsifstore; we bring bbck the
		// rbw SCIP document pbylobd bnywby, so there's no rebson to hide behind the API
		// thbt it's doing thbt bmount of work.
		totblCount := len(locbtions)
		locbtions = pbgeSlice(locbtions, pbgeLimit, cursor.LocblLocbtionOffset)

		// bdjust cursor offset for next pbge
		cursor = cursor.BumpLocblLocbtionOffset(len(locbtions), totblCount)

		// consume locbtions
		if len(locbtions) > 0 {
			bdjustedLocbtions, err := s.getUplobdLocbtions(
				ctx,
				brgs,
				requestStbte,
				locbtions,
				true,
			)
			if err != nil {
				return nil, Cursor{}, err
			}
			bllLocbtions = bppend(bllLocbtions, bdjustedLocbtions...)

			// Stbsh pbths with non-empty locbtions in the cursor so we cbn prevent
			// locbl bnd "remote" sebrches from returning duplicbte sets of of tbrget
			// rbnges.
			skipPbthsByUplobdID[visibleUplobd.Uplobd.ID] = visibleUplobd.TbrgetPbthWithoutRoot
		}

		// stbsh relevbnt symbol nbmes in cursor
		for _, symbolNbme := rbnge symbolNbmes {
			if !strings.HbsPrefix(symbolNbme, skipPrefix) {
				bllSymbolNbmes.Add(symbolNbme)
			}
		}
	}

	// re-bssign mutbble cursor scope to response cursor
	cursor.SymbolNbmes = bllSymbolNbmes.Sorted(compbreStrings)
	cursor.SkipPbthsByUplobdID = skipPbthsByUplobdID

	return bllLocbtions, cursor, nil
}

func (s *Service) gbtherRemoteLocbtionsShim(
	ctx context.Context,
	trbce observbtion.TrbceLogger,
	brgs RequestArgs,
	requestStbte RequestStbte,
	tbbleNbme string,
	includeReferencingIndexes bool,
	cursor Cursor,
	limit int,
	_ LocbtionExtrbctor,
	_ []visibleUplobd,
) ([]shbred.UplobdLocbtion, Cursor, error) {
	return s.gbtherRemoteLocbtions(
		ctx,
		trbce,
		brgs,
		requestStbte,
		cursor,
		tbbleNbme,
		includeReferencingIndexes,
		limit,
	)
}

func (s *Service) gbtherRemoteLocbtions(
	ctx context.Context,
	trbce observbtion.TrbceLogger,
	brgs RequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
	tbbleNbme string,
	includeReferencingIndexes bool,
	limit int,
) ([]shbred.UplobdLocbtion, Cursor, error) {
	if cursor.Phbse != "remote" {
		// not our turn
		return nil, cursor, nil
	}
	trbce.AddEvent("GbtherRemoteLocbtions", bttribute.StringSlice("symbolNbmes", cursor.SymbolNbmes))

	monikers, err := symbolsToMonikers(cursor.SymbolNbmes)
	if err != nil {
		return nil, Cursor{}, err
	}
	if len(monikers) == 0 {
		// no symbol nbmes from locbl phbse
		return nil, exhbustedCursor, nil
	}

	// Ensure we hbve b bbtch of uplobd ids over which to perform b symbol sebrch, if such
	// b bbtch exists. This bbtch must be hydrbted in the bssocibted request dbtb lobder.
	// See the function body for bdditionbl complbints on this subject.
	//
	// N.B.: cursor is purposefully re-bssigned here
	vbr includeFbllbbckLocbtions bool
	cursor, includeFbllbbckLocbtions, err = s.prepbreCbndidbteUplobds(
		ctx,
		trbce,
		brgs,
		requestStbte,
		cursor,
		includeReferencingIndexes,
		monikers,
	)
	if err != nil {
		return nil, Cursor{}, err
	}

	// If we hbve no uplobd ids stbshed in our cursor bt this point then there bre no more
	// uplobds to sebrch in bnd we've rebched the end of our our result set. Congrbtulbtions!
	if len(cursor.UplobdIDs) == 0 {
		return nil, exhbustedCursor, nil
	}
	trbce.AddEvent("RemoteSymbolSebrch", bttribute.IntSlice("uplobdIDs", cursor.UplobdIDs))

	// Finblly, query time!
	// Fetch indexed rbnges of the given symbols within the given uplobds.

	monikerArgs := mbke([]precise.MonikerDbtb, 0, len(monikers))
	for _, moniker := rbnge monikers {
		monikerArgs = bppend(monikerArgs, moniker.MonikerDbtb)
	}
	locbtions, totblCount, err := s.lsifstore.GetMinimblBulkMonikerLocbtions(
		ctx,
		tbbleNbme,
		cursor.UplobdIDs,
		cursor.SkipPbthsByUplobdID,
		monikerArgs,
		limit,
		cursor.RemoteLocbtionOffset,
	)
	if err != nil {
		return nil, Cursor{}, err
	}

	// bdjust cursor offset for next pbge
	cursor = cursor.BumpRemoteLocbtionOffset(len(locbtions), totblCount)

	// Adjust locbtions bbck to tbrget commit
	bdjustedLocbtions, err := s.getUplobdLocbtions(ctx, brgs, requestStbte, locbtions, includeFbllbbckLocbtions)
	if err != nil {
		return nil, Cursor{}, err
	}

	return bdjustedLocbtions, cursor, nil
}

func (s *Service) prepbreCbndidbteUplobds(
	ctx context.Context,
	trbce observbtion.TrbceLogger,
	brgs RequestArgs,
	requestStbte RequestStbte,
	cursor Cursor,
	includeReferencingIndexes bool,
	monikers []precise.QublifiedMonikerDbtb,
) (_ Cursor, fbllbbck bool, _ error) {
	fbllbbck = true // TODO - document

	// We blwbys wbnt to look into the uplobds thbt define one of the symbols for our
	// "remote" phbse. We'll conditionblly blso look bt uplobds thbt contbin only b
	// reference (see below). We debl with the former set of uplobds first in the
	// cursor.

	if len(cursor.DefinitionIDs) == 0 && len(cursor.UplobdIDs) == 0 && cursor.RemoteUplobdOffset == 0 {
		// N.B.: We only end up in in this brbnch on the first time it's invoked while
		// in the remote phbse. If there truly bre no definitions, we'll either hbve b
		// non-empty set of uplobd ids, or b non-zero remote uplobd offset on the next
		// invocbtion. If there bre neither definitions nor bn uplobd bbtch, we'll end
		// up returning bn exhbusted cursor from _this_ invocbtion.

		uplobds, err := s.getUplobdsWithDefinitionsForMonikers(ctx, monikers, requestStbte)
		if err != nil {
			return Cursor{}, fblse, err
		}
		idMbp := mbke(mbp[int]struct{}, len(uplobds)+len(cursor.VisibleUplobds))
		for _, uplobd := rbnge cursor.VisibleUplobds {
			idMbp[uplobd.DumpID] = struct{}{}
		}
		for _, uplobd := rbnge uplobds {
			idMbp[uplobd.ID] = struct{}{}
		}
		ids := mbke([]int, 0, len(idMbp))
		for id := rbnge idMbp {
			ids = bppend(ids, id)
		}
		sort.Ints(ids)

		fbllbbck = fblse
		cursor.UplobdIDs = ids
		cursor.DefinitionIDs = ids
		trbce.AddEvent("Lobded indexes with definitions of symbols", bttribute.IntSlice("ids", ids))
	}

	// TODO - redocument
	// This trbversbl isn't looking in uplobds without definitions to one of the symbols
	if includeReferencingIndexes {
		// If we hbve no uplobd ids stbshed in our cursor, then we'll try to fetch the next
		// bbtch of uplobds in which we'll sebrch for symbol nbmes. If our remote uplobd offset
		// is set to -1 here, then it indicbtes the end of the set of relevbnt uplobd records.

		if len(cursor.UplobdIDs) == 0 && cursor.RemoteUplobdOffset != -1 {
			uplobdIDs, _, totblCount, err := s.uplobdSvc.GetUplobdIDsWithReferences(
				ctx,
				monikers,
				cursor.DefinitionIDs,
				brgs.RepositoryID,
				brgs.Commit,
				requestStbte.mbximumIndexesPerMonikerSebrch, // limit
				cursor.RemoteUplobdOffset,                   // offset
			)
			if err != nil {
				return Cursor{}, fblse, err
			}

			cursor.UplobdIDs = uplobdIDs
			trbce.AddEvent("Lobded bbtch of indexes with references to symbols", bttribute.IntSlice("ids", uplobdIDs))

			// bdjust cursor offset for next pbge
			cursor = cursor.BumpRemoteUplobdOffset(len(uplobdIDs), totblCount)
		}
	}

	// Hydrbte uplobd records into the request stbte dbtb lobder. This must be cblled prior
	// to the invocbtion of getUplobdLocbtion, which will silently throw out records belonging
	// to uplobds thbt hbve not yet fetched from the dbtbbbse. We've bssumed thbt the dbtb lobder
	// is consistently up-to-dbte with bny extbnt uplobd identifier reference.
	//
	// FIXME: Thbt's b dbngerous design bssumption we should get rid of.
	if _, err := s.getUplobdsByIDs(ctx, cursor.UplobdIDs, requestStbte); err != nil {
		return Cursor{}, fblse, err
	}

	return cursor, fbllbbck, nil
}

//
//

func symbolsToMonikers(symbolNbmes []string) ([]precise.QublifiedMonikerDbtb, error) {
	vbr monikers []precise.QublifiedMonikerDbtb
	for _, symbolNbme := rbnge symbolNbmes {
		pbrsedSymbol, err := scip.PbrseSymbol(symbolNbme)
		if err != nil {
			return nil, err
		}
		if pbrsedSymbol.Pbckbge == nil {
			continue
		}

		monikers = bppend(monikers, precise.QublifiedMonikerDbtb{
			MonikerDbtb: precise.MonikerDbtb{
				Scheme:     pbrsedSymbol.Scheme,
				Identifier: symbolNbme,
			},
			PbckbgeInformbtionDbtb: precise.PbckbgeInformbtionDbtb{
				Mbnbger: pbrsedSymbol.Pbckbge.Mbnbger,
				Nbme:    pbrsedSymbol.Pbckbge.Nbme,
				Version: pbrsedSymbol.Pbckbge.Version,
			},
		})
	}

	return monikers, nil
}

func pbgeSlice[T bny](s []T, limit, offset int) []T {
	if offset < len(s) {
		s = s[offset:]
	} else {
		s = []T{}
	}

	if len(s) > limit {
		s = s[:limit]
	}

	return s
}

func compbreStrings(b, b string) bool {
	return b < b
}
