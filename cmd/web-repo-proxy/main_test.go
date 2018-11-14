// Unit tests for web-repo-proxy.
//
// To run without end-to-end tests (starting a server and cloning repositories
// from it) run go test with the -short option.

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	// Extract the http handler functions from their generators.
	createRepositoryHandler = handleCreateRepository()
	repositoryHandler       = handleRepository()
	listRepositoriesHandler = handleListRepositories()
)

func TestMain(m *testing.M) {
	// Create a temporary directory to store our repositories.
	repositoriesRoot, _ = ioutil.TempDir("", "web-repo-proxy-test")
	result := m.Run()
	os.RemoveAll(repositoriesRoot)
	os.Exit(result)
}

type sanitizedCreateRequestTest struct {
	description         string
	requestString       string
	expected            *CreateRepositoryRequest
	expectedErrorString string
}

var sanitizedCreateRequestTests = []sanitizedCreateRequestTest{
	sanitizedCreateRequestTest{
		description: "Decoding a simple request",
		requestString: `{
      "repositoryName": "simple_name",
      "fileContents": {
        "a_file": "file contents\n"
      }
    }`,
		expected: &CreateRepositoryRequest{
			RepositoryName: "simple_name",
			FileContents:   map[string]string{"a_file": "file contents\n"},
		},
	},
	sanitizedCreateRequestTest{
		description: "Repository name is encoded as a safe url string",
		requestString: `{
      "repositoryName": "stackoverflow.com/questions/1760757",
      "fileContents": {
        "a_file": "file contents\n"
      }
    }`,
		expected: &CreateRepositoryRequest{
			RepositoryName: "stackoverflow.com%2Fquestions%2F1760757",
			FileContents:   map[string]string{"a_file": "file contents\n"},
		},
	},
	sanitizedCreateRequestTest{
		description: "Reject filenames that use subpaths",
		requestString: `{
      "repositoryName": "simple_name",
      "fileContents": {
        "dir/file": "file contents\n"
      }
    }`,
		expectedErrorString: "Repository filenames can't contain '/'",
	},
}

func TestSanitizedCreateRequest(t *testing.T) {
	for _, test := range sanitizedCreateRequestTests {
		stringReader := strings.NewReader(test.requestString)
		result, err := sanitizedCreateRequest(stringReader)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("[%s] Got %+v, want %+v", test.description, result, test.expected)
			continue
		}
		errorString := ""
		if err != nil {
			errorString = err.Error()
		}
		if errorString != test.expectedErrorString {
			t.Errorf("[%s] Expected error [%s], got [%s]", test.description, test.expectedErrorString, errorString)
		}
	}
}

func TestCreateRepository(t *testing.T) {
	request := httptest.NewRequest("POST", "/create-repository", strings.NewReader(`{
    "repositoryName": "testRepo",
    "fileContents": {
      "a_file": "file contents\n"
    }
  }`))
	responseWriter := httptest.NewRecorder()
	createRepositoryHandler(responseWriter, request)

	result := responseWriter.Result()
	if result.StatusCode != http.StatusOK {
		t.Errorf("Create repository returned error code %v", result.StatusCode)
		return
	}
	response := CreateRepositoryResponse{}
	err := json.NewDecoder(result.Body).Decode(&response)
	if err != nil {
		t.Errorf("Couldn't decode create-repository response: %v", err)
		return
	}

	// Get the repository name from the returned path.
	repo := repository(strings.TrimPrefix(response.URLPath, "/repository/"))

	// Make sure the repo exists on disk with the given data.
	filePath := filepath.Join(repositoriesRoot, string(repo), "a_file")
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Errorf("Couldn't read repository file at path %v", filePath)
		return
	}
	if string(fileContents) != "file contents\n" {
		t.Errorf("Wrong file contents: {\n%v}", string(fileContents))
	}
}

// Verify that requests over 64KB produce an error.
func TestCreateRepositoryLargeRequest(t *testing.T) {
	request := httptest.NewRequest("POST", "/create-repository", strings.NewReader(`{
    "repositoryName": "testRepo",
    "fileContents": {
      "a_file": "`+strings.Repeat(" ", 66000)+`"
    }
  }`))
	responseWriter := httptest.NewRecorder()
	createRepositoryHandler(responseWriter, request)

	result := responseWriter.Result()
	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected error code 400 for overlong request, got %v", result.StatusCode)
	}
}

