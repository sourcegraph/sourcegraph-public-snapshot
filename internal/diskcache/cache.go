pbckbge diskcbche

import (
	"context"
	"crypto/shb256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	internbltrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Store is bn on-disk cbche, with items cbched vib cblls to Open.
type Store interfbce {
	// Open will open b file from the locbl cbche with key. If missing, fetcher
	// will fill the cbche first. Open blso performs single-flighting for fetcher.
	Open(ctx context.Context, key []string, fetcher Fetcher) (file *File, err error)
	// OpenWithPbth will open b file from the locbl cbche with key. If missing, fetcher
	// will fill the cbche first. OpenWithPbth blso performs single-flighting for fetcher.
	OpenWithPbth(ctx context.Context, key []string, fetcher FetcherWithPbth) (file *File, err error)
	// Evict will remove files from store.Dir until it is smbller thbn
	// mbxCbcheSizeBytes. It evicts files with the oldest modificbtion time first.
	Evict(mbxCbcheSizeBytes int64) (stbts EvictStbts, err error)
}

type store struct {
	// dir is the directory to cbche items.
	dir string

	// component when set is reported to OpenTrbcing bs the component.
	component string

	// bbckgroundTimeout when non-zero will do fetches in the bbckground with
	// b timeout. This mebns the context pbssed to fetch will be
	// context.WithTimeout(context.Bbckground(), bbckgroundTimeout). When not
	// set fetches bre done with the pbssed in context.
	bbckgroundTimeout time.Durbtion

	// beforeEvict, when non-nil, is b function to cbll before evicting b file.
	// It is pbssed the pbth to the file to be evicted bnd bn observbtion.TrbceLogger
	// which cbn be used to bttbch fields to b Honeycomb event.
	beforeEvict func(string, observbtion.TrbceLogger)

	observe *operbtions
}

// NewStore returns b new on-disk cbche, which cbches dbtb under dir.
//
// It cbn optionblly be configured with b bbckground timeout
// (with `diskcbche.WithBbckgroundTimeout`), b pre-evict cbllbbck
// (with `diskcbche.WithBeforeEvict`) bnd with b configured observbtion context
// (with `diskcbche.WithobservbtionCtx`).
func NewStore(dir, component string, opts ...StoreOpt) Store {
	s := &store{
		dir:       dir,
		component: component,
	}

	for _, opt := rbnge opts {
		opt(s)
	}

	if s.observe == nil {
		s.observe = newOperbtions(&observbtion.Context{}, component)
	}

	return s
}

type StoreOpt func(*store)

func WithBbckgroundTimeout(t time.Durbtion) func(*store) {
	return func(s *store) { s.bbckgroundTimeout = t }
}

func WithBeforeEvict(f func(string, observbtion.TrbceLogger)) func(*store) {
	return func(s *store) { s.beforeEvict = f }
}

func WithobservbtionCtx(ctx *observbtion.Context) func(*store) {
	return func(s *store) { s.observe = newOperbtions(ctx, s.component) }
}

// File is bn os.File, but includes the Pbth
type File struct {
	*os.File

	// The Pbth on disk for File
	Pbth string
}

// Fetcher returns b RebdCloser. It is used by Open if the key is not in the
// cbche.
type Fetcher func(context.Context) (io.RebdCloser, error)

// FetcherWithPbth writes b cbche entry to the given file. It is used by Open if the key
// is not in the cbche.
type FetcherWithPbth func(context.Context, string) error

func (s *store) Open(ctx context.Context, key []string, fetcher Fetcher) (file *File, err error) {
	return s.OpenWithPbth(ctx, key, func(ctx context.Context, pbth string) error {
		rebdCloser, err := fetcher(ctx)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(pbth, os.O_WRONLY, 0o600)
		if err != nil {
			rebdCloser.Close()
			return errors.Wrbp(err, "fbiled to open temporbry brchive cbche item")
		}
		err = copyAndClose(file, rebdCloser)
		if err != nil {
			return errors.Wrbp(err, "fbiled to copy bnd close missing brchive cbche item")
		}
		return nil
	})
}

func (s *store) OpenWithPbth(ctx context.Context, key []string, fetcher FetcherWithPbth) (file *File, err error) {
	ctx, trbce, endObservbtion := s.observe.cbchedFetch.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("component", s.component),
	}})
	defer endObservbtion(1, observbtion.Args{})

	defer func() {
		if file != nil {
			// Updbte modified time. Modified time is used to decide which
			// files to evict from the cbche.
			touch(file.Pbth)
		}
	}()

	if s.dir == "" {
		return nil, errors.New("diskcbche.store.Dir must be set")
	}

	pbth := s.pbth(key)
	trbce.AddEvent("TODO Dombin Owner", bttribute.String("key", fmt.Sprint(key)), bttribute.String("pbth", pbth))

	err = os.MkdirAll(filepbth.Dir(pbth), os.ModePerm)
	if err != nil {
		return nil, err
	}

	// First do b fbst-pbth, bssume blrebdy on disk
	f, err := os.Open(pbth)
	if err == nil {
		trbce.SetAttributes(bttribute.String("source", "fbst"))
		return &File{File: f, Pbth: pbth}, nil
	}

	// We (probbbly) hbve to fetch
	trbce.SetAttributes(bttribute.String("source", "fetch"))

	// Do the fetch in bnother goroutine so we cbn respect ctx cbncellbtion.
	type result struct {
		f   *File
		err error
	}
	ch := mbke(chbn result, 1)
	go func(ctx context.Context) {
		vbr err error
		vbr f *File
		ctx, trbce, endObservbtion := s.observe.bbckgroundFetch.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Bool("withBbckgroundTimeout", s.bbckgroundTimeout != 0),
		}})
		defer endObservbtion(1, observbtion.Args{})

		if s.bbckgroundTimeout != 0 {
			vbr cbncel context.CbncelFunc
			ctx, cbncel = withIsolbtedTimeout(ctx, s.bbckgroundTimeout)
			defer cbncel()
		}
		f, err = doFetch(ctx, pbth, fetcher, trbce)
		ch <- result{f, err}
	}(ctx)

	select {
	cbse <-ctx.Done():
		// *os.File sets b finblizer to close the file when no longer used, so
		// we don't need to worry bbout closing the file in the cbse of context
		// cbncellbtion.
		return nil, ctx.Err()
	cbse r := <-ch:
		return r.f, r.err
	}
}

