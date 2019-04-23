package main

import (
	"errors"
	"fmt"
)

const configurationSubjectFragment = `
fragment ConfigurationSubjectFields on ConfigurationSubject {
    id
    latestSettings {
        id
        configuration {
            ...ConfigurationFields
        }
        author {
            ...UserFields
        }
        createdAt
    }
    settingsURL
    viewerCanAdminister
}
`

type ConfigurationSubject struct {
	ID                   string
	LatestSettings       *Settings
	SettingsURL          string
	ViewerCanAdminister  bool
	ConfigurationCascade ConfigurationCascade
}

type Settings struct {
	ID            int32
	Configuration Configuration
	Author        *User
	CreatedAt     string
}

const configurationCascadeFragment = `
fragment ConfigurationCascadeFields on ConfigurationCascade {
    subjects {
        ...ConfigurationSubjectFields
    }
    merged {
        ...ConfigurationFields
    }
}
`

type ConfigurationCascade struct {
	Subjects []ConfigurationSubject
	Merged   Configuration
}

const configurationFragment = `
fragment ConfigurationFields on Configuration {
    contents
}
`

type Configuration struct {
	Contents string
}

const viewerConfigurationQuery = `query ViewerConfiguration {
  viewerConfiguration {
    ...ConfigurationCascadeFields
  }
}` + configurationCascadeFragment + configurationSubjectFragment + configurationFragment + userFragment

const configurationSubjectCascadeQuery = `query ConfigurationSubjectCascade($subject: ID!) {
  configurationSubject(id: $subject) {
    configurationCascade {
      ...ConfigurationCascadeFields
    }
  }
}` + configurationCascadeFragment + configurationSubjectFragment + configurationFragment + userFragment

type KeyPath struct {
	Property string `json:"property,omitempty"`
	Index    int    `json:"index,omitempty"`
}

func getViewerUserID() (string, error) {
	var result struct {
		CurrentUser *struct{ ID string }
	}
	req := &apiRequest{
		query: `
query ViewerUserID {
  currentUser {
    id
  }
}
`,
		result: &result,
	}
	if err := req.do(); err != nil {
		return "", err
	}
	if result.CurrentUser == nil || result.CurrentUser.ID == "" {
		return "", errors.New("unable to determine current user ID (see https://github.com/sourcegraph/src-cli#authentication)")
	}
	return result.CurrentUser.ID, nil
}

func getConfigurationSubjectLatestSettingsID(subjectID string) (*int, error) {
	var result struct {
		ConfigurationSubject *struct {
			LatestSettings *struct {
				ID int
			}
		}
	}
	req := &apiRequest{
		query: `
query ConfigurationSubjectLatestSettingsID($subject: ID!) {
  configurationSubject(id: $subject) {
    latestSettings {
      id
    }
  }
}
`,
		vars:   map[string]interface{}{"subject": subjectID},
		result: &result,
	}
	if err := req.do(); err != nil {
		return nil, err
	}
	if result.ConfigurationSubject == nil {
		return nil, fmt.Errorf("unable to find configuration subject with ID %s", subjectID)
	}
	if result.ConfigurationSubject.LatestSettings == nil {
		return nil, nil
	}
	return &result.ConfigurationSubject.LatestSettings.ID, nil
}
