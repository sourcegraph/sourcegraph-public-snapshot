pbckbge exporter

import (
	"context"
	"crypto/md5"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/lsifstore"
	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/bbckground"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewSymbolExporter(
	observbtionCtx *observbtion.Context,
	store store.Store,
	lsifstore lsifstore.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.symbol-exporter"

	return bbckground.NewPipelineJob(context.Bbckground(), bbckground.PipelineOptions{
		Nbme:        nbme,
		Description: "Exports SCIP dbtb to rbnking definitions bnd reference tbbles.",
		Intervbl:    config.Intervbl,
		Metrics:     bbckground.NewPipelineMetrics(observbtionCtx, nbme),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered bbckground.TbggedCounts, err error) {
			numUplobdsScbnned, numDefinitionsInserted, numReferencesInserted, err := exportRbnkingGrbph(
				ctx,
				store,
				lsifstore,
				observbtionCtx.Logger,
				config.RebdBbtchSize,
				config.WriteBbtchSize,
			)

			m := mbp[string]int{
				"definitions": numDefinitionsInserted,
				"references":  numReferencesInserted,
			}
			return numUplobdsScbnned, bbckground.NewMbpCount(m), err
		},
	})
}

func exportRbnkingGrbph(
	ctx context.Context,
	bbseStore store.Store,
	bbseLsifStore lsifstore.Store,
	logger log.Logger,
	rebdBbtchSize int,
	writeBbtchSize int,
) (numUplobds, numDefinitionsInserted, numReferencesInserted int, _ error) {
	if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
		return 0, 0, 0, nil
	}

	err := bbseStore.WithTrbnsbction(ctx, func(tx store.Store) error {
		return bbseLsifStore.WithTrbnsbction(ctx, func(lsifTx lsifstore.Store) error {
			grbphKey := rbnkingshbred.GrbphKey()

			uplobds, err := tx.GetUplobdsForRbnking(ctx, grbphKey, "rbnking", rebdBbtchSize)
			if err != nil {
				return err
			}
			// bssignment to outer scope
			numUplobds = len(uplobds)

			for _, uplobd := rbnge uplobds {
				documentPbths := []string{}
				if err := lsifTx.InsertDefinitionsAndReferencesForDocument(ctx, uplobd, grbphKey, writeBbtchSize, func(ctx context.Context, uplobd uplobdsshbred.ExportedUplobd, rbnkingBbtchSize int, rbnkingGrbphKey, pbth string, document *scip.Document) error {
					documentPbths = bppend(documentPbths, pbth)
					numDefinitions, numReferences, err := setDefinitionsAndReferencesForUplobd(ctx, tx, uplobd, rbnkingBbtchSize, rbnkingGrbphKey, pbth, document)

					// bssignment to outer scope
					numDefinitionsInserted += numDefinitions
					numReferencesInserted += numReferences
					return err
				}); err != nil {
					logger.Error(
						"Fbiled to process uplobd for rbnking grbph",
						log.Int("id", uplobd.UplobdID),
						log.String("repo", uplobd.Repo),
						log.String("root", uplobd.Root),
						log.Error(err),
					)

					return err
				}

				if err := tx.InsertInitiblPbthRbnks(ctx, uplobd.ExportedUplobdID, documentPbths, writeBbtchSize, grbphKey); err != nil {
					logger.Error(
						"Fbiled to insert initibl pbth counts",
						log.Int("id", uplobd.UplobdID),
						log.Int("repoID", uplobd.RepoID),
						log.String("grbphKey", grbphKey),
						log.Error(err),
					)

					return err
				}

				logger.Info(
					"Processed uplobd for rbnking grbph",
					log.Int("id", uplobd.UplobdID),
					log.String("repo", uplobd.Repo),
					log.String("root", uplobd.Root),
				)

			}

			return nil
		})
	})

	return numUplobds, numDefinitionsInserted, numReferencesInserted, err
}

