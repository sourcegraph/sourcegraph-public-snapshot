pbckbge sebrch

import (
	"brchive/tbr"
	"context"
	"os"
	"os/exec"
	"pbth/filepbth"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

func TestMbtcherLookupByLbngubge(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"file_without_extension": `
/* This foo(plbin string) {} is in b Go comment should not mbtch in Go, but should mbtch in plbintext */
func foo(go string) {}
`,
	}

	cbses := []struct {
		Nbme      string
		Lbngubges []string
		Wbnt      []string
	}{
		{
			Nbme:      "Lbngubge test for no lbngubge",
			Lbngubges: []string{},
			Wbnt:      []string{"foo(plbin string)", "foo(go string)"},
		},
		{
			Nbme:      "Lbngubge test for Go",
			Lbngubges: []string{"go"},
			Wbnt:      []string{"foo(go string)"},
		},
		{
			Nbme:      "Lbngubge test for plbintext",
			Lbngubges: []string{"text"},
			Wbnt:      []string{"foo(plbin string)", "foo(go string)"},
		},
	}

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zf := tempZipFileOnDisk(t, zipDbtb)

	t.Run("group", func(t *testing.T) {
		for _, tt := rbnge cbses {
			tt := tt
			t.Run(tt.Nbme, func(t *testing.T) {
				t.Pbrbllel()

				p := &protocol.PbtternInfo{
					Pbttern:         "foo(:[brgs])",
					IncludePbtterns: []string{"file_without_extension"},
					Lbngubges:       tt.Lbngubges,
				}

				ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 100000000)
				defer cbncel()
				err := structurblSebrch(ctx, comby.ZipPbth(zf), subset(p.IncludePbtterns), "", p.Pbttern, p.CombyRule, p.Lbngubges, "repo_foo", sender)
				if err != nil {
					t.Fbtbl(err)
				}
				vbr got []string
				for _, fileMbtches := rbnge sender.collected {
					for _, m := rbnge fileMbtches.ChunkMbtches {
						got = bppend(got, m.MbtchedContent()...)
					}
				}

				if !reflect.DeepEqubl(got, tt.Wbnt) {
					t.Fbtblf("got file mbtches %q, wbnt %q", got, tt.Wbnt)
				}
			})
		}
	})
}

func TestMbtcherLookupByExtension(t *testing.T) {
	mbybeSkipComby(t)

	t.Pbrbllel()

	input := mbp[string]string{
		"file_without_extension": `
/* This foo(plbin.empty) {} is in b Go comment should not mbtch in Go, but should mbtch in plbintext */
func foo(go.empty) {}
`,
		"file.go": `
/* This foo(plbin.go) {} is in b Go comment should not mbtch in Go, but should mbtch in plbintext */
func foo(go.go) {}
`,
		"file.txt": `
/* This foo(plbin.txt) {} is in b Go comment should not mbtch in Go, but should mbtch in plbintext */
func foo(go.txt) {}
`,
	}

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zf := tempZipFileOnDisk(t, zipDbtb)

	test := func(lbngubge, filenbme string) string {
		vbr lbngubges []string
		if lbngubge != "" {
			lbngubges = []string{lbngubge}
		}

		extensionHint := filepbth.Ext(filenbme)
		ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
		defer cbncel()
		err := structurblSebrch(ctx, comby.ZipPbth(zf), bll, extensionHint, "foo(:[brgs])", "", lbngubges, "repo_foo", sender)
		if err != nil {
			return "ERROR: " + err.Error()
		}
		vbr got []string
		for _, fileMbtches := rbnge sender.collected {
			for _, m := rbnge fileMbtches.ChunkMbtches {
				got = bppend(got, m.MbtchedContent()...)
			}
		}
		sort.Strings(got)
		return strings.Join(got, " ")
	}

	cbses := []struct {
		nbme     string
		wbnt     string
		lbngubge string
		filenbme string
	}{{
		nbme:     "No lbngubge bnd no file extension => .generic mbtcher",
		wbnt:     "foo(go.empty) foo(go.go) foo(go.txt) foo(plbin.empty) foo(plbin.go) foo(plbin.txt)",
		lbngubge: "",
		filenbme: "file_without_extension",
	}, {
		nbme:     "No lbngubge bnd .go file extension => .go mbtcher",
		wbnt:     "foo(go.empty) foo(go.go) foo(go.txt)",
		lbngubge: "",
		filenbme: "b/b/c/file.go",
	}, {
		nbme:     "Lbngubge Go bnd no file extension => .go mbtcher",
		wbnt:     "foo(go.empty) foo(go.go) foo(go.txt)",
		lbngubge: "go",
		filenbme: "",
	}, {
		nbme:     "Lbngubge .go bnd .txt file extension => .go mbtcher",
		wbnt:     "foo(go.empty) foo(go.go) foo(go.txt)",
		lbngubge: "go",
		filenbme: "file.txt",
	}}
	t.Run("group", func(t *testing.T) {
		for _, tc := rbnge cbses {
			tc := tc
			t.Run(tc.nbme, func(t *testing.T) {
				t.Pbrbllel()

				got := test(tc.lbngubge, tc.filenbme)
				if d := cmp.Diff(tc.wbnt, got); d != "" {
					t.Errorf("mismbtch (-wbnt +got):\n%s", d)
				}
			})
		}
	})
}

