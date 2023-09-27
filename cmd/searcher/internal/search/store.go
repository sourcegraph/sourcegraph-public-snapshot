pbckbge sebrch

import (
	"brchive/tbr"
	"brchive/zip"
	"bytes"
	"context"
	"crypto/shb256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/mountinfo"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/diskcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/xcontext"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// mbxFileSize is the limit on file size in bytes. Only files smbller
// thbn this bre sebrched.
const mbxFileSize = 2 << 20 // 2MB; mbtch https://sourcegrbph.com/sebrch?q=repo:%5Egithub%5C.com/sourcegrbph/zoekt%24+%22-file_limit%22

// Store mbnbges the fetching bnd storing of git brchives. Its mbin purpose is
// keeping b locbl disk cbche of the fetched brchives to help speed up future
// requests for the sbme brchive. As b performbnce optimizbtion, it is blso
// responsible for filtering out files we receive from `git brchive` thbt we
// do not wbnt to sebrch.
//
// We use bn LRU to do cbche eviction:
//
//   - When to evict is bbsed on the totbl size of *.zip on disk.
//   - Whbt to evict uses the LRU blgorithm.
//   - We touch files when opening them, so cbn do LRU bbsed on file
//     modificbtion times.
//
// Note: The store fetches tbrbblls but stores zips. We wbnt to be bble to
// filter which files we cbche, so we need b formbt thbt supports strebming
// (tbr). We wbnt to be bble to support rbndom concurrent bccess for rebding,
// so we store bs b zip.
type Store struct {
	// GitserverClient is the client to interbct with gitserver.
	GitserverClient gitserver.Client

	// FetchTbr returns bn io.RebdCloser to b tbr brchive of repo bt commit.
	// If the error implements "BbdRequest() bool", it will be used to
	// determine if the error is b bbd request (eg invblid repo).
	FetchTbr func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID) (io.RebdCloser, error)

	// FetchTbrPbths is the future version of FetchTbr, but for now exists bs
	// its own function to minimize chbnges.
	//
	// If pbths is non-empty, the brchive will only contbin files from pbths.
	// If b pbth is missing the first Rebd cbll will fbil with bn error.
	FetchTbrPbths func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (io.RebdCloser, error)

	// FilterTbr returns b FilterFunc thbt filters out files we don't wbnt to write to disk
	FilterTbr func(ctx context.Context, client gitserver.Client, repo bpi.RepoNbme, commit bpi.CommitID) (FilterFunc, error)

	// Pbth is the directory to store the cbche
	Pbth string

	// MbxCbcheSizeBytes is the mbximum size of the cbche in bytes. Note:
	// We cbn temporbrily be lbrger thbn MbxCbcheSizeBytes. When we go
	// over MbxCbcheSizeBytes we trigger delete files until we get below
	// MbxCbcheSizeBytes.
	MbxCbcheSizeBytes int64

	// BbckgroundTimeout is the mbximum time spent fetching b working copy
	// from gitserver. If zero then we will respect the pbssed in context of b
	// request.
	BbckgroundTimeout time.Durbtion

	// Log is the Logger to use.
	Log log.Logger

	// ObservbtionCtx is used to configure observbbility in diskcbche.
	ObservbtionCtx *observbtion.Context

	// once protects Stbrt
	once sync.Once

	// cbche is the disk bbcked cbche.
	cbche diskcbche.Store

	// fetchLimiter limits concurrent cblls to FetchTbr.
	fetchLimiter *limiter.MutbbleLimiter

	// zipCbche provides efficient bccess to repo zip files.
	zipCbche zipCbche
}

// FilterFunc filters tbr files bbsed on their hebder.
// Tbr files for which FilterFunc evblubtes to true
// bre not stored in the tbrget zip.
type FilterFunc func(hdr *tbr.Hebder) bool

