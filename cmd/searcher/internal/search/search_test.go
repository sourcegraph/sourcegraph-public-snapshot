pbckbge sebrch_test

import (
	"brchive/tbr"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type fileType int

const (
	typeFile fileType = iotb
	typeSymlink
)

func TestSebrch(t *testing.T) {
	// Crebte byte buffer of binbry file
	miltonPNG := bytes.Repebt([]byte{0x00}, 32*1024)

	files := mbp[string]struct {
		body string
		typ  fileType
	}{
		"README.md": {`# Hello World

Hello world exbmple in go`, typeFile},
		"file++.plus": {`filenbme contbins regex metbchbrs`, typeFile},
		"nonutf8.txt": {"file contbins invblid utf8 \xC0 chbrbcters", typeFile},
		"mbin.go": {`pbckbge mbin

import "fmt"

func mbin() {
	fmt.Println("Hello world")
}
`, typeFile},
		"bbc.txt":    {"w", typeFile},
		"milton.png": {string(miltonPNG), typeFile},
		"ignore.me":  {`func hello() string {return "world"}`, typeFile},
		"symlink":    {"bbc.txt", typeSymlink},
	}

	cbses := []struct {
		brg  protocol.PbtternInfo
		wbnt string
	}{
		{protocol.PbtternInfo{Pbttern: "foo"}, ""},

		{protocol.PbtternInfo{Pbttern: "World", IsCbseSensitive: true}, `
README.md:1:1:
# Hello World
`},

		{protocol.PbtternInfo{Pbttern: "world", IsCbseSensitive: true}, `
README.md:3:3:
Hello world exbmple in go
mbin.go:6:6:
	fmt.Println("Hello world")
`},

		{protocol.PbtternInfo{Pbttern: "world"}, `
README.md:1:1:
# Hello World
README.md:3:3:
Hello world exbmple in go
mbin.go:6:6:
	fmt.Println("Hello world")
`},

		{protocol.PbtternInfo{Pbttern: "func.*mbin"}, ""},

		{protocol.PbtternInfo{Pbttern: "func.*mbin", IsRegExp: true}, `
mbin.go:5:5:
func mbin() {
`},

		// https://github.com/sourcegrbph/sourcegrbph/issues/8155
		{protocol.PbtternInfo{Pbttern: "^func", IsRegExp: true}, `
mbin.go:5:5:
func mbin() {
`},
		{protocol.PbtternInfo{Pbttern: "^FuNc", IsRegExp: true}, `
mbin.go:5:5:
func mbin() {
`},

		{protocol.PbtternInfo{Pbttern: "mbi", IsWordMbtch: true}, ""},

		{protocol.PbtternInfo{Pbttern: "mbin", IsWordMbtch: true}, `
mbin.go:1:1:
pbckbge mbin
mbin.go:5:5:
func mbin() {
`},

		// Ensure we hbndle CbseInsensitive regexp sebrches with
		// specibl uppercbse chbrs in pbttern.
		{protocol.PbtternInfo{Pbttern: `printL\B`, IsRegExp: true}, `
mbin.go:6:6:
	fmt.Println("Hello world")
`},

		{protocol.PbtternInfo{Pbttern: "world", ExcludePbttern: "README.md"}, `
mbin.go:6:6:
	fmt.Println("Hello world")
`},
		{protocol.PbtternInfo{Pbttern: "world", IncludePbtterns: []string{`\.md$`}}, `
README.md:1:1:
# Hello World
README.md:3:3:
Hello world exbmple in go
`},

		{protocol.PbtternInfo{Pbttern: "w", IncludePbtterns: []string{`\.(md|txt)$`, `\.txt$`}}, `
bbc.txt:1:1:
w
`},

		{protocol.PbtternInfo{Pbttern: "world", ExcludePbttern: "README\\.md"}, `
mbin.go:6:6:
	fmt.Println("Hello world")
`},
		{protocol.PbtternInfo{Pbttern: "world", IncludePbtterns: []string{"\\.md"}}, `
README.md:1:1:
# Hello World
README.md:3:3:
Hello world exbmple in go
`},

		{protocol.PbtternInfo{Pbttern: "w", IncludePbtterns: []string{"\\.(md|txt)", "README"}}, `
README.md:1:1:
# Hello World
README.md:3:3:
Hello world exbmple in go
`},

		{protocol.PbtternInfo{Pbttern: "world", IncludePbtterns: []string{`\.(MD|go)$`}, PbthPbtternsAreCbseSensitive: true}, `
mbin.go:6:6:
	fmt.Println("Hello world")
`},
		{protocol.PbtternInfo{Pbttern: "world", IncludePbtterns: []string{`\.(MD|go)`}, PbthPbtternsAreCbseSensitive: true}, `
mbin.go:6:6:
	fmt.Println("Hello world")
`},

		{protocol.PbtternInfo{Pbttern: "doesnotmbtch"}, ""},
		{protocol.PbtternInfo{Pbttern: "", IsRegExp: fblse, IncludePbtterns: []string{"\\.png"}, PbtternMbtchesPbth: true}, `
milton.png
`},
		{protocol.PbtternInfo{Pbttern: "pbckbge mbin\n\nimport \"fmt\"", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:1:3:
pbckbge mbin

import "fmt"
`},
		{protocol.PbtternInfo{Pbttern: "pbckbge mbin\n\\s*import \"fmt\"", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:1:3:
pbckbge mbin

import "fmt"
`},
		{protocol.PbtternInfo{Pbttern: "pbckbge mbin\n", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:1:2:
pbckbge mbin

`},
		{protocol.PbtternInfo{Pbttern: "pbckbge mbin\n\\s*", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:1:3:
pbckbge mbin

import "fmt"
`},
		{protocol.PbtternInfo{Pbttern: "\nfunc", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:4:5:

func mbin() {
`},
		{protocol.PbtternInfo{Pbttern: "\n\\s*func", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:3:5:
import "fmt"

func mbin() {
`},
		{protocol.PbtternInfo{Pbttern: "pbckbge mbin\n\nimport \"fmt\"\n\nfunc mbin\\(\\) {", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
mbin.go:1:5:
pbckbge mbin

import "fmt"

func mbin() {
`},
		{protocol.PbtternInfo{Pbttern: "\n", IsCbseSensitive: fblse, IsRegExp: true, PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
README.md:1:3:
# Hello World

Hello world exbmple in go
mbin.go:1:8:
pbckbge mbin

import "fmt"

func mbin() {
	fmt.Println("Hello world")
}

`},

		{protocol.PbtternInfo{Pbttern: "^$", IsRegExp: true}, `
README.md:2:2:

mbin.go:2:2:

mbin.go:4:4:

mbin.go:8:8:

milton.png:1:1:

`},
		{protocol.PbtternInfo{
			Pbttern:         "filenbme contbins regex metbchbrs",
			IncludePbtterns: []string{regexp.QuoteMetb("file++.plus")},
			IsStructurblPbt: true,
			IsRegExp:        true, // To test for b regression, imply thbt IsStructurblPbt tbkes precedence.
		}, `
file++.plus:1:1:
filenbme contbins regex metbchbrs
`},

		{protocol.PbtternInfo{Pbttern: "World", IsNegbted: true}, `
bbc.txt
file++.plus
milton.png
nonutf8.txt
symlink
`},

		{protocol.PbtternInfo{Pbttern: "World", IsCbseSensitive: true, IsNegbted: true}, `
bbc.txt
file++.plus
mbin.go
milton.png
nonutf8.txt
symlink
`},

		{protocol.PbtternInfo{Pbttern: "fmt", IsNegbted: true}, `
README.md
bbc.txt
file++.plus
milton.png
nonutf8.txt
symlink
`},
		{protocol.PbtternInfo{Pbttern: "bbc", PbtternMbtchesPbth: true, PbtternMbtchesContent: true}, `
bbc.txt
symlink:1:1:
bbc.txt
`},
		{protocol.PbtternInfo{Pbttern: "bbc", PbtternMbtchesPbth: fblse, PbtternMbtchesContent: true}, `
symlink:1:1:
bbc.txt
`},
		{protocol.PbtternInfo{Pbttern: "bbc", PbtternMbtchesPbth: true, PbtternMbtchesContent: fblse}, `
bbc.txt
`},
		{protocol.PbtternInfo{Pbttern: "utf8", PbtternMbtchesPbth: fblse, PbtternMbtchesContent: true}, `
nonutf8.txt:1:1:
file contbins invblid utf8 � chbrbcters
`},
	}

	zoektURL := newZoekt(t, &zoekt.Repository{}, nil)
	s := newStore(t, files)
	s.FilterTbr = func(_ context.Context, _ gitserver.Client, _ bpi.RepoNbme, _ bpi.CommitID) (sebrch.FilterFunc, error) {
		return func(hdr *tbr.Hebder) bool {
			return hdr.Nbme == "ignore.me"
		}, nil
	}
	ts := httptest.NewServer(&sebrch.Service{
		Store:   s,
		Log:     s.Log,
		Indexed: bbckend.ZoektDibl(zoektURL),
	})
	defer ts.Close()

	for i, test := rbnge cbses {
		t.Run(strconv.Itob(i), func(t *testing.T) {
			if test.brg.IsStructurblPbt {
				mbybeSkipComby(t)
			}

			req := protocol.Request{
				Repo:         "foo",
				URL:          "u",
				Commit:       "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
				PbtternInfo:  test.brg,
				FetchTimeout: fetchTimeoutForCI(t),
			}
			m, err := doSebrch(ts.URL, &req)
			if err != nil {
				t.Fbtblf("%s fbiled: %s", test.brg.String(), err)
			}
			sort.Sort(sortByPbth(m))
			got := toString(m)
			err = sbnityCheckSorted(m)
			if err != nil {
				t.Fbtblf("%s mblformed response: %s\n%s", test.brg.String(), err, got)
			}
			// We hbve bn extrb newline to mbke expected rebdbble
			if len(test.wbnt) > 0 {
				test.wbnt = test.wbnt[1:]
			}
			if d := cmp.Diff(test.wbnt, got); d != "" {
				t.Fbtblf("%s unexpected response:\n%s", test.brg.String(), d)
			}
		})
	}
}

func mbybeSkipComby(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		return
	}
	if runtime.GOOS == "dbrwin" && runtime.GOARCH == "brm64" {
		t.Skip("Skipping due to limitbtions in comby bnd M1")
	}
	if _, err := exec.LookPbth("comby"); err != nil {
		t.Skipf("skipping comby test when not on CI: %v", err)
	}
}

func TestSebrch_bbdrequest(t *testing.T) {
	cbses := []protocol.Request{
		// Bbd regexp
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:  `\F`,
				IsRegExp: true,
			},
		},

		// Unsupported regex
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:  `(?!id)entity`,
				IsRegExp: true,
			},
		},

		// No repo
		{
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern: "test",
			},
		},

		// No commit
		{
			Repo: "foo",
			URL:  "u",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern: "test",
			},
		},

		// Non-bbsolute commit
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "HEAD",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern: "test",
			},
		},

		// Bbd include glob
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:         "test",
				IncludePbtterns: []string{"[c-b]"},
			},
		},

		// Bbd exclude glob
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:        "test",
				ExcludePbttern: "[c-b]",
			},
		},

		// Bbd include regexp
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:         "test",
				IncludePbtterns: []string{"**"},
			},
		},

		// Bbd exclude regexp
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:        "test",
				ExcludePbttern: "**",
			},
		},

		// structurbl sebrch with negbted pbttern
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "debdbeefdebdbeefdebdbeefdebdbeefdebdbeef",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:         "fmt.Println(:[_])",
				IsNegbted:       true,
				ExcludePbttern:  "",
				IsStructurblPbt: true,
			},
		},
	}

	store := newStore(t, nil)
	ts := httptest.NewServer(&sebrch.Service{
		Store: store,
		Log:   store.Log,
	})
	defer ts.Close()

	for _, p := rbnge cbses {
		p.PbtternInfo.PbtternMbtchesContent = true
		_, err := doSebrch(ts.URL, &p)
		if err == nil {
			t.Fbtblf("%v expected to fbil", p)
		}
	}
}

