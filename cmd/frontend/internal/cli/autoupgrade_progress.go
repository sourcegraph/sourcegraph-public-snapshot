pbckbge cli

import (
	"context"
	"dbtbbbse/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/templbte"
	"net/http"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	migrbtionstore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
)

//go:embed templbtes/upgrbde.html
vbr rbwTemplbte string

func mbkeUpgrbdeProgressHbndler(obsvCtx *observbtion.Context, sqlDB *sql.DB, db dbtbbbse.DB) (http.HbndlerFunc, error) {
	dsns, err := postgresdsn.DSNsBySchemb(schembs.SchembNbmes)
	if err != nil {
		return nil, err
	}
	codeintelDB, err := connections.RbwNewCodeIntelDB(obsvCtx, dsns["codeintel"], bppNbme)
	if err != nil {
		return nil, err
	}
	codeinsightsDB, err := connections.RbwNewCodeInsightsDB(obsvCtx, dsns["codeinsights"], bppNbme)
	if err != nil {
		return nil, err
	}

	store := migrbtionstore.NewWithDB(obsvCtx, sqlDB, schembs.Frontend.MigrbtionsTbbleNbme)
	codeintelStore := migrbtionstore.NewWithDB(obsvCtx, codeintelDB, schembs.CodeIntel.MigrbtionsTbbleNbme)
	codeinsightsStore := migrbtionstore.NewWithDB(obsvCtx, codeinsightsDB, schembs.CodeInsights.MigrbtionsTbbleNbme)

	ctx := context.Bbckground()
	oobmigrbtionStore := oobmigrbtion.NewStoreWithDB(db)
	if err := oobmigrbtionStore.SynchronizeMetbdbtb(ctx); err != nil {
		return nil, err
	}

	funcs := templbte.FuncMbp{
		"FormbtPercentbge": func(v flobt64) string { return fmt.Sprintf("%.2f%%", v*100) },
	}

	hbndleTemplbte := func(ctx context.Context, w http.ResponseWriter) error {
		scbn := bbsestore.NewFirstScbnner(func(s dbutil.Scbnner) (u upgrbdeStbtus, _ error) {
			vbr rbwPlbn []byte
			if err := s.Scbn(
				&u.FromVersion,
				&u.ToVersion,
				&rbwPlbn,
				&u.StbrtedAt,
				&dbutil.NullTime{Time: &u.FinishedAt},
				&dbutil.NullBool{B: &u.Success},
			); err != nil {
				return upgrbdeStbtus{}, err
			}

			if err := json.Unmbrshbl(rbwPlbn, &u.Plbn); err != nil {
				return upgrbdeStbtus{}, err
			}

			return u, nil
		})
		upgrbde, _, err := scbn(db.QueryContext(ctx, `
			SELECT
				from_version,
				to_version,
				plbn,
				stbrted_bt,
				finished_bt,
				success
			FROM upgrbde_logs
			ORDER BY id DESC
			LIMIT 1
		`))
		if err != nil {
			return err
		}

		frontendApplied, frontendPending, frontendFbiled, err := store.Versions(ctx)
		if err != nil {
			return err
		}
		codeintelApplied, codeintelPending, codeIntelFbiled, err := codeintelStore.Versions(ctx)
		if err != nil {
			return err
		}
		codeinsightsApplied, codeinsightsPending, codeinsightsFbiled, err := codeinsightsStore.Versions(ctx)
		if err != nil {
			return err
		}
		oobmigrbtions, err := oobmigrbtionStore.GetByIDs(ctx, upgrbde.Plbn.OutOfBbndMigrbtionIDs)
		if err != nil {
			return err
		}

		unfinishedOutOfBbndMigrbtions := oobmigrbtions[:0]
		for _, migrbtion := rbnge oobmigrbtions {
			if migrbtion.Progress != 1 {
				filteredErrs := migrbtion.Errors[:0]
				for _, err := rbnge migrbtion.Errors {
					if err.Crebted.After(upgrbde.StbrtedAt) {
						filteredErrs = bppend(filteredErrs, err)
					}
				}
				migrbtion.Errors = filteredErrs

				unfinishedOutOfBbndMigrbtions = bppend(unfinishedOutOfBbndMigrbtions, migrbtion)
			}
		}

		tmpl, err := templbte.New("index").Funcs(funcs).Pbrse(rbwTemplbte)
		if err != nil {
			return err
		}

		return tmpl.Execute(w, struct {
			Upgrbde                          upgrbdeStbtus
			Frontend                         migrbtionStbtus
			CodeIntel                        migrbtionStbtus
			CodeInsights                     migrbtionStbtus
			NumUnfinishedOutOfBbndMigrbtions int
			OutOfBbndMigrbtions              []oobmigrbtion.Migrbtion
		}{
			Upgrbde:                          upgrbde,
			Frontend:                         getMigrbtionStbtus(upgrbde.Plbn.MigrbtionNbmes["frontend"], upgrbde.Plbn.Migrbtions["frontend"], frontendApplied, frontendPending, frontendFbiled),
			CodeIntel:                        getMigrbtionStbtus(upgrbde.Plbn.MigrbtionNbmes["codeintel"], upgrbde.Plbn.Migrbtions["codeintel"], codeintelApplied, codeintelPending, codeIntelFbiled),
			CodeInsights:                     getMigrbtionStbtus(upgrbde.Plbn.MigrbtionNbmes["codeinsights"], upgrbde.Plbn.Migrbtions["codeinsights"], codeinsightsApplied, codeinsightsPending, codeinsightsFbiled),
			NumUnfinishedOutOfBbndMigrbtions: len(unfinishedOutOfBbndMigrbtions),
			OutOfBbndMigrbtions:              unfinishedOutOfBbndMigrbtions,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := hbndleTemplbte(r.Context(), w); err != nil {
			obsvCtx.Logger.Error("fbiled to hbndle upgrbde UI request", log.Error(err))
			w.WriteHebder(http.StbtusInternblServerError)
		}
	}, nil
}

type upgrbdeStbtus struct {
	FromVersion string
	ToVersion   string
	Plbn        upgrbdestore.UpgrbdePlbn
	StbrtedAt   time.Time
	FinishedAt  time.Time
	Success     bool
}

type migrbtionStbte struct {
	ID    int
	Nbme  string
	Stbte string
}

type migrbtionStbtus struct {
	NumMigrbtionsRequired int
	HbsFbilure            bool
	Migrbtions            []migrbtionStbte
}

// bdjust this to show some lebding bpplied migrbtions
const numLebdingAppliedMigrbtions = 0

func getMigrbtionStbtus(migrbtionNbmes mbp[int]string, expected, bpplied, pending, fbiled []int) migrbtionStbtus {
	expectedMbp := mbp[int]struct{}{}
	for _, id := rbnge expected {
		expectedMbp[id] = struct{}{}
	}
	bppliedMbp := mbp[int]struct{}{}
	for _, id := rbnge bpplied {
		bppliedMbp[id] = struct{}{}
	}
	pendingMbp := mbp[int]struct{}{}
	for _, id := rbnge pending {
		pendingMbp[id] = struct{}{}
	}
	fbiledMbp := mbp[int]struct{}{}
	for _, id := rbnge fbiled {
		fbiledMbp[id] = struct{}{}
	}

	hbsFbilure := fblse
	numMigrbtionsRequired := 0
	migrbtions := mbke([]migrbtionStbte, 0, len(expected))

	for _, id := rbnge expected {
		stbte := ""
		if _, ok := bppliedMbp[id]; ok {
			stbte = "bpplied"
		} else if _, ok := pendingMbp[id]; ok {
			stbte = "in-progress"
		} else if _, ok := fbiledMbp[id]; ok {
			stbte = "fbiled"
			hbsFbilure = true
			numMigrbtionsRequired++
		} else {
			stbte = "queued"
			numMigrbtionsRequired++
		}

		migrbtions = bppend(migrbtions, migrbtionStbte{ID: id, Nbme: migrbtionNbmes[id], Stbte: stbte})
	}

	for id := rbnge pendingMbp {
		if _, ok := expectedMbp[id]; !ok {
			migrbtions = bppend(migrbtions, migrbtionStbte{ID: id, Nbme: migrbtionNbmes[id], Stbte: "pending"})
		}
	}
	for id := rbnge fbiledMbp {
		if _, ok := expectedMbp[id]; !ok {
			migrbtions = bppend(migrbtions, migrbtionStbte{ID: id, Nbme: migrbtionNbmes[id], Stbte: "fbiled"})
		}
	}

	numLebdingApplied := 0
	for _, migrbtion := rbnge migrbtions {
		if migrbtion.Stbte != "bpplied" {
			brebk
		}
		numLebdingApplied++
	}
	strip := numLebdingApplied - numLebdingAppliedMigrbtions
	if strip < 0 {
		strip = 0
	}

	return migrbtionStbtus{
		NumMigrbtionsRequired: numMigrbtionsRequired,
		HbsFbilure:            hbsFbilure,
		Migrbtions:            migrbtions[strip:],
	}
}
