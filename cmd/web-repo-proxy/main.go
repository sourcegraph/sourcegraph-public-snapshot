package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	repositoriesRoot = env.Get("REPOSITORIES_ROOT", "./repositories", "Root path to store repository data")
	repositoryTTL    = mustParseDuration(env.Get("REPOSITORY_TTL", "30m", "How long to keep repositories, as a duration string"))

	listenerPort = env.Get("LISTENER_PORT", "4014", "The network port to listen on")

	maxRequestSize = 0x10000 // Repo creation requests must be at most 64KB.
)

func mustParseDuration(durationString string) time.Duration {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		panic(fmt.Sprintf("Unrecognized duration string '%v'", durationString))
	}
	return duration
}

// A repository is represented by its directory name under repositoriesRoot.
type repository string

// filePathForReponsitory returns the location of this repository on disk.
func (r *repository) filePath() string {
	return filepath.Join(repositoriesRoot, string(*r))
}

// urlPath returns the URL path to use when cloning this repository over http.
func (r *repository) urlPath() string {
	return "/repository/" + string(*r)
}

// goodUntil checks the repository's directory on disk and returns the time
// it will expire, or an error if the repository couldn't be read.
func (r *repository) goodUntil() (time.Time, error) {
	path := r.filePath()
	info, err := os.Lstat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime().Add(repositoryTTL), nil
}

func createRandomKey() string {
	keyIndex := rand.Int31()
	return strconv.FormatInt(int64(keyIndex), 16)
}

// createNewRepository creates a new repository with the given name and data.
// Returns a repository string if successful, otherwise returns error.
func createNewRepository(name string, data map[string]string) (*repository, error) {
	// Create a temporary directory to initialize with the repository data.
	tempRoot, err := ioutil.TempDir("", "web-repo-proxy")
	if err != nil {
		return nil, err
	}
	defer func() {
		// If there was an error, clean up what's left on disk.
		if err != nil {
			os.RemoveAll(tempRoot)
		}
	}()

	// Write out the given data to the file system.
	for filename, content := range data {
		filePath := filepath.Join(tempRoot, filename)
		err = ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return nil, err
		}
	}

	// Set up the git repository.
	gitCommands := [][]string{
		[]string{"init"},
		[]string{"add", "."},
		[]string{"commit", "-m", "Initial commit"},
		// update-server-info is needed to serve repositories over static http.
		[]string{"update-server-info"},
	}
	for _, command := range gitCommands {
		cmd := exec.Command("git", command...)
		cmd.Dir = tempRoot
		err = cmd.Run()
		if err != nil {
			return nil, err
		}
	}

	// Choose the real repository path and move it there.
	for i := 0; i < 3; i++ {
		// Retry up to 3 times to avoid ephemeral errors and name collisions.
		repository := repository(name + "-" + createRandomKey())
		repositoryRoot := repository.filePath()
		err = os.Rename(tempRoot, repositoryRoot)
		if err == nil {
			return &repository, nil
		}
	}
	// If we still failed after 3 tries, return the most recent error.
	return nil, err
}

// Check whether any repositories have expired, and delete them if so.
func deleteExpiredRepositories() {
	curTime := time.Now()
	files, err := ioutil.ReadDir(repositoriesRoot)
	if err != nil {
		log15.Error("Couldn't read repositories directory", "error", err)
		return
	}
	for _, file := range files {
		if file.ModTime().Add(repositoryTTL).Before(curTime) {
			// This repository has expired.
			tempDir, err := ioutil.TempDir("", "web-repo-proxy")
			if err != nil {
				log15.Error("Couldn't create temporary directory", "error", err)
				continue
			}
			fromPath := filepath.Join(repositoriesRoot, file.Name())
			toPath := filepath.Join(tempDir, file.Name())
			err = os.Rename(fromPath, toPath)
			if err != nil {
				log15.Error("Couldn't move expired repository entry", "error", err)
				continue
			}
			err = os.RemoveAll(tempDir)
			if err != nil {
				log15.Error("Couldn't delete temporary directory", "tempDir", tempDir)
			}
		}
	}
}

// Watches the repository deletion queue and removes repositories once
// their goodUntil field is past.
func repositoryDeleter() {
	for {
		deleteExpiredRepositories()
		time.Sleep(30 * time.Second)
	}
}

type CreateRepositoryRequest struct {
	RepositoryName string            `json:"repositoryName"`
	FileContents   map[string]string `json:"fileContents"`
}

