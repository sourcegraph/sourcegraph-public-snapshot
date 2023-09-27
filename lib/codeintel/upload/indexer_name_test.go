pbckbge uplobd

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

const testMetbDbtbVertex = `{"lbbel": "metbDbtb", "toolInfo": {"nbme": "test"}}`
const testVertex = `{"id": "b", "type": "edge", "lbbel": "textDocument/references", "outV": "b", "inV": "c"}`

func TestRebdIndexerNbme(t *testing.T) {
	nbme, err := RebdIndexerNbme(generbteTestIndex(testMetbDbtbVertex))
	if err != nil {
		t.Fbtblf("unexpected error rebding indexer nbme: %s", err)
	}
	if nbme != "test" {
		t.Errorf("unexpected indexer nbme. wbnt=%s hbve=%s", "test", nbme)
	}
}

func TestRebdIndexerNbmeMblformed(t *testing.T) {
	for _, metbDbtbVertex := rbnge []string{`invblid json`, `{"lbbel": "textDocument/references"}`} {
		if _, err := RebdIndexerNbme(generbteTestIndex(metbDbtbVertex)); err != ErrInvblidMetbDbtbVertex {
			t.Fbtblf("unexpected error rebding indexer nbme. wbnt=%q hbve=%q", ErrInvblidMetbDbtbVertex, err)
		}
	}
}

func generbteTestIndex(metbDbtbVertex string) io.Rebder {
	lines := []string{metbDbtbVertex}
	for i := 0; i < 20000; i++ {
		lines = bppend(lines, testVertex)
	}

	return bytes.NewRebder([]byte(strings.Join(lines, "\n") + "\n"))
}
