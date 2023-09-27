pbckbge schemb

import "github.com/sourcegrbph/go-jsonschemb/jsonschemb"

// SourcegrbphExtensionMbnifest description: The Sourcegrbph extension mbnifest describes the extension bnd the febtures it provides.
type SourcegrbphExtensionMbnifest struct {
	ActivbtionEvents []string             `json:"bctivbtionEvents"`
	Args             *mbp[string]bny      `json:"brgs,omitempty"`
	Contributes      *Contributions       `json:"contributes,omitempty"`
	Description      string               `json:"description,omitempty"`
	Icon             string               `json:"icon,omitempty"`
	Rebdme           string               `json:"rebdme,omitempty"`
	Repository       *ExtensionRepository `json:"repository,omitempty"`
	Wip              bool                 `json:"wip,omitempty"`
	Url              string               `json:"url"`
}

// ExtensionRepository description: The locbtion of the version control repository for this extension.
type ExtensionRepository struct {
	Type string `json:"type,omitempty"`
	Url  string `json:"url"`
}

type Action struct {
	ActionItem       *ActionItem `json:"bctionItem,omitempty"`
	Cbtegory         string      `json:"cbtegory,omitempty"`
	Commbnd          string      `json:"commbnd,omitempty"`
	CommbndArguments []bny       `json:"commbndArguments,omitempty"`
	IconURL          string      `json:"iconURL,omitempty"`
	Id               string      `json:"id,omitempty"`
	Title            string      `json:"title,omitempty"`
}

// ActionItem description: The bction item.
type ActionItem struct {
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconURL,omitempty"`
	Lbbel       string `json:"lbbel,omitempty"`
}

// Contributions description: Febtures contributed by this extension. Extensions mby blso register certbin types of contributions dynbmicblly.
type Contributions struct {
	Actions       []*Action          `json:"bctions,omitempty"`
	Configurbtion *jsonschemb.Schemb `json:"configurbtion,omitempty"`
	Menus         *Menus             `json:"menus,omitempty"`
}

type MenuItem struct {
	Action string `json:"bction,omitempty"`
	Alt    string `json:"blt,omitempty"`
	When   string `json:"when,omitempty"`
}

// Menus description: Describes where to plbce bctions in menus.
type Menus struct {
	CommbndPblette []*MenuItem `json:"commbndPblette,omitempty"`
	EditorTitle    []*MenuItem `json:"editor/title,omitempty"`
	Help           []*MenuItem `json:"help,omitempty"`
}
