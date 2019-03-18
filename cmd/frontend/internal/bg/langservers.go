package bg

import (
	"context"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/langservers"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// KeepLangServersAndGlobalSettingsInSync watches for configuration changes and
// updates global settings to reflect the enablement state of languages in the
// `langservers` array.
func KeepLangServersAndGlobalSettingsInSync(ctx context.Context) {
	conf.Watch(func() {
		config := conf.Get()
		if config == nil {
			return
		}

		// Don't bother getting the existing settings and a user ID if there are
		// no edits to make.
		if len(config.Langservers) == 0 {
			return
		}

		settings, err := db.Settings.GetLatest(context.Background(), api.ConfigurationSubject{Site: true})
		if err != nil {
			log15.Warn("error getting existing global settings", "error", err)
			return
		}
		var contents string
		var id *int32
		var authorUserID int32
		if settings == nil {
			contents = "{}"
			// HACK: Global settings are nil (probably because this is a brand new
			// instance), so there's no existing author user ID to reuse. Just take
			// any user's ID. The author isn't shown anywhere, anyway.
			users, err := db.Users.List(ctx, &db.UsersListOptions{
				LimitOffset: &db.LimitOffset{Limit: 1},
			})
			if err != nil {
				log15.Warn("error listing users in order to enable/disable a language extension", "error", err)
				return
			}
			if len(users) == 0 {
				log15.Warn("unable to obtain a user ID to enable/disable a language extension because there are no users in the database")
				return
			}
			authorUserID = users[0].ID
		} else {
			contents = settings.Contents
			authorUserID = settings.AuthorUserID
			id = &settings.ID
		}

		for _, language := range config.Langservers {
			edits, _, err := jsonx.ComputePropertyEdit(contents, jsonx.PropertyPath("extensions", "langserver/"+language.Language), !language.Disabled, nil, conf.FormatOptions)
			if err != nil {
				log15.Warn("error updating global settings to enable/disable a language extension", "error", err)
				return
			}

			contents, err = jsonx.ApplyEdits(contents, edits...)
			if err != nil {
				log15.Warn("error applying edits to global settings to enable/disable a language extension", "error", err)
				return
			}
		}

		_, err = db.Settings.CreateIfUpToDate(context.Background(), api.ConfigurationSubject{Site: true}, id, authorUserID, contents)
		if err != nil {
			log15.Warn("error updating global settings to enable/disable a language extension", "error", err)
			return
		}
	})
}

// RespectLangServersConfigUpdate is invoked inside of conf.Watch, but also
// sometimes manually when the caller needs to block untill the latest config
// has been respected.
func RespectLangServersConfigUpdate() {
	for _, language := range langservers.Languages {
		// Start language servers that were previously enabled.
		state, err := langservers.State(language)
		if err != nil {
			log15.Error("failed to get language server state", "language", language, "error", err)
			continue
		}
		if state == langservers.StateEnabled {
			// Start language server now.
			if err := langservers.Start(language); err != nil {
				log15.Error("failed to start language server", "language", language, "error", err)
			}
			continue
		}
		// Stop the language server if it is running, as it is not enabled
		// (StateDisabled or StateNone).
		_ = langservers.Stop(language)
	}
}

// StartLangServers should be invoked on startup, after DB initialization, in
// order to start up language servers, etc.
func StartLangServers(ctx context.Context) {
	if err := langservers.CanManage(); err != nil {
		return
	}

	startup := true
	conf.Watch(func() {
		defer func() {
			startup = false
		}()
		RespectLangServersConfigUpdate()

		// Do not run the below code to reflect docker state in the config,
		// except on server startup. Otherwise, we could introduce an
		// infinite loop due to langservers.SetDisabled writing to the
		// config and inherently firing conf.Watch again.
		if !startup {
			return
		}
		for _, language := range langservers.Languages {
			// We didn't start/stop the language server. If it is currently
			// running, this indicates that a server admin did so manually e.g. via
			// `docker run`. It is important that we mark this language as enabled
			// in the site config or else it would show up as "disabled" in the
			// admin UI and we would stop it on server shutdown and never start it
			// again.
			info, err := langservers.Info(language)
			if err != nil {
				log15.Error("failed to get language server info", "language", language, "error", err)
				continue
			}
			if !info.Running() {
				// No container for this language running.
				continue
			}

			// Set disabled=false in the site config.
			if err := langservers.SetDisabled(ctx, language, false); err != nil {
				log15.Error("failed to mark running language server as enabled", "language", language, "error", err)
				continue
			}
		}
	})
}