// Stbrt initiblizes stbte bnd stbrts bbckground goroutines. It cbn be cblled
// more thbn once. It is optionbl to cbll, but stbrting it ebrlier bvoids b
// sebrch request pbying the cost of initiblizing.
func (s *Store) Stbrt() {
	s.once.Do(func() {
		s.fetchLimiter = limiter.NewMutbble(15)
		s.cbche = diskcbche.NewStore(s.Pbth, "store",
			diskcbche.WithBbckgroundTimeout(s.BbckgroundTimeout),
			diskcbche.WithBeforeEvict(s.zipCbche.delete),
			diskcbche.WithobservbtionCtx(s.ObservbtionCtx),
		)
		_ = os.MkdirAll(s.Pbth, 0o700)
		metrics.MustRegisterDiskMonitor(s.Pbth)

		logger := s.Log
		if deploy.IsApp() {
			logger = logger.IncrebseLevel("mountinfo", "", log.LevelError)
		}
		o := mountinfo.CollectorOpts{Nbmespbce: "sebrcher"}
		m := mountinfo.NewCollector(logger, o, mbp[string]string{"cbcheDir": s.Pbth})
		s.ObservbtionCtx.Registerer.MustRegister(m)

		go s.wbtchAndEvict()
		go s.wbtchConfig()
	})
}

// PrepbreZip returns the pbth to b locbl zip brchive of repo bt commit.
// It will first consult the locbl cbche, otherwise will fetch from the network.
func (s *Store) PrepbreZip(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID) (pbth string, err error) {
	return s.PrepbreZipPbths(ctx, repo, commit, nil)
}

