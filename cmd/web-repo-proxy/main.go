package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	repositoriesRoot    = env.Get("REPOSITORIES_ROOT", "./repositories", "Root path to store repository data")
	repositoryTTLString = env.Get("REPOSITORY_TTL", "1800", "How long to keep repositories, in seconds")
	listenerPort        = env.Get("LISTENER_PORT", "4014", "The network port to listen on")
)

type repository struct {
	key        string
	pathPrefix string
	goodUntil  int64
	active     bool
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
			if !repository.active || curTime <= repository.goodUntil {
				// An inactive repository may not have a valid goodUntil field yet,
				// so try again later.
				break
			}
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
	Error     string `json:"error"`
	GoodUntil int64  `json:"good_until"`
}

func handleCreateRepository() http.HandlerFunc {
	repositoryTTL, err := strconv.ParseInt(repositoryTTLString, 10, 32)
	if err != nil {
		// Fall back on default of 30 minutes
		repositoryTTL = 30 * 60
	}
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		decoder := json.NewDecoder(request.Body)
		var createRepositoryRequest CreateRepositoryRequest
		err := decoder.Decode(&createRepositoryRequest)
		if err != nil {
			fmt.Printf("Couldn't decode request: %s\n", request.Body)
		} else {
			fmt.Printf("Decoded request: %#v\n", createRepositoryRequest)
			// TODO: sanitize / bounds-check all inputs.
			repo := createEmptyRepository()
			repo.pathPrefix = createRepositoryRequest.PathPrefix
			repo.goodUntil = time.Now().Unix() + int64(repositoryTTL)
			err = repo.writeToDisk(createRepositoryRequest.FileContents)
			if err != nil {
				fmt.Printf("Couldn't write the repository to disk: %v\n", err)
			} else {
				repo.active = true
				response := CreateRepositoryResponse{
					Path:      repo.urlRoot(),
					GoodUntil: repo.goodUntil}
				responseJSON, err := json.Marshal(response)
				if err != nil {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				} else {
					responseWriter.Write(responseJSON)
				}
			}
		}
	}
}

func handleRepository() http.HandlerFunc {
	repoKeyRegexp := regexp.MustCompile(`^/repository/([0-9a-f]+)/(.*)$`)

	return func(responseWriter http.ResponseWriter, request *http.Request) {
		pathComponents := repoKeyRegexp.FindStringSubmatch(request.URL.Path)
		if len(pathComponents) >= 3 {
			repoKey := pathComponents[1]
			repo := repositories[repoKey]
			if repo != nil && repo.active {
				if strings.HasPrefix(pathComponents[2], repo.pathPrefix) {
					truncatedPath := pathComponents[2][len(repo.pathPrefix):]
					filePath := repo.filePathForURLSubpath(truncatedPath)
					http.ServeFile(responseWriter, request, filePath)
				} else {
					fmt.Printf("Expected prefix [%s] for repo [%s]\n", repo.pathPrefix, repoKey)
				}
			} else {
				fmt.Printf("I don't recognize repository [%s]\n", repoKey)
			}
		} else {
			fmt.Printf("I don't recognize that as a repository command (%s)\n", request.URL.Path)
		}
	}
}

func handleListRepositories() http.HandlerFunc {
	type ListRepositoriesResponse struct {
		RepositoryPaths []string `json:"repository_paths"`
	}
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		repositoryPaths := []string{}
		for _, repo := range repositories {
			if repo.active {
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
	env.Lock()
	env.HandleHelpFlag()

	// Watch the repositories and delete them when they get too old.
	go repositoryDeleter()

	http.HandleFunc("/create-repository", handleCreateRepository())
	http.HandleFunc("/list-repositories", handleListRepositories())
	http.HandleFunc("/repository/", handleRepository())
	log.Fatal(http.ListenAndServe(":"+listenerPort, nil))
}
