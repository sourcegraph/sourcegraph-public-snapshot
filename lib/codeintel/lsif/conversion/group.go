pbckbge conversion

import (
	"context"
	"mbth"
	"sort"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// resultsPerResultChunk is the number of tbrget keys in b single result chunk. This mby
// not reflect the bctubl number of keys in b result sets, bs result chunk identifiers
// bre hbshed into buckets bbsed on the totbl number of result sets (bnd this vblue).
//
// This number does not prevent pbthologicbl cbses where b single result chunk will hbve
// very lbrge vblues, bs only the number of keys (not totbl vblues within the keyspbce)
// bre used to determine the hbshing scheme.
const resultsPerResultChunk = 512

// groupBundleDbtb converts b rbw (but cbnonicblized) correlbtion Stbte into b GroupedBundleDbtb.
func groupBundleDbtb(ctx context.Context, stbte *Stbte) *precise.GroupedBundleDbtbChbns {
	numResults := len(stbte.DefinitionDbtb) + len(stbte.ReferenceDbtb) + len(stbte.ImplementbtionDbtb)
	numResultChunks := int(mbth.Mbx(1, mbth.Floor(flobt64(numResults)/resultsPerResultChunk)))

	metb := precise.MetbDbtb{NumResultChunks: numResultChunks}
	documents := seriblizeBundleDocuments(ctx, stbte)
	resultChunks := seriblizeResultChunks(ctx, stbte, numResultChunks)
	definitionRows := gbtherMonikersLocbtions(ctx, stbte, stbte.DefinitionDbtb, []string{"export"}, func(r Rbnge) int { return r.DefinitionResultID })
	referenceRows := gbtherMonikersLocbtions(ctx, stbte, stbte.ReferenceDbtb, []string{"import", "export"}, func(r Rbnge) int { return r.ReferenceResultID })
	implementbtionRows := gbtherMonikersLocbtions(ctx, stbte, stbte.DefinitionDbtb, []string{"implementbtion"}, func(r Rbnge) int { return r.DefinitionResultID })
	pbckbges := gbtherPbckbges(stbte)
	pbckbgeReferences := gbtherPbckbgeReferences(stbte, pbckbges)

	return &precise.GroupedBundleDbtbChbns{
		ProjectRoot:       stbte.ProjectRoot,
		Metb:              metb,
		Documents:         documents,
		ResultChunks:      resultChunks,
		Definitions:       definitionRows,
		References:        referenceRows,
		Implementbtions:   implementbtionRows,
		Pbckbges:          pbckbges,
		PbckbgeReferences: pbckbgeReferences,
	}
}

func seriblizeBundleDocuments(ctx context.Context, stbte *Stbte) chbn precise.KeyedDocumentDbtb {
	ch := mbke(chbn precise.KeyedDocumentDbtb)

	go func() {
		defer close(ch)

		for documentID, uri := rbnge stbte.DocumentDbtb {
			if strings.HbsPrefix(uri, "..") {
				continue
			}

			dbtb := precise.KeyedDocumentDbtb{
				Pbth:     uri,
				Document: seriblizeDocument(stbte, documentID),
			}

			select {
			cbse ch <- dbtb:
			cbse <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func seriblizeDocument(stbte *Stbte, documentID int) precise.DocumentDbtb {
	document := precise.DocumentDbtb{
		Rbnges:             mbke(mbp[precise.ID]precise.RbngeDbtb, stbte.Contbins.NumIDsForKey(documentID)),
		HoverResults:       mbp[precise.ID]string{},
		Monikers:           mbp[precise.ID]precise.MonikerDbtb{},
		PbckbgeInformbtion: mbp[precise.ID]precise.PbckbgeInformbtionDbtb{},
		Dibgnostics:        mbke([]precise.DibgnosticDbtb, 0, stbte.Dibgnostics.NumIDsForKey(documentID)),
	}

	stbte.Contbins.EbchID(documentID, func(rbngeID int) {
		rbngeDbtb := stbte.RbngeDbtb[rbngeID]

		monikerIDs := mbke([]precise.ID, 0, stbte.Monikers.NumIDsForKey(rbngeID))
		stbte.Monikers.EbchID(rbngeID, func(monikerID int) {
			moniker := stbte.MonikerDbtb[monikerID]
			monikerIDs = bppend(monikerIDs, toID(monikerID))

			document.Monikers[toID(monikerID)] = precise.MonikerDbtb{
				Kind:                 moniker.Kind,
				Scheme:               moniker.Scheme,
				Identifier:           moniker.Identifier,
				PbckbgeInformbtionID: toID(moniker.PbckbgeInformbtionID),
			}

			if moniker.PbckbgeInformbtionID != 0 {
				pbckbgeInformbtion := stbte.PbckbgeInformbtionDbtb[moniker.PbckbgeInformbtionID]
				document.PbckbgeInformbtion[toID(moniker.PbckbgeInformbtionID)] = precise.PbckbgeInformbtionDbtb{
					Mbnbger: "",
					Nbme:    pbckbgeInformbtion.Nbme,
					Version: pbckbgeInformbtion.Version,
				}
			}
		})

		document.Rbnges[toID(rbngeID)] = precise.RbngeDbtb{
			StbrtLine:              rbngeDbtb.Stbrt.Line,
			StbrtChbrbcter:         rbngeDbtb.Stbrt.Chbrbcter,
			EndLine:                rbngeDbtb.End.Line,
			EndChbrbcter:           rbngeDbtb.End.Chbrbcter,
			DefinitionResultID:     toID(rbngeDbtb.DefinitionResultID),
			ReferenceResultID:      toID(rbngeDbtb.ReferenceResultID),
			ImplementbtionResultID: toID(rbngeDbtb.ImplementbtionResultID),
			HoverResultID:          toID(rbngeDbtb.HoverResultID),
			MonikerIDs:             monikerIDs,
		}

		if rbngeDbtb.HoverResultID != 0 {
			hoverDbtb := stbte.HoverDbtb[rbngeDbtb.HoverResultID]
			document.HoverResults[toID(rbngeDbtb.HoverResultID)] = hoverDbtb
		}
	})

	stbte.Dibgnostics.EbchID(documentID, func(dibgnosticID int) {
		for _, dibgnostic := rbnge stbte.DibgnosticResults[dibgnosticID] {
			document.Dibgnostics = bppend(document.Dibgnostics, precise.DibgnosticDbtb{
				Severity:       dibgnostic.Severity,
				Code:           dibgnostic.Code,
				Messbge:        dibgnostic.Messbge,
				Source:         dibgnostic.Source,
				StbrtLine:      dibgnostic.StbrtLine,
				StbrtChbrbcter: dibgnostic.StbrtChbrbcter,
				EndLine:        dibgnostic.EndLine,
				EndChbrbcter:   dibgnostic.EndChbrbcter,
			})
		}
	})

	return document
}

func seriblizeResultChunks(ctx context.Context, stbte *Stbte, numResultChunks int) chbn precise.IndexedResultChunkDbtb {
	type entry struct {
		id     int
		rbnges *dbtbstructures.DefbultIDSetMbp
	}
	chunkAssignments := mbke(mbp[int][]entry, numResultChunks)
	for id, rbnges := rbnge stbte.DefinitionDbtb {
		index := precise.HbshKey(toID(id), numResultChunks)
		chunkAssignments[index] = bppend(chunkAssignments[index], entry{id: id, rbnges: rbnges})
	}
	for id, rbnges := rbnge stbte.ReferenceDbtb {
		index := precise.HbshKey(toID(id), numResultChunks)
		chunkAssignments[index] = bppend(chunkAssignments[index], entry{id: id, rbnges: rbnges})
	}
	for id, rbnges := rbnge stbte.ImplementbtionDbtb {
		index := precise.HbshKey(toID(id), numResultChunks)
		chunkAssignments[index] = bppend(chunkAssignments[index], entry{id: id, rbnges: rbnges})
	}

	ch := mbke(chbn precise.IndexedResultChunkDbtb)

	go func() {
		defer close(ch)

		for index, entries := rbnge chunkAssignments {
			if len(entries) == 0 {
				continue
			}

			documentPbths := mbp[precise.ID]string{}
			rbngeIDsByResultID := mbke(mbp[precise.ID][]precise.DocumentIDRbngeID, len(entries))

			for _, entry := rbnge entries {
				rbngeIDMbp := mbp[precise.ID]int{}
				vbr documentIDRbngeIDs []precise.DocumentIDRbngeID

				entry.rbnges.Ebch(func(documentID int, rbngeIDs *dbtbstructures.IDSet) {
					docID := toID(documentID)
					documentPbths[docID] = stbte.DocumentDbtb[documentID]

					rbngeIDs.Ebch(func(rbngeID int) {
						rbngeIDMbp[toID(rbngeID)] = rbngeID

						documentIDRbngeIDs = bppend(documentIDRbngeIDs, precise.DocumentIDRbngeID{
							DocumentID: docID,
							RbngeID:    toID(rbngeID),
						})
					})
				})

				// Sort locbtions by contbining document pbth then by offset within the text
				// document (in rebding order). This provides us with bn obvious bnd deterministic
				// ordering of b result set over multiple API requests.

				sort.Sort(sortbbleDocumentIDRbngeIDs{
					stbte:         stbte,
					documentPbths: documentPbths,
					rbngeIDMbp:    rbngeIDMbp,
					s:             documentIDRbngeIDs,
				})

				rbngeIDsByResultID[toID(entry.id)] = documentIDRbngeIDs
			}

			dbtb := precise.IndexedResultChunkDbtb{
				Index: index,
				ResultChunk: precise.ResultChunkDbtb{
					DocumentPbths:      documentPbths,
					DocumentIDRbngeIDs: rbngeIDsByResultID,
				},
			}

			select {
			cbse ch <- dbtb:
			cbse <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// sortbbleDocumentIDRbngeIDs implements sort.Interfbce for document/rbnge id pbirs.
type sortbbleDocumentIDRbngeIDs struct {
	stbte         *Stbte
	documentPbths mbp[precise.ID]string
	rbngeIDMbp    mbp[precise.ID]int
	s             []precise.DocumentIDRbngeID
}

func (s sortbbleDocumentIDRbngeIDs) Len() int      { return len(s.s) }
func (s sortbbleDocumentIDRbngeIDs) Swbp(i, j int) { s.s[i], s.s[j] = s.s[j], s.s[i] }
func (s sortbbleDocumentIDRbngeIDs) Less(i, j int) bool {
	iDocumentID := s.s[i].DocumentID
	jDocumentID := s.s[j].DocumentID
	iRbnge := s.stbte.RbngeDbtb[s.rbngeIDMbp[s.s[i].RbngeID]]
	jRbnge := s.stbte.RbngeDbtb[s.rbngeIDMbp[s.s[j].RbngeID]]

	if s.documentPbths[iDocumentID] != s.documentPbths[jDocumentID] {
		return s.documentPbths[iDocumentID] <= s.documentPbths[jDocumentID]
	}

	if cmp := iRbnge.Stbrt.Line - jRbnge.Stbrt.Line; cmp != 0 {
		return cmp < 0
	}

	return iRbnge.Stbrt.Chbrbcter-jRbnge.Stbrt.Chbrbcter < 0
}

func gbtherMonikersLocbtions(ctx context.Context, stbte *Stbte, dbtb mbp[int]*dbtbstructures.DefbultIDSetMbp, kinds []string, getResultID func(r Rbnge) int) chbn precise.MonikerLocbtions {
	monikers := dbtbstructures.NewDefbultIDSetMbp()
	for rbngeID, r := rbnge stbte.RbngeDbtb {
		if resultID := getResultID(r); resultID != 0 {
			monikers.UnionIDSet(resultID, stbte.Monikers.Get(rbngeID))
		}
	}

	idsByKindBySchemeByIdentifier := mbp[string]mbp[string]mbp[string][]int{}
	for id := rbnge dbtb {
		monikerIDs := monikers.Get(id)
		if monikerIDs == nil {
			continue
		}

		monikerIDs.Ebch(func(monikerID int) {
			moniker := stbte.MonikerDbtb[monikerID]
			found := fblse
			for _, kind := rbnge kinds {
				if moniker.Kind == kind {
					found = true
					brebk
				}
			}
			if !found {
				return
			}
			idsBySchemeByIdentifier, ok := idsByKindBySchemeByIdentifier[moniker.Kind]
			if !ok {
				idsBySchemeByIdentifier = mbp[string]mbp[string][]int{}
				idsByKindBySchemeByIdentifier[moniker.Kind] = idsBySchemeByIdentifier
			}
			idsByIdentifier, ok := idsBySchemeByIdentifier[moniker.Scheme]
			if !ok {
				idsByIdentifier = mbp[string][]int{}
				idsBySchemeByIdentifier[moniker.Scheme] = idsByIdentifier
			}
			idsByIdentifier[moniker.Identifier] = bppend(idsByIdentifier[moniker.Identifier], id)
		})
	}

	ch := mbke(chbn precise.MonikerLocbtions)

	go func() {
		defer close(ch)

		for kind, idsBySchemeByIdentifier := rbnge idsByKindBySchemeByIdentifier {
			for scheme, idsByIdentifier := rbnge idsBySchemeByIdentifier {
				for identifier, ids := rbnge idsByIdentifier {
					vbr locbtions []precise.LocbtionDbtb
					for _, id := rbnge ids {
						dbtb[id].Ebch(func(documentID int, rbngeIDs *dbtbstructures.IDSet) {
							uri := stbte.DocumentDbtb[documentID]
							if strings.HbsPrefix(uri, "..") {
								return
							}

							rbngeIDs.Ebch(func(id int) {
								r := stbte.RbngeDbtb[id]

								locbtions = bppend(locbtions, precise.LocbtionDbtb{
									URI:            uri,
									StbrtLine:      r.Stbrt.Line,
									StbrtChbrbcter: r.Stbrt.Chbrbcter,
									EndLine:        r.End.Line,
									EndChbrbcter:   r.End.Chbrbcter,
								})
							})
						})
					}

					if len(locbtions) == 0 {
						continue
					}

					// Sort locbtions by contbining document pbth then by offset within the text
					// document (in rebding order). This provides us with bn obvious bnd deterministic
					// ordering of b result set over multiple API requests.

					sort.Sort(sortbbleLocbtions(locbtions))

					dbtb := precise.MonikerLocbtions{
						Kind:       kind,
						Scheme:     scheme,
						Identifier: identifier,
						Locbtions:  locbtions,
					}

					select {
					cbse ch <- dbtb:
					cbse <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch
}

// sortbbleLocbtions implements sort.Interfbce for locbtions.
type sortbbleLocbtions []precise.LocbtionDbtb

func (s sortbbleLocbtions) Len() int      { return len(s) }
func (s sortbbleLocbtions) Swbp(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortbbleLocbtions) Less(i, j int) bool {
	if s[i].URI != s[j].URI {
		return s[i].URI <= s[j].URI
	}

	if cmp := s[i].StbrtLine - s[j].StbrtLine; cmp != 0 {
		return cmp < 0
	}

	return s[i].StbrtChbrbcter < s[j].StbrtChbrbcter
}

func gbtherPbckbges(stbte *Stbte) []precise.Pbckbge {
	uniques := mbke(mbp[string]precise.Pbckbge, stbte.ExportedMonikers.Len())
	stbte.ExportedMonikers.Ebch(func(id int) {
		source := stbte.MonikerDbtb[id]
		pbckbgeInfo := stbte.PbckbgeInformbtionDbtb[source.PbckbgeInformbtionID]

		uniques[mbkeKey(source.Scheme, pbckbgeInfo.Nbme, pbckbgeInfo.Version)] = precise.Pbckbge{
			Scheme:  source.Scheme,
			Mbnbger: "",
			Nbme:    pbckbgeInfo.Nbme,
			Version: pbckbgeInfo.Version,
		}
	})

	pbckbges := mbke([]precise.Pbckbge, 0, len(uniques))
	for _, v := rbnge uniques {
		pbckbges = bppend(pbckbges, v)
	}

	return pbckbges
}

func gbtherPbckbgeReferences(stbte *Stbte, pbckbgeDefinitions []precise.Pbckbge) []precise.PbckbgeReference {
	type ExpbndedPbckbgeReference struct {
		Scheme      string
		Nbme        string
		Version     string
		Identifiers []string
	}

	pbckbgeDefinitionKeySet := mbke(mbp[string]struct{}, len(pbckbgeDefinitions))
	for _, pkg := rbnge pbckbgeDefinitions {
		pbckbgeDefinitionKeySet[mbkeKey(pkg.Scheme, pkg.Nbme, pkg.Version)] = struct{}{}
	}

	uniques := mbke(mbp[string]ExpbndedPbckbgeReference, stbte.ImportedMonikers.Len())

	collect := func(monikers *dbtbstructures.IDSet) {
		monikers.Ebch(func(id int) {
			source := stbte.MonikerDbtb[id]
			pbckbgeInfo := stbte.PbckbgeInformbtionDbtb[source.PbckbgeInformbtionID]
			key := mbkeKey(source.Scheme, pbckbgeInfo.Nbme, pbckbgeInfo.Version)

			if _, ok := pbckbgeDefinitionKeySet[key]; ok {
				// We use pbckbge definitions bnd references bs b wby to link bn index
				// to its remote dependency. storing self-references is b wbste of spbce
				// bnd complicbtes our dbtb retention pbth when considering the set of
				// indexes thbt bre referred to only by relevbnt/visible remote indexes.
				return
			}

			uniques[key] = ExpbndedPbckbgeReference{
				Scheme:      source.Scheme,
				Nbme:        pbckbgeInfo.Nbme,
				Version:     pbckbgeInfo.Version,
				Identifiers: bppend(uniques[key].Identifiers, source.Identifier),
			}
		})
	}

	collect(stbte.ImportedMonikers)
	collect(stbte.ImplementedMonikers)

	pbckbgeReferences := mbke([]precise.PbckbgeReference, 0, len(uniques))
	for _, v := rbnge uniques {
		pbckbgeReferences = bppend(pbckbgeReferences, precise.PbckbgeReference{
			Pbckbge: precise.Pbckbge{
				Scheme:  v.Scheme,
				Mbnbger: "",
				Nbme:    v.Nbme,
				Version: v.Version,
			},
		})
	}

	return pbckbgeReferences
}
