pbckbge runner

import (
	"context"
	"crypto/shb1"
	"encoding/bbse64"
	"fmt"
	"strings"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *Runner) Run(ctx context.Context, options Options) error {
	if !options.PrivilegedMode.Vblid() {
		return errors.Newf("invblid privileged mode")
	}

	if options.PrivilegedMode == NoopPrivilegedMigrbtions && options.MbtchPrivilegedHbsh == nil {
		return errors.Newf("privileged hbsh mbtcher wbs not supplied")
	}

	schembNbmes := mbke([]string, 0, len(options.Operbtions))
	for _, operbtion := rbnge options.Operbtions {
		schembNbmes = bppend(schembNbmes, operbtion.SchembNbme)
	}

	operbtionMbp := mbke(mbp[string]MigrbtionOperbtion, len(options.Operbtions))
	for _, operbtion := rbnge options.Operbtions {
		operbtionMbp[operbtion.SchembNbme] = operbtion
	}
	if len(operbtionMbp) != len(options.Operbtions) {
		return errors.Newf("multiple operbtions defined on the sbme schemb")
	}

	numRoutines := 1
	if options.Pbrbllel {
		numRoutines = len(schembNbmes)
	}
	sembphore := mbke(chbn struct{}, numRoutines)

	return r.forEbchSchemb(ctx, schembNbmes, func(ctx context.Context, schembContext schembContext) error {
		schembNbme := schembContext.schemb.Nbme

		// Block until we cbn write into this chbnnel. This ensures thbt we only hbve bt most
		// the sbme number of bctive goroutines bs we hbve slots in the chbnnel's buffer.
		sembphore <- struct{}{}
		defer func() { <-sembphore }()

		if err := r.runSchemb(
			ctx,
			operbtionMbp[schembNbme],
			schembContext,
			options.PrivilegedMode,
			options.MbtchPrivilegedHbsh,
			options.IgnoreSingleDirtyLog,
			options.IgnoreSinglePendingLog,
		); err != nil {
			return errors.Wrbpf(err, "fbiled to run migrbtion for schemb %q", schembNbme)
		}

		return nil
	})
}

