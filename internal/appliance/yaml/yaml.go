package yaml

import (
	"bytes"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// ConvertYAMLStringsToMultilineLiterals is intended to make large nested yaml
// strings more human-readable, and diffable (so mainly useful for tests). Note
// that in order to do this reliably, it removed trailing whitespace from all
// lines in multiline string fields.
//
// Do not use this function in contexts where that is problematic!
func ConvertYAMLStringsToMultilineLiterals(doc []byte) ([]byte, error) {
	var rootNode yaml.Node
	if err := yaml.Unmarshal(doc, &rootNode); err != nil {
		return nil, err
	}
	convertYAMLNodeToMultilineStringLiterals(&rootNode)

	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(&rootNode); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func convertYAMLNodeToMultilineStringLiterals(node *yaml.Node) {
	if node.Kind == yaml.ScalarNode && strings.Contains(node.Value, "\n") {
		node.Style = yaml.LiteralStyle

		// See comment on  convertYAMLNodeToMultilineStringLiterals - if we
		// don't do this, string fields containing trailing space will not be
		// represented literal-style, presumably to avoid ambiguity with yaml
		// parsers.
		node.Value = trimTrailingSpaceLines(node.Value)

		return
	}

	// We have a non-scalar node, recurse over its children
	for _, childNode := range node.Content {
		convertYAMLNodeToMultilineStringLiterals(childNode)
	}
}

func trimTrailingSpaceLines(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRightFunc(lines[i], unicode.IsSpace)
	}
	return strings.Join(lines, "\n")
}
