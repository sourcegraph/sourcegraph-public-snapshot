package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  List repositories:

    	$ src repos list

  List *all* repositories (may be slow!):

    	$ src repos list -first='-1'

  List repositories whose names match the query:

    	$ src repos list -query='myquery'

  Include repositories that are disabled:

    	$ src repos list -query='myquery' -disabled

  List only repositories that are disabled:

    	$ src repos list -disabled -enabled=false -query='github.com/slimsag/'

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
		enabledFlag         = flagSet.Bool("enabled", true, "Include enabled repositories.")
		disabledFlag        = flagSet.Bool("disabled", false, "Include disabled repositories.")
		clonedFlag          = flagSet.Bool("cloned", true, "Include cloned repositories.")
		cloneInProgressFlag = flagSet.Bool("clone-in-progress", true, "Include repositories that are currently being cloned.")
		notClonedFlag       = flagSet.Bool("not-cloned", true, "Include repositories that are not yet cloned and for which cloning is not in progress.")
		indexedFlag         = flagSet.Bool("indexed", true, "Include repositories that have a text search index.")
		notIndexedFlag      = flagSet.Bool("not-indexed", true, "Include repositories that do not have a text search index.")
		orderByFlag         = flagSet.String("order-by", "name", `How to order the results; possible choices are: "name", "created-at"`)
		descendingFlag      = flagSet.Bool("descending", false, "Whether or not results should be in descending order.")
		apiFlags            = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		query := `query Repositories(
  $first: Int,
  $query: String,
  $enabled: Boolean,
  $disabled: Boolean,
  $cloned: Boolean,
  $cloneInProgress: Boolean,
  $notCloned: Boolean,
  $indexed: Boolean,
  $notIndexed: Boolean,
  $orderBy: RepositoryOrderBy,
  $descending: Boolean,
) {
  repositories(
    first: $first,
    query: $query,
    enabled: $enabled,
    disabled: $disabled,
    cloned: $cloned,
    cloneInProgress: $cloneInProgress,
    notCloned: $notCloned,
    indexed: $indexed,
    notIndexed: $notIndexed,
    orderBy: $orderBy,
    descending: $descending,
  ) {
    nodes {
      name
    }
  }
}`

		var orderBy string
		switch *orderByFlag {
		case "name":
			orderBy = "REPO_URI"
		case "created-at":
			orderBy = "REPO_CREATED_AT"
		default:
			return fmt.Errorf("invalid -order-by flag value: %q", *orderByFlag)
		}

		var result struct {
			Repositories struct {
				Nodes []struct {
					Name string
				}
			}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"first":           nullInt(*firstFlag),
				"query":           nullString(*queryFlag),
				"enabled":         *enabledFlag,
				"disabled":        *disabledFlag,
				"cloned":          *clonedFlag,
				"cloneInProgress": *cloneInProgressFlag,
				"notCloned":       *notClonedFlag,
				"indexed":         *indexedFlag,
				"notIndexed":      *notIndexedFlag,
				"orderBy":         orderBy,
				"descending":      *descendingFlag,
			},
			result: &result,
			done: func() error {
				for _, repo := range result.Repositories.Nodes {
					fmt.Println(repo.Name)
				}
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
