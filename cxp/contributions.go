package cxp

// Contributions describes the functionality provided by an extension.
//
// See cxp-js for canonical documentation.
type Contributions struct {
	Commands []*CommandContribution `json:"commands,omitempty"`
	Menus    *MenuContributions     `json:"menus,omitempty"`
}

// CommandContribution is a command provided by the extension that can be invoked.
//
// See cxp-js for canonical documentation.
type CommandContribution struct {
	Command     string                          `json:"command"`
	Title       string                          `json:"title,omitempty"`
	Category    string                          `json:"category,omitempty"`
	Description string                          `json:"description,omitempty"`
	IconURL     string                          `json:"iconURL,omitempty"`
	ToolbarItem *CommandContributionToolbarItem `json:"toolbarItem,omitempty"`
}

type CommandContributionToolbarItem struct {
	Label           string `json:"label,omitempty"`
	Description     string `json:"description,omitempty"`
	Group           string `json:"group,omitempty"`
	IconURL         string `json:"iconURL,omitempty"`
	IconDescription string `json:"iconDescription,omitempty"`
}

// MenuContributions describes the menu items contributed by an extension.
//
// See cxp-js for canonical documentation.
type MenuContributions struct {
	CommandPalette []*MenuItemContribution `json:"commandPalette,omitempty"`
	EditorTitle    []*MenuItemContribution `json:"editor/title,omitempty"`
	GlobalNav      []*MenuItemContribution `json:"global/nav,omitempty"`
	DirectoryPage  []*MenuItemContribution `json:"directory/page,omitempty"`
	Help           []*MenuItemContribution `json:"help,omitempty"`
}

// MenuItemContribution is a menu item contributed by an extension.
//
// See cxp-js for canonical documentation.
type MenuItemContribution struct {
	Command string `json:"command"`
}
