package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/sourcegraph/src-cli/internal/api"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	usage := `
Examples:

  List repositories:

    	$ src repos list

  Print JSON description of repositories list:

    	$ src repos list -f '{{.|json}}'

  List *all* repositories (may be slow!):

    	$ src repos list -first='-1'

  List repositories whose names match the query:

    	$ src repos list -query='myquery'

`

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		firstFlag = flagSet.Int("first", 1000, "Returns the first n repositories from the list. (use -1 for unlimited)")
		queryFlag = flagSet.String("query", "", `Returns repositories whose names match the query. (e.g. "myorg/")`)
		// TODO: add support for "names" field.
		clonedFlag           = flagSet.Bool("cloned", true, "Include cloned repositories.")
		notClonedFlag        = flagSet.Bool("not-cloned", true, "Include repositories that are not yet cloned and for which cloning is not in progress.")
		indexedFlag          = flagSet.Bool("indexed", true, "Include repositories that have a text search index.")
		notIndexedFlag       = flagSet.Bool("not-indexed", true, "Include repositories that do not have a text search index.")
		orderByFlag          = flagSet.String("order-by", "name", `How to order the results; possible choices are: "name", "created-at"`)
		descendingFlag       = flagSet.Bool("descending", false, "Whether or not results should be in descending order.")
		namesWithoutHostFlag = flagSet.Bool("names-without-host", false, "Whether or not repository names should be printed without the hostname (or other first path component). If set, -f is ignored.")
		formatFlag           = flagSet.String("f", "{{.Name}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Name}}") or "{{.|json}}")`)
		apiFlags             = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := `query Repositories(
  $first: Int,
  $query: String,
  $cloned: Boolean,
  $notCloned: Boolean,
  $indexed: Boolean,
  $notIndexed: Boolean,
  $orderBy: RepositoryOrderBy,
  $descending: Boolean,
) {
  repositories(
    first: $first,
    query: $query,
    cloned: $cloned,
    notCloned: $notCloned,
    indexed: $indexed,
    notIndexed: $notIndexed,
    orderBy: $orderBy,
    descending: $descending,
  ) {
    nodes {
      ...RepositoryFields
    }
  }
}
` + repositoryFragment

		var orderBy string
		switch *orderByFlag {
		case "name":
			orderBy = "REPOSITORY_NAME"
		case "created-at":
			orderBy = "REPO_CREATED_AT"
		default:
			return errors.Newf("invalid -order-by flag value: %q", *orderByFlag)
		}

		var result struct {
			Repositories struct {
				Nodes []Repository
			}
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"first":      api.NullInt(*firstFlag),
			"query":      api.NullString(*queryFlag),
			"cloned":     *clonedFlag,
			"notCloned":  *notClonedFlag,
			"indexed":    *indexedFlag,
			"notIndexed": *notIndexedFlag,
			"orderBy":    orderBy,
			"descending": *descendingFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		for _, repo := range result.Repositories.Nodes {
			if *namesWithoutHostFlag {
				firstSlash := strings.Index(repo.Name, "/")
				fmt.Println(repo.Name[firstSlash+len("/"):])
				continue
			}

			if err := execTemplate(tmpl, repo); err != nil {
				return err
			}
		}
		return nil
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
