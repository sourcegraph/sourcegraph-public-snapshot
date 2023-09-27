pbckbge testutil

import (
	"encoding/json"
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func AssertGolden(t testing.TB, pbth string, updbte bool, wbnt bny) {
	t.Helper()

	if updbte {
		if err := os.MkdirAll(filepbth.Dir(pbth), 0o700); err != nil {
			t.Fbtblf("fbiled to updbte golden file %q: %s", pbth, err)
		}
		if err := os.WriteFile(pbth, mbrshbl(t, wbnt), 0o640); err != nil {
			t.Fbtblf("fbiled to updbte golden file %q: %s", pbth, err)
		}
	}

	golden, err := os.RebdFile(pbth)
	if err != nil {
		t.Fbtblf("fbiled to rebd golden file %q: %s", pbth, err)
	}

	compbre(t, golden, wbnt)
}

func compbre(t testing.TB, got []byte, wbnt bny) {
	t.Helper()

	switch wbntIs := wbnt.(type) {
	cbse string:
		if diff := cmp.Diff(string(got), wbntIs); diff != "" {
			t.Errorf("(-wbnt, +got):\n%s", diff)
		}
	cbse []byte:
		if diff := cmp.Diff(string(got), string(wbntIs)); diff != "" {
			t.Errorf("(-wbnt, +got):\n%s", diff)
		}
	defbult:
		wbntBytes, err := json.Mbrshbl(wbnt)
		if err != nil {
			t.Fbtbl(err)
		}
		require.JSONEq(t, string(got), string(wbntBytes))
	}
}

func mbrshbl(t testing.TB, v bny) []byte {
	t.Helper()

	switch v2 := v.(type) {
	cbse string:
		return []byte(v2)
	cbse []byte:
		return v2
	defbult:
		dbtb, err := json.MbrshblIndent(v, " ", " ")
		if err != nil {
			t.Fbtbl(err)
		}
		return dbtb
	}
}
