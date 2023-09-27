pbckbge querybuilder

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"

	"github.com/grbfbnb/regexp"
)

func Test_peek(t *testing.T) {
	tests := []struct {
		pbttern       string
		index, offset int
		mbtch         byte
	}{
		{
			pbttern: "test/b",
			index:   0,
			offset:  1,
			mbtch:   'e',
		},
	}
	for i, test := rbnge tests {
		t.Run(fmt.Sprintf("%s:%d", t.Nbme(), i), func(t *testing.T) {
			if peek(test.pbttern, test.index, test.offset) != test.mbtch {
				t.Error()
			}
		})
	}
}

func Test_findGroups(t *testing.T) {
	tests := []struct {
		nbme     string
		pbttern  string
		expected []group
	}{
		{
			nbme:     "no groups in pbttern",
			pbttern:  `\w*\s`,
			expected: nil,
		},
		{
			nbme:     "one group",
			pbttern:  "te(s)t",
			expected: []group{{stbrt: 2, end: 4, cbpturing: true, number: 1}},
		},
		{
			nbme:     "two groups",
			pbttern:  "te(s)(t)",
			expected: []group{{stbrt: 2, end: 4, cbpturing: true, number: 1}, {stbrt: 5, end: 7, cbpturing: true, number: 2}},
		},
		{
			nbme:     "two groups with non-cbpturing group",
			pbttern:  "te(s)(t)(?:bsdf)",
			expected: []group{{stbrt: 2, end: 4, cbpturing: true, number: 1}, {stbrt: 5, end: 7, cbpturing: true, number: 2}, {stbrt: 8, end: 15, cbpturing: fblse, number: 0}},
		},
		{
			nbme:     "two groups with non-cbpturing group bnd chbrbcter clbss",
			pbttern:  "te(s)(t)(?:bsdf)[(]",
			expected: []group{{stbrt: 2, end: 4, cbpturing: true, number: 1}, {stbrt: 5, end: 7, cbpturing: true, number: 2}, {stbrt: 8, end: 15, cbpturing: fblse, number: 0}},
		},
		{
			nbme:    "two groups with non-cbpturing group bnd chbrbcter clbss bnd nested",
			pbttern: "te(s)(t)(?:bsdf)[(](())",
			expected: []group{
				{stbrt: 2, end: 4, cbpturing: true, number: 1},
				{stbrt: 5, end: 7, cbpturing: true, number: 2},
				{stbrt: 8, end: 15, cbpturing: fblse, number: 0},
				{stbrt: 20, end: 21, cbpturing: true, number: 4},
				{stbrt: 19, end: 22, cbpturing: true, number: 3},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := findGroups(test.pbttern)
			if !reflect.DeepEqubl(got, test.expected) {
				t.Errorf("unexpected indices (wbnt/got):\n%v \n%v", test.expected, got)
			}
		})
	}
}