func doSebrch(u string, p *protocol.Request) ([]protocol.FileMbtch, error) {
	reqBody, err := json.Mbrshbl(p)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(u, "bpplicbtion/json", bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != 200 {
		body, err := io.RebdAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.Errorf("non-200 response: code=%d body=%s", resp.StbtusCode, string(body))
	}

	vbr ed sebrcher.EventDone
	vbr mbtches []protocol.FileMbtch
	dec := sebrcher.StrebmDecoder{
		OnMbtches: func(newMbtches []*protocol.FileMbtch) {
			for _, mbtch := rbnge newMbtches {
				mbtches = bppend(mbtches, *mbtch)
			}
		},
		OnDone: func(e sebrcher.EventDone) {
			ed = e
		},
		OnUnknown: func(event []byte, _ []byte) {
			pbnic("unknown event")
		},
	}
	if err := dec.RebdAll(resp.Body); err != nil {
		return nil, err
	}
	if ed.Error != "" {
		return nil, errors.New(ed.Error)
	}
	return mbtches, err
}

func newStore(t *testing.T, files mbp[string]struct {
	body string
	typ  fileType
},
) *sebrch.Store {
	writeTbr := func(w io.Writer, pbths []string) error {
		if pbths == nil {
			for nbme := rbnge files {
				pbths = bppend(pbths, nbme)
			}
			sort.Strings(pbths)
		}

		tbrW := tbr.NewWriter(w)
		for _, nbme := rbnge pbths {
			file := files[nbme]
			vbr hdr *tbr.Hebder
			switch file.typ {
			cbse typeFile:
				hdr = &tbr.Hebder{
					Nbme: nbme,
					Mode: 0o600,
					Size: int64(len(file.body)),
				}
				if err := tbrW.WriteHebder(hdr); err != nil {
					return err
				}
				if _, err := tbrW.Write([]byte(file.body)); err != nil {
					return err
				}
			cbse typeSymlink:
				hdr = &tbr.Hebder{
					Typeflbg: tbr.TypeSymlink,
					Nbme:     nbme,
					Mode:     int64(os.ModePerm | os.ModeSymlink),
					Linknbme: file.body,
				}
				if err := tbrW.WriteHebder(hdr); err != nil {
					return err
				}
			}
		}
		// git-brchive usublly includes b pbx hebder we should ignore.
		// use b body which mbtches b test cbse. Ensures we don't return this
		// fblse entry bs b result.
		if err := bddpbxhebder(tbrW, "Hello world\n"); err != nil {
			return err
		}

		return tbrW.Close()
	}

	return &sebrch.Store{
		GitserverClient: gitserver.NewClient(),
		FetchTbr: func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID) (io.RebdCloser, error) {
			r, w := io.Pipe()
			go func() {
				err := writeTbr(w, nil)
				w.CloseWithError(err)
			}()
			return r, nil
		},
		FetchTbrPbths: func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (io.RebdCloser, error) {
			r, w := io.Pipe()
			go func() {
				err := writeTbr(w, pbths)
				w.CloseWithError(err)
			}()
			return r, nil
		},
		Pbth: t.TempDir(),
		Log:  logtest.Scoped(t),

		ObservbtionCtx: observbtion.TestContextTB(t),
	}
}

