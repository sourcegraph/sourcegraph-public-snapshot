package cxp

import "github.com/sourcegraph/jsonx"

// Contributions describes the functionality provided by an extension.
type Contributions struct {
	Commands []*CommandContribution `json:"commands,omitempty"`
	Menus    *MenuContributions     `json:"menus,omitempty"`
}

// CommandContribution is a command provided by the extension that can be invoked.
type CommandContribution struct {
	// Command is an identifier for the command that is assumed to be unique. If another command
	// with the same identifier is defined (by this extension or another extension), the behavior is
	// undefined. To avoid collisions, the identifier conventionally is prefixed with
	// "${EXTENSION_NAME}.".
	Command string `json:"command"`

	Title string `json:"title,omitempty"` // a descriptive title

	IconURL string `json:"iconURL,omitempty"` // URL to an icon (base64: URIs are OK)

	// TODO(extensions): Because the CXP connection is (usually) stateless, commands can't modify
	// state. The second best option is for them to modify user settings. So, require commands to
	// define how they do so.
	ExperimentalSettingsAction *CommandContributionSettingsAction `json:"experimentalSettingsAction"`
}

// CommandContributionSettingsAction is the special action executed by a contributed command that
// modifies settings.
type CommandContributionSettingsAction struct {
	Path        jsonx.Path    `json:"path"`        // the key path to the value
	CycleValues []interface{} `json:"cycleValues"` // the values of the setting to cycle among
}

// MenuContributions describes the menu items contributed by an extension.
type MenuContributions struct {
	EditorTitle []*MenuItemContribution `json:"editor/title,omitempty"`
}

// MenuItemContribution is a menu item contributed by an extension.
type MenuItemContribution struct {
	Command string `json:"command"` // the command to execute when selected (== (CommandContribution).Command)
}
