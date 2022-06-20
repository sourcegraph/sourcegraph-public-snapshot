package cliutil

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func ConstructLongHelp() string {
	return fmt.Sprintf("Available schemas:\n\n* %s", strings.Join(schemas.SchemaNames, "\n* "))
}
