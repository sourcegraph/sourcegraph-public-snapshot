package backend

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/langservers"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/registry"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ListSynthesizedRegistryExtensions returns a list registry extensions that are synthesized from
// known language servers.
//
// BACKCOMPAT: This eases the transition to extensions from language servers configured in the site
// config "langservers" property.
func ListSynthesizedRegistryExtensions(ctx context.Context, opt db.RegistryExtensionsListOptions) []*registry.Extension {
	backcompatLangServerExtensionsMu.Lock()
	defer backcompatLangServerExtensionsMu.Unlock()
	return FilterRegistryExtensions(backcompatLangServerExtensions, opt)
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
	conf.Watch(func() {
		backcompatLangServerExtensionsMu.Lock()
		defer backcompatLangServerExtensionsMu.Unlock()

		for _, ls := range conf.EnabledLangservers() {
			info := langservers.StaticInfo[ls.Language]

			lang := ls.Language
			if info != nil {
				lang = info.DisplayName
			}
			description := `# ` + lang + ` language server` + "\n\n"
			if info != nil {
				var maybeExperimental string
				if info.Experimental {
					maybeExperimental = " **EXPERIMENTAL**"
				}
				repoName := strings.TrimPrefix(info.HomepageURL, "https://github.com/")
				description += `This extension provides code intelligence for ` + info.DisplayName + ` using the` + maybeExperimental + ` [` + repoName + ` language server](` + info.HomepageURL + `).` + "\n\n"
			}
			description += `This extension was automatically created from the Sourcegraph site configuration's ` + "`" + `langservers.` + ls.Language + "`" + ` setting. Site admins may delete this extension by removing that setting from site configuration.` + "\n\n"
			if info != nil {
				description += `More information:

* [Documentation and configuration options](` + info.DocsURL + `)
* [Source code and repository](` + info.HomepageURL + `)
* [Issue tracker](` + info.IssuesURL + `)`
			}

			x := schema.SourcegraphExtension{
				Title:       info.DisplayName + " code intelligence",
				Description: description,
				Platform: schema.ExtensionPlatform{
					Tcp: &schema.TCPTarget{
						Type:    "tcp",
						Address: strings.TrimPrefix(ls.Address, "tcp://"),
					},
				},
				ActivationEvents: []string{"*"},
			}
			if ls.InitializationOptions != nil {
				var v interface{} = ls.InitializationOptions
				x.Args = &v
			}
			data, err := json.MarshalIndent(x, "", "  ")
			if err != nil {
				panic(err)
			}
			dataStr := string(data)

			backcompatLangServerExtensions = append(backcompatLangServerExtensions, &registry.Extension{
				UUID:        uuid.NewSHA1(uuid.Nil, []byte(ls.Language)).String(),
				ExtensionID: "langserver/" + ls.Language,
				Publisher:   registry.Publisher{Name: "langserver"},
				Name:        ls.Language + "-langserver",
				Manifest:    &dataStr,
				URL:         info.HomepageURL,

				IsSynthesizedLocalExtension: true,
			})
		}
		sort.Slice(backcompatLangServerExtensions, func(i, j int) bool {
			return backcompatLangServerExtensions[i].ExtensionID < backcompatLangServerExtensions[j].ExtensionID
		})
	})
}
