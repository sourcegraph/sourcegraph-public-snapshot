package highlight

import (
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

func TestMultilineOccurrence(t *testing.T) {
	code := `namespace MyCompanyName.MyProjectName;

[DependsOn(
    // ABP Framework packages
    typeof(AbpAspNetCoreMvcModule)
)]`

	document := &scip.Document{
		Occurrences: []*scip.Occurrence{
			// Past end range occurrence
			{
				Range:      []int32{7, 0, 1},
				SyntaxKind: scip.SyntaxKind_BooleanLiteral,
			},
		},
	}

	rows := map[int32]bool{}
	lsifToHTML(code, document, func(row int32) {
		rows[row] = true
	}, func(kind scip.SyntaxKind, line string) {}, nil)

	if len(rows) != 6 {
		t.Error("Should only add once per row, and should skip the 7th row (since it doesn't exist)")
	}
}

func TestMultilineOccurrence2(t *testing.T) {
	code := `namespace MyCompanyName.MyProjectName;

[DependsOn(
    // ABP Framework packages
    typeof(AbpAspNetCoreMvcModule)
)]`

	document := &scip.Document{
		Occurrences: []*scip.Occurrence{
			// Valid range at the beginning, should be skipped.
			{
				Range:      []int32{0, 0, 8},
				SyntaxKind: scip.SyntaxKind_IdentifierNamespace,
			},

			// Past end range occurrence, should not cause panic
			{
				Range:      []int32{2, 0, 30, 4},
				SyntaxKind: scip.SyntaxKind_BooleanLiteral,
			},
		},
	}

	// Should stay false, because we should skip this identifier
	// due to the valid lines that is passed to lsifToHTML
	sawNamespaceIdentifier := false

	rowsSeen := map[int32]bool{}
	lsifToHTML(code, document, func(row int32) {
		rowsSeen[row] = true
	}, func(kind scip.SyntaxKind, line string) {
		if kind == scip.SyntaxKind_IdentifierNamespace {
			sawNamespaceIdentifier = true
		}

	}, map[int32]bool{
		0: false,
		1: false,
		2: true,
		3: false,
		4: false,
	})

	if len(rowsSeen) != 4 {
		t.Error("Should only add the rows from 2 until the end (due to weird multiline occurrence)")
	}

	if sawNamespaceIdentifier {
		t.Error("Should not have seen identifier for module, because line was skipped")
	}
}