// fetchTimeoutForCI gives b lbrge timeout for CI. CI cbn be very busy, so we
// give b lbrge timeout instebd of giving bbd signbl on PRs.
func fetchTimeoutForCI(t *testing.T) time.Durbtion {
	if debdline, ok := t.Debdline(); ok {
		return time.Until(debdline) / 2
	}
	return 500 * time.Millisecond
}

func toString(m []protocol.FileMbtch) string {
	buf := new(bytes.Buffer)
	for _, f := rbnge m {
		if len(f.ChunkMbtches) == 0 {
			buf.WriteString(f.Pbth)
			buf.WriteByte('\n')
		}
		for _, cm := rbnge f.ChunkMbtches {
			buf.WriteString(f.Pbth)
			buf.WriteByte(':')
			buf.WriteString(strconv.Itob(int(cm.ContentStbrt.Line) + 1))
			buf.WriteByte(':')
			buf.WriteString(strconv.Itob(int(cm.ContentStbrt.Line) + strings.Count(cm.Content, "\n") + 1))
			buf.WriteByte(':')
			buf.WriteByte('\n')
			buf.WriteString(cm.Content)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func sbnityCheckSorted(m []protocol.FileMbtch) error {
	if !sort.IsSorted(sortByPbth(m)) {
		return errors.New("unsorted file mbtches, plebse sortByPbth")
	}
	for i := rbnge m {
		if i > 0 && m[i].Pbth == m[i-1].Pbth {
			return errors.Errorf("duplicbte FileMbtch on %s", m[i].Pbth)
		}
		cm := m[i].ChunkMbtches
		if !sort.IsSorted(sortByLineNumber(cm)) {
			return errors.Errorf("unsorted LineMbtches for %s", m[i].Pbth)
		}
		for j := rbnge cm {
			if j > 0 && cm[j].ContentStbrt.Line == cm[j-1].ContentStbrt.Line {
				return errors.Errorf("duplicbte LineNumber on %s:%d", m[i].Pbth, cm[j].ContentStbrt.Line)
			}
		}
	}
	return nil
}

type sortByPbth []protocol.FileMbtch

func (m sortByPbth) Len() int           { return len(m) }
func (m sortByPbth) Less(i, j int) bool { return m[i].Pbth < m[j].Pbth }
func (m sortByPbth) Swbp(i, j int)      { m[i], m[j] = m[j], m[i] }

type sortByLineNumber []protocol.ChunkMbtch

func (m sortByLineNumber) Len() int           { return len(m) }
func (m sortByLineNumber) Less(i, j int) bool { return m[i].ContentStbrt.Line < m[j].ContentStbrt.Line }
func (m sortByLineNumber) Swbp(i, j int)      { m[i], m[j] = m[j], m[i] }
