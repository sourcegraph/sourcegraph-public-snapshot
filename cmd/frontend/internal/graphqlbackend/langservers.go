package graphqlbackend

import (
	"context"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/langservers"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type langServerResolver struct {
	language                                     string
	homepageURL, issuesURL, docsURL              string
	dataCenter                                   bool
	enabled                                      bool
	pending                                      bool
	canEnable, canDisable, canRestart, canUpdate bool
	healthy                                      bool
}

func (c *langServerResolver) Language(ctx context.Context) string    { return c.language }
func (c *langServerResolver) HomepageURL(ctx context.Context) string { return c.homepageURL }
func (c *langServerResolver) IssuesURL(ctx context.Context) string   { return c.issuesURL }
func (c *langServerResolver) DocsURL(ctx context.Context) string     { return c.docsURL }
func (c *langServerResolver) DataCenter(ctx context.Context) bool    { return c.dataCenter }
func (c *langServerResolver) Enabled(ctx context.Context) bool       { return c.enabled }
func (c *langServerResolver) Pending(ctx context.Context) bool       { return c.pending }
func (c *langServerResolver) CanEnable(ctx context.Context) bool     { return c.canEnable }
func (c *langServerResolver) CanDisable(ctx context.Context) bool    { return c.canDisable }
func (c *langServerResolver) CanRestart(ctx context.Context) bool    { return c.canRestart }
func (c *langServerResolver) CanUpdate(ctx context.Context) bool     { return c.canUpdate }
func (c *langServerResolver) Healthy(ctx context.Context) bool       { return c.healthy }

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

		if conf.IsDataCenter(conf.DeployType()) {
			// Running in Data Center. We cannot execute Docker commands, so we
			// have less information.
			results = append(results, &langServerResolver{
				language:    language,
				homepageURL: langservers.URLs[language].Homepage,
				issuesURL:   langservers.URLs[language].Issues,
				docsURL:     langservers.URLs[language].Docs,
				dataCenter:  true,
				enabled:     state == langservers.StateEnabled,
				pending:     false,
				canEnable:   false,
				canDisable:  false,
				canRestart:  false,
				canUpdate:   false,
				healthy:     false,
			})
			continue
		}

		info, err := langservers.Info(language)
		if err != nil {
			return nil, errors.Wrap(err, "langservers.Info")
		}

		results = append(results, &langServerResolver{
			language:    language,
			homepageURL: langservers.URLs[language].Homepage,
			issuesURL:   langservers.URLs[language].Issues,
			docsURL:     langservers.URLs[language].Docs,
			dataCenter:  false,
			enabled:     state == langservers.StateEnabled,
			pending:     info.Pulling || info.Status == langservers.StatusStarting,
			canEnable:   isSiteAdmin || state == langservers.StateNone,
			canDisable:  isSiteAdmin,
			canRestart:  isSiteAdmin,
			canUpdate:   isSiteAdmin,
			healthy:     info.Pulling || info.Status != langservers.StatusUnhealthy,
		})
	}
	return results, nil
}

type langServersResolver struct{}

func (s *schemaResolver) LangServers(ctx context.Context) *langServersResolver {
	return &langServersResolver{}
}

func (c *langServersResolver) Enable(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDataCenter(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.enable) in Data Center mode")
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
	if err := langservers.SetDisabled(args.Language, false); err != nil {
		return nil, errors.Wrap(err, "langservers.SetDisabled")
	}
	return &EmptyResponse{}, nil
}

func (c *langServersResolver) Disable(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDataCenter(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.disable) in Data Center mode")
	}

	// 🚨 SECURITY: Only admins may disable language servers.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// Set disabled=true in the site config.
	if err := langservers.SetDisabled(args.Language, true); err != nil {
		return nil, errors.Wrap(err, "langservers.SetDisabled")
	}
	return &EmptyResponse{}, nil
}

func (c *langServersResolver) Restart(ctx context.Context, args *struct{ Language string }) (*EmptyResponse, error) {
	if conf.IsDataCenter(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.restart) in Data Center mode")
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
	if conf.IsDataCenter(conf.DeployType()) {
		return nil, errors.New("cannot use this API (langServers.update) in Data Center mode")
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