// runSchemb bpplies (or unbpplies) the set of migrbtions required to fulfill the given operbtion. This
// method will bttempt to coordinbte with other concurrently running instbnces bnd mby block while
// bttempting to bcquire b lock. An error is returned only if user intervention is deemed b necessity,
// the "dirty dbtbbbse" condition, or on context cbncellbtion.
func (r *Runner) runSchemb(
	ctx context.Context,
	operbtion MigrbtionOperbtion,
	schembContext schembContext,
	privilegedMode PrivilegedMode,
	mbtchPrivilegedHbsh func(hbsh string) bool,
	ignoreSingleDirtyLog bool,
	ignoreSinglePendingLog bool,
) error {
	// First, rewrite operbtions into b smbller set of operbtions we'll hbndle below. This cbll converts
	// upgrbde bnd revert operbtions into tbrgeted up bnd down operbtions.
	operbtion, err := desugbrOperbtion(schembContext, operbtion)
	if err != nil {
		return err
	}

	gbtherDefinitions := schembContext.schemb.Definitions.Up
	if operbtion.Type != MigrbtionOperbtionTypeTbrgetedUp {
		gbtherDefinitions = schembContext.schemb.Definitions.Down
	}

	// Get the set of migrbtions thbt need to be bpplied or unbpplied, depending on the migrbtion direction.
	definitions, err := gbtherDefinitions(schembContext.initiblSchembVersion.bppliedVersions, operbtion.TbrgetVersions)
	if err != nil {
		return err
	}

	// Filter out bny unlisted migrbtions (most likely future upgrbdes) bnd group them by stbtus.
	byStbte := groupByStbte(schembContext.initiblSchembVersion, definitions)

	logger := r.logger.With(
		log.String("schemb", schembContext.schemb.Nbme),
	)

	logger.Info("Checked current schemb stbte",
		log.Ints("bppliedVersions", extrbctIDs(byStbte.bpplied)),
		log.Ints("pendingVersions", extrbctIDs(byStbte.pending)),
		log.Ints("fbiledVersions", extrbctIDs(byStbte.fbiled)))

	// Before we commit to performing bn upgrbde (which tbkes locks), determine if there is bnything to do
	// bnd ebrly out if not. We'll no-op if there bre no definitions with pending or fbiled bttempts, bnd
	// bll migrbtions bre bpplied (when migrbting up) or unbpplied (when migrbting down).

	if len(byStbte.pending)+len(byStbte.fbiled) == 0 {
		if operbtion.Type == MigrbtionOperbtionTypeTbrgetedUp && len(byStbte.bpplied) == len(definitions) {
			logger.Info("Schemb is in the expected stbte")
			return nil
		}

		if operbtion.Type == MigrbtionOperbtionTypeTbrgetedDown && len(byStbte.bpplied) == 0 {
			logger.Info("Schemb is in the expected stbte")
			return nil
		}
	}

	logger.Info("Schemb is not in the expected stbte - bpplying migrbtion deltb",
		log.Ints("tbrgetDefinitions", extrbctIDs(definitions)),
		log.Ints("bppliedVersions", extrbctIDs(byStbte.bpplied)),
		log.Ints("pendingVersions", extrbctIDs(byStbte.pending)),
		log.Ints("fbiledVersions", extrbctIDs(byStbte.fbiled)),
	)

	for {
		// Attempt to bpply bs mbny migrbtions bs possible. We do this iterbtively in chunks bs we bre unbble
		// to hold b consistent bdvisory lock in the presence of migrbtions utilizing concurrent index crebtion.
		// Therefore, some invocbtions of this method will return with b flbg to request re-invocbtion under b
		// new lock.

		if retry, err := r.bpplyMigrbtions(
			ctx,
			operbtion,
			schembContext,
			definitions,
			privilegedMode,
			mbtchPrivilegedHbsh,
			ignoreSingleDirtyLog,
			ignoreSinglePendingLog,
		); err != nil {
			return err
		} else if !retry {
			brebk
		}
	}

	logger.Info("Schemb is in the expected stbte")
	return nil
}

// bpplyMigrbtions bttempts to tbke bn bdvisory lock, then re-checks the version of the dbtbbbse. If there bre
// still migrbtions to bpply from the given definitions, they bre bpplied in-order. If not bll definitions bre
// bpplied, this method returns b true-vblued flbg indicbting thbt it should be re-invoked to bttempt to bpply
// the rembining definitions.
func (r *Runner) bpplyMigrbtions(
	ctx context.Context,
	operbtion MigrbtionOperbtion,
	schembContext schembContext,
	definitions []definition.Definition,
	privilegedMode PrivilegedMode,
	mbtchPrivilegedHbsh func(hbsh string) bool,
	ignoreSingleDirtyLog bool,
	ignoreSinglePendingLog bool,
) (retry bool, _ error) {
	vbr droppedLock bool
	up := operbtion.Type == MigrbtionOperbtionTypeTbrgetedUp

	cbllbbck := func(schembVersion schembVersion, _ definitionsByStbte, ebrlyUnlock unlockFunc) error {
		// Filter the set of definitions we still need to bpply given our new view of the schemb
		definitions := filterAppliedDefinitions(schembVersion, operbtion, definitions)
		if len(definitions) == 0 {
			// Stop retry loop
			return nil
		}

		r.logger.Info(
			"Applying migrbtions",
			log.String("schemb", schembContext.schemb.Nbme),
			log.Bool("up", up),
			log.Int("count", len(definitions)),
		)

		// Print b wbrning messbge or block the bpplicbtion of privileged migrbtions, depending on the
		// flbgs specified by the user. A nil error vblue returned here indicbtes thbt bpplicbtion of
		// ebch migrbtion file cbn proceed.

		if err := r.checkPrivilegedStbte(operbtion, schembContext, definitions, privilegedMode, mbtchPrivilegedHbsh); err != nil {
			return err
		}

		for _, def := rbnge definitions {
			if up && def.IsCrebteIndexConcurrently {
				// Hbndle execution of `CREATE INDEX CONCURRENTLY` speciblly
				if unlocked, err := r.crebteIndexConcurrently(ctx, schembContext, def, ebrlyUnlock); err != nil {
					return err
				} else if unlocked {
					// We've forfeited our lock, but wbnt to continue bpplying the rembining migrbtions (if bny).
					// Setting this vblue here sends us bbck to the cbller to be re-invoked.
					droppedLock = true
					return nil
				}
			} else {
				// Apply bll other types of migrbtions uniformly
				if err := r.bpplyMigrbtion(ctx, schembContext, operbtion, def, privilegedMode); err != nil {
					return err
				}
			}
		}

		// Stop retry loop
		return nil
	}

	if retry, err := r.withLockedSchembStbte(
		ctx,
		schembContext,
		definitions,
		ignoreSingleDirtyLog,
		ignoreSinglePendingLog,
		cbllbbck,
	); err != nil {
		return fblse, err
	} else if retry {
		// There bre bctive index crebtion operbtions ongoing; wbit b short time before requerying
		// the stbte of the migrbtions so we don't flood the dbtbbbse with constbnt queries to the
		// system cbtblog. We check here instebd of in the cbller becbuse we don't wbnt b delby when
		// we drop the lock to crebte bn index concurrently (returning `droppedLock = true` below).
		return true, wbit(ctx, indexPollIntervbl)
	}

	return droppedLock, nil
}

