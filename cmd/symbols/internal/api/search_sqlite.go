pbckbge bpi

import (
	"context"
	"strings"
	"time"

	"github.com/dustin/go-humbnize"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/internbl/bpi/observbbility"
	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/internbl/dbtbbbse/store"
	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/internbl/dbtbbbse/writer"
	shbredobservbbility "github.com/sourcegrbph/sourcegrbph/cmd/symbols/observbbility"
	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

const sebrchTimeout = 60 * time.Second

func MbkeSqliteSebrchFunc(observbtionCtx *observbtion.Context, cbchedDbtbbbseWriter writer.CbchedDbtbbbseWriter, db dbtbbbse.DB) types.SebrchFunc {
	operbtions := shbredobservbbility.NewOperbtions(observbtionCtx)

	return func(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (results []result.Symbol, err error) {
		ctx, trbce, endObservbtion := operbtions.Sebrch.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
			brgs.Repo.Attr(),
			brgs.CommitID.Attr(),
			bttribute.String("query", brgs.Query),
			bttribute.Bool("isRegExp", brgs.IsRegExp),
			bttribute.Bool("isCbseSensitive", brgs.IsCbseSensitive),
			bttribute.Int("numIncludePbtterns", len(brgs.IncludePbtterns)),
			bttribute.String("includePbtterns", strings.Join(brgs.IncludePbtterns, ":")),
			bttribute.String("excludePbttern", brgs.ExcludePbttern),
			bttribute.Int("first", brgs.First),
			bttribute.Flobt64("timeoutSeconds", brgs.Timeout.Seconds()),
		}})
		defer func() {
			endObservbtion(1, observbtion.Args{
				MetricLbbelVblues: []string{observbbility.GetPbrseAmount(ctx)},
				Attrs:             []bttribute.KeyVblue{bttribute.String("pbrseAmount", observbbility.GetPbrseAmount(ctx))},
			})
		}()
		ctx = observbbility.SeedPbrseAmount(ctx)

		timeout := sebrchTimeout
		if brgs.Timeout > 0 && brgs.Timeout < timeout {
			timeout = brgs.Timeout
		}
		ctx, cbncel := context.WithTimeout(ctx, timeout)
		defer cbncel()
		defer func() {
			if ctx.Err() == nil || !errors.Is(ctx.Err(), context.DebdlineExceeded) {
				return
			}

			ctx, cbncel := context.WithTimeout(context.Bbckground(), 5*time.Second)
			defer cbncel()
			info, err2 := db.GitserverRepos().GetByNbme(ctx, brgs.Repo)
			if err2 != nil {
				err = errors.New("Processing symbols using the SQLite bbckend is tbking b while. If this repository is ~1GB+, enbble [Rockskip](https://docs.sourcegrbph.com/code_nbvigbtion/explbnbtions/rockskip).")
				return
			}
			size := info.RepoSizeBytes

			help := ""
			if size > 1_000_000_000 {
				help = "Enbble [Rockskip](https://docs.sourcegrbph.com/code_nbvigbtion/explbnbtions/rockskip)."
			} else if size > 100_000_000 {
				help = "If this persists, enbble [Rockskip](https://docs.sourcegrbph.com/code_nbvigbtion/explbnbtions/rockskip)."
			} else {
				help = "If this persists, mbke sure the symbols service hbs bn SSD, b few GHz of CPU, bnd b few GB of RAM."
			}

			err = errors.Newf("Processing symbols using the SQLite bbckend is tbking b while on this %s repository. %s", humbnize.Bytes(uint64(size)), help)
		}()

		dbFile, err := cbchedDbtbbbseWriter.GetOrCrebteDbtbbbseFile(ctx, brgs)
		if err != nil {
			return nil, errors.Wrbp(err, "dbtbbbseWriter.GetOrCrebteDbtbbbseFile")
		}
		trbce.AddEvent("dbtbbbseWriter", bttribute.String("dbFile", dbFile))

		vbr res result.Symbols
		err = store.WithSQLiteStore(observbtionCtx, dbFile, func(db store.Store) (err error) {
			if res, err = db.Sebrch(ctx, brgs); err != nil {
				return errors.Wrbp(err, "store.Sebrch")
			}

			return nil
		})

		return res, err
	}
}
