package watcher

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher interface {
	Run()
	UpdateVersion(newVersion string)
}

func New() Watcher {
	versionFile := os.Getenv("VERSION_CONFIG_FILE")
	if versionFile == "" {
		log.Fatal("env var VERSION_CONFIG_FILE is not set")
	}

	return &watcher{
		versionFile: versionFile,
	}
}

type watcher struct {
	versionFile    string
	currentVersion string
}

func (w *watcher) Run() {
	go w.watch()
}

func (w *watcher) UpdateVersion(newVersion string) {
	w.currentVersion = newVersion
}

// Watches for version changes in config file and restarts
func (w *watcher) watch() {
	log.Println("Watching", w.versionFile, "for changes")

	var reload = func(currentVersion string, newVersion string) {
		fmt.Println("Version changed",
			"from", currentVersion,
			"to", newVersion,
			"! Restarting...")
		os.Exit(0)
	}

	w.currentVersion = w.load()

	var checkVersion = func() {
		newVersion := w.load()
		if newVersion != w.currentVersion {
			reload(w.currentVersion, newVersion)
		}
	}

	// Initialize watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					checkVersion()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(w.versionFile)
	if err != nil {
		log.Fatal(err)
	}

	// Periodically check the file in case no events are triggered
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			checkVersion()
		}
	}()

	<-done
}

func (w *watcher) load() string {
	data, err := os.ReadFile(w.versionFile)
	if err != nil {
		log.Fatal("cannot read config file", err)
	}
	version := strings.Split(string(data), "\n")[0]
	log.Println("version (config file):", version)
	return version
}
