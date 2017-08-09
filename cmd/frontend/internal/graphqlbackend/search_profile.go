package graphqlbackend

import (
	"context"
	"sync"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type searchProfile struct {
	name        string
	description string

	// repos are already resolved repos
	repos []*sourcegraph.Repo

	// uris are repos that need to be fetched
	uris []string
}

func (p *searchProfile) Name() string {
	return p.name
}

func (p *searchProfile) Description() *string {
	if p.description == "" {
		return nil
	}
	return &p.description
}

func (p *searchProfile) Repositories(ctx context.Context) ([]*repositoryResolver, error) {
	resolvers := make([]*repositoryResolver, 0, len(p.repos)+len(p.uris))
	for _, repo := range p.repos {
		resolvers = append(resolvers, &repositoryResolver{repo: repo})
	}

	extra, err := repositoriesByURLs(ctx, p.uris)
	if err != nil {
		return nil, err
	}
	for _, resolver := range extra {
		if resolver != nil {
			resolvers = append(resolvers, resolver)
		}
	}
	return resolvers, nil
}

func (*rootResolver) SearchProfiles(ctx context.Context) ([]*searchProfile, error) {
	active, inactive, err := listActiveAndInactive(ctx)
	if err != nil {
		return nil, err
	}
	profiles := []*searchProfile{}
	if len(active) > 0 {
		profiles = append(profiles, &searchProfile{
			name:        "Active",
			description: "Repositories that are active.",
			repos:       active,
		})
	}
	if len(inactive) > 0 {
		profiles = append(profiles, &searchProfile{
			name:        "Inactive",
			description: "Repositories that are marked inactive.",
			repos:       inactive,
		})
	}

	if len(profiles) == 0 {
		// Only add examples groups if we don't have active or inactive (unauthed user)
		oss := ossProfiles()
		profiles = append(profiles, oss...)
	}

	return profiles, nil
}

func ossProfiles() []*searchProfile {
	return []*searchProfile{{
		name:        "Go",
		description: "Repositories part of the Go Programming Language project.",
		uris:        []string{"github.com/golang/go", "github.com/golang/net", "github.com/golang/tools", "github.com/golang/crypto", "github.com/golang/sys", "github.com/golang/arch", "github.com/golang/sync"},
	}, {
		name:        "Angular",
		description: "Repositories part of the Angular 2 Framework project.",
		uris:        []string{"github.com/angular/angular", "github.com/angular/material2", "github.com/angular/angular-cli"},
	}, {
		name:        "vscode",
		description: "Repositories related to the Visual Studio Code project. Taken from https://github.com/Microsoft/vscode/wiki/Related-Projects",
		uris: []string{"github.com/Microsoft/vscode",
			// Core Repositories
			"github.com/Microsoft/monaco-editor", "github.com/Microsoft/vscode-node-debug2", "github.com/Microsoft/vscode-filewatcher-windows", "github.com/Microsoft/vscode-extension-vscode", "github.com/Microsoft/vscode-languageserver-node", "github.com/Microsoft/vscode-textmate", "github.com/Microsoft/vscode-loader",
			// SDK Tools
			"github.com/Microsoft/vscode-generator-code", "github.com/Microsoft/vscode-vsce",
			// Documentation
			"github.com/Microsoft/vscode-docs",
			// Languages
			"github.com/Microsoft/language-server-protocol", "github.com/OmniSharp/omnisharp-vscode", "github.com/Microsoft/vscode-go", "github.com/Microsoft/vscode-latex", "github.com/Microsoft/vscode-css-languageservice", "github.com/Microsoft/vscode-json-languageservice", "github.com/Microsoft/vscode-html-languageservice",
			// Linters
			"github.com/Microsoft/vscode-jscs", "github.com/Microsoft/vscode-tslint", "github.com/Microsoft/vscode-eslint", "github.com/Microsoft/vscode-jshint",
			// Themes
			"github.com/Microsoft/vscode-themes",
		},
	}, {
		name:        "Dropwizard",
		description: "Repositories related to the Dropwizard project.",
		uris: []string{"github.com/dropwizard/dropwizard",
			// Jetty for HTTP servin".
			"github.com/eclipse/jetty.project",
			// Jersey for REST modelin".
			"github.com/jersey/jersey",
			// Jackson for JSON parsin" and generatin".
			"github.com/FasterXML/jackson-core", "github.com/FasterXML/jackson-annotations", "github.com/FasterXML/jackson-databind",
			// Logback for loggin".
			"github.com/qos-ch/logback",
			// Hibernate Validator for validatin".
			"github.com/hibernate/hibernate-validator",
			// Metrics for figurin" out what your application is doin" in production.
			"github.com/dropwizard/metrics",
			// JDBI and Hibernate for databasin".
			"github.com/jdbi/jdbi", "github.com/hibernate/hibernate-orm",
			// Liquibase for migratin".
			"github.com/liquibase/liquibase",
		},
	}, {
		name:        "Kubernetes",
		description: "Projects part of the Kubernetes Container Orchestartion project.",
		uris:        []string{"github.com/kubernetes/kubernetes", "github.com/kubernetes/contrib", "github.com/kubernetes/charts", "github.com/kubernetes/client-go"},
	}}
}

// repositoriesByURLs will resolve a list of repos URIs into a list of repositoryResolver.
// The order of the repos URI list is respected. Repos that 404 will have a nil repositoryResolver.
func repositoriesByURLs(ctx context.Context, uris []string) ([]*repositoryResolver, error) {
	var (
		wg        sync.WaitGroup
		errOnce   sync.Once
		reposErr  error
		resolvers = make([]*repositoryResolver, len(uris))
	)
	for i, uri := range uris {
		i, uri := i, uri
		wg.Add(1)
		go func() {
			defer wg.Done()

			repo, err := localstore.Repos.GetByURI(ctx, uri)
			if err != nil {
				if isNotFound(err) { // ignore repos we haven't cloned
					return
				}
				errOnce.Do(func() {
					reposErr = err
				})
				return
			}
			resolvers[i] = &repositoryResolver{repo: repo}
		}()
	}

	wg.Wait()
	if reposErr != nil {
		return nil, reposErr
	}
	return resolvers, nil
}

func isNotFound(err error) bool {
	return legacyerr.ErrCode(err) == legacyerr.NotFound
}
