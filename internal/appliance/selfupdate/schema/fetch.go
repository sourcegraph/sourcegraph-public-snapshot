package schema

import (
	"log"
	"os"
)

// TODO(nelsona): Add proxy/socks support.
var SELF_UPDATE_URL = []string{
	os.Getenv("SELF_UPDATE_URL"),
	// TODO(nelsona): Make a forward
	"https://sourcegraph.com/REGISTRY-GOES-here/....",
}

func Fetch() (*SelfUpdateDefinition, error) {
	log.Println("Fetching manifest from", SELF_UPDATE_URL)
	return &SelfUpdateDefinition{
		SelfUpdate: ComponentUpdateInformation{
			Version:     "2.3.4",
			Name:        "selfupdater",
			DisplayName: "Self Updater",
			UpdateUrl:   "http://nowhere/blah/blah",
		},
		Components: []ComponentUpdateInformation{
			{
				Name:        "component-one",
				Version:     "1.1.1",
				DisplayName: "Component O-N-E",
				UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
			},
			{
				Name:        "component-two",
				Version:     "1.1.1",
				DisplayName: "Component T-W-O",
				UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
			},
			{
				Name:        "component-three",
				Version:     "1.1.1",
				DisplayName: "Component III",
				UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
			},
		},
	}, nil
}
