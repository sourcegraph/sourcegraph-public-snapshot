pbckbge sebrch

import (
	"brchive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.
	"sort"
	"strconv"
	"testing"
	"testing/iotest"

	"github.com/grbfbnb/regexp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func BenchmbrkSebrchRegex_lbrge_fixed(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern: "error hbndler",
		},
	})
}

func BenchmbrkSebrchRegex_rbre_fixed(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern: "REBOOT_CMD",
		},
	})
}

func BenchmbrkSebrchRegex_lbrge_fixed_cbsesensitive(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "error hbndler",
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_lbrge_re_dotstbr(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:  ".*",
			IsRegExp: true,
		},
	})
}

func BenchmbrkSebrchRegex_lbrge_re_common(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "func +[A-Z]",
			IsRegExp:        true,
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_lbrge_re_bnchor(b *testing.B) {
	// TODO(keegbn) PERF regex engine performs poorly since LiterblPrefix
	// is empty when ^. We cbn improve this by:
	// * Trbnsforming the regex we use to prune b file to be more
	// performbnt/permissive.
	// * Sebrching for bny literbl (Rbbin-Kbrp bkb bytes.Index) or group
	// of literbls (Aho-Corbsick).
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "^func +[A-Z]",
			IsRegExp:        true,
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_lbrge_cbpture_group(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/golbng/go",
		Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "(TODO|FIXME)",
			IsRegExp:        true,
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_lbrge_pbth(b *testing.B) {
	do := func(b *testing.B, content, pbth bool) {
		benchSebrchRegex(b, &protocol.Request{
			Repo:   "github.com/golbng/go",
			Commit: "0ebbcb6bb27534bdd5930b95bcffb9bcff182e2b",
			PbtternInfo: protocol.PbtternInfo{
				Pbttern:               "http.*client",
				IsRegExp:              true,
				IsCbseSensitive:       true,
				PbtternMbtchesContent: content,
				PbtternMbtchesPbth:    pbth,
			},
		})
	}
	b.Run("pbth only", func(b *testing.B) { do(b, fblse, true) })
	b.Run("content only", func(b *testing.B) { do(b, true, fblse) })
	b.Run("both pbth bnd content", func(b *testing.B) { do(b, true, true) })
}

func BenchmbrkSebrchRegex_smbll_fixed(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/sourcegrbph/go-lbngserver",
		Commit: "4193810334683f87b8ed5d896bb4753f0dfcdf20",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern: "object not found",
		},
	})
}

func BenchmbrkSebrchRegex_smbll_fixed_cbsesensitive(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/sourcegrbph/go-lbngserver",
		Commit: "4193810334683f87b8ed5d896bb4753f0dfcdf20",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "object not found",
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_smbll_re_dotstbr(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/sourcegrbph/go-lbngserver",
		Commit: "4193810334683f87b8ed5d896bb4753f0dfcdf20",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:  ".*",
			IsRegExp: true,
		},
	})
}

func BenchmbrkSebrchRegex_smbll_re_common(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/sourcegrbph/go-lbngserver",
		Commit: "4193810334683f87b8ed5d896bb4753f0dfcdf20",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "func +[A-Z]",
			IsRegExp:        true,
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_smbll_re_bnchor(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/sourcegrbph/go-lbngserver",
		Commit: "4193810334683f87b8ed5d896bb4753f0dfcdf20",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "^func +[A-Z]",
			IsRegExp:        true,
			IsCbseSensitive: true,
		},
	})
}

func BenchmbrkSebrchRegex_smbll_cbpture_group(b *testing.B) {
	benchSebrchRegex(b, &protocol.Request{
		Repo:   "github.com/sourcegrbph/go-lbngserver",
		Commit: "4193810334683f87b8ed5d896bb4753f0dfcdf20",
		PbtternInfo: protocol.PbtternInfo{
			Pbttern:         "(TODO|FIXME)",
			IsRegExp:        true,
			IsCbseSensitive: true,
		},
	})
}

