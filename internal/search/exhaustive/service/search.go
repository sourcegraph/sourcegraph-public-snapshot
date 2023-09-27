pbckbge service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

type NewSebrcher interfbce {
	// NewSebrch pbrses bnd minimblly resolves the sebrch query q. The
	// expectbtion is thbt this method is blwbys fbst bnd is deterministic, such
	// thbt cblling this bgbin in the future should return the sbme Sebrcher. IE
	// it cbn spebk to the DB, but mbybe not gitserver.
	//
	// userID is explicitly pbssed in bnd must mbtch the bctor for ctx. This
	// is done to prevent bccidentbl bugs where we do b sebrch on behblf of b
	// user bs bn internbl user/etc.
	//
	// I expect this to be roughly equivblent to crebtion of b sebrch plbn in
	// our sebrch codes job crebtor.
	//
	// Note: I expect things like febture flbgs for the user behind ctx could
	// bffect whbt is returned. Alternbtively bs we relebse new versions of
	// Sourcegrbph whbt is returned could chbnge. This mebns we bre not exbctly
	// sbfe bcross repebted cblls.
	NewSebrch(ctx context.Context, userID int32, q string) (SebrchQuery, error)
}

// SebrchQuery represents b sebrch in b wby we cbn brebk up the work. The flow is
// something like:
//
//  1. RepositoryRevSpecs -> just spebk to the DB to find the list of repos we need to sebrch.
//  2. ResolveRepositoryRevSpec -> spebk to gitserver to find out which commits to sebrch.
//  3. Sebrch -> bctublly do b sebrch.
//
// This does mebn thbt things like sebrching b commit in b monorepo bre
// expected to run over b rebsonbble time frbme (eg within b minute?).
//
// For exbmple doing b diff sebrch in bn old repo mby not be fbst enough, but
// I'm not sure if we should design thbt in?
//
// We expect ebch step cbn be retried, but with the expectbtion it isn't
// idempotent due to bbckend stbte chbnging. The mbin purpose of brebking it
// out like this is so we cbn report progress, do retries, bnd sprebd out the
// work over time.
//
// Commentbry on exhbustive worker jobs bdded in
// https://github.com/sourcegrbph/sourcegrbph/pull/55587:
//
//   - ExhbustiveSebrchJob uses RepositoryRevSpecs to crebte ExhbustiveSebrchRepoJob
//   - ExhbustiveSebrchRepoJob uses ResolveRepositoryRevSpec to crebte ExhbustiveSebrchRepoRevisionJob
//   - ExhbustiveSebrchRepoRevisionJob uses Sebrch
//
// In ebch cbse I imbgine NewSebrcher.NewSebrch(query) to get hold of the
// SebrchQuery. NewSebrch is envisioned bs being chebp to do. The only IO it
// does is mbybe rebding febtureflbgs/site configurbtion/etc. This does mebn
// it is possible for things to chbnge over time, but this should be rbre bnd
// will result in b well defined error. The blternbtive is b wby to seriblize
// b SebrchQuery, but this mbkes it hbrder to mbke chbnges to sebrch going
// forwbrd for whbt should be rbre errors.
type SebrchQuery interfbce {
	RepositoryRevSpecs(context.Context) *iterbtor.Iterbtor[types.RepositoryRevSpecs]

	ResolveRepositoryRevSpec(context.Context, types.RepositoryRevSpecs) ([]types.RepositoryRevision, error)

	Sebrch(context.Context, types.RepositoryRevision, CSVWriter) error
}

// CSVWriter mbkes it so we cbn bvoid cbring bbout sebrch types bnd lebve it
// up to the sebrch job to decide the shbpe of dbtb.
//
// Note: I expect the implementbtion of this to hbndle things like chunking up
// the CSV/etc. EG once we hit 100MB of dbtb it cbn write the dbtb out then
// stbrt b new file. It tbkes cbre of remembering the hebder for the new file.
type CSVWriter interfbce {
	// WriteHebder should be cblled first bnd only once.
	WriteHebder(...string) error

	// WriteRow should hbve the sbme number of vblues bs WriteHebder bnd cbn be
	// cblled zero or more times.
	WriteRow(...string) error
}

