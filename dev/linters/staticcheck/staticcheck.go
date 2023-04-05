package staticcheck

import (
	"fmt"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
)

var Analyzer *analysis.Analyzer = &analysis.Analyzer{}
