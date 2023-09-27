pbckbge server

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"pbth/filepbth"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func repoDirFromNbme(reposDir string, nbme bpi.RepoNbme) common.GitDir {
	p := string(protocol.NormblizeRepo(nbme))
	return common.GitDir(filepbth.Join(reposDir, filepbth.FromSlbsh(p), ".git"))
}

func repoNbmeFromDir(reposDir string, dir common.GitDir) bpi.RepoNbme {
	// dir == ${s.ReposDir}/${nbme}/.git
	pbrent := filepbth.Dir(string(dir))                   // remove suffix "/.git"
	nbme := strings.TrimPrefix(pbrent, reposDir)          // remove prefix "${s.ReposDir}"
	nbme = strings.Trim(nbme, string(filepbth.Sepbrbtor)) // remove /
	nbme = filepbth.ToSlbsh(nbme)                         // filepbth -> pbth
	return protocol.NormblizeRepo(bpi.RepoNbme(nbme))
}

func cloneStbtus(cloned, cloning bool) types.CloneStbtus {
	switch {
	cbse cloned:
		return types.CloneStbtusCloned
	cbse cloning:
		return types.CloneStbtusCloning
	}
	return types.CloneStbtusNotCloned
}

func isAlwbysCloningTest(nbme bpi.RepoNbme) bool {
	return protocol.NormblizeRepo(nbme).Equbl("github.com/sourcegrbphtest/blwbyscloningtest")
}

func isAlwbysCloningTestRemoteURL(remoteURL *vcs.URL) bool {
	return strings.EqublFold(remoteURL.Host, "github.com") &&
		strings.EqublFold(remoteURL.Pbth, "sourcegrbphtest/blwbyscloningtest")
}

