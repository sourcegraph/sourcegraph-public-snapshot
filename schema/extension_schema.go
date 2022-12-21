package schema

import "github.com/sourcegraph/go-jsonschema/jsonschema"

// SourcegraphExtensionManifest description: The Sourcegraph extension manifest describes the extension and the features it provides.
type SourcegraphExtensionManifest struct {
	ActivationEvents []string             `json:"activationEvents"`
	Args             *map[string]any      `json:"args,omitempty"`
	Contributes      *Contributions       `json:"contributes,omitempty"`
	Description      string               `json:"description,omitempty"`
	Icon             string               `json:"icon,omitempty"`
	Readme           string               `json:"readme,omitempty"`
	Repository       *ExtensionRepository `json:"repository,omitempty"`
	Wip              bool                 `json:"wip,omitempty"`
	Url              string               `json:"url"`
}

// ExtensionRepository description: The location of the version control repository for this extension.
type ExtensionRepository struct {
	Type string `json:"type,omitempty"`
	Url  string `json:"url"`
}

type Action struct {
	ActionItem       *ActionItem `json:"actionItem,omitempty"`
	Category         string      `json:"category,omitempty"`
	Command          string      `json:"command,omitempty"`
	CommandArguments []any       `json:"commandArguments,omitempty"`
	IconURL          string      `json:"iconURL,omitempty"`
	Id               string      `json:"id,omitempty"`
	Title            string      `json:"title,omitempty"`
}

// ActionItem description: The action item.
type ActionItem struct {
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconURL,omitempty"`
	Label       string `json:"label,omitempty"`
}

// Contributions description: Features contributed by this extension. Extensions may also register certain types of contributions dynamically.
type Contributions struct {
	Actions       []*Action          `json:"actions,omitempty"`
	Configuration *jsonschema.Schema `json:"configuration,omitempty"`
	Menus         *Menus             `json:"menus,omitempty"`
}

type MenuItem struct {
	Action string `json:"action,omitempty"`
	Alt    string `json:"alt,omitempty"`
	When   string `json:"when,omitempty"`
}

// Menus description: Describes where to place actions in menus.
type Menus struct {
	CommandPalette []*MenuItem `json:"commandPalette,omitempty"`
	EditorTitle    []*MenuItem `json:"editor/title,omitempty"`
	Help           []*MenuItem `json:"help,omitempty"`
}