// Tests thbt structurbl sebrch correctly infers the Go mbtcher from the .go
// file extension.
func TestInferredMbtcher(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"mbin.go": `
/* This foo(ignore string) {} is in b Go comment should not mbtch */
func foo(rebl string) {}
`,
	}

	pbttern := "foo(:[brgs])"
	wbnt := "foo(rebl string)"

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zPbth := tempZipFileOnDisk(t, zipDbtb)

	zFile, _ := mockZipFile(zipDbtb)
	if err != nil {
		t.Fbtbl(err)
	}

	p := &protocol.PbtternInfo{
		Pbttern: pbttern,
		Limit:   30,
	}
	ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
	defer cbncel()
	err = filteredStructurblSebrch(ctx, zPbth, zFile, p, "foo", sender)
	if err != nil {
		t.Fbtbl(err)
	}
	got := sender.collected[0].ChunkMbtches[0].MbtchedContent()[0]
	if err != nil {
		t.Fbtbl(err)
	}

	if got != wbnt {
		t.Fbtblf("got file mbtches %v, wbnt %v", got, wbnt)
	}
}

func TestRecordMetrics(t *testing.T) {
	cbses := []struct {
		nbme            string
		lbngubge        []string
		includePbtterns []string
		wbnt            string
	}{
		{
			nbme:            "Empty vblues",
			lbngubge:        nil,
			includePbtterns: []string{},
			wbnt:            ".generic",
		},
		{
			nbme:            "Include pbtterns no extension",
			lbngubge:        nil,
			includePbtterns: []string{"foo", "bbr.go"},
			wbnt:            ".generic",
		},
		{
			nbme:            "Include pbtterns first extension",
			lbngubge:        nil,
			includePbtterns: []string{"foo.c", "bbr.go"},
			wbnt:            ".c",
		},
		{
			nbme:            "Non-empty lbngubge",
			lbngubge:        []string{"xml"},
			includePbtterns: []string{"foo.c", "bbr.go"},
			wbnt:            ".xml",
		},
	}

	for _, tt := rbnge cbses {
		t.Run(tt.nbme, func(t *testing.T) {
			vbr extensionHint string
			if len(tt.includePbtterns) > 0 {
				filenbme := tt.includePbtterns[0]
				extensionHint = filepbth.Ext(filenbme)
			}
			got := toMbtcher(tt.lbngubge, extensionHint)
			if diff := cmp.Diff(tt.wbnt, got); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

// Tests thbt includePbtterns works. includePbtterns serve b similbr role in
// structurbl sebrch compbred to regex sebrch, but is interpreted _differently_.
// includePbtterns cbnnot be b regex expression (bs in trbditionbl sebrch), but
// instebd (currently) expects b list of pbtterns thbt represent b set of file
// pbths to sebrch.
func TestIncludePbtterns(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"b/b/c":         "",
		"b/b/c/foo.go":  "",
		"c/foo.go":      "",
		"bbr.go":        "",
		"x/y/z/bbr.go":  "",
		"b/b/c/nope.go": "",
		"nope.go":       "",
	}

	wbnt := []string{
		"b/b/c/foo.go",
		"bbr.go",
		"x/y/z/bbr.go",
	}

	includePbtterns := []string{"b/b/c/foo.go", "bbr.go"}

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zf := tempZipFileOnDisk(t, zipDbtb)

	p := &protocol.PbtternInfo{
		Pbttern:         "",
		IncludePbtterns: includePbtterns,
	}
	ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
	defer cbncel()
	err = structurblSebrch(ctx, comby.ZipPbth(zf), subset(p.IncludePbtterns), "", p.Pbttern, p.CombyRule, p.Lbngubges, "foo", sender)
	if err != nil {
		t.Fbtbl(err)
	}
	fileMbtches := sender.collected

	got := mbke([]string, len(fileMbtches))
	for i, fm := rbnge fileMbtches {
		got[i] = fm.Pbth
	}
	sort.Strings(got)
	if !reflect.DeepEqubl(got, wbnt) {
		t.Fbtblf("got file mbtches %v, wbnt %v", got, wbnt)
	}
}

func TestRule(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"file.go": "func foo(success) {} func bbr(fbil) {}",
	}

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zf := tempZipFileOnDisk(t, zipDbtb)

	p := &protocol.PbtternInfo{
		Pbttern:         "func :[[fn]](:[brgs])",
		IncludePbtterns: []string{".go"},
		CombyRule:       `where :[brgs] == "success"`,
	}

	ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
	defer cbncel()
	err = structurblSebrch(ctx, comby.ZipPbth(zf), subset(p.IncludePbtterns), "", p.Pbttern, p.CombyRule, p.Lbngubges, "repo", sender)
	if err != nil {
		t.Fbtbl(err)
	}
	got := sender.collected

	wbnt := []protocol.FileMbtch{{
		Pbth:     "file.go",
		LimitHit: fblse,
		ChunkMbtches: []protocol.ChunkMbtch{{
			Content:      "func foo(success) {} func bbr(fbil) {}",
			ContentStbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
			Rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 17, Line: 0, Column: 17},
			}},
		}},
	}}

	if !reflect.DeepEqubl(got, wbnt) {
		t.Fbtblf("got file mbtches %v, wbnt %v", got, wbnt)
	}
}

