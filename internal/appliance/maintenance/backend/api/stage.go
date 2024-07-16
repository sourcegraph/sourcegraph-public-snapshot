package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var maintenanceEndpoint = os.Getenv("MAINTENANCE_ENDPOINT")

func init() {
	if adminPassword == "" {
		fmt.Println("Variable MAINTENANCE_PASSWORD is missing.")
		os.Exit(1)
	}
}

type status struct {
	Stage          Stage    `json:"stage"`
	CurrentVersion *string  `json:"version"`     // current version, nil if not installed
	NextVersion    *string  `json:"nextVersion"` // version being installed/upgraded nil if not being installed/upgraded
	Tasks          []Task   `json:"tasks"`
	Errors         []string `json:"errors"`
}

type Stage string

const (
	StageUnknown         Stage = "unknown"
	StageIdle            Stage = "idle"
	StageInstall         Stage = "install"
	StageInstalling      Stage = "installing"
	StageUpgrading       Stage = "upgrading"
	StageWaitingForAdmin Stage = "wait-for-admin"
	StageRefresh         Stage = "refresh"
)

type Feature struct {
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Version  string `json:"version"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type FeatureResponse struct {
	Feature Feature `json:"feature"`
}

type StageResponse struct {
	Stage string `json:"stage"`
	Data  string `json:"data"`
}

var epoch = time.Unix(0, 0)

var currentStage Stage = StageInstall
var switchToAdminTime time.Time = epoch

func init() {
	fmt.Println("Initial stage:", currentStage)
}

func StageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request for stage")

	status, _ := GetSearchStatus()
	fmt.Println("GetSearchStatus", status)

	switch status {
	case "installing":
		currentStage = StageInstalling
	case "ready":
		fmt.Println("ready!", switchToAdminTime, currentStage)
		if switchToAdminTime == time.Unix(0, 0) {
			if currentStage != StageRefresh && currentStage != StageWaitingForAdmin {
				switchToAdminTime = time.Now().Add(5 * time.Second)
			}
		} else {
			if time.Now().After(switchToAdminTime) {
				switchToAdminTime = epoch
				currentStage = StageWaitingForAdmin
			}
		}
	case "unknown":
		break
	}

	result := StageResponse{
		Stage: string(currentStage),
	}

	switch currentStage {
	case StageRefresh:
		currentStage = StageUnknown
	}

	fmt.Println("Sending current stage", result)
	sendJson(w, result)
}

func SetStageHandlerForTesting(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request to set stage")

	var request StageResponse
	receiveJson(w, r, &request)

	fmt.Println("Setting stage to", request.Stage)
	currentStage = Stage(request.Stage)

	fmt.Println(installTasks)

	switch currentStage {
	case StageInstalling:
		installError = ""
		installTasks = createInstallTasks()
		installVersion = request.Data
		go func() {
			_, err := EnableSearch()
			if err != nil {
				installError = err.Error()
			}
		}()
	case StageUpgrading:
		installError = ""
		installTasks = createFakeUpgradeTasks()
		installVersion = request.Data
	}
}

func GetSearchStatus() (string, error) {
	data := []byte(`{"name": "search"}`)
	url := fmt.Sprintf("http://%s:8734/operator.maintenance.v1.MaintenanceService/GetFeature", maintenanceEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	var feature FeatureResponse
	if err := json.Unmarshal(body, &feature); err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	fmt.Println("install feature", feature)

	if feature.Feature.Status == "installing" {
		installTasks[InstallTaskWaitForCluster].Finished = true
		installTasks[InstallTaskSetup].Started = true
		installTasks[InstallTaskSetup].Finished = false
		installTasks[InstallTaskSetup].Progress = feature.Feature.Progress
	}

	if feature.Feature.Status == "ready" {
		installTasks[InstallTaskSetup].Finished = true
		installTasks[InstallTaskStart].Started = true
	}

	return feature.Feature.Status, nil
}

func EnableSearch() (string, error) {
	data := []byte(`{"feature":{"name":"search", "enabled":true}}`)
	url := fmt.Sprintf("http://%s:8734/operator.maintenance.v1.MaintenanceService/EnableFeature", maintenanceEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error %v", err)
		return "", err
	}

	var feature FeatureResponse
	if err := json.Unmarshal(body, &feature); err != nil {
		fmt.Printf("Error %v", err)
		installError = err.Error()
		return "", err
	}

	installTasks[InstallTaskWaitForCluster].Finished = true
	installTasks[InstallTaskWaitForCluster].Progress = 100

	return feature.Feature.Status, nil
}
