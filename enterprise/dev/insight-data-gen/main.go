package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

type DefinitionFile struct {
	Definitions []Definition `json:"definitions,omitempty"`
}

type Definition struct {
	RepoPath string             `json:"repo_path,omitempty"`
	Symbols  []SymbolDefinition `json:"symbols,omitempty"`
}

type SymbolDefinition struct {
	Symbol    string     `json:"symbol,omitempty"`
	Snapshots []Snapshot `json:"snapshots,omitempty"`
}

type Snapshot struct {
	Instant time.Time `json:"instant"`
	Count   int       `json:"count,omitempty"`
}

type SymbolCount struct {
	Symbol string
	count  int
}

type SnapshotContent struct {
	Repo    string
	Instant time.Time
	Symbols []SymbolCount
}

var generatedFilename = "/files/findme.txt"
var generatedFolder = "/files"

var inputFile = flag.String("manifest", "", "path to a manifest json file describing what should be generated")

func main() {
	flag.Parse()
	log.SetFlags(0)

	if inputFile == nil || len(*inputFile) == 0 {
		log.Fatal(errors.Errorf("unable to read manifest file: %v", *inputFile))
	}

	jsonFile, err := os.Open(*inputFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "unable to open manifest file"))
	}
	log.Printf("Generating from manifest %v...", jsonFile.Name())

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "unable to read manifest file"))
	}

	var definitionFile DefinitionFile
	err = json.Unmarshal(bytes, &definitionFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "unable to unmarshal json from manifest file"))
	}

	contents := make([]SnapshotContent, 0)
	for _, repoDef := range definitionFile.Definitions {
		contents = append(contents, mapToDates(repoDef.RepoPath, repoDef.Symbols)...)
	}

	sort.Slice(contents, func(i, j int) bool {
		return contents[i].Instant.Before(contents[j].Instant)
	})

	for _, snapshot := range contents {
		content := generateFileContent(snapshot)

		err := preparePath(snapshot)
		if err != nil {
			log.Fatal(errors.Wrap(err, "unable to prepare repo path"))
		}
		path := buildPath(snapshot)

		log.Printf("Writing content: %v @ %v", path, snapshot.Instant)
		err = writeContent(path, content)
		if err != nil {
			log.Fatal(err)
		}
		err = inDir(snapshot.Repo, func() error {
			log.Printf("Adding...")
			err := run("git", "add", "files/findme.txt")
			if err != nil {
				return errors.Wrap(err, "failed to add file to git repository")
			}
			log.Printf("Committing...")
			err = commit(snapshot.Instant)
			if err != nil {
				return errors.Wrap(err, "failed to commit to git repository")
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func commit(commitTime time.Time) error {
	AD := fmt.Sprintf("GIT_AUTHOR_DATE=\"%s\"", commitTime.Format(time.RFC3339))
	CD := fmt.Sprintf("GIT_COMMITTER_DATE=\"%s\"", commitTime.Format(time.RFC3339))
	return runWithEnv(AD, CD)("git", "commit", "-m", "autogen", "--allow-empty")
}

func mapToDates(repo string, defs []SymbolDefinition) []SnapshotContent {
	mapped := make(map[time.Time][]SymbolCount)
	for _, def := range defs {
		text := def.Symbol
		for _, snapshot := range def.Snapshots {
			mapped[snapshot.Instant] = append(mapped[snapshot.Instant], SymbolCount{
				Symbol: text,
				count:  snapshot.Count,
			})
		}
	}

	results := make([]SnapshotContent, 0)
	for at, counts := range mapped {
		results = append(results, SnapshotContent{
			Instant: at,
			Symbols: counts,
			Repo:    repo,
		})
	}

	return results
}

func preparePath(snapshot SnapshotContent) error {
	if _, err := os.Stat(snapshot.Repo + generatedFolder); errors.Is(err, os.ErrNotExist) {
		// the race here is fine
		log.Printf("Creating path: %v", snapshot.Repo+generatedFolder)
		return os.MkdirAll(snapshot.Repo+generatedFolder, 0755)
	}
	log.Printf("path found: %v", snapshot.Repo+generatedFolder)
	return nil
}

func buildPath(snapshot SnapshotContent) string {
	return snapshot.Repo + generatedFilename
}

func writeContent(path string, content string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	_, err = f.WriteString(content)
	if err != nil {
		return errors.Wrap(err, "failed to write string content")
	}
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "failed to close file")
	}
	return nil
}

func generateFileContent(snapshot SnapshotContent) string {
	var b strings.Builder
	for _, symbol := range snapshot.Symbols {
		for i := 0; i < symbol.count; i++ {
			b.WriteString(fmt.Sprintf("%s\n", symbol.Symbol))
		}
	}
	return b.String()
}

// run executes an external command.
func run(args ...string) error {
	return runWithEnv()(args...)
}

// runWithEnv executes an external command with additional environment variables given as variable=value strings
func runWithEnv(vars ...string) func(args ...string) error {
	return func(args ...string) error {
		cmd := exec.Command(args[0], args[1:]...)

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, vars...)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "output: %s", out)
		}
		return nil
	}
}

// inDir runs function f in directory d.
func inDir(d string, f func() error) (err error) {
	d0, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "getting working dir: %s", d0)
	}
	defer func() {
		if chdirErr := os.Chdir(d0); chdirErr != nil {
			err = errors.Wrapf(chdirErr, "changing dir to %s", d0)
		}
	}()
	if err := os.Chdir(d); err != nil {
		return errors.Wrapf(err, "changing dir to %s", d)
	}
	return f()
}
