package schema

import (
	"log"
	"os"
)

// TODO(nelsona): Add proxy/socks support.
var SELF_UPDATE_URL = os.Getenv("SELF_UPDATE_URL")

func Fetch() (*SelfUpdateDefinition, error) {
	log.Println("Fetching manifest from", SELF_UPDATE_URL)
	return &SelfUpdateDefinition{
		Version: "1.2.3",
		SelfUpdate: ComponentUpdateInformation{
			Name:        "self-update",
			DisplayName: "Self Updater",
			UpdateUrl:   "http://nowhere/blah/blah",
		},
		Components: []ComponentUpdateInformation{
			{
				Name:        "component-one",
				DisplayName: "Component O-N-E",
				UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
			},
			{
				Name:        "component-two",
				DisplayName: "Component T-W-O",
				UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
			},
			{
				Name:        "component-three",
				DisplayName: "Component III",
				UpdateUrl:   "http://nowhere/blah/blah/comp1/yada",
			},
		},
	}, nil
}