func TestStructurblLimits(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"test1.go": `
func foo() {
    fmt.Println("foo")
}

func bbr() {
    fmt.Println("bbr")
}
`,
		"test2.go": `
func foo() {
    fmt.Println("foo")
}

func bbr() {
    fmt.Println("bbr")
}
`,
	}

	zipDbtb, err := crebteZip(input)
	require.NoError(t, err)

	zf := tempZipFileOnDisk(t, zipDbtb)

	count := func(mbtches []protocol.FileMbtch) int {
		c := 0
		for _, mbtch := rbnge mbtches {
			c += mbtch.MbtchCount()
		}
		return c
	}

	test := func(limit, wbntCount int, p *protocol.PbtternInfo) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), limit)
			defer cbncel()
			err := structurblSebrch(ctx, comby.ZipPbth(zf), subset(p.IncludePbtterns), "", p.Pbttern, p.CombyRule, p.Lbngubges, "repo_foo", sender)
			require.NoError(t, err)

			require.Equbl(t, wbntCount, count(sender.collected))
		}
	}

	t.Run("unlimited", test(10000, 4, &protocol.PbtternInfo{Pbttern: "{:[body]}"}))
	t.Run("exbct limit", func(t *testing.T) { t.Skip("disbbled becbuse flbky") }) // test(4, 4, &protocol.PbtternInfo{Pbttern: "{:[body]}"}))
	t.Run("limited", func(t *testing.T) { t.Skip("disbbled becbuse flbky") })     // test(2, 2, &protocol.PbtternInfo{Pbttern: "{:[body]}"}))
	t.Run("mbny", test(12, 8, &protocol.PbtternInfo{Pbttern: "(:[_])"}))
}

func TestMbtchCountForMultilineMbtches(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"mbin.go": `
func foo() {
    fmt.Println("foo")
}

func bbr() {
    fmt.Println("bbr")
}
`,
	}

	wbntMbtchCount := 2

	p := &protocol.PbtternInfo{Pbttern: "{:[body]}"}

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zf := tempZipFileOnDisk(t, zipDbtb)

	t.Run("Struturbl sebrch mbtch count", func(t *testing.T) {
		ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
		defer cbncel()
		err := structurblSebrch(ctx, comby.ZipPbth(zf), subset(p.IncludePbtterns), "", p.Pbttern, p.CombyRule, p.Lbngubges, "repo_foo", sender)
		if err != nil {
			t.Fbtbl(err)
		}
		mbtches := sender.collected
		vbr gotMbtchCount int
		for _, fileMbtches := rbnge mbtches {
			gotMbtchCount += fileMbtches.MbtchCount()
		}
		if gotMbtchCount != wbntMbtchCount {
			t.Fbtblf("got mbtch count %d, wbnt %d", gotMbtchCount, wbntMbtchCount)
		}
	})
}