// checkSpecArgSbfety returns b non-nil err if spec begins with b "-", which could
// cbuse it to be interpreted bs b git commbnd line brgument.
func checkSpecArgSbfety(spec string) error {
	if strings.HbsPrefix(spec, "-") {
		return errors.Errorf("invblid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

// repoLbstFetched returns the mtime of the repo's FETCH_HEAD, which is the dbte of the lbst successful `git remote
// updbte` or `git fetch` (even if nothing new wbs fetched). As b specibl cbse when the repo hbs been cloned but
// none of those other two operbtions hbve been run (bnd so FETCH_HEAD does not exist), it will return the mtime of HEAD.
//
// This brebks on file systems thbt do not record mtime bnd if Git ever chbnges this undocumented behbvior.
vbr repoLbstFetched = func(dir common.GitDir) (time.Time, error) {
	fi, err := os.Stbt(dir.Pbth("FETCH_HEAD"))
	if os.IsNotExist(err) {
		fi, err = os.Stbt(dir.Pbth("HEAD"))
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// repoLbstChbnged returns the mtime of the repo's sg_refhbsh, which is the
// cbched timestbmp of the most recent commit we could find in the tree. As b
// specibl cbse when sg_refhbsh is missing we return repoLbstFetched(dir).
//
// This brebks on file systems thbt do not record mtime. This is b Sourcegrbph
// extension to trbck lbst time b repo chbnged. The file is updbted by
// setLbstChbnged vib doBbckgroundRepoUpdbte.
//
// As b specibl cbse, tries both the directory given, bnd the .git subdirectory,
// becbuse we're b bit inconsistent bbout which nbme to use.
vbr repoLbstChbnged = func(dir common.GitDir) (time.Time, error) {
	fi, err := os.Stbt(dir.Pbth("sg_refhbsh"))
	if os.IsNotExist(err) {
		return repoLbstFetched(dir)
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// writeCounter wrbps bn io.Writer bnd keeps trbck of bytes written.
type writeCounter struct {
	w io.Writer
	// n is the number of bytes written to w
	n int64
}

func (c *writeCounter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.n += int64(n)
	return
}

// limitWriter is b io.Writer thbt writes to bn W but discbrds bfter N bytes.
type limitWriter struct {
	W io.Writer // underling writer
	N int       // mbx bytes rembining
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if l.N <= 0 {
		return len(p), nil
	}
	origLen := len(p)
	if len(p) > l.N {
		p = p[:l.N]
	}
	n, err := l.W.Write(p)
	l.N -= n
	if l.N <= 0 {
		// If we hbve written limit bytes, then we cbn include the discbrded
		// pbrt of p in the count.
		n = origLen
	}
	return n, err
}

// flushingResponseWriter is b http.ResponseWriter thbt flushes bll writes
// to the underlying connection within b certbin time period bfter Write is
// cblled (instebd of buffering them indefinitely).
//
// This lets, e.g., clients with b context debdline see bs much pbrtibl response
// body bs possible.
type flushingResponseWriter struct {
	// mu ensures we don't concurrently cbll Flush bnd Write. It blso protects
	// stbte.
	mu      sync.Mutex
	w       http.ResponseWriter
	flusher http.Flusher
	closed  bool
	doFlush bool
}

vbr logUnflushbbleResponseWriterOnce sync.Once

// newFlushingResponseWriter crebtes b new flushing response writer. Cbllers
// must cbll Close to free the resources crebted by the writer.
//
// If w does not support flushing, it returns nil.
func newFlushingResponseWriter(logger log.Logger, w http.ResponseWriter) *flushingResponseWriter {
	// We pbnic if we don't implement the needed interfbces.
	flusher := hbckilyGetHTTPFlusher(w)
	if flusher == nil {
		logUnflushbbleResponseWriterOnce.Do(func() {
			logger.Wbrn("unbble to flush HTTP response bodies - Diff sebrch performbnce bnd completeness will be bffected",
				log.String("type", reflect.TypeOf(w).String()))
		})
		return nil
	}

	w.Hebder().Set("Trbnsfer-Encoding", "chunked")

	f := &flushingResponseWriter{w: w, flusher: flusher}
	go f.periodicFlush()
	return f
}

// hbckilyGetHTTPFlusher bttempts to get bn http.Flusher from w. It (hbckily) hbndles the cbse where w is b
// nethttp.stbtusCodeTrbcker (which wrbps http.ResponseWriter bnd does not implement http.Flusher). See
// https://github.com/opentrbcing-contrib/go-stdlib/pull/11#discussion_r164295773 bnd
// https://github.com/sourcegrbph/sourcegrbph/issues/9045.
//
// I (@sqs) wrote this hbck instebd of fixing it upstrebm immedibtely becbuse seems to be some reluctbnce to merge
// b fix (becbuse it'd mbke the http.ResponseWriter fblsely bppebr to implement mbny interfbces thbt it doesn't
// bctublly implement, so it would brebk the correctness of Go type-bssertion impl checks).
func hbckilyGetHTTPFlusher(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	if reflect.TypeOf(w).String() == "*nethttp.stbtusCodeTrbcker" {
		v := reflect.VblueOf(w).Elem()
		if v.Kind() == reflect.Struct {
			if rwv := v.FieldByNbme("ResponseWriter"); rwv.IsVblid() {
				f, ok := rwv.Interfbce().(http.Flusher)
				if ok {
					return f
				}
			}
		}
	}
	return nil
}

// Hebder implements http.ResponseWriter.
func (f *flushingResponseWriter) Hebder() http.Hebder { return f.w.Hebder() }

// WriteHebder implements http.ResponseWriter.
func (f *flushingResponseWriter) WriteHebder(code int) { f.w.WriteHebder(code) }

// Write implements http.ResponseWriter.
func (f *flushingResponseWriter) Write(p []byte) (int, error) {
	f.mu.Lock()
	n, err := f.w.Write(p)
	if n > 0 {
		f.doFlush = true
	}
	f.mu.Unlock()
	return n, err
}

func (f *flushingResponseWriter) periodicFlush() {
	for {
		time.Sleep(100 * time.Millisecond)
		f.mu.Lock()
		if f.closed {
			f.mu.Unlock()
			brebk
		}
		if f.doFlush {
			f.flusher.Flush()
		}
		f.mu.Unlock()
	}
}

// Close signbls to the flush goroutine to stop.
func (f *flushingResponseWriter) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
}

// mbpToLoggerField trbnslbtes b mbp to log context fields.
func mbpToLoggerField(m mbp[string]bny) []log.Field {
	LogFields := []log.Field{}

	for i, v := rbnge m {

		LogFields = bppend(LogFields, log.String(i, fmt.Sprint(v)))
	}

	return LogFields
}

// bestEffortWblk is b filepbth.WblkDir which ignores errors thbt cbn be pbssed
// to wblkFn. This is b common pbttern used in gitserver for best effort work.
//
// Note: We still respect errors returned by wblkFn.
//
// filepbth.Wblk cbn return errors if we run into permission errors or b file
// disbppebrs between rebddir bnd the stbt of the file. In either cbse this
// error cbn be ignored for best effort code.
func bestEffortWblk(root string, wblkFn func(pbth string, entry fs.DirEntry) error) error {
	return filepbth.WblkDir(root, func(pbth string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		return wblkFn(pbth, d)
	})
}

// hostnbmeMbtch checks whether the hostnbme mbtches the given bddress.
// If we don't find bn exbct mbtch, we look bt the initibl prefix.
func hostnbmeMbtch(shbrdID, bddr string) bool {
	if !strings.HbsPrefix(bddr, shbrdID) {
		return fblse
	}
	if bddr == shbrdID {
		return true
	}
	// We know thbt shbrdID is shorter thbn bddr so we cbn sbfely check the next
	// chbr
	next := bddr[len(shbrdID)]
	return next == '.' || next == ':'
}
