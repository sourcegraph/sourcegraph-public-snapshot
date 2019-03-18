package graphqlbackend

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/bg"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/langservers"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

type langServerResolver struct {
	language                                     string
	displayName                                  string
	homepageURL, issuesURL, docsURL              string
	isClusterDeployment                          bool
	custom                                       bool
	experimental                                 bool
	state                                        langservers.ConfigState
	pending                                      bool
	downloading                                  bool
	canEnable, canDisable, canRestart, canUpdate bool
	healthy                                      bool
}

func (c *langServerResolver) Language(ctx context.Context) string    { return c.language }
func (c *langServerResolver) DisplayName(ctx context.Context) string { return c.displayName }
func (c *langServerResolver) HomepageURL(ctx context.Context) *string {
	return nullString(c.homepageURL)
}
func (c *langServerResolver) IssuesURL(ctx context.Context) *string {
	return nullString(c.issuesURL)
}
func (c *langServerResolver) DocsURL(ctx context.Context) *string { return nullString(c.docsURL) }
func (c *langServerResolver) IsClusterDeployment(ctx context.Context) bool {
	return c.isClusterDeployment
}
func (c *langServerResolver) Custom(ctx context.Context) bool       { return c.custom }
func (c *langServerResolver) Experimental(ctx context.Context) bool { return c.experimental }
func (c *langServerResolver) State(ctx context.Context) string {
	switch c.state {
	case langservers.StateNone:
		return "LANG_SERVER_STATE_NONE"
	case langservers.StateEnabled:
		return "LANG_SERVER_STATE_ENABLED"
	case langservers.StateDisabled:
		return "LANG_SERVER_STATE_DISABLED"
	default:
		panic("invalid state")
	}
}
func (c *langServerResolver) Pending(ctx context.Context) bool     { return c.pending }
func (c *langServerResolver) Downloading(ctx context.Context) bool { return c.downloading }
func (c *langServerResolver) CanEnable(ctx context.Context) bool   { return c.canEnable }
func (c *langServerResolver) CanDisable(ctx context.Context) bool  { return c.canDisable }
func (c *langServerResolver) CanRestart(ctx context.Context) bool  { return c.canRestart }
func (c *langServerResolver) CanUpdate(ctx context.Context) bool   { return c.canUpdate }
func (c *langServerResolver) Healthy(ctx context.Context) bool     { return c.healthy }

func (s *siteResolver) LangServer(ctx context.Context, args struct{ Language string }) (*langServerResolver, error) {
	langServers, err := s.LangServers(ctx)
	if err != nil {
		return nil, err
	}
	for _, langServer := range langServers {
		if langServer.language == args.Language {
			return langServer, nil
		}
	}
	return nil, nil
}

func (s *siteResolver) LangServers(ctx context.Context) ([]*langServerResolver, error) {
	// Note: This only affects whether or not the client displays
	// enable/disable/restart/update buttons. It does not affect security.
	isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx) == nil

	var results []*langServerResolver
	for _, language := range langservers.Languages {
		state, err := langservers.State(language)
		if err != nil {
			return nil, errors.Wrap(err, "langservers.State")
		}

		var (
			info      *langservers.LangInfo
			infoErr   error
			canManage = langservers.CanManage() == nil
		)
		if canManage {
			info, infoErr = langservers.Info(language)
		}

		results = append(results, &langServerResolver{
			language:            language,
			displayName:         langservers.StaticInfo[language].DisplayName,
			homepageURL:         langservers.StaticInfo[language].HomepageURL,
			issuesURL:           langservers.StaticInfo[language].IssuesURL,
			docsURL:             langservers.StaticInfo[language].DocsURL,
			isClusterDeployment: conf.IsDeployTypeKubernetesCluster(conf.DeployType()),
			custom:              false,
			experimental:        langservers.StaticInfo[language].Experimental,
			state:               state,
			pending:             canManage && infoErr == nil && (info.Pulling || info.Status == langservers.StatusStarting),
			downloading:         canManage && infoErr == nil && info.Pulling,
			canEnable:           canManage && infoErr == nil && (isSiteAdmin || state == langservers.StateNone),
			canDisable:          canManage && infoErr == nil && isSiteAdmin,
			canRestart:          canManage && infoErr == nil && isSiteAdmin && state == langservers.StateEnabled,
			canUpdate:           canManage && infoErr == nil && isSiteAdmin,
			healthy:             canManage && infoErr == nil && (info.Pulling || info.Running()),
		})
	}

	// Also add in custom language servers that were added to the site
	// configuration. These are language servers that do not come with
	// Sourcegraph by default, and we cannot manage them via Docker etc.
	for _, ls := range conf.Get().Langservers {
		_, builtin := langservers.StaticInfo[ls.Language]
		if builtin {
			continue
		}
		state := langservers.StateEnabled
		if ls.Disabled {
			state = langservers.StateDisabled
		}

		result := &langServerResolver{
			language:            strings.ToLower(ls.Language),
			displayName:         strings.Title(ls.Language),
			isClusterDeployment: conf.IsDeployTypeKubernetesCluster(conf.DeployType()),
			custom:              true,
			state:               state,
			canEnable:           isSiteAdmin,
			canDisable:          isSiteAdmin,
			canRestart:          false,
			canUpdate:           false,
			healthy:             false,
		}

		if ls.Metadata != nil {
			result.homepageURL = ls.Metadata.HomepageURL
			result.issuesURL = ls.Metadata.IssuesURL
			result.docsURL = ls.Metadata.DocsURL
			// Experimental language servers can only be added through the site configuration
			result.experimental = ls.Metadata.Experimental
		}

		results = append(results, result)
	}

	return results, nil
}