type CreateRepositoryResponse struct {
	URLPath   string `json:"urlPath"`
	GoodUntil int64  `json:"goodUntil"`
}

// Given a Reader with a create-repository request, return the corresponding
// CreateRepositoryRequest with guaranteed-safe values, or an error.
func sanitizedCreateRequest(requestBody io.Reader) (*CreateRepositoryRequest, error) {
	var request CreateRepositoryRequest
	err := json.NewDecoder(requestBody).Decode(&request)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't decode create-repository request")
	}
	// The repository name needs to be a valid URL path substring.
	request.RepositoryName = url.PathEscape(request.RepositoryName)
	// File keys with subdirectories are not supported, so immediately reject
	// anything that has a slash in it.
	// This is not enough to ensure that a filename is valid, but unsupported
	// filename strings are caught in the file creation stage.
	for filename, _ := range request.FileContents {
		if strings.Contains(filename, "/") {
			return nil, errors.New("Repository filenames can't contain '/'")
		}
	}
	return &request, nil
}

func handleCreateRepository() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		// Decode the request.
		requestBody := http.MaxBytesReader(responseWriter, request.Body, int64(maxRequestSize))
		defer requestBody.Close()

		// Sanitize and error-check the request fields.
		createRepositoryRequest, err := sanitizedCreateRequest(requestBody)
		if err != nil {
			log15.Warn("/create-repository request failed", "error", err)
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
		// Empty repositories are considered invalid requests.
		if len(createRepositoryRequest.FileContents) == 0 {
			log15.Warn("/create-repository called with empty file contents")
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create the new repository.
		repo, err := createNewRepository(createRepositoryRequest.RepositoryName, createRepositoryRequest.FileContents)
		if err != nil {
			log15.Error("Couldn't create new repository", "error", err)
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		goodUntil, _ := repo.goodUntil()
		// Send successful response to client.
		response := CreateRepositoryResponse{
			URLPath:   repo.urlPath(),
			GoodUntil: goodUntil.Unix(),
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log15.Error("/create-repository couldn't marshal response for client", "error", err)
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseWriter.Write(responseJSON)
	}
}

func handleRepository() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		if !strings.HasPrefix(request.URL.Path, "/repository/") {
			// This should never happen, but let's make sure.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Malformed repository request", "path", request.URL.Path)
			return
		}
		trailingPath := request.URL.Path[12:]

		// Trim the repository string off the start of the path.
		pathComponents := strings.SplitN(trailingPath, "/", 2)
		if len(pathComponents) < 2 {
			// A malformed URL: we need at least a repository name and file subpath.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Malformed repository request", "path", request.URL.Path)
			return
		}

		// Serve the requested file.
		repo := repository(pathComponents[0])
		repositoryRoot := repo.filePath()
		// ".git" because we serve as a bare repository.
		filePath := filepath.Join(repositoryRoot, ".git", filepath.Clean(pathComponents[1]))
		http.ServeFile(responseWriter, request, filePath)
	}
}

func handleListRepositories() http.HandlerFunc {
	type ListRepositoriesResponse struct {
		RepositoryPaths []string `json:"repository_paths"`
	}
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		files, err := ioutil.ReadDir(repositoriesRoot)
		if err != nil {
			log15.Error("Couldn't read repositories directory", "error", err)
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		repositoryPaths := []string{}
		for _, file := range files {
			repo := repository(file.Name())
			repositoryPaths = append(repositoryPaths, repo.urlPath())
		}
		response := ListRepositoriesResponse{repositoryPaths}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			responseWriter.WriteHeader(http.StatusInternalServerError)
		} else {
			responseWriter.Write(responseJSON)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	env.Lock()
	env.HandleHelpFlag()

	err := os.MkdirAll(repositoriesRoot, 0755)
	if err != nil {
		panic(fmt.Sprintf("Couldn't access repositories root path '%s'", repositoriesRoot))
	}

	// Watch the repositories and delete them when they get too old.
	go repositoryDeleter()

	http.HandleFunc("/create-repository", handleCreateRepository())
	http.HandleFunc("/repository/", handleRepository())
	if env.InsecureDev {
		// list-repositories is for testing, not production.
		http.HandleFunc("/list-repositories", handleListRepositories())
	}

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, listenerPort)
	log15.Info("web-repo-proxy: listening", "addr", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
}
