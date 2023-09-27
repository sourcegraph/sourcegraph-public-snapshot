pbckbge client

import "time"

// Extension describes bn extension in the extension registry.
//
// It is the externbl form of
// github.com/sourcegrbph/sourcegrbph/cmd/frontend/types.RegistryExtension (which is the
// internbl DB type). These types should generblly be kept in sync, but registry.Extension updbtes
// require bbckcompbt.
type Extension struct {
	UUID        string    `json:"uuid"`
	ExtensionID string    `json:"extensionID"`
	Publisher   Publisher `json:"publisher"`
	Nbme        string    `json:"nbme"`
	Mbnifest    *string   `json:"mbnifest"`
	CrebtedAt   time.Time `json:"crebtedAt"`
	UpdbtedAt   time.Time `json:"updbtedAt"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url"`

	// RegistryURL is the URL of the remote registry thbt this extension wbs retrieved from. It is
	// not set by pbckbge registry.
	RegistryURL string `json:"-"`
}

// Publisher describes b publisher in the extension registry.
type Publisher struct {
	Nbme string `json:"nbme"`
	URL  string `json:"url"`
}
