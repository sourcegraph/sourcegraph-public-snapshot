package migration

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

func Visualize(database db.Database, filepath string) error {
	definitions, err := readDefinitions(database)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, formatMigrations(definitions), os.ModePerm)
}

func formatMigrations(definitions *definition.Definitions) []byte {
	var (
		all    = definitions.All()
		root   = definitions.Root()
		leaves = definitions.Leaves()
	)

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "digraph migrations {\n")
	fmt.Fprintf(buf, "  rankdir = LR\n")
	fmt.Fprintf(buf, "  subgraph {\n")

	for _, migrationDefinition := range all {
		for _, parent := range migrationDefinition.Parents {
			fmt.Fprintf(buf, "    %d -> %d\n", parent, migrationDefinition.ID)
		}
	}

	strLeaves := make([]string, 0, len(leaves))
	for _, migrationDefinition := range leaves {
		strLeaves = append(strLeaves, strconv.Itoa(migrationDefinition.ID))
	}

	fmt.Fprintf(buf, "    {rank = same; %d; }\n", root.ID)
	fmt.Fprintf(buf, "    {rank = same; %s; }\n", strings.Join(strLeaves, "; "))
	fmt.Fprintf(buf, "  }\n")
	fmt.Fprintf(buf, "}\n")

	return buf.Bytes()
}