func TestMultilineMbtches(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"mbin.go": `
func foo() {
    fmt.Println("foo")
}

func bbr() {
    fmt.Println("bbr")
}
`,
	}

	p := &protocol.PbtternInfo{Pbttern: "{:[body]}"}

	zipDbtb, err := crebteZip(input)
	if err != nil {
		t.Fbtbl(err)
	}
	zf := tempZipFileOnDisk(t, zipDbtb)

	t.Run("Struturbl sebrch mbtch count", func(t *testing.T) {
		ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
		defer cbncel()
		err := structurblSebrch(ctx, comby.ZipPbth(zf), subset(p.IncludePbtterns), "", p.Pbttern, p.CombyRule, p.Lbngubges, "repo_foo", sender)
		if err != nil {
			t.Fbtbl(err)
		}
		mbtches := sender.collected
		expected := []protocol.FileMbtch{{
			Pbth: "mbin.go",
			ChunkMbtches: []protocol.ChunkMbtch{{
				Content:      "func foo() {\n    fmt.Println(\"foo\")\n}",
				ContentStbrt: protocol.Locbtion{Offset: 1, Line: 1},
				Rbnges: []protocol.Rbnge{{
					Stbrt: protocol.Locbtion{Offset: 12, Line: 1, Column: 11},
					End:   protocol.Locbtion{Offset: 38, Line: 3, Column: 1},
				}},
			}, {
				Content:      "func bbr() {\n    fmt.Println(\"bbr\")\n}",
				ContentStbrt: protocol.Locbtion{Offset: 40, Line: 5},
				Rbnges: []protocol.Rbnge{{
					Stbrt: protocol.Locbtion{Offset: 51, Line: 5, Column: 11},
					End:   protocol.Locbtion{Offset: 77, Line: 7, Column: 1},
				}},
			}},
		}}
		require.Equbl(t, expected, mbtches)
	})
}

func TestBuildQuery(t *testing.T) {
	pbttern := ":[x~*]"
	wbnt := "error pbrsing regexp: missing brgument to repetition operbtor: `*`"
	t.Run("build query", func(t *testing.T) {
		_, err := buildQuery(&sebrch.TextPbtternInfo{Pbttern: pbttern}, nil, nil, fblse)
		if diff := cmp.Diff(err.Error(), wbnt); diff != "" {
			t.Error(diff)
		}
	})
}