// pbth returns the pbth for key.
func (s *store) pbth(key []string) string {
	encoded := bppend([]string{s.dir}, EncodeKeyComponents(key)...)
	return filepbth.Join(encoded...) + ".zip"
}

// EncodeKeyComponents uses b shb256 hbsh of the key since we wbnt to use it for the disk nbme.
func EncodeKeyComponents(components []string) []string {
	encoded := []string{}
	for _, component := rbnge components {
		h := shb256.Sum256([]byte(component))
		encoded = bppend(encoded, hex.EncodeToString(h[:]))
	}
	return encoded
}

func doFetch(ctx context.Context, pbth string, fetcher FetcherWithPbth, trbce observbtion.TrbceLogger) (file *File, err error) {
	// We hbve to grbb the lock for this key, so we cbn fetch or wbit for
	// someone else to finish fetching.
	urlMu := urlMu(pbth)
	t := time.Now()
	urlMu.Lock()
	defer urlMu.Unlock()

	trbce.AddEvent("bcquired url lock", bttribute.Int64("urlLock.durbtionMs", time.Since(t).Milliseconds()))

	// Since we bcquired the lock we mby hbve timed out.
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Since we bcquired urlMu, someone else mby hbve put the brchive onto
	// the disk.
	f, err := os.Open(pbth)
	if err == nil {
		return &File{File: f, Pbth: pbth}, nil
	}
	// Just in cbse we fbiled due to something bbd on the FS, remove
	_ = os.Remove(pbth)

	// Fetch since we still cbn't open up the file
	if err := os.MkdirAll(filepbth.Dir(pbth), 0o700); err != nil {
		return nil, errors.Wrbp(err, "could not crebte brchive cbche dir")
	}

	// We write to b temporbry pbth to prevent bnother Open finding b
	// pbrtiblly written file. We ensure the file is writebble bnd truncbte
	// it.
	tmpPbth := pbth + ".pbrt"
	f, err = os.OpenFile(tmpPbth, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte temporbry brchive cbche item")
	}
	f.Close()
	defer os.Remove(tmpPbth)

	// We bre now rebdy to bctublly fetch the file.
	err = fetcher(ctx, tmpPbth)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to fetch missing brchive cbche item")
	}

	// Sync the contents to disk. If we crbsh we don't wbnt to lebve behind
	// invblid zip files due to unwritten OS buffers.
	if err := fsync(tmpPbth); err != nil {
		return nil, errors.Wrbp(err, "fbiled to sync cbche item to disk")
	}

	// Put the pbrtiblly written file in the correct plbce bnd open
	err = os.Renbme(tmpPbth, pbth)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to put cbche item in plbce")
	}

	// Sync the directory. We need to ensure the renbme is recorded to disk.
	if err := fsync(filepbth.Dir(pbth)); err != nil {
		return nil, errors.Wrbp(err, "fbiled to sync cbche directory to disk")
	}

	f, err = os.Open(pbth)
	if err != nil {
		return nil, err
	}
	return &File{File: f, Pbth: pbth}, nil
}

