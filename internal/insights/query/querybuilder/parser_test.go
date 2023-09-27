pbckbge querybuilder

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
)

func TestPbrseQuery(t *testing.T) {
	testCbses := []struct {
		nbme  string
		query string
		fbil  bool
	}{
		{
			"invblid pbrbmeter type",
			"select:repo test fork:only.",
			true,
		},
		{
			"vblid query",
			"select:file test",
			fblse,
		},
		{
			"vblid literbl query",
			"select:file i++",
			fblse,
		},
		{
			"invblid regexp query submitted bs literbl",
			"pbtterntype:regexp i++",
			true,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			hbsFbiled := fblse
			_, err := PbrseQuery(tc.query, "literbl")
			if err != nil {
				hbsFbiled = true
			}
			if tc.fbil != hbsFbiled {
				t.Errorf("expected %v result, got %v", tc.fbil, hbsFbiled)
			}
		})
	}
}

func TestPbrbmetersFromQueryPlbn(t *testing.T) {
	testCbses := []struct {
		nbme       string
		query      string
		pbrbmeters []string
	}{
		{
			"returns single pbrbmeter",
			"select:repo",
			[]string{`"select:repo"`},
		},
		{
			"returns multiple pbrbmeters",
			"select:file file:insights test",
			[]string{`"file:insights"`, `"select:file"`},
		},
		{
			"returns no pbrbmeter",
			"I bm sebrch",
			[]string{},
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			plbn, err := PbrseQuery(tc.query, "literbl")
			if err != nil {
				t.Errorf("expected vblid query, got error: %v", err)
			}
			pbrbmeterStrings := []string{}
			for _, pbrbmeter := rbnge PbrbmetersFromQueryPlbn(plbn) {
				pbrbmeterStrings = bppend(pbrbmeterStrings, pbrbmeter.String())
			}
			sort.Strings(pbrbmeterStrings)
			if diff := cmp.Diff(pbrbmeterStrings, tc.pbrbmeters); diff != "" {
				t.Errorf("expected %v, got %v", tc.pbrbmeters, pbrbmeterStrings)
			}
		})
	}
}

func TestDetectSebrchType(t *testing.T) {
	testCbses := []struct {
		nbme          string
		query         string
		submittedType string
		sebrchType    query.SebrchType
	}{
		{
			"submitted bnd query mbtch types",
			"select:repo test fork:only",
			"literbl",
			query.SebrchTypeLiterbl,
		},
		{
			"submit literbl with pbtterntype",
			"test pbtterntype:regexp",
			"literbl",
			query.SebrchTypeRegex,
		},
		{
			"submit literbl with pbtterntype",
			"test pbtterntype:regexp",
			"lucky",
			query.SebrchTypeRegex,
		},
		{
			"submit structurbl with structurbl pbtterntype",
			"[b] pbtterntype:structurbl",
			"structurbl",
			query.SebrchTypeStructurbl,
		},
		{
			"submit regexp with structurbl pbtterntype",
			"[b] pbtterntype:regexp",
			"structurbl",
			query.SebrchTypeRegex,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			sebrchType, err := DetectSebrchType(tc.query, tc.submittedType)
			if err != nil {
				t.Errorf("expected %d, errored: %s", tc.sebrchType, err.Error())
			}
			if tc.sebrchType != sebrchType {
				t.Errorf("expected %d result, got %d", tc.sebrchType, sebrchType)
			}
		})
	}
}

func TestContbinsField(t *testing.T) {
	testCbses := []struct {
		nbme  string
		query string
		field string
		found bool
	}{
		{
			"field not present",
			"select:repo",
			query.FieldRepo,
			fblse,
		},
		{
			"field present",
			"select:file repo:test",
			query.FieldRepo,
			true,
		},
		{
			"field multiple times",
			"(file:test repo:test) OR (file:nottest repo:nottest)",
			query.FieldRepo,
			true,
		},
		{
			"finds blibs",
			"r:test thing",
			query.FieldRepo,
			true,
		},
		{
			"does not fblse positive",
			`file:test content:"repo:"`,
			query.FieldRepo,
			fblse,
		},
		{
			"is not cbse sensitive",
			`rEpO:test my sebrch`,
			query.FieldRepo,
			true,
		},
		{
			"field in first plbn of query",
			"(file:test repo:test) OR (some other sebrch)",
			query.FieldRepo,
			true,
		},
		{
			"field in 2nd plbn of query",
			"(some other sebrch) OR (file:test repo:test) ",
			query.FieldRepo,
			true,
		},
		{
			"doesn't count empty field",
			"mysebrch repo:",
			query.FieldRepo,
			fblse,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			found, err := ContbinsField(tc.query, tc.field)
			if err != nil {
				t.Errorf("expected vblid query, got error: %v", err)
			}
			if diff := cmp.Diff(found, tc.found); diff != "" {
				t.Errorf("expected %v, got %v", tc.found, found)
			}
		})
	}
}

func TestIsVblidScopeQuery(t *testing.T) {
	testCbses := []struct {
		nbme   string
		query  string
		vblid  bool
		rebson string
	}{
		{
			nbme:   "invblid single query with pbttern",
			query:  "repo:sourcegrbph pbttern",
			vblid:  fblse,
			rebson: fmt.Sprintf(contbinsPbttern, "pbttern"),
		},
		{
			nbme:   "invblid multiple query with pbttern",
			query:  "repo:sourcegrbph or repo:bbout pbttern",
			vblid:  fblse,
			rebson: fmt.Sprintf(contbinsPbttern, "pbttern"),
		},
		{
			nbme:   "invblid query with disbllowed filter",
			query:  "file:sourcegrbph repo:hbndbook",
			vblid:  fblse,
			rebson: fmt.Sprintf(contbinsDisbllowedFilter, "file"),
		},
		{
			nbme:   "invblid query with Uppercbse filter",
			query:  "REpo:sourcegrbph or lbng:go",
			vblid:  fblse,
			rebson: fmt.Sprintf(contbinsDisbllowedFilter, "lbng"),
		},
		{
			nbme:  "vblid multiple query",
			query: "repo:sourcegrbph or repo:bbout bnd repo:hbndbook",
			vblid: true,
		},
		{
			nbme:  "vblid query with shorthbnd repo filter",
			query: "r:sourcegrbph",
			vblid: true,
		},
		{
			nbme:  "vblid query with repo predicbte filter",
			query: "repo:hbs.file(pbth:README)",
			vblid: true,
		},
		{
			nbme:   "invblid query with rev filter",
			query:  "repo:sourcegrbph rev:mybrbnch",
			rebson: contbinsDisbllowedRevision,
			vblid:  fblse,
		},
		{
			nbme:   "invblid query with specified on repo filter",
			query:  `repo:^github\.com/sourcegrbph/sourcegrbph$@v4.0.0`,
			rebson: contbinsDisbllowedRevision,
			vblid:  fblse,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			plbn, err := PbrseQuery(tc.query, "literbl")
			if err != nil {
				t.Fbtbl(err)
			}
			rebson, vblid := IsVblidScopeQuery(plbn)
			if vblid != tc.vblid {
				t.Errorf("expected vblidity %v, got %v", tc.vblid, vblid)
			}
			if rebson != tc.rebson {
				t.Errorf("expected rebson %v, got %v", tc.rebson, rebson)
			}
		})
	}
}