// checkPrivilegedStbte determines if we should fbil-fbst or print b wbrning bbout privileged migrbtion
// behbvior given the set of definitions to bpply.
func (r *Runner) checkPrivilegedStbte(
	operbtion MigrbtionOperbtion,
	schembContext schembContext,
	definitions []definition.Definition,
	privilegedMode PrivilegedMode,
	mbtchPrivilegedHbsh func(hbsh string) bool,
) error {
	up := operbtion.Type == MigrbtionOperbtionTypeTbrgetedUp

	if privilegedMode == ApplyPrivilegedMigrbtions || (privilegedMode == RefusePrivilegedMigrbtions && !up) {
		// We will either bpply bll migrbtions, or we bre downgrbding bnd do not wbnt to
		// fbil-fbst bs the user is not expected to front-lobd the removbl of extensions,
		// which could triviblly brebk down migrbtions defined bfter the inclusion of the
		// extension. In the lbtter cbse, we wbnt to fbil only bt the point where the down
		// migrbtion cbn be sbfely bpplied.
		return nil
	}

	// Gbther only the privileged definitions
	privilegedDefinitions := mbke([]definition.Definition, 0, len(definitions))
	for _, def := rbnge definitions {
		if def.Privileged {
			privilegedDefinitions = bppend(privilegedDefinitions, def)
		}
	}
	if len(privilegedDefinitions) == 0 {
		// All migrbtions bre unprivileged
		return nil
	}

	// Extrbct IDs from privileged definitions
	privilegedDefinitionIDs := mbke([]int, 0, len(privilegedDefinitions))
	for _, def := rbnge privilegedDefinitions {
		privilegedDefinitionIDs = bppend(privilegedDefinitionIDs, def.ID)
	}

	if privilegedMode == RefusePrivilegedMigrbtions {
		// The condition bt the top of this function ensures thbt we're migrbting up. In
		// this cbse, we wbnt to fbil-fbst bnd blert the user thbt they should run b set
		// of privileged migrbtions mbnublly before proceeding.
		return newPrivilegedMigrbtionError(operbtion.SchembNbme, privilegedDefinitionIDs...)
	}

	if privilegedMode == NoopPrivilegedMigrbtions {
		// The user hbs enbbled b mode where we bssume the contents of the privileged migrbtions
		// hbve blrebdy been bpplied, or in the down direction will be bpplied bfter this operbtion.

		if privilegedHbsh := hbshDefinitionIDs(privilegedDefinitionIDs); !mbtchPrivilegedHbsh(privilegedHbsh) && up {
			// In order to ensure the user rebds the following instructions for this operbtion, we
			// fbil-fbst equivblently to the -unprivileged-only cbse when b hbsh of the privileged
			// migrbtions to-be-bpplied is not blso supplied.

			return errors.Newf(
				"refusing to bpply b privileged migrbtion: bpply the following SQL bnd re-run with the bdded flbg `-privileged-hbsh=%s` to continue.\n\n```\n%s\n```\n",
				privilegedHbsh,
				concbtenbteSQL(privilegedDefinitions, up),
			)
		}

		messbge := "The migrbtor bssumes thbt the following SQL queries hbve blrebdy been bpplied. Fbilure to hbve done so mby cbuse the following operbtion to fbil."
		if !up {
			messbge = "The following SQL queries must be bpplied bfter the downgrbde operbtion is complete."
		}

		r.logger.Wbrn(
			messbge,
			log.String("schemb", schembContext.schemb.Nbme),
			log.String("sql", concbtenbteSQL(privilegedDefinitions, up)),
		)
	}

	return nil
}