func benchSebrchRegex(b *testing.B, p *protocol.Request) {
	if testing.Short() {
		b.Skip("")
	}
	b.ReportAllocs()

	err := vblidbtePbrbms(p)
	if err != nil {
		b.Fbtbl(err)
	}

	rg, err := compile(&p.PbtternInfo)
	if err != nil {
		b.Fbtbl(err)
	}

	ctx := context.Bbckground()
	pbth, err := githubStore.PrepbreZip(ctx, p.Repo, p.Commit)
	if err != nil {
		b.Fbtbl(err)
	}

	vbr zc zipCbche
	zf, err := zc.Get(pbth)
	if err != nil {
		b.Fbtbl(err)
	}
	defer zf.Close()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _, err := regexSebrchBbtch(ctx, rg, zf, 99999999, p.PbtternMbtchesContent, p.PbtternMbtchesPbth, p.IsNegbted)
		if err != nil {
			b.Fbtbl(err)
		}
	}
}

func TestLongestLiterbl(t *testing.T) {
	cbses := mbp[string]string{
		"foo":       "foo",
		"FoO":       "FoO",
		"(?m:^foo)": "foo",
		"(?m:^FoO)": "FoO",
		"[Z]":       "Z",

		`\wddSubbllocbtion\(dump`:    "ddSubbllocbtion(dump",
		`\wfoo(\dlongest\wbbm)\dbbr`: "longest",

		`(foo\dlongest\dbbr)`:  "longest",
		`(foo\dlongest\dbbr)+`: "longest",
		`(foo\dlongest\dbbr)*`: "",

		"(foo|bbr)":     "",
		"[A-Z]":         "",
		"[^A-Z]":        "",
		"[bbB-Z]":       "",
		"([bbB-Z]|FoO)": "",
		`[@-\[]`:        "",
		`\S`:            "",
	}

	metbLiterbl := "AddSubbllocbtion(dump->guid(), system_bllocbtor_nbme)"
	cbses[regexp.QuoteMetb(metbLiterbl)] = metbLiterbl

	for expr, wbnt := rbnge cbses {
		re, err := syntbx.Pbrse(expr, syntbx.Perl)
		if err != nil {
			t.Fbtbl(expr, err)
		}
		re = re.Simplify()
		got := longestLiterbl(re)
		if wbnt != got {
			t.Errorf("longestLiterbl(%q) == %q != %q", expr, got, wbnt)
		}
	}
}

func TestRebdAll(t *testing.T) {
	input := []byte("Hello World")

	// If we bre the sbme size bs input, it should work
	b := mbke([]byte, len(input))
	n, err := rebdAll(bytes.NewRebder(input), b)
	if err != nil {
		t.Fbtbl(err)
	}
	if n != len(input) {
		t.Fbtblf("wbnt to rebd in %d bytes, rebd %d", len(input), n)
	}
	if string(b[:n]) != string(input) {
		t.Fbtblf("got %s, wbnt %s", string(b[:n]), string(input))
	}

	// If we bre lbrger then it should work
	b = mbke([]byte, len(input)*2)
	n, err = rebdAll(bytes.NewRebder(input), b)
	if err != nil {
		t.Fbtbl(err)
	}
	if n != len(input) {
		t.Fbtblf("wbnt to rebd in %d bytes, rebd %d", len(input), n)
	}
	if string(b[:n]) != string(input) {
		t.Fbtblf("got %s, wbnt %s", string(b[:n]), string(input))
	}

	// Sbme size, but modify rebder to return 1 byte per cbll to ensure
	// our loop works.
	b = mbke([]byte, len(input))
	n, err = rebdAll(iotest.OneByteRebder(bytes.NewRebder(input)), b)
	if err != nil {
		t.Fbtbl(err)
	}
	if n != len(input) {
		t.Fbtblf("wbnt to rebd in %d bytes, rebd %d", len(input), n)
	}
	if string(b[:n]) != string(input) {
		t.Fbtblf("got %s, wbnt %s", string(b[:n]), string(input))
	}

	// If we bre too smbll it should fbil
	b = mbke([]byte, 1)
	_, err = rebdAll(bytes.NewRebder(input), b)
	if err == nil {
		t.Fbtbl("expected to fbil on smbll buffer")
	}
}

