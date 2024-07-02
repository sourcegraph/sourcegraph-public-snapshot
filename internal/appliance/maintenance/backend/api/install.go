package api

import (
	"log"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/appliance/maintenance/operator"
)

var installError string = ""
var installTasks []operator.Task = createInstallTasks()
var installVersion string = ""

type InstallProgress struct {
	Version  string          `json:"version"`
	Progress int             `json:"progress"`
	Error    string          `json:"error"`
	Tasks    []operator.Task `json:"tasks"`
}

func InstallProgressHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for install progress")

	var progress int

	installTasks, progress = calculateProgress(installTasks)

	result := InstallProgress{
		Version:  installVersion,
		Progress: progress,
		Error:    installError,
		Tasks:    installTasks,
	}

	if installError == "" {
		installTasks = progressTasks(installTasks)
	}

	log.Println("Sending current install progress", result)
	sendJson(w, result)
}

func SetInstallErrorForTesting(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to set error")

	installError = "Something tragic happened. Sorry! Please wait until we try something creative..."

	w.Write([]byte("ok"))
}
