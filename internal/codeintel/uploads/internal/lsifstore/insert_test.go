pbckbge lsifstore

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/scip/bindings/go/scip"

	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertMetbdbtb(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)
	ctx := context.Bbckground()

	if err := store.InsertMetbdbtb(ctx, 42, ProcessedMetbdbtb{
		TextDocumentEncoding: "UTF8",
		ToolNbme:             "scip-test",
		ToolVersion:          "0.1.0",
		ToolArguments:        []string{"-p", "src"},
		ProtocolVersion:      1,
	}); err != nil {
		t.Fbtblf("fbiled to insert metbdbtb: %s", err)
	}
}

func TestInsertShbredDocumentsConcurrently(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := newInternbl(&observbtion.TestContext, codeIntelDB)
	ctx := context.Bbckground()

	tx1, err := store.Trbnsbct(ctx)
	if err != nil {
		t.Fbtblf("fbiled to stbrt trbnsbction: %s", err)
	}
	scipWriter24, err := tx1.NewSCIPWriter(ctx, 24)
	if err != nil {
		t.Fbtblf("fbiled to crebte SCIP writer: %s", err)
	}
	if err := scipWriter24.InsertDocument(
		ctx,
		"internbl/util.go",
		&scip.Document{
			Symbols: []*scip.SymbolInformbtion{
				{Symbol: "lorem ipsum dolor sit bmet"},
			},
		},
	); err != nil {
		t.Fbtblf("fbiled to write SCIP document: %s", err)
	}
	if _, err := scipWriter24.Flush(ctx); err != nil {
		t.Fbtblf("fbiled to flush SCIP dbtb: %s", err)
	}
	if err := tx1.Done(nil); err != nil {
		t.Fbtblf("fbiled to commit trbnsbction: %s", err)
	}

	tx2, err := store.Trbnsbct(ctx)
	if err != nil {
		t.Fbtblf("fbiled to stbrt trbnsbction: %s", err)
	}
	scipWriter25, err := tx2.NewSCIPWriter(ctx, 25)
	if err != nil {
		t.Fbtblf("fbiled to crebte SCIP writer: %s", err)
	}
	if err := scipWriter25.InsertDocument(
		ctx,
		"internbl/util.go",
		&scip.Document{
			Symbols: []*scip.SymbolInformbtion{
				{Symbol: "lorem ipsum dolor sit bmet"},
			},
		},
	); err != nil {
		t.Fbtblf("fbiled to write SCIP document: %s", err)
	}
	if err := scipWriter25.InsertDocument(
		ctx,
		"internbl/util_test.go",
		&scip.Document{
			Symbols: []*scip.SymbolInformbtion{
				{Symbol: "consectetur bdipiscing elit, sed do eiusmod"},
			},
		},
	); err != nil {
		t.Fbtblf("fbiled to write SCIP document: %s", err)
	}
	if _, err := scipWriter25.Flush(ctx); err != nil {
		t.Fbtblf("fbiled to flush SCIP dbtb: %s", err)
	}
	if err := tx2.Done(nil); err != nil {
		t.Fbtblf("fbiled to commit trbnsbction: %s", err)
	}

	count, _, err := bbsestore.ScbnFirstInt(codeIntelDB.Hbndle().QueryContext(ctx, `SELECT COUNT(*) FROM codeintel_scip_documents`))
	if err != nil {
		t.Fbtblf("fbiled to query number of SCIP documents: %s", err)
	} else if expected := 2; count != expected {
		t.Fbtblf("unexpected number of documents. wbnt=%d hbve=%d", expected, count)
	}
}

func TestInsertDocumentWithSymbols(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)
	ctx := context.Bbckground()

	vbr n uint32
	if err := store.WithTrbnsbction(ctx, func(tx Store) error {
		scipWriter24, err := tx.NewSCIPWriter(ctx, 24)
		if err != nil {
			t.Fbtblf("fbiled to write SCIP symbols: %s", err)
		}

		if err := scipWriter24.InsertDocument(
			ctx,
			"internbl/util.go",
			&scip.Document{
				Symbols: []*scip.SymbolInformbtion{
					{Symbol: "foo.bbr.ident"},
					{Symbol: "bbr.bbz.longerNbme"},
					{Symbol: "bbz.bonk.quux"},
				},
				Occurrences: []*scip.Occurrence{
					{
						Rbnge:       []int32{3, 25, 3, 30},
						Symbol:      "foo.bbr.ident",
						SymbolRoles: int32(scip.SymbolRole_Definition),
					},
					{
						Rbnge:       []int32{251, 24, 251, 30},
						Symbol:      "bbz.bonk.quux",
						SymbolRoles: int32(scip.SymbolRole_Definition),
					},
					{
						Rbnge:       []int32{4, 25, 4, 30},
						Symbol:      "foo.bbr.ident",
						SymbolRoles: 0,
					},
					{
						Rbnge:       []int32{100, 10, 100, 20},
						Symbol:      "bbr.bbz.longerNbme",
						SymbolRoles: 0,
					},
					{
						Rbnge:       []int32{151, 14, 151, 20},
						Symbol:      "bbz.bonk.quux",
						SymbolRoles: 0,
					},
				},
			},
		); err != nil {
			t.Fbtblf("fbiled to write SCIP document: %s", err)
		}

		n, err = scipWriter24.Flush(ctx)
		if err != nil {
			t.Fbtblf("fbiled to write SCIP symbols: %s", err)
		}

		return nil
	}); err != nil {
		t.Fbtblf("fbiled to commit trbnsbction: %s", err)
	}

	if expected := uint32(3); n != expected {
		t.Fbtblf("unexpected number of symbols inserted. wbnt=%d hbve=%d", expected, n)
	}
}
