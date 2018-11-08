package main

import (
	"container/list"
	"encoding/json"
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
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	repositoriesRoot    = env.Get("REPOSITORIES_ROOT", "./repositories", "Root path to store repository data")
	repositoryTTLString = env.Get("REPOSITORY_TTL", "1800", "How long to keep repositories, in seconds")
	listenerPort        = env.Get("LISTENER_PORT", "4014", "The network port to listen on")

	maxRequestSize = 0x10000 // Repo creation requests must be at most 64KB.
)

type repository struct {
	key         string
	pathPrefix  string
	goodUntil   int64
	savedToDisk bool
}

func (r *repository) filePathRoot() string {
	return filepath.Join(repositoriesRoot, r.key)
}

func (r *repository) urlRoot() string {
	root := "/repository/" + r.key
	if r.pathPrefix != "" {
		return root + "/" + r.pathPrefix
	}
	return root
}

func (r *repository) filePathForURLSubpath(subpath string) string {
	root := r.filePathRoot()
	// To serve a repository over static http we need to serve files from its
	// .git subdirectory.
	return filepath.Join(root, ".git", filepath.Clean(subpath))
}

func (r *repository) writeToDisk(data map[string]string) error {
	path := r.filePathRoot()
	// Delete the target directory and its contents if they exist.
	os.RemoveAll(path)
	// Create a new empty directory.
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	// Write out the given data to the file system.
	for filename, content := range data {
		filePath := filepath.Join(path, filename)
		err = ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return err
		}
	}
	// Initialize the new repo and add the files to it.
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		return err
	}
	// This command builds info/refs in the repository, which we need to serve
	// repos as static files.
	cmd = exec.Command("git", "update-server-info")
	cmd.Dir = path
	return cmd.Run()
}

var (
	repositories            = map[string]*repository{}
	repositoryDeletionQueue = list.New()
	repositoriesMutex       = &sync.Mutex{}
)

func deleteRepository(r *repository) {
	repositoriesMutex.Lock()
	delete(repositories, r.key)
	repositoriesMutex.Unlock()
	os.RemoveAll(r.filePathRoot())
}

func repositoryDeleter() {
	// Watches the repository deletion queue and removes repositories once
	// their goodUntil field is past.
	for {
		curTime := time.Now().Unix()
		for r := repositoryDeletionQueue.Front(); r != nil; r = repositoryDeletionQueue.Front() {
			repository := r.Value.(*repository)
			if !repository.savedToDisk || curTime <= repository.goodUntil {
				// If a repository hasn't been saved to disk yet then its goodUntil
				// field may not be initialized, so try again later.
				break
			}
			log15.Info("Deleting expired repository", "repositoryKey", repository.key)
			repositoriesMutex.Lock()
			delete(repositories, repository.key)
			repositoryDeletionQueue.Remove(r)
			repositoriesMutex.Unlock()
			os.RemoveAll(repository.filePathRoot())
		}
		time.Sleep(time.Minute)
	}
}

func createRandomKey() string {
	keyIndex := rand.Int31()
	return strconv.FormatInt(int64(keyIndex), 16)
}

// Generates a unique key, adds a repository with that key to the global
// list, and returns the new repository.
func createEmptyRepository() *repository {
	key := createRandomKey()

	repositoriesMutex.Lock()
	for repositories[key] != nil {
		repositoriesMutex.Unlock()
		key = createRandomKey()
		repositoriesMutex.Lock()
	}
	repository := &repository{key: key}
	repositories[key] = repository
	repositoryDeletionQueue.PushBack(repository)
	repositoriesMutex.Unlock()
	return repository
}

type CreateRepositoryRequest struct {
	PathPrefix   string            `json:"path_prefix"`
	FileContents map[string]string `json:"file_contents"`
}

type CreateRepositoryResponse struct {
	Path      string `json:"path"`
	GoodUntil int64  `json:"good_until"`
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
	// The path prefix needs to be a valid URL substring.
	request.PathPrefix = url.PathEscape(request.PathPrefix)
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
	repositoryTTL, err := strconv.ParseInt(repositoryTTLString, 10, 32)
	if err != nil {
		// Fall back on default of 30 minutes
		repositoryTTL = 30 * 60
	}
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		// Decode the request.
		createRepositoryRequest, err := sanitizedCreateRequest(request.Body)
		if err != nil {
			log15.Warn("/create-repository request failed", "error", err)
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}
		// Create a new repository and save it to disk.
		repo := createEmptyRepository()
		repo.pathPrefix = createRepositoryRequest.PathPrefix
		repo.goodUntil = time.Now().Unix() + int64(repositoryTTL)
		err = repo.writeToDisk(createRepositoryRequest.FileContents)
		if err != nil {
			log15.Error("/create-repository couldn't write repository to disk", "error", err)
			responseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		repo.savedToDisk = true
		// Send successful response to client.
		response := CreateRepositoryResponse{
			Path:      repo.urlRoot(),
			GoodUntil: repo.goodUntil}
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
		repo := repositories[repoKey]
		if repo == nil || !repo.savedToDisk {
			// This repository doesn't exist or hasn't been initialized yet.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Request for unknown repository", "path", request.URL.Path, "repoKey", repoKey)
			return
		}
		if !strings.HasPrefix(pathComponents[2], repo.pathPrefix) {
			// All repository file requests must include the repository path prefix.
			responseWriter.WriteHeader(http.StatusNotFound)
			log15.Warn("Repository request doesn't include path prefix", "path", request.URL.Path, "repoKey", repoKey, "pathPrefix", repo.pathPrefix)
			return
		}
		truncatedPath := pathComponents[2][len(repo.pathPrefix):]
		filePath := repo.filePathForURLSubpath(truncatedPath)
		http.ServeFile(responseWriter, request, filePath)
	}
}

func handleListRepositories() http.HandlerFunc {
	type ListRepositoriesResponse struct {
		RepositoryPaths []string `json:"repository_paths"`
	}
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		repositoryPaths := []string{}
		for _, repo := range repositories {
			if repo.savedToDisk {
				repositoryPaths = append(repositoryPaths, repo.urlRoot())
			}
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