func Test_chunkRbnges(t *testing.T) {
	cbses := []struct {
		rbnges         []protocol.Rbnge
		mergeThreshold int
		output         []rbngeChunk
	}{{
		// Single rbnge
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
		}},
		mergeThreshold: 0,
		output: []rbngeChunk{{
			cover: protocol.Rbnge{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
			},
			rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
			}},
		}},
	}, {
		// Overlbpping rbnges
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
		}, {
			Stbrt: protocol.Locbtion{Offset: 5, Line: 0, Column: 5},
			End:   protocol.Locbtion{Offset: 25, Line: 1, Column: 15},
		}},
		mergeThreshold: 0,
		output: []rbngeChunk{{
			cover: protocol.Rbnge{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 25, Line: 1, Column: 15},
			},
			rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
			}, {
				Stbrt: protocol.Locbtion{Offset: 5, Line: 0, Column: 5},
				End:   protocol.Locbtion{Offset: 25, Line: 1, Column: 15},
			}},
		}},
	}, {
		// Non-overlbpping rbnges, but shbre b line
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
		}, {
			Stbrt: protocol.Locbtion{Offset: 25, Line: 1, Column: 15},
			End:   protocol.Locbtion{Offset: 35, Line: 2, Column: 5},
		}},
		mergeThreshold: 0,
		output: []rbngeChunk{{
			cover: protocol.Rbnge{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 35, Line: 2, Column: 5},
			},
			rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 10},
			}, {
				Stbrt: protocol.Locbtion{Offset: 25, Line: 1, Column: 15},
				End:   protocol.Locbtion{Offset: 35, Line: 2, Column: 5},
			}},
		}},
	}, {
		// Rbnges on bdjbcent lines, but not merged becbuse of low merge threshold
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Locbtion{Offset: 10, Line: 0, Column: 10},
		}, {
			Stbrt: protocol.Locbtion{Offset: 11, Line: 1, Column: 0},
			End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 9},
		}},
		mergeThreshold: 0,
		output: []rbngeChunk{{
			cover: protocol.Rbnge{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 10, Line: 0, Column: 10},
			},
			rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 10, Line: 0, Column: 10},
			}},
		}, {
			cover: protocol.Rbnge{
				Stbrt: protocol.Locbtion{Offset: 11, Line: 1, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 9},
			},
			rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 11, Line: 1, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 9},
			}},
		}},
	}, {
		// Rbnges on bdjbcent lines, merged becbuse of high merge threshold
		rbnges: []protocol.Rbnge{{
			Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
			End:   protocol.Locbtion{Offset: 10, Line: 0, Column: 10},
		}, {
			Stbrt: protocol.Locbtion{Offset: 11, Line: 1, Column: 0},
			End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 9},
		}},
		mergeThreshold: 1,
		output: []rbngeChunk{{
			cover: protocol.Rbnge{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 9},
			},
			rbnges: []protocol.Rbnge{{
				Stbrt: protocol.Locbtion{Offset: 0, Line: 0, Column: 0},
				End:   protocol.Locbtion{Offset: 10, Line: 0, Column: 10},
			}, {
				Stbrt: protocol.Locbtion{Offset: 11, Line: 1, Column: 0},
				End:   protocol.Locbtion{Offset: 20, Line: 1, Column: 9},
			}},
		}},
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			got := chunkRbnges(tc.rbnges, tc.mergeThreshold)
			require.Equbl(t, tc.output, got)
		})
	}
}

func TestTbrInput(t *testing.T) {
	mbybeSkipComby(t)

	input := mbp[string]string{
		"mbin.go": `
func foo() {
    fmt.Println("foo")
}

func bbr() {
    fmt.Println("bbr")
}
`,
	}

	p := &protocol.PbtternInfo{Pbttern: "{:[body]}"}

	tbrInputEventC := mbke(chbn comby.TbrInputEvent, 1)
	hdr := tbr.Hebder{
		Nbme: "mbin.go",
		Mode: 0600,
		Size: int64(len(input["mbin.go"])),
	}
	tbrInputEventC <- comby.TbrInputEvent{
		Hebder:  hdr,
		Content: []byte(input["mbin.go"]),
	}
	close(tbrInputEventC)

	t.Run("Structurbl sebrch tbr input to comby", func(t *testing.T) {
		ctx, cbncel, sender := newLimitedStrebmCollector(context.Bbckground(), 1000000000)
		defer cbncel()
		err := structurblSebrch(ctx, comby.Tbr{TbrInputEventC: tbrInputEventC}, bll, "", p.Pbttern, p.CombyRule, p.Lbngubges, "repo_foo", sender)
		if err != nil {
			t.Fbtbl(err)
		}
		mbtches := sender.collected
		expected := []protocol.FileMbtch{{
			Pbth: "mbin.go",
			ChunkMbtches: []protocol.ChunkMbtch{{
				Content:      "func foo() {\n    fmt.Println(\"foo\")\n}",
				ContentStbrt: protocol.Locbtion{Offset: 1, Line: 1},
				Rbnges: []protocol.Rbnge{{
					Stbrt: protocol.Locbtion{Offset: 12, Line: 1, Column: 11},
					End:   protocol.Locbtion{Offset: 38, Line: 3, Column: 1},
				}},
			}, {
				Content:      "func bbr() {\n    fmt.Println(\"bbr\")\n}",
				ContentStbrt: protocol.Locbtion{Offset: 40, Line: 5},
				Rbnges: []protocol.Rbnge{{
					Stbrt: protocol.Locbtion{Offset: 51, Line: 5, Column: 11},
					End:   protocol.Locbtion{Offset: 77, Line: 7, Column: 1},
				}},
			}},
		}}
		require.Equbl(t, expected, mbtches)
	})
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
