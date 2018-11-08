package main

import (
	"container/list"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// Extract the http handler functions from their generators.
var (
	createRepositoryHandler = handleCreateRepository()
	repositoryHandler       = handleRepository()
	listRepositoriesHandler = handleListRepositories()
)

// This clears out all cached repository data. It does not delete the on-disk
// repositories (that is done at the end of the full test).
func resetRepositoriesForTest() {
	repositoriesMutex.Lock()
	repositories = map[string]*repository{}
	repositoryDeletionQueue = list.New()
	repositoriesMutex.Unlock()
}

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
      "repository_name": "simple_name",
      "file_contents": {
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
      "repository_name": "stackoverflow.com/questions/1760757",
      "file_contents": {
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
      "repository_name": "simple_name",
      "file_contents": {
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
    "repository_name": "testRepo",
    "file_contents": {
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
	responseDecoder := json.NewDecoder(result.Body)
	err := responseDecoder.Decode(&response)
	if err != nil {
		t.Errorf("Couldn't decode create-repository response: %v", err)
		return
	}

	// Get the repository key from the returned path.
	repoKeyRegexp := regexp.MustCompile(`^/repository/([0-9a-f]+)`)
	pathComponents := repoKeyRegexp.FindStringSubmatch(response.Path)
	if len(pathComponents) < 2 {
		t.Errorf("Didn't recognize repository key in create-repository response")
		return
	}
	repoKey := pathComponents[1]
	repo := repositories[repoKey]
	if repo == nil {
		t.Errorf("Returned repository [%s] doesn't exist", repoKey)
		return
	}

	// Make sure the repo exists on disk with the given data.
	filePath := filepath.Join(repositoriesRoot, repoKey, "a_file")
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
    "repository_name": "testRepo",
    "file_contents": {
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