// Verify that repositories with empty file contents produce an error.
func TestCreateEmptyRepository(t *testing.T) {
	request := httptest.NewRequest("POST", "/create-repository", strings.NewReader(`{
    "repositoryName": "emptyRepo"
  }`))
	responseWriter := httptest.NewRecorder()
	createRepositoryHandler(responseWriter, request)

	result := responseWriter.Result()
	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected error code 400 for repository with empty file contents, got %v", result.StatusCode)
	}
}

func TestRepositoryRequest(t *testing.T) {
	// Real repositories are hard to unit test because our server gives access
	// to the bare repositories, so realistically we'd need to clone from them.
	// Instead, create a fake repository entry where we know the contents.
	// For an end-to-end test that uses cloning, see e2e_test.go.
	repository := repository("testRepo-abcdef")
	dirPath := filepath.Join(repository.filePath(), ".git", "testdir")
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		t.Errorf("Couldn't create test repository directory '%s'", dirPath)
		return
	}
	filePath := filepath.Join(dirPath, "testfile")
	err = ioutil.WriteFile(filePath, []byte("test\ncontents\n"), 0644)
	if err != nil {
		t.Errorf("Couldn't create test repository file '%s'", filePath)
		return
	}

	// Now we can try fetching "testfile" through the repository handler.
	request := httptest.NewRequest("GET", repository.urlPath()+"/testdir/testfile", nil)
	responseWriter := httptest.NewRecorder()
	repositoryHandler(responseWriter, request)

	result := responseWriter.Result()
	if result.StatusCode != http.StatusOK {
		t.Errorf("Repository file request returned error code %v", result.StatusCode)
		return
	}
	body, _ := ioutil.ReadAll(result.Body)
	if string(body) != "test\ncontents\n" {
		t.Errorf("Wrong file contents: want:\n{test\ncontents\n}\ngot:\n{%s}", string(body))
	}
}

func TestEndToEnd(t *testing.T) {
	// This test starts the server and clones a git repository over the network,
	// so don't run in short mode.
	if testing.Short() {
		t.Skip()
	}

	// Build the server. We can't start it directly with "go run" because
	// that runs the server as a forked child and we can't stop the process
	// directly.
	err := exec.Command("go", "build").Run()
	if err != nil {
		t.Errorf("Couldn't build server: %v", err)
		return
	}

	// Start up a full server.
	context, cancel := context.WithCancel(context.Background())
	server := exec.CommandContext(context, "./web-repo-proxy")
	server.Env = append(os.Environ(),
		"REPOSITORIES_ROOT="+repositoriesRoot)
	err = server.Start()
	if err != nil {
		t.Errorf("Couldn't start repository server: %v", err)
		return
	}
	// Shutdown the server when finished
	defer server.Wait()
	defer cancel()

	// Delay to let the server start listening.
	time.Sleep(time.Second / 10)

	// The server is running, send it a repository creation request.
	request := strings.NewReader(`{
    "repositoryName": "endToEnd",
    "fileContents": {
      "a_file": "file contents\n"
    }
  }`)
	response, err := http.Post("http://localhost:4014/create-repository", "application/json", request)
	if err != nil {
		t.Errorf("create-repository request failed: %v", err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("create-repository returned error code %v", response.StatusCode)
		return
	}

	// We got a successful response, extract the new repository path.
	parsedResponse := CreateRepositoryResponse{}
	err = json.NewDecoder(response.Body).Decode(&parsedResponse)
	if err != nil {
		t.Errorf("Couldn't parse create-repository response: %v", err)
		return
	}
	url, _ := url.Parse("http://localhost:4014" + parsedResponse.URLPath)

	// Try cloning from the server into a new temporary directory.
	tempDir, _ := ioutil.TempDir("", "web-repo-proxy-endtoend")
	defer os.RemoveAll(tempDir)
	gitCmd := exec.Command("git", "clone", url.String())
	gitCmd.Dir = tempDir
	err = gitCmd.Run()
	if err != nil {
		t.Errorf("Couldn't run git-clone: %v", err)
		return
	}

	// Check that the repository has the right contents.
	repoName := path.Base(url.Path)
	filePath := filepath.Join(tempDir, repoName, "a_file")
	fileContents, _ := ioutil.ReadFile(filePath)
	if string(fileContents) != "file contents\n" {
		t.Errorf("Wrong file contents, expected:\n{\nfile contents\n}, got:\n{\n%v}", string(fileContents))
	}
}
