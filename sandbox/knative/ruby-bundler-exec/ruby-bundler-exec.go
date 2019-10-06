package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Params struct {
	ArchiveURL string
	Dir        string
	Commands   [][]string // TODO!(sqs): this allows arbitrary execution

	IncludeFiles []string // paths of files (relative to Dir) whose contents to return in Response
}

type Result struct {
	Commands []CommandResult   `json:"commands"`
	Files    map[string]string `json:"files"`
}

type CommandResult struct {
	CombinedOutput string `json:"combinedOutput"`
	Ok             bool   `json:"ok"`
	Error          string `json:"error,omitempty"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Allow CORS.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With")

	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	var params Params
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO!(sqs): ensure dir is not ".." to avoid executing in arbitrary directories
	params.Dir = filepath.Clean(params.Dir)

	log.Printf("Start request: %+v", params)
	start := time.Now()
	defer func() { log.Printf("Finish request: %+v (%s)", params, time.Since(start)) }()

	if params.ArchiveURL == "" || len(params.Commands) == 0 {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	// Fetch and unzip archive.
	var tempDir string
	{
		var err error
		tempDir, err = ioutil.TempDir("", "workdir")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(tempDir)
		req, err := http.NewRequest("GET", params.ArchiveURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Accept", "application/x-tar")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tempFile, err := ioutil.TempFile("", "archive-zip")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := ioutil.WriteFile(tempFile.Name(), body, 0600); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tempFile.Name())

		cmd := exec.Command("tar", "x", "-C", tempDir, "-f", tempFile.Name())
		if out, err := cmd.CombinedOutput(); err != nil {
			http.Error(w, fmt.Sprintf("%s\n\n%s", err, out), http.StatusInternalServerError)
			return
		}
	}

	{
		// HACK: Many Gemfiles assume that the current directory is a Git repository (they run `git
		// ls-files`). Fake this.
		if err := os.Mkdir(filepath.Join(tempDir, ".git"), 0700); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := ioutil.WriteFile(filepath.Join(tempDir, ".git", "HEAD"), []byte("ref: refs/heads/master\n"), 0600); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.Mkdir(filepath.Join(tempDir, ".git", "objects"), 0700); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.Mkdir(filepath.Join(tempDir, ".git", "refs"), 0700); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Another solution... Also, alpine doesn't include Git, so make a fake `git` binary.
		tempPathDir, err := ioutil.TempDir("", "git-path")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(tempPathDir)
		if err := ioutil.WriteFile(filepath.Join(tempPathDir, "git"), []byte(`#!/bin/sh
find # mimic 'git ls-files'
`), 0700); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+tempPathDir); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	workDir := filepath.Join(tempDir, params.Dir)
	result := Result{
		Commands: make([]CommandResult, len(params.Commands)),
		Files:    make(map[string]string, len(params.IncludeFiles)),
	}
	for i, args := range params.Commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			result.Commands[i].Error = fmt.Sprintf("%s (command: %v)", err, args)
			log.Printf("Error running command %v in %q:%s\n%s", args, params.ArchiveURL, err, out)
		}
		result.Commands[i].CombinedOutput = string(out)
		result.Commands[i].Ok = err == nil
	}

	for _, includeFile := range params.IncludeFiles {
		// TODO!(sqs): ensure file is inside our workdir to avoid security problem of exposing all
		// file contents
		includeFile = filepath.Clean(includeFile)
		data, err := ioutil.ReadFile(filepath.Join(tempDir, params.Dir, includeFile))
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result.Files[includeFile] = string(data)
	}

	respBody, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(respBody)
	w.Write([]byte("\n"))
}

func main() {
	log.Print("ruby-bundler-exec started")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
