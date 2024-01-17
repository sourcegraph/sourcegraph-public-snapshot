package ipynb

import (
	"bytes"
	"fmt"

	"github.com/bevzzz/nb"

	"github.com/sourcegraph/sourcegraph/internal/htmlutil"
)

// Render renders Jupyter Notebook file (.ipynb) to sanitized HTML that is safe to run anywhere.
func Render(content string) (string, error) {
	var buf bytes.Buffer
	if err := nb.Convert(&buf, []byte(content)); err != nil {
		return "", fmt.Errorf("ipynb.Render: %w", err)
	}
	return htmlutil.SanitizeReader(&buf).String(), nil
}
