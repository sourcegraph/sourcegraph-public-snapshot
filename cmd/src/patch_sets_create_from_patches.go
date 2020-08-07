package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns"
)

func init() {
	usage := `
Create a patch set from a set of patches (in unified diff format) to repository branches.

Standard input is expected to be a JSON array of {repository: string, baseRef: string, baseRevision: string, patch: string}. The repository value is the repository's GraphQL ID (which you can look up given a repository name using 'src repos get -name=...').

Examples:

  Create a patch set from my.patch applied to a repository's master branch:

		$ patch=$(jq -M -R -s . < my.patch)
		$ repo=$(src repo get -name=github.com/sourcegraph/src-cli -f '{{.ID|json}}')
		$ echo '[{"repository": $repo, "baseRef": "refs/heads/master", "baseRevision": "f00b4r", "patch": $patch}]' |\
		  src campaigns patchset create-from-patches

  Create a patch set from patches.json produced by 'src actions exec':

		$ src actions exec -f action.json > patches.json
		$ src campaigns patchset create-from-patches < patches.json

  Create a patch set by piping output of 'src actions exec' into 'src patchset create-from-patches':

		$ src actions exec -f action.json | src patchset create-from-patches < patches.json

`

	flagSet := flag.NewFlagSet("create-from-patches", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns patchsets %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		patchesFlag = flagSet.Int("patches", 1000, "Returns the first n patches in the patch set.")
		formatFlag  = flagSet.String("f", "{{friendlyPatchSetCreatedMessage .}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{len .Patches}} patches") or "{{.|json}}")`)
		apiFlags    = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		if isatty.IsTerminal(os.Stdin.Fd()) {
			log.Println("# Waiting for JSON patches input on stdin...")
		}

		var patches []campaigns.PatchInput
		if err := json.NewDecoder(os.Stdin).Decode(&patches); err != nil {
			return errors.Wrap(err, "invalid JSON patches input")
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		return createPatchSetFromPatches(ctx, client, patches, tmpl, *patchesFlag)
	}

	// Register the command.
	campaignPatchSetsCommands = append(campaignPatchSetsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const createPatchSetMutation = `
mutation CreatePatchSetFromPatches($patches: [PatchInput!]!) {
  createPatchSetFromPatches(patches: $patches) {
    ...PatchSetFields
  }
}
`

func createPatchSetFromPatches(
	ctx context.Context,
	client api.Client,
	patches []campaigns.PatchInput,
	tmpl *template.Template,
	numChangesets int,
) error {
	query := createPatchSetMutation + patchSetFragment(numChangesets)

	var result struct {
		CreatePatchSetFromPatches PatchSet
	}

	version, err := getSourcegraphVersion(ctx, client)
	if err != nil {
		return err
	}
	supportsBaseRef, err := sourcegraphVersionCheck(version, ">= 3.14.0", "2020-03-11")
	if err != nil {
		return err
	}

	// If we're on Sourcegraph >=3.14 the GraphQL API is "fixed" and accepts
	// patches with `BaseRevision` and `BaseRef` fields. <3.14 expects a ref
	// (e.g. "refs/heads/master") in `BaseRevision`, so we need to copy the
	// value over.
	if !supportsBaseRef {
		patchesWithoutBaseRef := make([]campaigns.PatchInput, len(patches))
		for i, p := range patches {
			patchesWithoutBaseRef[i] = campaigns.PatchInput{
				Repository:   p.Repository,
				BaseRevision: p.BaseRef,
				BaseRef:      "IGNORE-THIS",
				Patch:        p.Patch,
			}
		}
		patches = patchesWithoutBaseRef
	}

	if ok, err := client.NewRequest(query, map[string]interface{}{
		"patches": patches,
	}).Do(ctx, &result); err != nil || !ok {
		return err
	}

	return execTemplate(tmpl, result.CreatePatchSetFromPatches)
}