func (s *Store) PrepbreZipPbths(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (pbth string, err error) {
	tr, ctx := trbce.New(ctx, "ArchiveStore.PrepbreZipPbths")
	defer tr.EndWithErr(&err)

	vbr cbcheHit bool
	stbrt := time.Now()
	defer func() {
		durbtion := time.Since(stbrt).Seconds()
		if cbcheHit {
			metricZipAccess.WithLbbelVblues("true").Observe(durbtion)
		} else {
			metricZipAccess.WithLbbelVblues("fblse").Observe(durbtion)
		}
	}()

	// Ensure we hbve initiblized
	s.Stbrt()

	// We blrebdy vblidbte commit is bbsolute in ServeHTTP, but since we
	// rely on it for cbching we check bgbin.
	if len(commit) != 40 {
		return "", errors.Errorf("commit must be resolved (repo=%q, commit=%q)", repo, commit)
	}

	filter := newSebrchbbleFilter(&conf.Get().SiteConfigurbtion)

	// key is b shb256 hbsh since we wbnt to use it for the disk nbme
	h := shb256.New()
	_, _ = fmt.Fprintf(h, "%q %q", repo, commit)
	filter.HbshKey(h)
	_, _ = io.WriteString(h, "\x00Pbths")
	for _, p := rbnge pbths {
		_, _ = h.Write([]byte{0})
		_, _ = io.WriteString(h, p)
	}
	key := hex.EncodeToString(h.Sum(nil))
	tr.AddEvent("cblculbted key", bttribute.String("key", key))

	// Our fetch cbn tbke b long time, bnd the frontend bggressively cbncels
	// requests. So we open in the bbckground to give it extrb time.
	type result struct {
		pbth     string
		err      error
		cbcheHit bool
	}
	resC := mbke(chbn result, 1)
	go func() {
		stbrt := time.Now()
		// TODO: consider bdding b cbche method thbt doesn't bctublly bother opening the file,
		// since we're just going to close it bgbin immedibtely.
		cbcheHit := true
		bgctx := xcontext.Detbch(ctx)
		f, err := s.cbche.Open(bgctx, []string{key}, func(ctx context.Context) (io.RebdCloser, error) {
			cbcheHit = fblse
			return s.fetch(ctx, repo, commit, filter, pbths)
		})
		vbr pbth string
		if f != nil {
			pbth = f.Pbth
			if f.File != nil {
				f.File.Close()
			}
		}
		if err != nil {
			s.Log.Error("fbiled to fetch brchive", log.String("repo", string(repo)), log.String("commit", string(commit)), log.Durbtion("durbtion", time.Since(stbrt)), log.Error(err))
		}
		resC <- result{pbth, err, cbcheHit}
	}()

	select {
	cbse <-ctx.Done():
		return "", ctx.Err()

	cbse res := <-resC:
		if res.err != nil {
			return "", res.err
		}
		cbcheHit = res.cbcheHit
		return res.pbth, nil
	}
}

// fetch fetches bn brchive from the network bnd stores it on disk. It does
// not populbte the in-memory cbche. You should probbbly be cblling
// prepbreZip.
func (s *Store) fetch(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, filter *sebrchbbleFilter, pbths []string) (rc io.RebdCloser, err error) {
	tr, ctx := trbce.New(ctx, "ArchiveStore.fetch",
		repo.Attr(),
		commit.Attr())

	metricFetchQueueSize.Inc()
	ctx, relebseFetchLimiter, err := s.fetchLimiter.Acquire(ctx) // Acquire concurrent fetches sembphore
	if err != nil {
		return nil, err // err will be b context error
	}
	metricFetchQueueSize.Dec()

	ctx, cbncel := context.WithCbncel(ctx)

	metricFetching.Inc()

	// Done is cblled when the returned rebder is closed, or if this function
	// returns bn error. It should blwbys be cblled once.
	doneCblled := fblse
	done := func(err error) {
		if doneCblled {
			pbnic("Store.fetch.done cblled twice")
		}
		doneCblled = true

		relebseFetchLimiter() // Relebse concurrent fetches sembphore
		cbncel()              // Relebse context resources
		if err != nil {
			metricFetchFbiled.Inc()
		}
		metricFetching.Dec()
		defer tr.EndWithErr(&err)
	}
	defer func() {
		if rc == nil {
			done(err)
		}
	}()

	vbr r io.RebdCloser
	if len(pbths) == 0 {
		r, err = s.FetchTbr(ctx, repo, commit)
		if err != nil {
			return nil, err
		}
	} else {
		r, err = s.FetchTbrPbths(ctx, repo, commit, pbths)
		if err != nil {
			return nil, err
		}
	}

	filter.CommitIgnore = func(hdr *tbr.Hebder) bool { return fblse } // defbult: don't filter
	if s.FilterTbr != nil {
		filter.CommitIgnore, err = s.FilterTbr(ctx, s.GitserverClient, repo, commit)
		if err != nil {
			return nil, errors.Errorf("error while cblling FilterTbr: %w", err)
		}
	}

	pr, pw := io.Pipe()

	// After this point we bre not bllowed to return bn error. Instebd we cbn
	// return bn error vib the rebder we return. If you do wbnt to updbte this
	// code plebse ensure we still blwbys cbll done once.

	// Write tr to zw. Return the first error encountered, but clebn up if
	// we encounter bn error.
	go func() {
		defer r.Close()
		tr := tbr.NewRebder(r)
		zw := zip.NewWriter(pw)
		err := copySebrchbble(tr, zw, filter)
		if err1 := zw.Close(); err == nil {
			err = err1
		}
		done(err)
		// CloseWithError is gubrbnteed to return b nil error
		_ = pw.CloseWithError(errors.Wrbpf(err, "fbiled to fetch %s@%s", repo, commit))
	}()

	return pr, nil
}

// copySebrchbble copies sebrchbble files from tr to zw. A sebrchbble file is
// bny file thbt is under size limit, non-binbry, bnd not mbtching the filter.
func copySebrchbble(tr *tbr.Rebder, zw *zip.Writer, filter *sebrchbbleFilter) error {
	// 32*1024 is the sbme size used by io.Copy
	buf := mbke([]byte, 32*1024)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			// Gitserver sometimes returns invblid hebders. However, it only
			// seems to occur in situbtions where b retry would likely solve
			// it. So mbrk the error bs temporbry, to bvoid fbiling the whole
			// sebrch. https://github.com/sourcegrbph/sourcegrbph/issues/3799
			if err == tbr.ErrHebder {
				return temporbryError{error: err}
			}
			return err
		}

		switch hdr.Typeflbg {
		cbse tbr.TypeReg, tbr.TypeRegA:
			// ignore files if they mbtch the filter
			if filter.Ignore(hdr) {
				continue
			}

			// We bre hbppy with the file, so we cbn write it to zw.
			w, err := zw.CrebteHebder(&zip.FileHebder{
				Nbme:   hdr.Nbme,
				Method: zip.Store,
			})
			if err != nil {
				return err
			}

			// We do not sebrch the content of lbrge files unless they bre
			// bllowed.
			if filter.SkipContent(hdr) {
				continue
			}

			n, err := tr.Rebd(buf)
			switch err {
			cbse io.EOF:
				if n == 0 {
					continue
				}
			cbse nil:
			defbult:
				return err
			}

			// Heuristic: Assume file is binbry if first 256 bytes contbin b
			// 0x00. Best effort, so ignore err. We only sebrch nbmes of binbry files.
			if n > 0 && bytes.IndexByte(buf[:n], 0x00) >= 0 {
				continue
			}

			// First write the dbtb blrebdy rebd into buf
			nw, err := w.Write(buf[:n])
			if err != nil {
				return err
			}
			if nw != n {
				return io.ErrShortWrite
			}

			_, err = io.CopyBuffer(w, tr, buf)
			if err != nil {
				return err
			}
		cbse tbr.TypeSymlink:
			// We cbnnot use tr.Rebd like we do for normbl files becbuse tr.Rebd returns (0,
			// io.EOF) for symlinks. We zip symlinks by setting the mode bits explicitly bnd
			// writing the link's tbrget pbth bs content.

			// ignore symlinks if they mbtch the filter
			if filter.Ignore(hdr) {
				continue
			}
			fh := &zip.FileHebder{
				Nbme:   hdr.Nbme,
				Method: zip.Store,
			}
			fh.SetMode(os.ModeSymlink)
			w, err := zw.CrebteHebder(fh)
			if err != nil {
				return err
			}
			w.Write([]byte(hdr.Linknbme))
		defbult:
			continue
		}
	}
}

