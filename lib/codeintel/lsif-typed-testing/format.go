package lsif_typed_testing

import (
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func FormatSnapshots(
	index *lsif_typed.Index,
	commentSyntax string,
	symbolFormatter lsif_typed.SymbolFormatter,
) ([]*lsif_typed.SourceFile, error) {
	var result []*lsif_typed.SourceFile
	projectRoot, err := url.Parse(index.Metadata.ProjectRoot)
	if err != nil {
		return nil, err
	}
	for _, document := range index.Documents {
		snapshot, err := FormatSnapshot(document, index, commentSyntax, symbolFormatter)
		if err != nil {
			return nil, err
		}
		sourceFile := lsif_typed.NewSourceFile(
			filepath.Join(projectRoot.Path, document.RelativePath),
			document.RelativePath,
			snapshot,
		)
		result = append(result, sourceFile)
	}
	return result, nil
}

func FormatSnapshot(
	x *lsif_typed.Document,
	index *lsif_typed.Index,
	commentSyntax string,
	formatter lsif_typed.SymbolFormatter,
) (string, error) {
	b := strings.Builder{}
	uri, err := url.Parse(filepath.Join(index.Metadata.ProjectRoot, x.RelativePath))
	if err != nil {
		return "", err
	}
	if uri.Scheme != "file" {
		return "", errors.New("expected url scheme 'file', obtained " + uri.Scheme)
	}
	data, err := os.ReadFile(uri.Path)
	if err != nil {
		return "", err
	}
	symtab := x.SymbolTable()
	sort.SliceStable(x.Occurrences, func(i, j int) bool {
		return isRangeLess(x.Occurrences[i].Range, x.Occurrences[j].Range)
	})
	var formattingError error
	formatSymbol := func(symbol string) string {
		formatted, err := formatter.Format(symbol)
		if err != nil {
			formattingError = errors.CombineErrors(formattingError, errors.Wrapf(err, symbol))
			return err.Error()
		}
		return formatted
	}
	i := 0
	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		b.WriteString(strings.Repeat(" ", len(commentSyntax)))
		b.WriteString(strings.ReplaceAll(line, "\t", " "))
		b.WriteString("\n")
		for i < len(x.Occurrences) && x.Occurrences[i].Range[0] == int32(lineNumber) {
			occ := x.Occurrences[i]
			pos := lsif_typed.NewRangePositionFromLsif(occ.Range)
			if !pos.IsSingleLine() {
				continue
			}
			b.WriteString(commentSyntax)
			for indent := 0; indent < pos.Start.Character; indent++ {
				b.WriteRune(' ')
			}
			length := pos.End.Character - pos.Start.Character
			for caret := 0; caret < length; caret++ {
				b.WriteRune('^')
			}
			b.WriteRune(' ')
			role := "reference"
			isDefinition := occ.SymbolRoles&int32(lsif_typed.SymbolRole_Definition) > 0
			if isDefinition {
				role = "definition"
			}
			b.WriteString(role)
			b.WriteRune(' ')
			b.WriteString(formatSymbol(occ.Symbol))

			if info, ok := symtab[occ.Symbol]; ok && isDefinition {
				prefix := "\n" + commentSyntax + strings.Repeat(" ", pos.Start.Character)
				for _, documentation := range info.Documentation {
					b.WriteString(prefix)
					b.WriteString("documentation ")
					truncatedDocumentation := documentation
					newlineIndex := strings.Index(documentation, "\n")
					if newlineIndex >= 0 {
						truncatedDocumentation = documentation[0:newlineIndex]
					}
					b.WriteString(truncatedDocumentation)
				}
				sort.SliceStable(info.Relationships, func(i, j int) bool {
					return info.Relationships[i].Symbol < info.Relationships[j].Symbol
				})
				for _, relationship := range info.Relationships {
					b.WriteString(prefix)
					b.WriteString("relationship ")
					b.WriteString(formatSymbol(relationship.Symbol))
					if relationship.IsImplementation {
						b.WriteString(" implementation")
					}
					if relationship.IsReference {
						b.WriteString(" reference")
					}
					if relationship.IsTypeDefinition {
						b.WriteString(" type_definition")
					}
				}
			}

			b.WriteString("\n")
			i++
		}
	}
	return b.String(), formattingError
}

func isRangeLess(a []int32, b []int32) bool {
	if a[0] != b[0] { // start line
		return a[0] < b[0]
	}
	if a[1] != b[1] { // start character
		return a[1] < b[1]
	}
	if len(a) != len(b) { // is one of these multiline
		return len(a) < len(b)
	}
	if a[2] != b[2] { // end line
		return a[2] < b[1]
	}
	// end character
	return a[3] < b[2]
}