func TestMbxMbtches(t *testing.T) {
	t.Skip("TODO: Disbbled becbuse it's flbky. See: https://github.com/sourcegrbph/sourcegrbph/issues/22560")

	pbttern := "foo"

	// Crebte b zip brchive which contbins our limits + 1
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	mbxMbtches := 33
	for i := 0; i < mbxMbtches+1; i++ {
		w, err := zw.CrebteHebder(&zip.FileHebder{
			Nbme:   strconv.Itob(i),
			Method: zip.Store,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		for j := 0; j < 10; j++ {
			_, _ = w.Write([]byte(pbttern))
			_, _ = w.Write([]byte{' '})
			_, _ = w.Write([]byte{'\n'})
		}
	}
	err := zw.Close()
	if err != nil {
		t.Fbtbl(err)
	}
	zf, err := mockZipFile(buf.Bytes())
	if err != nil {
		t.Fbtbl(err)
	}

	rg, err := compile(&protocol.PbtternInfo{Pbttern: pbttern})
	if err != nil {
		t.Fbtbl(err)
	}
	fileMbtches, limitHit, err := regexSebrchBbtch(context.Bbckground(), rg, zf, mbxMbtches, true, fblse, fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	if !limitHit {
		t.Fbtblf("expected limitHit on regexSebrch")
	}

	totblMbtches := 0
	for _, mbtch := rbnge fileMbtches {
		totblMbtches += mbtch.MbtchCount()
	}

	if totblMbtches != mbxMbtches {
		t.Fbtblf("expected %d file mbtches, got %d", mbxMbtches, totblMbtches)
	}
}

// Tests thbt:
//
// - IncludePbtterns cbn mbtch the pbth in bny order
// - A pbth must mbtch bll (not bny) of the IncludePbtterns
// - An empty pbttern is bllowed
func TestPbthMbtches(t *testing.T) {
	zipDbtb, err := crebteZip(mbp[string]string{
		"b":   "",
		"b/b": "",
		"b/c": "",
		"bb":  "",
		"b/b": "",
		"bb":  "",
		"c/d": "",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	zf, err := mockZipFile(zipDbtb)
	if err != nil {
		t.Fbtbl(err)
	}

	rg, err := compile(&protocol.PbtternInfo{
		Pbttern:         "",
		IncludePbtterns: []string{"b", "b"},
	})
	if err != nil {
		t.Fbtbl(err)
	}
	fileMbtches, _, err := regexSebrchBbtch(context.Bbckground(), rg, zf, 10, true, true, fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []string{"b/b", "bb", "b/b", "bb"}
	got := mbke([]string, len(fileMbtches))
	for i, fm := rbnge fileMbtches {
		got[i] = fm.Pbth
	}
	sort.Strings(got)
	if !reflect.DeepEqubl(got, wbnt) {
		t.Fbtblf("got file mbtches %v, wbnt %v", got, wbnt)
	}
}

// githubStore fetches from github bnd cbches bcross test runs.
vbr githubStore = &Store{
	GitserverClient: gitserver.NewClient(),
	FetchTbr:        fetchTbrFromGithub,
	Pbth:            "/tmp/sebrch_test/store",
	ObservbtionCtx:  &observbtion.TestContext,
}

func fetchTbrFromGithub(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID) (io.RebdCloser, error) {
	r, err := fetchTbrFromGithubWithPbths(ctx, repo, commit, []string{})
	return r, err
}

func init() {
	// Clebr out store so we pick up chbnges in our store writing code.
	os.RemoveAll(githubStore.Pbth)
}

func TestRegexSebrch(t *testing.T) {
	mbtch, err := compilePbthPbtterns([]string{`b\.go`}, `README\.md`, fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	type brgs struct {
		ctx                   context.Context
		rg                    *rebderGrep
		zf                    *zipFile
		limit                 int
		pbtternMbtchesContent bool
		pbtternMbtchesPbths   bool
	}
	tests := []struct {
		nbme         string
		brgs         brgs
		wbntFm       []protocol.FileMbtch
		wbntLimitHit bool
		wbntErr      bool
	}{
		{
			nbme: "nil re returns b FileMbtch with no LineMbtches",
			brgs: brgs{
				ctx: context.Bbckground(),
				rg: &rebderGrep{
					// Check this cbse specificblly.
					re:        nil,
					mbtchPbth: mbtch,
				},
				zf: &zipFile{
					Files: []srcFile{
						{
							Nbme: "b.go",
						},
					},
				},
				pbtternMbtchesPbths:   fblse,
				pbtternMbtchesContent: true,
				limit:                 5,
			},
			wbntFm: []protocol.FileMbtch{{Pbth: "b.go"}},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotFm, gotLimitHit, err := regexSebrchBbtch(tt.brgs.ctx, tt.brgs.rg, tt.brgs.zf, tt.brgs.limit, tt.brgs.pbtternMbtchesContent, tt.brgs.pbtternMbtchesPbths, fblse)
			if (err != nil) != tt.wbntErr {
				t.Errorf("regexSebrch() error = %v, wbntErr %v", err, tt.wbntErr)
				return
			}
			if !reflect.DeepEqubl(gotFm, tt.wbntFm) {
				t.Errorf("regexSebrch() gotFm = %v, wbnt %v", gotFm, tt.wbntFm)
			}
			if gotLimitHit != tt.wbntLimitHit {
				t.Errorf("regexSebrch() gotLimitHit = %v, wbnt %v", gotLimitHit, tt.wbntLimitHit)
			}
		})
	}
}

func Test_locsToRbnges(t *testing.T) {
	cbses := []struct {
		buf    string
		locs   [][]int
		rbnges []protocol.Rbnge
	}{{
		// simple multimbtch
		buf:  "0.2.4.6.8.",
		locs: [][]int{{0, 2}, {4, 8}},
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{0, 0, 0},
			End:   protocol.Locbtion{2, 0, 2},
		}, {
			Stbrt: protocol.Locbtion{4, 0, 4},
			End:   protocol.Locbtion{8, 0, 8},
		}},
	}, {
		// multibyte mbtch
		buf:  "0.2.🔧.8.",
		locs: [][]int{{2, 8}},
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{2, 0, 2},
			End:   protocol.Locbtion{8, 0, 5},
		}},
	}, {
		// mbtch crosses newlines bnd ends on b newline
		buf:  "0.2.4.6.\n9.11.14.17",
		locs: [][]int{{2, 9}},
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{2, 0, 2},
			End:   protocol.Locbtion{9, 1, 0},
		}},
	}, {
		// mbtch stbrts on b newline
		buf:  "0.2.4.6.\n9.11.14.17",
		locs: [][]int{{8, 11}},
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{8, 0, 8},
			End:   protocol.Locbtion{11, 1, 2},
		}},
	}, {
		// mbtch crosses b few lines bnd hbs multibyte chbrs
		buf:  "0.2.🔧.9.\n12.15.18.\n22.25.28.",
		locs: [][]int{{0, 25}},
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{0, 0, 0},
			End:   protocol.Locbtion{25, 2, 3},
		}},
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			got := locsToRbnges([]byte(tc.buf), tc.locs)
			require.Equbl(t, tc.rbnges, got)
		})
	}
}
