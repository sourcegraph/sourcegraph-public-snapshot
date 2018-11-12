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
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	repositoriesRoot    = env.Get("REPOSITORIES_ROOT", "./repositories", "Root path to store repository data")
	repositoryTTLString = env.Get("REPOSITORY_TTL", "30m", "How long to keep repositories, as a duration string")
	repositoryTTL       = time.Duration(0)

	listenerPort = env.Get("LISTENER_PORT", "4014", "The network port to listen on")

	maxRequestSize = 0x10000 // Repo creation requests must be at most 64KB.
)

type repository struct {
	key  string
	name string
}

// filePathRoot returns the location of this repository on disk.
func (r *repository) filePathRoot() string {
	return filepath.Join(repositoriesRoot, fmt.Sprintf("%s-%s", r.key, r.name))
}

// urlRoot returns the URL path to use when cloning this repository over http.
func (r *repository) urlRoot() string {
	root := "/repository/" + r.key
	if r.name != "" {
		return root + "/" + r.name
	}
	return root
}

// goodUntil checks the repository's directory on disk and returns the time
// it will expire, or an error if the repository couldn't be read.
func (r *repository) goodUntil() (time.Time, error) {
	info, err := os.Lstat(r.filePathRoot())
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime().Add(repositoryTTL), nil
}

// serveFile handles an http file request within the given repository.
func (r *repository) serveFile(responseWriter http.ResponseWriter, request *http.Request, path string) {
	if r.name != "" {
		if !strings.HasPrefix(path, r.name+"/") {
			// All repository file requests must include the repository name
			// if it has one.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Repository request doesn't include repository name", "path", request.URL.Path, "repository key", r.key, "repository name", r.name)
			return
		}
		// Trim the name off the path.
		truncatedPath := path[len(r.name):]
		// To serve a repository over static http we need to serve files from its
		// .git subdirectory.
		filePath := filepath.Join(r.filePathRoot(), ".git", filepath.Clean(truncatedPath))
		http.ServeFile(responseWriter, request, filePath)
	}
}

// Returns repository objects for every directory matching the given key.
// Typically this should return 0 or 1 result.
func repositoriesForKey(key string) []*repository {
	files, err := ioutil.ReadDir(repositoriesRoot)
	if err != nil {
		return nil
	}
	matches := []*repository{}
	for _, file := range files {
		fileName := file.Name()
		if strings.HasPrefix(fileName, key+"-") {
			matches = append(matches, &repository{
				key:  key,
				name: fileName[len(key)+1:],
			})
		}
	}
	return matches
}

func createRandomKey() string {
	keyIndex := rand.Int31()
	return strconv.FormatInt(int64(keyIndex), 16)
}

// createNewRepository creates a new repository with the given name and data.
// Returns a repository object if successful, otherwise returns error.
func createNewRepository(name string, data map[string]string) (*repository, error) {
	// Create a temporary directory to initialize with the repository data.
	repositoryRoot, err := ioutil.TempDir("", "web-repo-proxy")
	if err != nil {
		return nil, err
	}

	// Write out the given data to the file system.
	for filename, content := range data {
		filePath := filepath.Join(repositoryRoot, filename)
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
		cmd.Dir = repositoryRoot
		err = cmd.Run()
		if err != nil {
			return nil, err
		}
	}

	// Choose the real repository path and move it there.
	for {
		repository := repository{
			key:  createRandomKey(),
			name: name,
		}
		newRoot := repository.filePathRoot()
		err = os.Rename(repositoryRoot, newRoot)
		if err != nil {
			return nil, err
		}
		repositoryRoot = newRoot
		// Corner case: only finish if we picked a unique key.
		repos := repositoriesForKey(repository.key)
		if len(repos) <= 1 {
			return &repository, nil
		}
	}
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
	UrlPath   string `json:"urlPath"`
	GoodUntil int64  `json:"goodUntil"`
}

// Given a Reader with a create-repository request, return the corresponding
// CreateRepositoryRequest with guaranteed-safe values, or an error.
func sanitizedCreateRequest(requestBody io.Reader) (*CreateRepositoryRequest, error) {
	decoder := json.NewDecoder(requestBody)
	var request CreateRepositoryRequest
	err := decoder.Decode(&request)
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
		createRepositoryRequest, err := sanitizedCreateRequest(requestBody)
		if err != nil {
			log15.Warn("/create-repository request failed", "error", err)
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
			UrlPath:   repo.urlRoot(),
			GoodUntil: goodUntil.Unix()}
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
	// Right now repo keys must be hexadecimal strings.
	repoKeyRegexp := regexp.MustCompile(`^/repository/([0-9a-f]+)/(.*)$`)

	return func(responseWriter http.ResponseWriter, request *http.Request) {
		pathComponents := repoKeyRegexp.FindStringSubmatch(request.URL.Path)
		if len(pathComponents) < 3 {
			// A malformed URL: we need at least a repository key and file subpath.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Malformed repository request", "path", request.URL.Path)
			return
		}
		repoKey := pathComponents[1]
		repositories := repositoriesForKey(repoKey)
		if len(repositories) == 0 {
			// No matching repository was found.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Request for unknown repository", "path", request.URL.Path, "repoKey", repoKey)
			return
		}
		repo := repositories[0]
		repo.serveFile(responseWriter, request, pathComponents[2])
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
			components := strings.SplitN(file.Name(), "-", 2)
			if len(components) != 2 {
				log15.Warn("Unrecognized file in repositories root", "fileName", file.Name())
				continue
			}
			repo := repository{key: components[0], name: components[1]}
			repositoryPaths = append(repositoryPaths, repo.urlRoot())
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

	// Make sure the TTL and repository root are valid.
	ttl, err := time.ParseDuration(repositoryTTLString)
	if err != nil {
		panic(fmt.Sprintf("Unrecognized repository TTL '%v'", repositoryTTLString))
	}
	repositoryTTL = ttl
	err = os.MkdirAll(repositoriesRoot, 0755)
	if err != nil {
		panic(fmt.Sprintf("Couldn't access repositories root path", "repositoriesRoot", repositoriesRoot))
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
