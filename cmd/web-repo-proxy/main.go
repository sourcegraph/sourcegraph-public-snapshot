package main

import (
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
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

/*var (
	addr = flag.String
)*/

var (
	repositoriesRoot = "/Users/fae/gitproxy"
	repositoryTTL    = 30 * 60 // 30 minutes
)

type CreateRepositoryRequest struct {
	PathPrefix   string            `json:"path_prefix"`
	FileContents map[string]string `json:"file_contents"`
}

type CreateRepositoryResponse struct {
	Path      string `json:"path"`
	Error     string `json:"error"`
	GoodUntil int64  `json:"good_until"`
}

type repository struct {
	key        string
	pathPrefix string
	goodUntil  int64
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
	// Create path if it doesn't exist
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	for filename, content := range data {
		filePath := filepath.Join(path, filename)
		err = ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return err
		}
	}
	repo, err := git.PlainInit(path, false)
	if err != nil {
		fmt.Printf("PlainInit failed\n")
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Worktree failed\n")
		return err
	}
	for filename := range data {
		_, err = worktree.Add(filename)
		if err != nil {
			fmt.Printf("Add [%s] failed\n", filename)
			return err
		}
	}
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Sourcegraph git proxy",
			Email: "sourcegraph@sourcegraph.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Printf("Commit failed\n")
		return err
	}

	// This command builds info/refs in
	// the repository, which we need to serve repos as static files.
	// go-git doesn't support update-server-info (see
	// https://github.com/src-d/go-git/blob/master/COMPATIBILITY.md) so we
	// need to do it the old-fashioned way.
	cmd := exec.Command("git", "update-server-info")
	cmd.Dir = path
	return cmd.Run()
}

var (
	repositories      = map[string]*repository{}
	repositoriesMutex = &sync.Mutex{}
)

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
	repositoriesMutex.Unlock()
	return repository
}

func handleCreateRepository() http.HandlerFunc {
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
			if repo != nil {
				fmt.Printf("I have repository [%s]\n", repoKey)
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
		// Trim "/repository" off the path.
		/*path := request.URL.Path[11:]
		path := path.Clean(request.URL.Path)
		pathComponents := strings.Split(path, "/")*/
	}
}

func handleListRepositories() http.HandlerFunc {
	type ListRepositoriesResponse struct {
		RepositoryPaths []string `json:"repository_paths"`
	}
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		repositoryPaths := []string{}
		for key, repo := range repositories {
			path := key
			if repo.pathPrefix != "" {
				path += "/" + repo.pathPrefix
			}
			repositoryPaths = append(repositoryPaths, path)
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

	http.HandleFunc("/create-repository", handleCreateRepository())
	http.HandleFunc("/repository/", handleRepository())
	http.HandleFunc("/list-repositories", handleListRepositories())
	log.Fatal(http.ListenAndServe(":4014", nil))
}