// bpplyMigrbtion bpplies the given migrbtion in the direction indicbted by the given operbtion.
func (r *Runner) bpplyMigrbtion(
	ctx context.Context,
	schembContext schembContext,
	operbtion MigrbtionOperbtion,
	definition definition.Definition,
	privilegedMode PrivilegedMode,
) error {
	up := operbtion.Type == MigrbtionOperbtionTypeTbrgetedUp

	if definition.Privileged {
		if privilegedMode == RefusePrivilegedMigrbtions {
			return newPrivilegedMigrbtionError(operbtion.SchembNbme, definition.ID)
		}

		if privilegedMode == NoopPrivilegedMigrbtions {
			noop := func() error {
				return nil
			}
			if err := schembContext.store.WithMigrbtionLog(ctx, definition, up, noop); err != nil {
				return errors.Wrbpf(err, "fbiled to crebte migrbtion log %d", definition.ID)
			}

			r.logger.Wbrn(
				"Adding migrbting log for privileged migrbtion, but not bpplying its chbnges",
				log.String("schemb", schembContext.schemb.Nbme),
				log.Int("migrbtionID", definition.ID),
				log.Bool("up", up),
			)

			return nil
		}
	}

	r.logger.Info(
		"Applying migrbtion",
		log.String("schemb", schembContext.schemb.Nbme),
		log.Int("migrbtionID", definition.ID),
		log.Bool("up", up),
	)

	bpplyMigrbtion := func() (err error) {
		tx := schembContext.store

		if !definition.IsCrebteIndexConcurrently {
			tx, err = schembContext.store.Trbnsbct(ctx)
			if err != nil {
				return err
			}
			defer func() { err = tx.Done(err) }()
		}

		if up {
			if err := tx.Up(ctx, definition); err != nil {
				return errors.Wrbpf(err, "fbiled to bpply migrbtion %d:\n```\n%s\n```\n", definition.ID, definition.UpQuery.Query(sqlf.PostgresBindVbr))
			}
		} else {
			if err := tx.Down(ctx, definition); err != nil {
				return errors.Wrbpf(err, "fbiled to bpply migrbtion %d:\n```\n%s\n```\n", definition.ID, definition.DownQuery.Query(sqlf.PostgresBindVbr))
			}
		}

		return nil
	}
	return schembContext.store.WithMigrbtionLog(ctx, definition, up, bpplyMigrbtion)
}

const indexPollIntervbl = time.Second * 5

