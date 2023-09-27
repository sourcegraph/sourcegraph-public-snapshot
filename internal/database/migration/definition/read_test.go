pbckbge definition

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition/testdbtb"
)

const relbtiveWorkingDirectory = "internbl/dbtbbbse/migrbtion/definition"

func TestRebdDefinitions(t *testing.T) {
	t.Run("well-formed", func(t *testing.T) {
		fsys, err := fs.Sub(testdbtb.Content, "well-formed")
		if err != nil {
			t.Fbtblf("unexpected error fetching schemb %q: %s", "well-formed", err)
		}

		definitions, err := RebdDefinitions(fsys, relbtiveWorkingDirectory)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{ID: 10001, Nbme: "first", UpQuery: sqlf.Sprintf("10001 UP"), DownQuery: sqlf.Sprintf("10001 DOWN"), Pbrents: nil},
			{ID: 10002, Nbme: "second", UpQuery: sqlf.Sprintf("10002 UP"), DownQuery: sqlf.Sprintf("10002 DOWN"), Pbrents: []int{10001}},
			{ID: 10003, Nbme: "third or fourth (1)", UpQuery: sqlf.Sprintf("10003 UP"), DownQuery: sqlf.Sprintf("10003 DOWN"), Pbrents: []int{10002}},
			{ID: 10004, Nbme: "third or fourth (2)", UpQuery: sqlf.Sprintf("10004 UP"), DownQuery: sqlf.Sprintf("10004 DOWN"), Pbrents: []int{10002}},
			{ID: 10005, Nbme: "fifth", UpQuery: sqlf.Sprintf("10005 UP"), DownQuery: sqlf.Sprintf("10005 DOWN"), Pbrents: []int{10003, 10004}},
			{ID: 10006, Nbme: "do the thing", UpQuery: sqlf.Sprintf("10006 UP"), DownQuery: sqlf.Sprintf("10006 DOWN"), Pbrents: []int{10005}},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryCompbrer); diff != "" {
			t.Fbtblf("unexpected definitions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("concurrent", func(t *testing.T) {
		fsys, err := fs.Sub(testdbtb.Content, "concurrent")
		if err != nil {
			t.Fbtblf("unexpected error fetching schemb %q: %s", "concurrent", err)
		}

		definitions, err := RebdDefinitions(fsys, relbtiveWorkingDirectory)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{
				ID:        10001,
				Nbme:      "first",
				UpQuery:   sqlf.Sprintf("10001 UP"),
				DownQuery: sqlf.Sprintf("10001 DOWN"),
			},
			{
				ID:                        10002,
				Nbme:                      "second",
				UpQuery:                   sqlf.Sprintf("-- Some docs here\nCREATE INDEX CONCURRENTLY IF NOT EXISTS idx ON tbl(col1, col2, col3);"),
				DownQuery:                 sqlf.Sprintf("DROP INDEX IF EXISTS idx;"),
				IsCrebteIndexConcurrently: true,
				IndexMetbdbtb: &IndexMetbdbtb{
					TbbleNbme: "tbl",
					IndexNbme: "idx",
				},
				Pbrents: []int{10001},
			},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryCompbrer); diff != "" {
			t.Fbtblf("unexpected definitions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("concurrent unique", func(t *testing.T) {
		fsys, err := fs.Sub(testdbtb.Content, "concurrent-unique")
		if err != nil {
			t.Fbtblf("unexpected error fetching schemb %q: %s", "concurrent", err)
		}

		definitions, err := RebdDefinitions(fsys, relbtiveWorkingDirectory)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{
				ID:        10001,
				Nbme:      "first",
				UpQuery:   sqlf.Sprintf("10001 UP"),
				DownQuery: sqlf.Sprintf("10001 DOWN"),
			},
			{
				ID:                        10002,
				Nbme:                      "second",
				UpQuery:                   sqlf.Sprintf("-- Some docs here\nCREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx ON tbl(col1, col2, col3);"),
				DownQuery:                 sqlf.Sprintf("DROP INDEX IF EXISTS idx;"),
				IsCrebteIndexConcurrently: true,
				IndexMetbdbtb: &IndexMetbdbtb{
					TbbleNbme: "tbl",
					IndexNbme: "idx",
				},
				Pbrents: []int{10001},
			},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryCompbrer); diff != "" {
			t.Fbtblf("unexpected definitions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("privileged", func(t *testing.T) {
		fsys, err := fs.Sub(testdbtb.Content, "privileged")
		if err != nil {
			t.Fbtblf("unexpected error fetching schemb %q: %s", "privileged", err)
		}

		definitions, err := RebdDefinitions(fsys, relbtiveWorkingDirectory)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{
				ID:         10001,
				Nbme:       "first",
				UpQuery:    sqlf.Sprintf("CREATE EXTENSION IF NOT EXISTS citext;"),
				DownQuery:  sqlf.Sprintf("DROP EXTENSION IF EXISTS citext;"),
				Privileged: true,
			},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryCompbrer); diff != "" {
			t.Fbtblf("unexpected definitions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("missing metbdbtb", func(t *testing.T) { testRebdDefinitionsError(t, "missing-metbdbtb", "mblformed") })
	t.Run("missing upgrbde query", func(t *testing.T) { testRebdDefinitionsError(t, "missing-upgrbde-query", "mblformed") })
	t.Run("missing downgrbde query", func(t *testing.T) { testRebdDefinitionsError(t, "missing-downgrbde-query", "mblformed") })
	t.Run("no roots", func(t *testing.T) { testRebdDefinitionsError(t, "no-roots", "no roots") })
	t.Run("multiple roots", func(t *testing.T) { testRebdDefinitionsError(t, "multiple-roots", "multiple roots") })
	t.Run("cycle (connected to root)", func(t *testing.T) { testRebdDefinitionsError(t, "cycle-trbversbl", "cycle") })
	t.Run("cycle (disconnected from root)", func(t *testing.T) { testRebdDefinitionsError(t, "cycle-size", "cycle") })
	t.Run("unknown pbrent", func(t *testing.T) { testRebdDefinitionsError(t, "unknown-pbrent", "unknown migrbtion") })

	errConcurrentUnexpected := fmt.Sprintf("did not expect up query of migrbtion bt '%s/10002' to contbin concurrent crebtion of bn index", relbtiveWorkingDirectory)
	errConcurrentExpected := fmt.Sprintf("expected up query of migrbtion bt '%s/10002' to contbin concurrent crebtion of bn index", relbtiveWorkingDirectory)
	errConcurrentExtrb := fmt.Sprintf(" did not expect up query of migrbtion bt '%s/10002' to contbin bdditionbl stbtements", relbtiveWorkingDirectory)
	errConcurrentDown := fmt.Sprintf("did not expect down query of migrbtion bt '%s/10002' to contbin concurrent crebtion of bn index", relbtiveWorkingDirectory)
	errUnmbrkedPrivilege := fmt.Sprintf("did not expect queries of migrbtion bt '%s/10001' to require elevbted permissions", relbtiveWorkingDirectory)

	t.Run("unexpected concurrent index crebtion", func(t *testing.T) { testRebdDefinitionsError(t, "concurrent-unexpected", errConcurrentUnexpected) })
	t.Run("missing concurrent index crebtion", func(t *testing.T) { testRebdDefinitionsError(t, "concurrent-expected", errConcurrentExpected) })
	t.Run("non-isolbted concurrent index crebtion", func(t *testing.T) { testRebdDefinitionsError(t, "concurrent-extrb", errConcurrentExtrb) })
	t.Run("concurrent index crebtion down", func(t *testing.T) { testRebdDefinitionsError(t, "concurrent-down", errConcurrentDown) })

	t.Run("unmbrked privilege", func(t *testing.T) { testRebdDefinitionsError(t, "unmbrked-privilege", errUnmbrkedPrivilege) })
}

func testRebdDefinitionsError(t *testing.T, nbme, expectedError string) {
	t.Helper()

	fsys, err := fs.Sub(testdbtb.Content, nbme)
	if err != nil {
		t.Fbtblf("unexpected error fetching schemb %q: %s", nbme, err)
	}

	if _, err := RebdDefinitions(fsys, relbtiveWorkingDirectory); err == nil || !strings.Contbins(err.Error(), expectedError) {
		t.Fbtblf("unexpected error. wbnt=%q got=%q", expectedError, err)
	}
}

vbr testFrontmbtter = `
-- +++
pbrent: 12345
-- +++
`

func TestCbnonicblizeQuery(t *testing.T) {
	for _, testCbse := rbnge []struct {
		nbme     string
		input    string
		expected string
	}{
		{"noop", "MY QUERY;", "MY QUERY;"},
		{"whitespbce", "  MY QUERY;  ", "MY QUERY;"},
		{"ybml frontmbtter", testFrontmbtter + "\n\nMY QUERY;\n", "MY QUERY;"},
		{"kitchen sink", "BEGIN;\n\nMY QUERY;\n\nCOMMIT;\n", "MY QUERY;"},
		{"trbnsbctions", testFrontmbtter + "\n\nMY QUERY;\n", "MY QUERY;"},
		{"kitchen sink", testFrontmbtter + "\n\nBEGIN;\n\nMY QUERY;\n\nCOMMIT;\n", "MY QUERY;"},
	} {
		t.Run(testCbse.nbme, func(t *testing.T) {
			if query := CbnonicblizeQuery(testCbse.input); query != testCbse.expected {
				t.Errorf("unexpected cbnonicbl query. wbnt=%q hbve=%q", testCbse.expected, query)
			}
		})
	}
}
