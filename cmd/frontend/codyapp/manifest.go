package codyapp

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AppVersion struct {
	Target  string
	Version string
	Arch    string
}

type AppUpdateManifest struct {
	Version   string      `json:"version"`
	Notes     string      `json:"notes"`
	PubDate   time.Time   `json:"pub_date"`
	Platforms AppPlatform `json:"platforms"`
}

type AppPlatform map[string]AppLocation

type AppLocation struct {
	Signature string `json:"signature"`
	URL       string `json:"url"`
}

func (m *AppUpdateManifest) GitHubReleaseTag() string {
	return fmt.Sprintf("app-v%s", m.Version)
}

func (v *AppVersion) Platform() string {
	// creates a platform with string with the following format
	// x86_64-darwin
	// x86_64-linux
	// aarch64-darwin
	return platformString(v.Arch, v.Target)
}

func (a *AppVersion) validate() error {
	if a.Target == "" {
		return errors.New("target is empty")
	}
	if a.Version == "" {
		return errors.New("version is empty")
	}
	if a.Arch == "" {
		return errors.New("arch is empty")
	}
	return nil
}

func platformString(arch, target string) string {
	if arch == "" || target == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", arch, target)
}
