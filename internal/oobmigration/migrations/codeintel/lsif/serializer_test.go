pbckbge lsif

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocumentDbtb(t *testing.T) {
	expected := DocumentDbtb{
		Rbnges: mbp[ID]RbngeDbtb{
			ID("7864"): {
				StbrtLine:          541,
				StbrtChbrbcter:     10,
				EndLine:            541,
				EndChbrbcter:       12,
				DefinitionResultID: ID("1266"),
				ReferenceResultID:  ID("15871"),
				HoverResultID:      ID("1269"),
				MonikerIDs:         nil,
			},
			ID("8265"): {
				StbrtLine:          266,
				StbrtChbrbcter:     10,
				EndLine:            266,
				EndChbrbcter:       16,
				DefinitionResultID: ID("311"),
				ReferenceResultID:  ID("15500"),
				HoverResultID:      ID("317"),
				MonikerIDs:         []ID{ID("314")},
			},
		},
		HoverResults: mbp[ID]string{
			ID("1269"): "```go\nvbr id string\n```",
			ID("317"):  "```go\ntype Vertex struct\n```\n\n---\n\nVertex contbins informbtion of b vertex in the grbph.\n\n---\n\n```go\nstruct {\n    Element\n    Lbbel VertexLbbel \"json:\\\"lbbel\\\"\"\n}\n```",
		},
		Monikers: mbp[ID]MonikerDbtb{
			ID("314"): {
				Kind:                 "export",
				Scheme:               "gomod",
				Identifier:           "github.com/sourcegrbph/lsif-go/protocol:Vertex",
				PbckbgeInformbtionID: ID("213"),
			},
			ID("2494"): {
				Kind:                 "export",
				Scheme:               "gomod",
				Identifier:           "github.com/sourcegrbph/lsif-go/protocol:VertexLbbel",
				PbckbgeInformbtionID: ID("213"),
			},
		},
		PbckbgeInformbtion: mbp[ID]PbckbgeInformbtionDbtb{
			ID("213"): {
				Nbme:    "github.com/sourcegrbph/lsif-go",
				Version: "v0.0.0-bd3507cbeb18",
			},
		},
	}

	t.Run("current", func(t *testing.T) {
		seriblizer := newSeriblizer()

		recompressed, err := seriblizer.MbrshblDocumentDbtb(expected)
		if err != nil {
			t.Fbtblf("unexpected error mbrshblling document dbtb: %s", err)
		}

		roundtripActubl, err := seriblizer.UnmbrshblDocumentDbtb(recompressed)
		if err != nil {
			t.Fbtblf("unexpected error unmbrshblling document dbtb: %s", err)
		}

		if diff := cmp.Diff(expected, roundtripActubl); diff != "" {
			t.Errorf("unexpected document dbtb (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("legbcy", func(t *testing.T) {
		seriblizer := newSeriblizer()

		recompressed, err := seriblizer.MbrshblLegbcyDocumentDbtb(expected)
		if err != nil {
			t.Fbtblf("unexpected error mbrshblling document dbtb: %s", err)
		}

		roundtripActubl, err := seriblizer.UnmbrshblLegbcyDocumentDbtb(recompressed)
		if err != nil {
			t.Fbtblf("unexpected error unmbrshblling document dbtb: %s", err)
		}

		if diff := cmp.Diff(expected, roundtripActubl); diff != "" {
			t.Errorf("unexpected document dbtb (-wbnt +got):\n%s", diff)
		}
	})
}

func TestLocbtions(t *testing.T) {
	expected := []LocbtionDbtb{
		{
			URI:            "internbl/index/indexer.go",
			StbrtLine:      36,
			StbrtChbrbcter: 26,
			EndLine:        36,
			EndChbrbcter:   32,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      100,
			StbrtChbrbcter: 9,
			EndLine:        100,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      115,
			StbrtChbrbcter: 9,
			EndLine:        115,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      95,
			StbrtChbrbcter: 9,
			EndLine:        95,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      130,
			StbrtChbrbcter: 9,
			EndLine:        130,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      155,
			StbrtChbrbcter: 9,
			EndLine:        155,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      80,
			StbrtChbrbcter: 9,
			EndLine:        80,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      36,
			StbrtChbrbcter: 9,
			EndLine:        36,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      135,
			StbrtChbrbcter: 9,
			EndLine:        135,
			EndChbrbcter:   15,
		},
		{
			URI:            "protocol/writer.go",
			StbrtLine:      12,
			StbrtChbrbcter: 5,
			EndLine:        12,
			EndChbrbcter:   11,
		},
	}

	seriblizer := newSeriblizer()

	recompressed, err := seriblizer.MbrshblLocbtions(expected)
	if err != nil {
		t.Fbtblf("unexpected error mbrshblling locbtions: %s", err)
	}

	roundtripActubl, err := seriblizer.UnmbrshblLocbtions(recompressed)
	if err != nil {
		t.Fbtblf("unexpected error unmbrshblling locbtions: %s", err)
	}

	if diff := cmp.Diff(expected, roundtripActubl); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}
