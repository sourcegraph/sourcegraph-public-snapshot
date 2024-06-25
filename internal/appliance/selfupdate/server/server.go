package server

import (
	"log"
	"time"

	"github.com/sourcegraph/sourcegraph/appliance/selfupdate"
	"github.com/sourcegraph/sourcegraph/appliance/selfupdate/schema"
)

const CHECK_INTERVAL = 5 * time.Second

type Server interface {
	Run()
}

func New(updater selfupdate.SelfUpdater) Server {
	return &server{
		updater: updater,
	}
}

type server struct {
	updater selfupdate.SelfUpdater
}

func (s *server) Run() {
	log.Println("Running self-update check loop every", CHECK_INTERVAL)
	timer := time.NewTicker(CHECK_INTERVAL)
	defer timer.Stop()

	go func() {
		for {
			log.Println("Waiting a little before checking manifest online")
			<-timer.C
			// Fetch from the web
			config, err := schema.Fetch()
			if err != nil {
				log.Println("Failed to download self-update manifest", err.Error())
			}
			if err := s.updater.Start(config); err != nil {
				log.Println("Failed to update the system", err.Error())
			}
		}
	}()

	time.Sleep(24 * time.Hour)
}
