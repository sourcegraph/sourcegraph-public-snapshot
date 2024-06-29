package server

import (
	"log"
	"os"
	"time"
)

const CHECK_INTERVAL = 5 * time.Second

var exitOnSelfUpdate = os.Getenv("EXIT_ON_SELF_UPDATE")

type Server interface {
	Run()
}

type UpdaterResult int

const UpdaterResultUpToDate UpdaterResult = 0
const UpdaterResultUpgraded UpdaterResult = 1
const UpdaterResultFailed UpdaterResult = 2

type Updater interface {
	Start() (UpdaterResult, error)
}

func New(updater Updater) Server {
	return &server{
		updater: updater,
	}
}

type server struct {
	updater Updater
}

func (s *server) Run() {
	log.Println("Executing self-update check on startup")
	s.execute()

	log.Println("Running self-update check loop every", CHECK_INTERVAL)
	timer := time.NewTicker(CHECK_INTERVAL)
	defer timer.Stop()

	go func() {
		for {
			log.Println("Waiting a little before checking manifest online")
			<-timer.C
			s.execute()
		}
	}()

	time.Sleep(24 * time.Hour)
}

func (s *server) execute() {
	// Fetch from the web
	if result, err := s.updater.Start(); err != nil {
		log.Println("Failed to update the system", err.Error())
	} else {
		switch result {
		case UpdaterResultUpToDate:
			log.Println("System is up to date")
		case UpdaterResultUpgraded:
			log.Println("System updated successfully")
			if exitOnSelfUpdate == "1" {
				log.Println("Restarting the self-updater")
				os.Exit(0)
			} else {
				log.Println("Continuing to run")
			}
		case UpdaterResultFailed:
			log.Println("Failed to update the system")
		}
	}
}
