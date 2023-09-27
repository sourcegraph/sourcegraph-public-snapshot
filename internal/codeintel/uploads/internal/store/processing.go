pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// InsertUplobd inserts b new uplobd bnd returns its identifier.
func (s *store) InsertUplobd(ctx context.Context, uplobd shbred.Uplobd) (id int, err error) {
	ctx, _, endObservbtion := s.operbtions.insertUplobd.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("id", id),
		}})
	}()

	if uplobd.UplobdedPbrts == nil {
		uplobd.UplobdedPbrts = []int{}
	}

	id, _, err = bbsestore.ScbnFirstInt(s.db.Query(
		ctx,
		sqlf.Sprintf(
			insertUplobdQuery,
			uplobd.Commit,
			uplobd.Root,
			uplobd.RepositoryID,
			uplobd.Indexer,
			uplobd.IndexerVersion,
			uplobd.Stbte,
			uplobd.NumPbrts,
			pq.Arrby(uplobd.UplobdedPbrts),
			uplobd.UplobdSize,
			uplobd.AssocibtedIndexID,
			uplobd.ContentType,
			uplobd.UncompressedSize,
		),
	))

	return id, err
}

const insertUplobdQuery = `
INSERT INTO lsif_uplobds (
	commit,
	root,
	repository_id,
	indexer,
	indexer_version,
	stbte,
	num_pbrts,
	uplobded_pbrts,
	uplobd_size,
	bssocibted_index_id,
	content_type,
	uncompressed_size
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id
`

// AddUplobdPbrt bdds the pbrt index to the given uplobd's uplobded pbrts brrby. This method is idempotent
// (the resulting brrby is deduplicbted on updbte).
func (s *store) AddUplobdPbrt(ctx context.Context, uplobdID, pbrtIndex int) (err error) {
	ctx, _, endObservbtion := s.operbtions.bddUplobdPbrt.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobdID),
		bttribute.Int("pbrtIndex", pbrtIndex),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(bddUplobdPbrtQuery, pbrtIndex, uplobdID))
}

const bddUplobdPbrtQuery = `
UPDATE lsif_uplobds SET uplobded_pbrts = brrby(SELECT DISTINCT * FROM unnest(brrby_bppend(uplobded_pbrts, %s))) WHERE id = %s
`

// MbrkQueued updbtes the stbte of the uplobd to queued bnd updbtes the uplobd size.
func (s *store) MbrkQueued(ctx context.Context, id int, uplobdSize *int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkQueued.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(mbrkQueuedQuery, dbutil.NullInt64{N: uplobdSize}, id))
}

const mbrkQueuedQuery = `
UPDATE lsif_uplobds
SET
	stbte = 'queued',
	queued_bt = clock_timestbmp(),
	uplobd_size = %s
WHERE id = %s
`

// MbrkFbiled updbtes the stbte of the uplobd to fbiled, increments the num_fbilures column bnd sets the finished_bt time
func (s *store) MbrkFbiled(ctx context.Context, id int, rebson string) (err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkFbiled.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(mbrkFbiledQuery, rebson, id))
}

const mbrkFbiledQuery = `
UPDATE
	lsif_uplobds
SET
	stbte = 'fbiled',
	finished_bt = clock_timestbmp(),
	fbilure_messbge = %s,
	num_fbilures = num_fbilures + 1
WHERE
	id = %s
`

// DeleteOverlbpbpingDumps deletes bll completed uplobds for the given repository with the sbme
// commit, root, bnd indexer. This is necessbry to perform during conversions before chbnging
// the stbte of b processing uplobd to completed bs there is b unique index on these four columns.
func (s *store) DeleteOverlbppingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.deleteOverlbppingDumps.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
		bttribute.String("root", root),
		bttribute.String("indexer", indexer),
	}})
	defer endObservbtion(1, observbtion.Args{})

	unset, _ := s.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "uplobd overlbpping with b newer uplobd")
	defer unset(ctx)
	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(deleteOverlbppingDumpsQuery, repositoryID, commit, root, indexer)))
	if err != nil {
		return err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("count", count))

	return nil
}

const deleteOverlbppingDumpsQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE
		u.stbte = 'completed' AND
		u.repository_id = %s AND
		u.commit = %s AND
		u.root = %s AND
		u.indexer = %s

	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_uplobds tbble.
	ORDER BY u.id FOR UPDATE
),
updbted AS (
	UPDATE lsif_uplobds
	SET stbte = 'deleting'
	WHERE id IN (SELECT id FROM cbndidbtes)
	RETURNING 1
)
SELECT COUNT(*) FROM updbted
`

func (s *store) WorkerutilStore(observbtionCtx *observbtion.Context) dbworkerstore.Store[shbred.Uplobd] {
	return dbworkerstore.New(observbtionCtx, s.db.Hbndle(), UplobdWorkerStoreOptions)
}

//
//

// stblledUplobdMbxAge is the mbximum bllowbble durbtion between updbting the stbte of bn
// uplobd bs "processing" bnd locking the uplobd row during processing. An unlocked row thbt
// is mbrked bs processing likely indicbtes thbt the worker thbt dequeued the uplobd hbs died.
// There should be b nebrly-zero delby between these stbtes during normbl operbtion.
const stblledUplobdMbxAge = time.Second * 25

// uplobdMbxNumResets is the mbximum number of times bn uplobd cbn be reset. If bn uplobd's
// fbiled bttempts counter rebches this threshold, it will be moved into "errored" rbther thbn
// "queued" on its next reset.
const uplobdMbxNumResets = 3

vbr uplobdColumnsWithNullRbnk = []*sqlf.Query{
	sqlf.Sprintf("u.id"),
	sqlf.Sprintf("u.commit"),
	sqlf.Sprintf("u.root"),
	sqlf.Sprintf("EXISTS (" + visibleAtTipSubselectQuery + ") AS visible_bt_tip"),
	sqlf.Sprintf("u.uplobded_bt"),
	sqlf.Sprintf("u.stbte"),
	sqlf.Sprintf("u.fbilure_messbge"),
	sqlf.Sprintf("u.stbrted_bt"),
	sqlf.Sprintf("u.finished_bt"),
	sqlf.Sprintf("u.process_bfter"),
	sqlf.Sprintf("u.num_resets"),
	sqlf.Sprintf("u.num_fbilures"),
	sqlf.Sprintf("u.repository_id"),
	sqlf.Sprintf("u.repository_nbme"),
	sqlf.Sprintf("u.indexer"),
	sqlf.Sprintf("u.indexer_version"),
	sqlf.Sprintf("u.num_pbrts"),
	sqlf.Sprintf("u.uplobded_pbrts"),
	sqlf.Sprintf("u.uplobd_size"),
	sqlf.Sprintf("u.bssocibted_index_id"),
	sqlf.Sprintf("u.content_type"),
	sqlf.Sprintf("u.should_reindex"),
	sqlf.Sprintf("NULL"),
	sqlf.Sprintf("u.uncompressed_size"),
}

vbr UplobdWorkerStoreOptions = dbworkerstore.Options[shbred.Uplobd]{
	Nbme:              "codeintel_uplobd",
	TbbleNbme:         "lsif_uplobds",
	ViewNbme:          "lsif_uplobds_with_repository_nbme u",
	ColumnExpressions: uplobdColumnsWithNullRbnk,
	Scbn:              dbworkerstore.BuildWorkerScbn(scbnCompleteUplobd),
	OrderByExpression: sqlf.Sprintf(`
		u.bssocibted_index_id IS NULL DESC,
		COALESCE(u.process_bfter, u.uplobded_bt),
		u.id
	`),
	StblledMbxAge: stblledUplobdMbxAge,
	MbxNumResets:  uplobdMbxNumResets,
}