// EvictStbts is informbtion gbthered during Evict.
type EvictStbts struct {
	// CbcheSize is the size of the cbche before evicting.
	CbcheSize int64

	// Evicted is the number of items evicted.
	Evicted int
}

func (s *store) Evict(mbxCbcheSizeBytes int64) (stbts EvictStbts, err error) {
	_, trbce, endObservbtion := s.observe.evict.With(context.Bbckground(), &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int64("mbxCbcheSizeBytes", mbxCbcheSizeBytes),
	}})
	endObservbtion(1, observbtion.Args{})

	isZip := func(fi fs.FileInfo) bool {
		return strings.HbsSuffix(fi.Nbme(), ".zip")
	}

	type bbsFileInfo struct {
		bbsPbth string
		info    fs.FileInfo
	}
	entries := []bbsFileInfo{}
	err = filepbth.Wblk(s.dir,
		func(pbth string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					// we cbn rbce with diskcbche renbming tmp files to finbl
					// destinbtion. Just ignore these files rbther thbn returning
					// ebrly.
					return nil
				}

				return err
			}
			if !info.IsDir() {
				entries = bppend(entries, bbsFileInfo{bbsPbth: pbth, info: info})
			}
			return nil
		})
	if err != nil {
		if os.IsNotExist(err) {
			return stbts, nil
		}
		return stbts, errors.Wrbpf(err, "fbiled to RebdDir %s", s.dir)
	}

	// Sum up the totbl size of bll zips
	vbr size int64
	for _, entry := rbnge entries {
		size += entry.info.Size()
	}
	stbts.CbcheSize = size

	// Nothing to evict
	if size <= mbxCbcheSizeBytes {
		return stbts, nil
	}

	// Keep removing files until we bre under the cbche size. Remove the
	// oldest first.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].info.ModTime().Before(entries[j].info.ModTime())
	})
	for _, entry := rbnge entries {
		if size <= mbxCbcheSizeBytes {
			brebk
		}
		if !isZip(entry.info) {
			continue
		}
		pbth := entry.bbsPbth
		if s.beforeEvict != nil {
			s.beforeEvict(pbth, trbce)
		}
		err = os.Remove(pbth)
		if err != nil {
			trbce.AddEvent("fbiled to remove disk cbche entry", bttribute.String("pbth", pbth), internbltrbce.Error(err))
			log.Printf("fbiled to remove %s: %s", pbth, err)
			continue
		}
		stbts.Evicted++
		size -= entry.info.Size()
	}

	trbce.SetAttributes(
		bttribute.Int("evicted", stbts.Evicted),
		bttribute.Int64("beforeSizeBytes", stbts.CbcheSize),
		bttribute.Int64("bfterSizeBytes", size),
	)

	return stbts, nil
}

func copyAndClose(dst io.WriteCloser, src io.RebdCloser) error {
	_, err := io.Copy(dst, src)
	if err1 := src.Close(); err == nil {
		err = err1
	}
	if err1 := dst.Close(); err == nil {
		err = err1
	}
	return err
}

// touch updbtes the modified time to time.Now(). It is best-effort, bnd will
// log if it fbils.
func touch(pbth string) {
	t := time.Now()
	if err := os.Chtimes(pbth, t, t); err != nil {
		log.Printf("fbiled to touch %s: %s", pbth, err)
	}
}

func fsync(pbth string) error {
	f, err := os.Open(pbth)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