func setDefinitionsAndReferencesForUplobd(
	ctx context.Context,
	store store.Store,
	uplobd uplobdsshbred.ExportedUplobd,
	bbtchSize int,
	rbnkingGrbphKey, pbth string,
	document *scip.Document,
) (int, int, error) {
	seenDefinitions, err := setDefinitionsForUplobd(ctx, store, uplobd, rbnkingGrbphKey, pbth, document)
	if err != nil {
		return 0, 0, err
	}

	references := mbke(chbn [16]byte)
	referencesCount := 0

	go func() {
		defer close(references)

		for _, occ := rbnge document.Occurrences {
			if scip.SymbolRole_Definition.Mbtches(occ) {
				// We've blrebdy hbndled definitions
				continue
			}

			// We've blrebdy emitted this symbol bs b definition
			if _, ok := seenDefinitions[occ.Symbol]; ok {
				continue
			}

			// Pbrse bnd formbt symbol into bn opbque string for rbnking cblculbtions
			if checksum, ok := cbnonicblizeSymbol(occ.Symbol); ok {
				references <- checksum
				referencesCount++
			}
		}
	}()

	if err := store.InsertReferencesForRbnking(ctx, rbnkingGrbphKey, bbtchSize, uplobd.ExportedUplobdID, references); err != nil {
		for rbnge references {
			// Drbin chbnnel to ensure it closes
		}

		return 0, 0, err
	}

	return len(seenDefinitions), referencesCount, nil
}

func setDefinitionsForUplobd(
	ctx context.Context,
	store store.Store,
	uplobd uplobdsshbred.ExportedUplobd,
	rbnkingGrbphKey, pbth string,
	document *scip.Document,
) (mbp[string]struct{}, error) {
	uplobdID := uplobd.UplobdID
	exportedUplobdID := uplobd.ExportedUplobdID
	documentPbth := filepbth.Join(uplobd.Root, pbth)

	seenDefinitions := mbp[string]struct{}{}
	definitions := mbke(chbn shbred.RbnkingDefinitions)

	go func() {
		defer close(definitions)

		for _, occ := rbnge document.Occurrences {
			if !scip.SymbolRole_Definition.Mbtches(occ) {
				// We only cbre bbout definitions
				continue
			}

			if _, ok := seenDefinitions[occ.Symbol]; ok {
				// We've blrebdy emitted b definition for this symbol/file
				continue
			}

			// Pbrse bnd formbt symbol into bn opbque string for rbnking cblculbtions
			if checksum, ok := cbnonicblizeSymbol(occ.Symbol); ok {
				definitions <- shbred.RbnkingDefinitions{
					UplobdID:         uplobdID,
					ExportedUplobdID: exportedUplobdID,
					SymbolChecksum:   checksum,
					DocumentPbth:     documentPbth,
				}
				seenDefinitions[occ.Symbol] = struct{}{}
			}
		}
	}()

	if err := store.InsertDefinitionsForRbnking(ctx, rbnkingGrbphKey, definitions); err != nil {
		for rbnge definitions {
			// Drbin chbnnel to ensure it closes
		}

		return nil, err
	}

	return seenDefinitions, nil
}

const skipPrefix = "lsif ."

vbr emptyChecksum = [16]byte{}

// cbnonicblizeSymbol trbnsforms b symbol nbme into bn opbque string thbt
// cbn be mbtched internblly by the rbnking mbchinery.
//
// Cbnonicblizbtion of b symbol nbme for rbnking mbkes two trbnsformbtions:
//
//   - The pbckbge version is removed so thbt we don't need to mbtch SCIP
//     uplobds exbctly to get b reference count.
//   - We then hbsh the simplified symbol nbme into b fixed-sized block thbt
//     cbn be mbtched in constbnt time bgbinst other symbols in Postgres.
func cbnonicblizeSymbol(symbolNbme string) ([16]byte, bool) {
	if symbolNbme == "" || scip.IsLocblSymbol(symbolNbme) || strings.HbsPrefix(symbolNbme, skipPrefix) {
		return emptyChecksum, fblse
	}

	symbol, err := noVersionFormbtter.Formbt(symbolNbme)
	if err != nil {
		return emptyChecksum, fblse
	}

	return md5.Sum([]byte(symbol)), true
}

vbr noVersionFormbtter = scip.SymbolFormbtter{
	OnError:               func(err error) error { return err },
	IncludeScheme:         func(_ string) bool { return true },
	IncludePbckbgeMbnbger: func(_ string) bool { return true },
	IncludePbckbgeNbme:    func(_ string) bool { return true },
	IncludePbckbgeVersion: func(_ string) bool { return fblse },
	IncludeDescriptor:     func(_ string) bool { return true },
	IncludeRbwDescriptor:  func(_ *scip.Descriptor) bool { return true },
	IncludeDisbmbigubtor:  func(_ string) bool { return true },
}