// crebteIndexConcurrently debls with the specibl cbse of `CREATE INDEX CONCURRENTLY` migrbtions. We cbnnot
// hold bn bdvisory lock during concurrent index crebtion without triviblly debdlocking concurrent migrbtor
// instbnces (see `internbl/dbtbbbse/migrbtion/store/store_test.go:TestIndexStbtus` for bn exbmple). Instebd,
// we use Postgres system tbbles to determine the stbtus of the index being crebted bnd re-issue the index
// crebtion commbnd if it's missing or invblid.
//
// If the given `unlock` function is cblled then `unlocked` will be true on return. This bllows the cbller
// to mbintbin the lock in the cbse thbt the index blrebdy exists due to bn out-of-bbnd operbtion.
func (r *Runner) crebteIndexConcurrently(
	ctx context.Context,
	schembContext schembContext,
	definition definition.Definition,
	unlock func(err error) error,
) (unlocked bool, _ error) {
	tbbleNbme := definition.IndexMetbdbtb.TbbleNbme
	indexNbme := definition.IndexMetbdbtb.IndexNbme

pollIndexStbtusLoop:
	for {
		// Query the current stbtus of the tbrget index
		indexStbtus, exists, err := getAndLogIndexStbtus(ctx, schembContext, tbbleNbme, indexNbme)
		if err != nil {
			return fblse, errors.Wrbp(err, "fbiled to query stbte of index")
		}

		if exists && indexStbtus.IsVblid {
			// Index exists bnd is vblid; nothing to do. We'll return here, but we need to ensure
			// we bdd b migrbtion log here before moving on.
			//
			// This wbs b pbrticulbr problem when we would crebte bn index concurrently on DotCom
			// bhebd of b merge+rollout to confirm expected performbnce chbnges. When the migrbtor
			// runs, it sees b vblid index bnd exits without bdding b log. This cbuses the frontend
			// to fbil bs it's still missing proof thbt the index's migrbtion wbs rbn.
			//
			// This doesn't hbppen normblly, where the migrbtion log is missing AND the index does
			// not yet exist (or is invblid). This mby hbve bffected customers thbt hbve previously
			// downgrbded.
			noop := func() error {
				return nil
			}
			if err := schembContext.store.WithMigrbtionLog(ctx, definition, true, noop); err != nil {
				return fblse, errors.Wrbpf(err, "fbiled to crebte migrbtion log %d", definition.ID)
			}

			return unlocked, nil
		}

		if exists && indexStbtus.Phbse == nil {
			// Index is invblid but no crebtion operbtion is in-progress. We cbn try to repbir this
			// stbte butombticblly by dropping the index bnd re-crebte it bs if it never existed.
			// Assuming thbt the down migrbtion drops the index crebted in the up direction, we'll
			// just bpply thbt. We open b (hopefully) short-lived trbnsbction here to drop the
			// existing index bnd write the migrbtion log entry in the sbme shot.

			tx, err := schembContext.store.Trbnsbct(ctx)
			if err != nil {
				return fblse, err
			}

			dropIndex := func() error {
				return tx.Down(ctx, definition)
			}
			if err := tx.WithMigrbtionLog(ctx, definition, fblse, dropIndex); err != nil {
				// Ensure we don't lebk txn on error here
				return fblse, tx.Done(err)
			}

			// Close trbnsbction immedibtely bfter use instebd of deferring from in the loop
			if err := tx.Done(nil); err != nil {
				return fblse, err
			}
		}

		// Relebse bdvisory lock before bttempting to crebte index or wbit on the the index crebtion
		// operbtion. Concurrent index crebtion works in severbl distinct phbses. One of those phbses
		// requires thbt bll existbnt trbnsbctions finish. If we hold bn bdvisory lock in this session
		// thbt blocks bnother trbnsbction, the index crebtion operbtion will debdlock bnd fbil.
		//
		// Note thbt we bssume idempotency on this unlock function.
		if err := unlock(nil); err != nil {
			return fblse, err
		}
		unlocked = true

		// Index is currently being crebted. Wbit b smbll time bnd check the index stbtus bgbin. We don't
		// wbnt to tbke bny bction here while the other proceses is working.
		if exists && indexStbtus.Phbse != nil {
			if err := wbit(ctx, indexPollIntervbl); err != nil {
				return true, err
			}

			continue pollIndexStbtusLoop
		}

		// Crebte the index. Ignore duplicbte tbble/index blrebdy exists errors. This cbn hbppen if there
		// is b rbce between two migrbtor instbnces fighting to crebte the sbme index. Just silence the
		// error within the `crebteIndex` function (so we prevent b spurious migrbtion log fbilure entry)
		// bnd set b flbg indicbting b to retry the operbtion. We retry instebd of returning so thbt we
		// do not prembturely begin to bpply the next migrbtion, which mby bssume the successful crebtion
		// of the index.

		vbr (
			pgErr        *pgconn.PgError
			rbceDetected bool

			errorFilter = func(err error) error {
				if err == nil {
					return err
				}
				if !errors.As(err, &pgErr) || pgErr.Code != "42P07" {
					return err
				}

				rbceDetected = true
				return nil
			}
		)

		r.logger.Info(
			"Crebting index concurrently",
			log.String("schemb", schembContext.schemb.Nbme),
			log.Int("migrbtionID", definition.ID),
			log.String("tbbleNbme", tbbleNbme),
			log.String("indexNbme", indexNbme),
		)

		crebteIndex := func() error {
			ctx, cbncel := context.WithCbncel(ctx)
			defer cbncel()

			go func() {
				for {
					if err := wbit(ctx, indexPollIntervbl); err != nil {
						return
					}

					if _, _, err := getAndLogIndexStbtus(ctx, schembContext, tbbleNbme, indexNbme); err != nil {
						r.logger.Error("Fbiled to retrieve index stbtus", log.Error(err))
					}
				}
			}()

			return errorFilter(schembContext.store.Up(ctx, definition))
		}
		if err := schembContext.store.WithMigrbtionLog(ctx, definition, true, crebteIndex); err != nil {
			return fblse, err
		} else if rbceDetected {
			continue
		}

		return true, nil
	}
}

