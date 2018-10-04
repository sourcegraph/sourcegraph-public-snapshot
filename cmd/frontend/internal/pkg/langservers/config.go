package langservers

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/pkg/api"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ConfigState represents the current state of a language server in the
// configuration.
type ConfigState int

const (
	// StateNone represents that code intelligence for a given language is
	// neither disabled nor enabled. If the language server is not experimental,
	// plain users can enable it when in this state. Only admins can enable
	// experimental language servers.
	StateNone ConfigState = iota

	// StateEnabled represents that code intelligence for a given language
	// is enabled by a plain user or admin.
	StateEnabled ConfigState = iota

	// StateDisabled represents that code intelligence for a given language
	// is disabled by an admin. It cannot be enabled by a plain user when in
	// this state, but rather only by an admin.
	StateDisabled ConfigState = iota
)

// State gets the current state for the given language.
func State(language string) (ConfigState, error) {
	// Check if the language is supported.
	if err := checkSupported(language); err != nil {
		return StateNone, err
	}

	for _, langserver := range conf.Get().Langservers {
		if language == strings.ToLower(langserver.Language) {
			if langserver.Disabled {
				return StateDisabled, nil
			}
			return StateEnabled, nil
		}
	}
	return StateNone, nil
}

func setDisabledInGlobalSettings(language string, disabled bool) error {
	settings, err := db.Settings.GetLatest(context.Background(), api.ConfigurationSubject{Site: true})
	if err != nil {
		return err
	}
	if settings == nil {
		return nil
	}
	edits, _, err := jsonx.ComputePropertyEdit(settings.Contents, jsonx.PropertyPath("extensions", "langserver/"+language), !disabled, nil, conf.FormatOptions)
	if err != nil {
		return err
	}
	newSettings, err := jsonx.ApplyEdits(settings.Contents, edits...)
	if err != nil {
		return err
	}
	_, err = db.Settings.CreateIfUpToDate(context.Background(), api.ConfigurationSubject{Site: true}, &settings.ID, settings.AuthorUserID, newSettings)
	if err != nil {
		return err
	}
	return nil
}

// SetDisabled sets the state of the language server for the specified language.
//
// This is done by updating the site configuration, and as such should never be
// invoked in response to a conf.Watch callback, etc.
//
// This also enables/disables the corresponding Sourcegraph Extension in global
// settings (not site configuration).
func SetDisabled(language string, disabled bool) error {
	// Check if the language specified is for a custom language server or not.
	customLangserver := checkSupported(language) != nil

	err := setDisabledInGlobalSettings(language, disabled)
	if err != nil {
		return errors.Wrap(err, "setDisabledInGlobalSettings")
	}

	return conf.Edit(func(current *schema.SiteConfiguration, raw string) ([]jsonx.Edit, error) {
		// Copy the langservers slice, since we intend to edit it.
		newLangservers := make([]*schema.Langservers, 0, len(current.Langservers))

		foundExisting := false
		for _, existing := range current.Langservers {
			if language == strings.ToLower(existing.Language) {
				// Already exists, so we should only update the Disabled field.
				existing.Disabled = disabled
				foundExisting = true
			}
			newLangservers = append(newLangservers, existing)
		}
		if !foundExisting {
			// Doesn't already exist, so add a new entry.
			var newLangserver *schema.Langservers
			if !customLangserver {
				newLangserver = &StaticInfo[language].SiteConfig
			} else {
				// best effort
				newLangserver = &schema.Langservers{Language: language}
			}
			newLangserver.Disabled = disabled
			newLangservers = append(newLangservers, newLangserver)
		}

		// Replace the langservers property with our new list.
		edits, _, err := jsonx.ComputePropertyEdit(
			raw,
			jsonx.PropertyPath("langservers"),
			newLangservers,
			nil,
			conf.FormatOptions,
		)
		return edits, err
	})
}
