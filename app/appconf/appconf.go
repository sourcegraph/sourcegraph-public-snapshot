// Package appconf holds app configuration. The global config in this
// package is consulted by many app handlers, and if running in CLI
// mode, the config's values are set based on the CLI flags provided.
//
// This package is separate from package app to avoid import cycles
// when internal subpackages import it.
package appconf

import (
	"html/template"
	"time"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Flags configure the app. The values are set by CLI flags (or during testing).
var Flags struct {
	NoUIBuild bool `long:"app.no-ui-build" description:"disable manual building of repositories from the UI"`

	// DisableTreeEntryCommits is sometimes necessary for extremely
	// large Git repositories when `git log -- PATH` takes a very long
	// time.
	DisableTreeEntryCommits bool `long:"app.disable-tree-entry-commits" description:"do not show the latest commit for tree files/dirs"`

	RepoBadgesAndCounters bool `long:"app.repo-badges-counters" description:"enable repo badges and counters"`

	DisableRepoTreeSearch bool `long:"app.disable-repo-tree-search" description:"do not show repo fulltext search results (only defs) (slower for large repos)"`

	DisableGlobalSearch bool `long:"app.disable-global-search" description:"if set, only allow searching within a single repository at a time"`

	DisableSearch bool `long:"app.disable-search" description:"if set, search will be entirely disabled / never allowed"`

	DisableApps bool `long:"app.disable-apps" description:"if set, disable the changes and issues applications"`

	DisableIntegrations bool `long:"app.disable-integrations" description:"disable integrations with third-party services that are accessible from the user settings page"`

	DisableCloneURL bool `long:"app.disable-clone-url" description:"if set, disable display of the git clone URL"`

	DisableUserContent bool `long:"app.disable-user-content" description:"if set, disable ability to upload user content (e.g., pasting images into textareas)"`

	EnableGitHubRepoShortURIAliases bool `long:"app.enable-github-repo-short-uri-aliases" description:"if set, redirect 'user/repo' URLs (with no 'github.com/') to '/github.com/user/repo'"`

	EnableGitHubStyleUserPaths bool `long:"app.enable-github-style-user-paths" description:"redirect GitHub paths like '/user' to valid ones like '/~user' (disables single-path repos)"`

	CustomLogo template.HTML `long:"app.custom-logo" description:"custom logo to display in the top nav bar (HTML)"`

	CustomNavLayout template.HTML `long:"app.custom-nav-layout" description:"custom layout to display in place of the search form (HTML)"`

	// MOTD is a message of the day that is shown in a ribbon at the
	// top of the page. It can be hidden on a per-response basis by
	// setting the tmpl.Common HideMOTD field to true.
	MOTD template.HTML `long:"app.motd" description:"show a custom message to all users beneath the top nav bar (HTML)" env:"SG_NAV_MSG"`

	GoogleAnalyticsTrackingID string `long:"app.google-analytics-tracking-id" description:"Google Analytics tracking ID (UA-########-#)" env:"GOOGLE_ANALYTICS_TRACKING_ID"`

	HeapAnalyticsID string `long:"app.heap-analytics-id" description:"Heap Analytics ID" env:"HEAP_ANALYTICS_ID"`

	CustomFeedbackForm template.HTML `long:"app.custom-feedback-form" description:"custom feedback form to display (HTML)" env:"CUSTOM_FEEDBACK_FORM"`

	CheckForUpdates time.Duration `long:"app.check-for-updates" description:"rate at which to check for updates and display a notification (not download/install) (0 to disable)" default:"30m"`

	Blog bool `long:"app.blog" description:"Enable the Sourcegraph blog, must also set $SG_TUMBLR_API_KEY"`

	DisableExternalLinks bool `long:"app.disable-external-links" description:"Disable links to external websites"`

	ReloadAssets bool `long:"reload" description:"(development mode only) reload app templates and other assets on each request"`

	ExtraHeadHTML template.HTML `long:"app.extra-head-html" description:"extra HTML (<script> tags, etc.) to insert before the </head> tag"`
	ExtraBodyHTML template.HTML `long:"app.extra-body-html" description:"extra HTML (<script> tags, etc.) to insert before the </body> tag"`

	MirrorRepoUpdateRate              time.Duration `long:"app.mirror-repo-update-rate" description:"rate at which to update mirrored repositories" default:"3s"`
	DisableMirrorRepoBackgroundUpdate bool          `long:"app.disable-mirror-repo-bg-update" description:"disable updating mirrored repos in the background"`

	DisableGitNotify bool `long:"app.disable-git-notify" description:"disable git notifications"`
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("App", "App flags", &Flags)
	})
}