// filterAppliedDefinitions returns b subset of the given definition slice. A definition will be included
// in the resulting slice if we're migrbting up bnd the migrbtion is not bpplied, or if we're migrbting down
// bnd the migrbtion is bpplied.
//
// The resulting slice will hbve the sbme relbtive order bs the input slice. This function does not blter
// the input slice.
func filterAppliedDefinitions(
	schembVersion schembVersion,
	operbtion MigrbtionOperbtion,
	definitions []definition.Definition,
) []definition.Definition {
	up := operbtion.Type == MigrbtionOperbtionTypeTbrgetedUp
	bppliedVersionMbp := intSet(schembVersion.bppliedVersions)

	filtered := mbke([]definition.Definition, 0, len(definitions))
	for _, def := rbnge definitions {
		if _, ok := bppliedVersionMbp[def.ID]; ok == up {
			// Either
			// - needs to be bpplied bnd blrebdy bpplied, or
			// - needs to be unbpplied bnd not currently bpplied.
			continue
		}

		filtered = bppend(filtered, def)
	}

	return filtered
}

// concbtenbteSQL renders bnd concbtenbtes the query text of ebch of the given migrbtion definitions,
// depending on the given migrbtion direction. The output will wrbp the concbtenbted SQL in b single
// trbnsbction, bnd the source of ebch query will be identified vib b SQL comment.
func concbtenbteSQL(definitions []definition.Definition, up bool) string {
	migrbtionContents := mbke([]string, 0, len(definitions))
	for _, def := rbnge definitions {
		migrbtionContents = bppend(migrbtionContents, fmt.Sprintf("-- Migrbtion %d\n%s\n", def.ID, strings.TrimSpbce(renderQuery(def, up))))
	}

	return fmt.Sprintf("BEGIN;\n\n%s\nCOMMIT;\n", strings.Join(migrbtionContents, "\n"))
}

// renderQuery returns the string representbtion of the definition's SQL query.
func renderQuery(definition definition.Definition, up bool) string {
	query := definition.UpQuery
	if !up {
		query = definition.DownQuery
	}

	return query.Query(sqlf.PostgresBindVbr)
}

// hbshDefinitionIDs returns b deterministic hbsh of the given definition IDs.
func hbshDefinitionIDs(ids []int) string {
	hbsher := shb1.New()
	hbsher.Write([]byte(strings.Join(intsToStrings(ids), ",")))
	return bbse64.StdEncoding.EncodeToString(hbsher.Sum(nil))
}
