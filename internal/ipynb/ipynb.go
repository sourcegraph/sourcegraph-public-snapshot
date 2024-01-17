package ipynb

import (
	"fmt"
	"strings"

	"github.com/bevzzz/nb"
)

// Render renders Jupyter Notebook file (.ipynb) to sanitized HTML that is safe to run anywhere.
func Render(content string) (string, error) {
	var w strings.Builder
	if err := nb.Convert(&w, []byte(content)); err != nil {
		return "", fmt.Errorf("ipynb.Render: %w", err)
	}
	return w.String(), nil
}
