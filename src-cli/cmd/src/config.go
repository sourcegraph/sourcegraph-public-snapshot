package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
)

var configCommands commander

func init() {
	usage := `'src config' is a tool that manages global, organization, and user settings on a Sourcegraph instance.

The effective setting is computed by shallow-merging the following settings, in order from lowest to highest precedence:

- Global settings (site-wide)
- Organization settings for the user's organizations (if any)
- User settings
- Client settings (when using a Sourcegraph browser or editor extension)

For unauthenticated requests, the organization and user settings are empty.

Usage:

	src config command [command options]

The commands are:

	get       gets the effective (merged) settings
	edit      updates settings
	list      lists the partial settings (that, when merged, yield the effective settings)

Use "src config [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("config", flag.ExitOnError)
	handler := func(args []string) error {
		configCommands.run(flagSet, "src config", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

const settingsSubjectFragment = `
fragment SettingsSubjectFields on SettingsSubject {
    id
    latestSettings {
        id
        contents
        author {
            ...UserFields
        }
        createdAt
    }
    settingsURL
    viewerCanAdminister
}
`

type SettingsSubject struct {
	ID                  string
	LatestSettings      *Settings
	SettingsURL         string
	ViewerCanAdminister bool
	SettingsCascade     SettingsCascade
}

type Settings struct {
	ID        int32
	Contents  string
	Author    *User
	CreatedAt string
}

const settingsCascadeFragment = `
fragment SettingsCascadeFields on SettingsCascade {
    subjects {
        ...SettingsSubjectFields
    }
    final
}
`

type SettingsCascade struct {
	Subjects []SettingsSubject
	Final    string
}

const viewerSettingsQuery = `query ViewerSettings {
  viewerSettings {
    ...SettingsCascadeFields
  }
}` + settingsCascadeFragment + settingsSubjectFragment + userFragment

const settingsSubjectCascadeQuery = `query SettingsSubjectCascade($subject: ID!) {
  settingsSubject(id: $subject) {
    settingsCascade {
      ...SettingsCascadeFields
    }
  }
}` + settingsCascadeFragment + settingsSubjectFragment + userFragment

type KeyPath struct {
	Property string `json:"property,omitempty"`
	Index    int    `json:"index,omitempty"`
}

func getViewerUserID(ctx context.Context, client api.Client) (string, error) {
	query := `
query ViewerUserID {
  currentUser {
    id
  }
}
`

	var result struct {
		CurrentUser *struct{ ID string }
	}

	if _, err := client.NewQuery(query).Do(ctx, &result); err != nil {
		return "", err
	}

	if result.CurrentUser == nil || result.CurrentUser.ID == "" {
		return "", errors.New("unable to determine current user ID (see https://github.com/sourcegraph/src-cli#authentication)")
	}
	return result.CurrentUser.ID, nil
}

func getSettingsSubjectLatestSettingsID(ctx context.Context, client api.Client, subjectID string) (*int, error) {
	query := `
query SettingsSubjectLatestSettingsID($subject: ID!) {
  settingsSubject(id: $subject) {
    latestSettings {
      id
    }
  }
}
`

	var result struct {
		SettingsSubject *struct {
			LatestSettings *struct {
				ID int
			}
		}
	}

	if _, err := client.NewRequest(query, map[string]interface{}{
		"subject": subjectID,
	}).Do(ctx, &result); err != nil {
		return nil, err
	}

	if result.SettingsSubject == nil {
		return nil, errors.Newf("unable to find settings subject with ID %s", subjectID)
	}
	if result.SettingsSubject.LatestSettings == nil {
		return nil, nil
	}
	return &result.SettingsSubject.LatestSettings.ID, nil
}
