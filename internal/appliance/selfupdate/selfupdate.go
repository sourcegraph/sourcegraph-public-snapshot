package selfupdate

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/schema"
)

// TODO: Make sure this gets injected by the build
var currentVersion string = os.Getenv("VERSION")
var exitOnSelfUpdate bool = os.Getenv("EXIT_ON_SELF_UPDATE") != "0"

type SelfUpdater interface {
	Start(config *schema.SelfUpdateDefinition) error
}

type ComponentUpdate interface {
	Update(comp *schema.ComponentUpdateInformation) (*semver.Version, error)
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
	if !exitOnSelfUpdate {
		fmt.Println("Self-update complete, but EXIT_ON_SELF_UPDATE=0. Not terminating.")
		return
	}
	os.Exit(0)
}

func (s *selfupdater) Start(
	config *schema.SelfUpdateDefinition, // latest definitions from server
) error {
	if config.SelfUpdate.Version != s.currentVersion {
		log.Println("Upgrading self-updater",
			"from", s.currentVersion,
			"to", config.SelfUpdate.Version,
			"@", config.SelfUpdate.UpdateUrl)
		if updated, err := s.updater.Update(&config.SelfUpdate); err != nil {
			log.Println("Failed to update self-updater", err)
			return err
		} else {
			if updated != nil {
				log.Println("Upgraded self-updater to", updated)
			} else {
				log.Println("Self-updater is already up-to-date")
			}
		}
		s.exitHandler()
		return nil
	} else {
		log.Println("Self-Updater is up to date.")
	}

	log.Println("Updating components")
	for _, c := range config.Components {
		log.Println("Updating", c.Name, "from", c.UpdateUrl)
		if updated, err := s.updater.Update(&c); err != nil {
			log.Println("Failed to update", c.Name, c.DisplayName, err)
			return err
		} else {
			if updated != nil {
				log.Println("Upgraded", c.Name, c.DisplayName, "to", updated)
			} else {
				log.Println("Component", c.Name, c.DisplayName, "is already up-to-date")
			}
		}
	}

	log.Println("Success!")
	return nil
}