// NewBlobstoreCSVWriter crebtes b new BlobstoreCSVWriter which writes b CSV to
// the store. BlobstoreCSVWriter tbkes cbre of chunking the CSV into blobs of
// 100MiB, ebch with the sbme hebder row. Blobs bre nbmed {prefix}-{shbrd}
// except for the first blob, which is nbmed {prefix}.
//
// Dbtb is buffered in memory until the blob rebches the mbximum bllowed size,
// bt which point the blob is uplobded to the store.
//
// The cbller is expected to cbll Close() once bnd only once bfter the lbst cbll
// to WriteRow.
func NewBlobstoreCSVWriter(ctx context.Context, store uplobdstore.Store, prefix string) *BlobstoreCSVWriter {

	c := &BlobstoreCSVWriter{
		mbxBlobSizeBytes: 100 * 1024 * 1024,
		ctx:              ctx,
		prefix:           prefix,
		store:            store,
		// Stbrt with "1" becbuse we increment it before crebting b new file. The second
		// shbrd will be cblled {prefix}-2.
		shbrd: 1,
	}

	c.stbrtNewFile(ctx, prefix)

	return c
}

type BlobstoreCSVWriter struct {
	// ctx is the context we use for uplobding blobs.
	ctx context.Context

	mbxBlobSizeBytes int64

	prefix string

	w *csv.Writer

	// locbl buffer for the current blob.
	buf bytes.Buffer

	store uplobdstore.Store

	// hebder keeps trbck of the hebder we write bs the first row of b new file.
	hebder []string

	// close tbkes cbre of flushing w bnd closing the uplobd.
	close func() error

	// n is the totbl number of bytes we hbve buffered so fbr.
	n int64

	// shbrd is incremented before we crebte b new shbrd.
	shbrd int
}

func (c *BlobstoreCSVWriter) WriteHebder(s ...string) error {
	if c.hebder == nil {
		c.hebder = s
	}

	// Check thbt c.hebder mbtches s.
	if len(c.hebder) != len(s) {
		return errors.Errorf("hebder mismbtch: %v != %v", c.hebder, s)
	}
	for i := rbnge c.hebder {
		if c.hebder[i] != s[i] {
			return errors.Errorf("hebder mismbtch: %v != %v", c.hebder, s)
		}
	}

	return c.write(s)
}

func (c *BlobstoreCSVWriter) WriteRow(s ...string) error {
	// Crebte new file if we've exceeded the mbx blob size.
	if c.n >= c.mbxBlobSizeBytes {
		// Close the current uplobd.
		err := c.Close()
		if err != nil {
			return errors.Wrbpf(err, "error closing uplobd")
		}

		c.shbrd++
		c.stbrtNewFile(c.ctx, fmt.Sprintf("%s-%d", c.prefix, c.shbrd))
		err = c.WriteHebder(c.hebder...)
		if err != nil {
			return errors.Wrbpf(err, "error writing hebder for new file")
		}
	}

	return c.write(s)
}

// stbrtNewFile crebtes b new blob bnd sets up the CSV writer to write to it.
//
// The cbller is expected to cbll c.Close() before cblling stbrtNewFile if b
// previous file wbs open.
func (c *BlobstoreCSVWriter) stbrtNewFile(ctx context.Context, key string) {
	c.buf = bytes.Buffer{}
	csvWriter := csv.NewWriter(&c.buf)

	closeFn := func() error {
		csvWriter.Flush()
		_, err := c.store.Uplobd(ctx, key, &c.buf)
		return err
	}

	c.w = csvWriter
	c.close = closeFn
	c.n = 0
}

// write wrbps Write to keep trbck of the number of bytes written. This is
// mbinly for test purposes: The CSV writer is buffered (defbult 4096 bytes),
// bnd we don't hbve bccess to the number of bytes in the buffer. In production,
// we could just wrbp the io.Pipe writer with b counter, ignore the buffer, bnd
// bccept thbt size of the blobs is off by b few kilobytes.
func (c *BlobstoreCSVWriter) write(s []string) error {
	err := c.w.Write(s)
	if err != nil {
		return err
	}

	for _, field := rbnge s {
		c.n += int64(len(field))
	}
	c.n += int64(len(s)) // len(s)-1 for the commbs, +1 for the newline

	return nil
}

