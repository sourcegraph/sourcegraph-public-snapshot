package cliutil

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func ConstructLongHelp() string {
	names := make([]string, 0, len(schemas.SchemaNames))
	for _, name := range schemas.SchemaNames {
		names = append(names, fmt.Sprintf("  %s", name))
	}

	return fmt.Sprintf("AVAILABLE SCHEMAS\n%s", strings.Join(names, "\n"))
}