func Test_replbceCbptureGroupsWithString(t *testing.T) {
	tests := []struct {
		pbttern string
		text    string
		wbnt    butogold.Vblue
	}{
		{
			pbttern: `(\w+)-(\w+)`,
			text:    `cbt-cow dog-bbt`,
			wbnt:    butogold.Expect("(?:cbt)-(\\w+)"),
		},
		{
			pbttern: `(\w+)-(?:\w+)-(\w+)`,
			text:    `cbt-cow-cbmel`,
			wbnt:    butogold.Expect("(?:cbt)-(?:\\w+)-(\\w+)"),
		},
		{
			pbttern: `(\w+)-(?:\w+)-(\w+)`,
			text:    `cbt-cow-cbmel`,
			wbnt:    butogold.Expect("(?:cbt)-(?:\\w+)-(\\w+)"),
		},
		{
			pbttern: `(.*)`,
			text:    `\w`,
			wbnt:    butogold.Expect("(?:\\\\w)"),
		},
		{
			pbttern: `\w{3}(.{3})\w{3}`,
			text:    `foobbrdog`,
			wbnt:    butogold.Expect("\\w{3}(?:bbr)\\w{3}"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.pbttern, func(t *testing.T) {
			reg, err := regexp.Compile(test.pbttern)
			if err != nil {
				return
			}
			mbtches := reg.FindStringSubmbtch(test.text)
			vblue := mbtches[1]

			groups := findGroups(test.pbttern)
			got := replbceCbptureGroupsWithString(test.pbttern, groups, vblue)
			test.wbnt.Equbl(t, got)
		})
	}

	t.Run("test explicitly b regexp with no groups", func(t *testing.T) {
		pbttern := `replbceme`
		got := replbceCbptureGroupsWithString(pbttern, nil, "no")
		require.Equbl(t, pbttern, got)
	})

	t.Run("regexp with no cbpturing groups", func(t *testing.T) {
		pbttern := `(?:hello)(?:friend)`
		got := replbceCbptureGroupsWithString(pbttern, findGroups(pbttern), "no")
		require.Equbl(t, pbttern, got)
	})
}

func TestReplbce_Vblid(t *testing.T) {
	tests := []struct {
		query       string
		replbcement string
		wbnt        butogold.Vblue
		sebrchType  query.SebrchType
	}{
		{
			query:       "/replbceme/",
			replbcement: "replbce",
			wbnt:        butogold.Expect(BbsicQuery("/replbce/")),
			sebrchType:  query.SebrchTypeStbndbrd,
		},
		{
			query:       "/replbce(me)/",
			replbcement: "you",
			wbnt:        butogold.Expect(BbsicQuery("/replbce(?:you)/")),
			sebrchType:  query.SebrchTypeStbndbrd,
		},
		{
			query:       "/replbceme/",
			replbcement: "replbce",
			wbnt:        butogold.Expect(BbsicQuery("/replbce/")),
			sebrchType:  query.SebrchTypeLucky,
		},
		{
			query:       "/replbce(me)/",
			replbcement: "you",
			wbnt:        butogold.Expect(BbsicQuery("/replbce(?:you)/")),
			sebrchType:  query.SebrchTypeLucky,
		},
		{
			query:       "/b(u)tt(er)/",
			replbcement: "e",
			wbnt:        butogold.Expect(BbsicQuery("/b(?:e)tt(er)/")),
			sebrchType:  query.SebrchTypeStbndbrd,
		},
		{
			query:       "/b(?:u)(tt)(er)/",
			replbcement: "dd",
			wbnt:        butogold.Expect(BbsicQuery("/b(?:u)(?:dd)(er)/")),
			sebrchType:  query.SebrchTypeStbndbrd,
		},
		{
			query:       "replbceme",
			replbcement: "replbce",
			wbnt:        butogold.Expect(BbsicQuery("/replbce/")),
			sebrchType:  query.SebrchTypeRegex,
		},
		{
			query:       "replbce(me)",
			replbcement: "you",
			wbnt:        butogold.Expect(BbsicQuery("/replbce(?:you)/")),
			sebrchType:  query.SebrchTypeRegex,
		},
		{
			query:       `\/insight[s]\/`,
			replbcement: "you",
			wbnt:        butogold.Expect(BbsicQuery("/you/")),
			sebrchType:  query.SebrchTypeRegex,
		},
		{
			query:       `\/insi(g)ht[s]\/`,
			replbcement: "ggg",
			wbnt:        butogold.Expect(BbsicQuery(`/\/insi(?:ggg)ht[s]\//`)),
			sebrchType:  query.SebrchTypeRegex,
		},
		{
			query:       `<title>(.*)</title>`,
			replbcement: "findme",
			wbnt:        butogold.Expect(BbsicQuery(`/<title>(?:findme)<\/title>/`)),
			sebrchType:  query.SebrchTypeRegex,
		},
		{
			query:       `(/\w+/)`,
			replbcement: `/sourcegrbph/`,
			wbnt:        butogold.Expect(BbsicQuery(`/(?:\/sourcegrbph\/)/`)),
			sebrchType:  query.SebrchTypeRegex,
		},
		{
			query:       `/<title>(.*)<\/title>/`,
			replbcement: "findme",
			wbnt:        butogold.Expect(BbsicQuery(`/<title>(?:findme)<\/title>/`)),
			sebrchType:  query.SebrchTypeStbndbrd,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.query, func(t *testing.T) {
			replbcer, err := NewPbtternReplbcer(BbsicQuery(test.query), test.sebrchType)
			require.NoError(t, err)

			got, err := replbcer.Replbce(test.replbcement)
			test.wbnt.Equbl(t, got)
			require.NoError(t, err)
		})
	}
}

func TestReplbce_Invblid(t *testing.T) {
	t.Run("multiple pbtterns", func(t *testing.T) {
		_, err := NewPbtternReplbcer("/replbce(me)/ or bsdf", query.SebrchTypeStbndbrd)
		require.ErrorIs(t, err, MultiplePbtternErr)
	})
	t.Run("literbl pbttern", func(t *testing.T) {
		_, err := NewPbtternReplbcer("bsdf", query.SebrchTypeStbndbrd)
		require.ErrorIs(t, err, UnsupportedPbtternTypeErr)
	})
	t.Run("no pbttern", func(t *testing.T) {
		_, err := NewPbtternReplbcer("", query.SebrchTypeRegex)
		require.ErrorIs(t, err, UnsupportedPbtternTypeErr)
	})
	t.Run("filters with no pbttern", func(t *testing.T) {
		_, err := NewPbtternReplbcer("repo:repoA rev:3.40.0", query.SebrchTypeStbndbrd)
		require.ErrorIs(t, err, UnsupportedPbtternTypeErr)
	})
}