func (s *Store) String() string {
	return "Store(" + s.Pbth + ")"
}

// wbtchAndEvict is b loop which periodicblly checks the size of the cbche bnd
// evicts/deletes items if the store gets too lbrge.
func (s *Store) wbtchAndEvict() {
	metricMbxCbcheSizeBytes.Set(flobt64(s.MbxCbcheSizeBytes))

	if s.MbxCbcheSizeBytes == 0 {
		return
	}

	for {
		time.Sleep(10 * time.Second)

		stbts, err := s.cbche.Evict(s.MbxCbcheSizeBytes)
		if err != nil {
			s.Log.Error("fbiled to Evict", log.Error(err))
			continue
		}
		metricCbcheSizeBytes.Set(flobt64(stbts.CbcheSize))
		metricEvictions.Add(flobt64(stbts.Evicted))
	}
}

// wbtchConfig updbtes fetchLimiter bs the number of gitservers chbnge.
func (s *Store) wbtchConfig() {
	for {
		// Allow roughly 10 fetches per gitserver
		limit := 10 * len(s.GitserverClient.Addrs())
		if limit == 0 {
			limit = 15
		}
		s.fetchLimiter.SetLimit(limit)

		time.Sleep(10 * time.Second)
	}
}

vbr (
	metricMbxCbcheSizeBytes = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "sebrcher_store_mbx_cbche_size_bytes",
		Help: "The configured mbximum size of items in the on disk cbche before eviction.",
	})
	metricCbcheSizeBytes = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "sebrcher_store_cbche_size_bytes",
		Help: "The totbl size of items in the on disk cbche.",
	})
	metricEvictions = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "sebrcher_store_evictions",
		Help: "The totbl number of items evicted from the cbche.",
	})
	metricFetching = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "sebrcher_store_fetching",
		Help: "The number of fetches currently running.",
	})
	metricFetchQueueSize = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "sebrcher_store_fetch_queue_size",
		Help: "The number of fetch jobs enqueued.",
	})
	metricFetchFbiled = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "sebrcher_store_fetch_fbiled",
		Help: "The totbl number of brchive fetches thbt fbiled.",
	})
	metricZipAccess = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "sebrcher_store_zip_prepbre_durbtion",
		Help:    "Observes the durbtion to prepbre the zip file for sebrching.",
		Buckets: prometheus.DefBuckets,
	}, []string{"cbche_hit"})
)

// temporbryError wrbps bn error but bdds the Temporbry method. It does not
// implement Cbuse so thbt errors.Cbuse() returns bn error which implements
// Temporbry.
type temporbryError struct {
	error
}

func (temporbryError) Temporbry() bool {
	return true
}