type langServersResolver struct{}

func (s *schemaResolver) LangServers(ctx context.Context) *langServersResolver {
	return &langServersResolver{}
}

func (c *langServersResolver) Enable(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDeployTypeKubernetesCluster(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.enable) on cluster deployments")
	}

	// For custom (non-builtin) language servers, Enable/Disable is just
	// updating the site config. We do not do anything else. Only admins
	// can perform this action, period.
	info, builtin := langservers.StaticInfo[args.Language]
	if !builtin {
		// 🚨 SECURITY: Only admins can enable/disable custom language servers.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
		// Set disabled=false in the site config.
		if err := langservers.SetDisabled(ctx, args.Language, false); err != nil {
			return nil, errors.Wrap(err, "langservers.SetDisabled")
		}
		return &EmptyResponse{}, nil
	}

	if info.Experimental {
		// 🚨 SECURITY: Only admins can enable/disable experimental language servers.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
	}

	state, err := langservers.State(args.Language)
	if err != nil {
		return nil, errors.Wrap(err, "langservers.State")
	}
	if state == langservers.StateEnabled {
		// Code intelligence is already enabled for this language.
		return &EmptyResponse{}, nil
	}
	if state == langservers.StateDisabled {
		// 🚨 SECURITY: Code intelligence for this language was explicitly disabled
		// by an admin. Only admin users can re-enable it.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
	}

	// At this point, code intelligence is not enabled, and either we are an
	// admin OR it was not explicitly disabled by an admin and any user can
	// enable it.

	// Set disabled=false in the site config.
	if err := langservers.SetDisabled(ctx, args.Language, false); err != nil {
		return nil, errors.Wrap(err, "langservers.SetDisabled")
	}

	// Wait for the new configuration to be respected. Usually this is not required,
	// but in the case of enabling a language server here if we returned without
	// waiting for the new configuration to be respected the language server container
	// would be in a langserver.StateNone status (i.e. no container exists) until the
	// config change was detected -- which is rightfully displayed in the UI as
	// 'Unhealthy'. We prefer to avoid that when starting language servers.
	bg.RespectLangServersConfigUpdate()
	return &EmptyResponse{}, nil
}

func (c *langServersResolver) Disable(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDeployTypeKubernetesCluster(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.disable) on cluster deployments")
	}

	// Note: For custom language servers, we do not need to do anything special
	// since unlike the enable action, only admins can disable language servers
	// regardless of whether or not they are custom.

	// 🚨 SECURITY: Only admins may disable language servers.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// Set disabled=true in the site config.
	if err := langservers.SetDisabled(ctx, args.Language, true); err != nil {
		return nil, errors.Wrap(err, "langservers.SetDisabled")
	}
	return &EmptyResponse{}, nil
}

func (c *langServersResolver) Restart(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDeployTypeKubernetesCluster(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.restart) on cluster deployments")
	}
	_, builtin := langservers.StaticInfo[args.Language]
	if !builtin {
		return nil, errors.New("cannot use this API (langServers.restart) on custom language servers")
	}

	// 🚨 SECURITY: Only admins may restart language servers.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// Restart language server now.
	if err := langservers.Restart(args.Language); err != nil {
		return nil, errors.Wrap(err, "langservers.Restart")
	}
	return &EmptyResponse{}, nil
}

func (c *langServersResolver) Update(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDeployTypeKubernetesCluster(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.update) on cluster deployments")
	}
	_, builtin := langservers.StaticInfo[args.Language]
	if !builtin {
		return nil, errors.New("cannot use this API (langServers.update) on custom language servers")
	}

	// 🚨 SECURITY: Only admins may update language servers.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// Update language server now.
	if err := langservers.Update(args.Language); err != nil {
		return nil, errors.Wrap(err, "langservers.Update")
	}
	return &EmptyResponse{}, nil
}

// nullString returns nil if s == "", otherwise it returns a pointer to s.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *siteResolver) LanguageServerManagementStatus(ctx context.Context) (*languageServerManagementStatusResolver, error) {
	// 🚨 SECURITY: Only admins may see this information because it's unnecessary for other users to
	// see it, and the reason may contain sensitive data.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &languageServerManagementStatusResolver{}, nil
}

type languageServerManagementStatusResolver struct{}

func (r *languageServerManagementStatusResolver) SiteCanManage() bool {
	err := langservers.CanManage()
	return err == nil
}

func (r *languageServerManagementStatusResolver) Reason(ctx context.Context) (*string, error) {
	// 🚨 SECURITY: Only admins may see this information because it's unnecessary for other users to
	// see it, and the reason may contain sensitive data.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	err := langservers.CanManage()
	if err == nil {
		return nil, nil
	}
	return nullString(err.Error()), nil
}
