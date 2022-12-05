package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  Update the value for a key on a repository:

    	$ src repos update-kvp -repo=repoID -key=my-key -value=new-value

  Omitting -value will set the value of the key to null.
`

	flagSet := flag.NewFlagSet("update-kvp", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag  = flagSet.String("repo", "", `The ID of the repo with the key to be updated (required)`)
		keyFlag   = flagSet.String("key", "", `The name of the key to be updated (required)`)
		valueFlag = flagSet.String("value", "", `The new value of the key to be set. Defaults to null.`)
		apiFlags  = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}
		if *repoFlag == "" {
			return errors.New("error: repo is required")
		}

		keyFlag = nil
		valueFlag = nil
		flagSet.Visit(func(f *flag.Flag) {
			if f.Name == "key" {
				key := f.Value.String()
				keyFlag = &key
			}

			if f.Name == "value" {
				value := f.Value.String()
				valueFlag = &value
			}
		})
		if keyFlag == nil {
			return errors.New("error: key is required")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation updateKVP(
  $repo: ID!,
  $key: String!,
  $value: String,
) {
  updateRepoKeyValuePair(
    repo: $repo,
    key: $key,
    value: $value,
  ) {
    alwaysNil
  }
}`

		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repo":  *repoFlag,
			"key":   *keyFlag,
			"value": valueFlag,
		}).Do(context.Background(), nil); err != nil || !ok {
			return err
		}

		if valueFlag != nil {
			fmt.Printf("Value of key '%s' updated to '%v'\n", *keyFlag, *valueFlag)
		} else {
			fmt.Printf("Value of key '%s' updated to <nil>\n", *keyFlag)
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
