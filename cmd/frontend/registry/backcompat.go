package registry

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/langservers"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// listSynthesizedRegistryExtensions returns a list of registry extensions that are synthesized from
// known language servers (matching the query).
//
// BACKCOMPAT: This eases the transition to extensions from language servers configured in the site
// config "langservers" property.
func listSynthesizedRegistryExtensions(ctx context.Context, query string) []*registry.Extension {
	backcompatLangServerExtensionsMu.Lock()
	defer backcompatLangServerExtensionsMu.Unlock()
	return FilterRegistryExtensions(backcompatLangServerExtensions, query)
}

func getSynthesizedRegistryExtension(ctx context.Context, field, value string) (*registry.Extension, error) {
	backcompatLangServerExtensionsMu.Lock()
	defer backcompatLangServerExtensionsMu.Unlock()
	return FindRegistryExtension(backcompatLangServerExtensions, field, value), nil
}

var (
	backcompatLangServerExtensionsMu sync.Mutex
	backcompatLangServerExtensions   []*registry.Extension
)

func init() {
	// Synthesize extensions for language server in the site config "langservers" property, and keep
	// them in sync.
	var lastEnabledLangServers []*schema.Langservers
	conf.Watch(func() {
		enabledLangServers := conf.EnabledLangservers()

		// Nothing to do if the relevant config value didn't change.
		if reflect.DeepEqual(enabledLangServers, lastEnabledLangServers) {
			return
		}
		lastEnabledLangServers = enabledLangServers

		backcompatLangServerExtensionsMu.Lock()
		defer backcompatLangServerExtensionsMu.Unlock()
		backcompatLangServerExtensions = make([]*registry.Extension, 0, len(enabledLangServers))
		for _, ls := range enabledLangServers {
			info := langservers.StaticInfo[ls.Language]

			lang := ls.Language
			if info != nil {
				lang = info.DisplayName
			}
			title := lang
			readme := `# ` + lang + ` language server` + "\n\n"
			var description string
			if info != nil {
				var maybeExperimental string
				if info.Experimental {
					maybeExperimental = " **EXPERIMENTAL**"
				}
				repoName := strings.TrimPrefix(info.HomepageURL, "https://github.com/")
				description = info.DisplayName + " code intelligence using the " + repoName + " language server"
				readme += `This extension provides code intelligence for ` + info.DisplayName + ` using the` + maybeExperimental + ` [` + repoName + ` language server](` + info.HomepageURL + `).` + "\n\n"
			}
			readme += `This extension was automatically created from the Sourcegraph site configuration's ` + "`" + `langservers.` + ls.Language + "`" + ` setting. Site admins may delete this extension by removing that setting from site configuration.` + "\n\n"
			if info != nil {
				readme += `More information:

* [Documentation and configuration options](` + info.DocsURL + `)
* [Source code and repository](` + info.HomepageURL + `)
* [Issue tracker](` + info.IssuesURL + `)`
			}

			var url string
			if info != nil {
				url = info.HomepageURL
			}

			var addr string
			if ls.Address != "" {
				// Address is specified in site config; prefer that.
				addr = ls.Address
			} else if info.SiteConfig.Address != "" {
				// Use the default TCP address. This is necessary to know the address on Data
				// Center, because it is not necessary to specify the address in site config on Data
				// Center for builtin lang servers.
				//
				// TODO(sqs): The better way to obtain the address on Kubernetes would be to use
				// the LANGSERVER_xyz vars, which are only set on the lsp-proxy deployment. That
				// would get the correct address even when it is changed from the default in
				// deploy-sourcegraph.
				addr = info.SiteConfig.Address
			}
			if addr == "" {
				title += " (unavailable)"
				readme += "\n\n## Status: unavailable\nThis language server is unavailable because no TCP address is specified for it in site configuration."
			}

			x := schema.SourcegraphExtensionManifest{
				Title:       title,
				Description: description,
				Readme:      readme,
				// The same extension is used for each language server (for now). It is built from
				// https://github.com/sourcegraph/sourcegraph-langserver-http.
				Url:              "https://storage.googleapis.com/sourcegraph-cx-dev/sourcegraph-langserver-http.4.js",
				ActivationEvents: []string{"onLanguage:" + ls.Language},
				Contributes: &schema.Contributions{
					Actions: []*schema.Action{
						&schema.Action{
							Id: "langserver.status",
							ActionItem: &schema.ActionItem{
								Description: "Code intelligence active for ${resource.language}",
								// This is a data URI for an SVG icon of the green plug.
								IconURL: "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2aWV3Qm94PSIwIDAgMjQgMjQiPgogICAgICAgICAgICAgICAgICAgPHN2ZyBjbGFzcz0ibWRpLWljb24gaWNvbi1pbmxpbmUiIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgZmlsbD0iIzM3YjI0ZCIgdmlld0JveD0iMCAwIDI0IDI0Ij4KICAgICAgICAgICAgICAgICAgICAgPHBhdGggZD0iTTE2LDdWM0gxNFY3SDEwVjNIOFY3SDhDNyw3IDYsOCA2LDlWMTQuNUw5LjUsMThWMjFIMTQuNVYxOEwxOCwxNC41VjlDMTgsOCAxNyw3IDE2LDdaIj48L3BhdGg+CiAgICAgICAgICAgICAgICAgICA8L3N2Zz4KICAgICAgICAgICAgICAgICA8L3N2Zz4KICAgICAgICAgICAgICAgIA==",
							},
						},
					},
					Menus: &schema.Menus{
						EditorTitle: []*schema.MenuItem{
							&schema.MenuItem{
								Action: "langserver.status",
								When:   "resource",
							},
						},
					},
				},
			}
			if ls.InitializationOptions != nil {
				x.Args = &ls.InitializationOptions
			}
			data, err := json.MarshalIndent(x, "", "  ")
			if err != nil {
				log15.Error("Parsing the JSON manifest for builtin language server failed. Omitting.", "lang", lang, "err", err)
				continue
			}
			dataStr := string(data)

			backcompatLangServerExtensions = append(backcompatLangServerExtensions, &registry.Extension{
				UUID:        uuid.NewSHA1(uuid.Nil, []byte(ls.Language)).String(),
				ExtensionID: "langserver/" + ls.Language,
				Publisher:   registry.Publisher{Name: "langserver"},
				Name:        ls.Language + "-langserver",
				Manifest:    &dataStr,
				URL:         url,

				IsSynthesizedLocalExtension: true,
			})
		}
		sort.Slice(backcompatLangServerExtensions, func(i, j int) bool {
			return backcompatLangServerExtensions[i].ExtensionID < backcompatLangServerExtensions[j].ExtensionID
		})
	})
}
