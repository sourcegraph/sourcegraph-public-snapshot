package selfupdate

import (
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/schema"
)

// TODO: Make sure this gets injected by the build
var currentVersion string = os.Getenv("VERSION")

type SelfUpdater interface {
	Start(config *schema.SelfUpdateDefinition) error
}

type ComponentUpdate interface {
	Update(comp *schema.ComponentUpdateInformation) error
}

func New(updater ComponentUpdate) SelfUpdater {
	return &selfupdater{
		currentVersion: currentVersion,
		updater:        updater,
		exitHandler:    defaultExitHandler,
	}
}

type selfupdater struct {
	currentVersion string
	updater        ComponentUpdate
	exitHandler    func()
}

func defaultExitHandler() {
	os.Exit(0)
}

func (s *selfupdater) Start(
	config *schema.SelfUpdateDefinition, // latest definitions from server
) error {
	if config.Version != s.currentVersion {
		log.Println("Upgrading self-updater from", s.currentVersion, "to", config.Version, "from", config.SelfUpdate)
		if err := s.updater.Update(&config.SelfUpdate); err != nil {
			log.Println("Failed to update self-updater", err)
			return err
		}
		s.exitHandler()
		return nil
	}

	log.Println("Updating components")
	for _, c := range config.Components {
		log.Println("Updating", c.Name, "from", c.UpdateUrl)
		if err := s.updater.Update(&c); err != nil {
			log.Println("Failed to update", c.Name, c.DisplayName, err)
			return err
		}
	}

	log.Println("Success!")
	return nil
}