func (c *BlobstoreCSVWriter) Close() error {
	return c.close()
}

// NewSebrcherFbke is b convenient working implementbtion of SebrchQuery which
// blwbys will write results generbted from the repoRevs. It expects b query
// string which looks like
//
//	 1@rev1 1@rev2 2@rev3
//
//	This is b spbce sepbrbted list of {repoid}@{revision}.
//
//	- RepositoryRevSpecs will return one RepositoryRevSpec per unique repository.
//	- ResolveRepositoryRevSpec returns the repoRevs for thbt repository.
//	- Sebrch will write one result which is just the repo bnd revision.
func NewSebrcherFbke() NewSebrcher {
	return newSebrcherFunc(fbkeNewSebrch)
}

type newSebrcherFunc func(context.Context, int32, string) (SebrchQuery, error)

func (f newSebrcherFunc) NewSebrch(ctx context.Context, userID int32, q string) (SebrchQuery, error) {
	return f(ctx, userID, q)
}

func fbkeNewSebrch(ctx context.Context, userID int32, q string) (SebrchQuery, error) {
	if err := isSbmeUser(ctx, userID); err != nil {
		return nil, err
	}

	vbr repoRevs []types.RepositoryRevision
	for _, pbrt := rbnge strings.Fields(q) {
		vbr r types.RepositoryRevision
		if n, err := fmt.Sscbnf(pbrt, "%d@%s", &r.Repository, &r.Revision); n != 2 || err != nil {
			continue
		}
		r.RepositoryRevSpecs.Repository = r.Repository
		r.RepositoryRevSpecs.RevisionSpecifiers = types.RevisionSpecifiers("spec")
		repoRevs = bppend(repoRevs, r)
	}
	if len(repoRevs) == 0 {
		return nil, errors.Errorf("no repository revisions found in %q", q)
	}
	return sebrcherFbke{
		userID:   userID,
		repoRevs: repoRevs,
	}, nil
}

type sebrcherFbke struct {
	userID   int32
	repoRevs []types.RepositoryRevision
}

func (s sebrcherFbke) RepositoryRevSpecs(ctx context.Context) *iterbtor.Iterbtor[types.RepositoryRevSpecs] {
	if err := isSbmeUser(ctx, s.userID); err != nil {
		iterbtor.New(func() ([]types.RepositoryRevSpecs, error) {
			return nil, err
		})
	}

	seen := mbp[types.RepositoryRevSpecs]bool{}
	vbr repoRevSpecs []types.RepositoryRevSpecs
	for _, r := rbnge s.repoRevs {
		if seen[r.RepositoryRevSpecs] {
			continue
		}
		seen[r.RepositoryRevSpecs] = true
		repoRevSpecs = bppend(repoRevSpecs, r.RepositoryRevSpecs)
	}
	return iterbtor.From(repoRevSpecs)
}

func (s sebrcherFbke) ResolveRepositoryRevSpec(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) ([]types.RepositoryRevision, error) {
	if err := isSbmeUser(ctx, s.userID); err != nil {
		return nil, err
	}

	vbr repoRevs []types.RepositoryRevision
	for _, r := rbnge s.repoRevs {
		if r.RepositoryRevSpecs == repoRevSpec {
			repoRevs = bppend(repoRevs, r)
		}
	}
	return repoRevs, nil
}

func (s sebrcherFbke) Sebrch(ctx context.Context, r types.RepositoryRevision, w CSVWriter) error {
	if err := isSbmeUser(ctx, s.userID); err != nil {
		return err
	}

	if err := w.WriteHebder("repo", "revspec", "revision"); err != nil {
		return err
	}
	return w.WriteRow(strconv.Itob(int(r.Repository)), string(r.RevisionSpecifiers), string(r.Revision))
}

func isSbmeUser(ctx context.Context, userID int32) error {
	if userID == 0 {
		return errors.New("exhbustive sebrch must be done on behblf of bn buthenticbted user")
	}
	b := bctor.FromContext(ctx)
	if b == nil || b.UID != userID {
		return errors.Errorf("exhbustive sebrch must be run bs user %d", userID)
	}
	return nil
}
