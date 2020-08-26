package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src campaigns repositories' works out the repositories that a campaign spec
would apply to.

Usage:

    src campaigns repositories -f FILE

Examples:

    $ src campaigns repositories -f campaign.spec.yaml

`

	flagSet := flag.NewFlagSet("repositories", flag.ExitOnError)

	var (
		fileFlag = flagSet.String("f", "", "The campaign spec file to read.")
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		specFile, err := campaignsOpenFileFlag(fileFlag)
		if err != nil {
			return err
		}
		defer specFile.Close()

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		svc := campaigns.NewService(&campaigns.ServiceOpts{Client: client})
		spec, _, err := svc.ParseCampaignSpec(specFile)
		if err != nil {
			return errors.Wrap(err, "parsing campaign spec")
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		if err := campaignsValidateSpec(out, spec); err != nil {
			return err
		}

		tmpl, err := parseTemplate(campaignsRepositoriesTemplate)
		if err != nil {
			return err
		}

		seen := map[string]struct{}{}
		final := []*graphql.Repository{}
		finalMax := 0
		for _, on := range spec.On {
			repos, err := svc.ResolveRepositoriesOn(ctx, &on)
			if err != nil {
				return errors.Wrapf(err, "Resolving %q", on.String())
			}

			max := 0
			for _, repo := range repos {
				if len(repo.Name) > max {
					max = len(repo.Name)
				}

				if _, ok := seen[repo.ID]; !ok {
					seen[repo.ID] = struct{}{}
					final = append(final, repo)
				}
			}

			if max > finalMax {
				finalMax = max
			}

			if *verbose {
				if err := execTemplate(tmpl, campaignsRepositoryTemplateInput{
					Max:                 max,
					Query:               on.String(),
					RepoCount:           len(repos),
					Repos:               repos,
					SourcegraphEndpoint: cfg.Endpoint,
				}); err != nil {
					return err
				}
			}
		}

		return execTemplate(tmpl, campaignsRepositoryTemplateInput{
			Max:                 finalMax,
			RepoCount:           len(final),
			Repos:               final,
			SourcegraphEndpoint: cfg.Endpoint,
		})
	}

	campaignsCommands = append(campaignsCommands, &command{
		flagSet: flagSet,
		aliases: []string{"repos"},
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

const campaignsRepositoriesTemplate = `
{{- color "logo" -}}✱{{- color "nc" -}}
{{- " " -}}
{{- if eq .RepoCount 0 -}}
    {{- color "warning" -}}
{{- else -}}
    {{- color "success" -}}
{{- end -}}
{{- .RepoCount }} result{{ if ne .RepoCount 1 }}s{{ end }}{{- color "nc" -}}
{{- if ne (len .Query) 0 -}}
    {{- " for " -}}{{- color "search-query"}}"{{.Query}}"{{ color "nc" -}}
{{- end -}}
{{- "\n" -}}

{{- range .Repos -}}
    {{- "  "}}{{ color "success" }}{{ padRight .Name $.Max " " }}{{ color "nc" -}}
    {{- color "search-border"}}{{" ("}}{{color "nc" -}}
    {{- color "search-repository"}}{{$.SourcegraphEndpoint}}{{.URL}}{{color "nc" -}}
    {{- color "search-border"}}{{")\n"}}{{color "nc" -}}
{{- end -}}
`

type campaignsRepositoryTemplateInput struct {
	Max                 int
	Query               string
	RepoCount           int
	Repos               []*graphql.Repository
	SourcegraphEndpoint string
}
