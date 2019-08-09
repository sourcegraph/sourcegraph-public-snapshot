package schema

import "github.com/sourcegraph/go-jsonschema/jsonschema"

// TODO: This file is manually updated and must remain in sync with extension.schema.json. It does
// not need to contain all fields, only those used by Go code.

// SourcegraphExtensionManifest description: The Sourcegraph extension manifest describes the extension and the features it provides.
type SourcegraphExtensionManifest struct {
	ActivationEvents []string                `json:"activationEvents"`
	Args             *map[string]interface{} `json:"args,omitempty"`
	Contributes      *Contributions          `json:"contributes,omitempty"`
	Description      string                  `json:"description,omitempty"`
	Icon             string                  `json:"icon,omitempty"`
	Readme           string                  `json:"readme,omitempty"`
	Repository       *ExtensionRepository    `json:"repository,omitempty"`
	Wip              bool                    `json:"wip,omitempty"`
	Url              string                  `json:"url"`
}

// ExtensionRepository description: The location of the version control repository for this extension.
type ExtensionRepository struct {
	Type string `json:"type,omitempty"`
	Url  string `json:"url"`
}

type Action struct {
	ActionItem       *ActionItem   `json:"actionItem,omitempty"`
	Category         string        `json:"category,omitempty"`
	Command          string        `json:"command,omitempty"`
	CommandArguments []interface{} `json:"commandArguments,omitempty"`
	IconURL          string        `json:"iconURL,omitempty"`
	Id               string        `json:"id,omitempty"`
	Title            string        `json:"title,omitempty"`
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_978(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
