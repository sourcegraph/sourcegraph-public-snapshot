package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/upload"
	"github.com/sourcegraph/src-cli/internal/codeintel"
)

var lsifUploadFlags struct {
	file string

	// UploadRecordOptions
	repo              string
	commit            string
	root              string
	indexer           string
	associatedIndexID int

	// SourcegraphInstanceOptions
	uploadRoute      string
	gitHubToken      string
	maxPayloadSizeMb int64

	// Output and error behavior
	ignoreUploadFailures bool
	noProgress           bool
	verbosity            int
	json                 bool
	open                 bool
}

var lsifUploadFlagSet = flag.NewFlagSet("upload", flag.ExitOnError)

func init() {
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.file, "file", "./dump.lsif", `The path to the LSIF dump file.`)

	// UploadRecordOptions
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.repo, "repo", "", `The name of the repository (e.g. github.com/gorilla/mux). By default, derived from the origin remote.`)
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.commit, "commit", "", `The 40-character hash of the commit. Defaults to the currently checked-out commit.`)
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.root, "root", "", `The path in the repository that matches the LSIF projectRoot (e.g. cmd/project1). Defaults to the directory where the dump file is located.`)
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.indexer, "indexer", "", `The name of the indexer that generated the dump. This will override the 'toolInfo.name' field in the metadata vertex of the LSIF dump file. This must be supplied if the indexer does not set this field (in which case the upload will fail with an explicit message).`)
	lsifUploadFlagSet.IntVar(&lsifUploadFlags.associatedIndexID, "associated-index-id", -1, "ID of the associated index record for this upload. For internal use only.")

	// SourcegraphInstanceOptions
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.uploadRoute, "upload-route", "/.api/lsif/upload", "The path of the upload route. For internal use only.")
	lsifUploadFlagSet.StringVar(&lsifUploadFlags.gitHubToken, "github-token", "", `A GitHub access token with 'public_repo' scope that Sourcegraph uses to verify you have access to the repository.`)
	lsifUploadFlagSet.Int64Var(&lsifUploadFlags.maxPayloadSizeMb, "max-payload-size", 100, `The maximum upload size (in megabytes). Indexes exceeding this limit will be uploaded over multiple HTTP requests.`)

	// Output and error behavior
	lsifUploadFlagSet.BoolVar(&lsifUploadFlags.ignoreUploadFailures, "ignore-upload-failure", false, `Exit with status code zero on upload failure.`)
	lsifUploadFlagSet.BoolVar(&lsifUploadFlags.noProgress, "no-progress", false, `Do not display progress updates.`)
	lsifUploadFlagSet.IntVar(&lsifUploadFlags.verbosity, "trace", 0, "-trace=0 shows no logs; -trace=1 shows requests and response metadata; -trace=2 shows headers, -trace=3 shows response body")
	lsifUploadFlagSet.BoolVar(&lsifUploadFlags.json, "json", false, `Output relevant state in JSON on success.`)
	lsifUploadFlagSet.BoolVar(&lsifUploadFlags.open, "open", false, `Open the LSIF upload page in your browser.`)
}

// parseAndValidateLSIFUploadFlags calls lsifUploadFlagSet.Parse, then infers values for
// missing flags, normalizes supplied values, and validates the state of the lsifUploadFlags
// object.
//
// On success, the global lsifUploadFlags object will be populated with valid values. An
// error is returned on failure.
func parseAndValidateLSIFUploadFlags(args []string) error {
	if err := lsifUploadFlagSet.Parse(args); err != nil {
		return err
	}

	if inferenceErrors := inferMissingLSIFUploadFlags(); len(inferenceErrors) > 0 {
		return errorWithHint{
			err: inferenceErrors[0].err, hint: strings.Join([]string{
				fmt.Sprintf(
					"Unable to determine %s from environment. Check your working directory or supply -%s={value} explicitly",
					inferenceErrors[0].argument,
					inferenceErrors[0].argument,
				),
			}, "\n"),
		}
	}

	if err := validateLSIFUploadFlags(); err != nil {
		return err
	}

	return nil
}

type argumentInferenceError struct {
	argument string
	err      error
}

// inferMissingLSIFUploadFlags updates the flags values which were not explicitly
// supplied by the user with default values inferred from the current git state and
// filesystem.
//
// Note: This function must not be called before lsifUploadFlagSet.Parse.
func inferMissingLSIFUploadFlags() (inferErrors []argumentInferenceError) {
	if _, err := os.Stat(lsifUploadFlags.file); os.IsNotExist(err) {
		inferErrors = append(inferErrors, argumentInferenceError{"file", err})
	}

	if err := inferUnsetFlag("repo", &lsifUploadFlags.repo, codeintel.InferRepo); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("commit", &lsifUploadFlags.commit, codeintel.InferCommit); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("root", &lsifUploadFlags.root, inferIndexRoot); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("indexer", &lsifUploadFlags.indexer, readIndexerName); err != nil {
		inferErrors = append(inferErrors, *err)
	}

	return inferErrors
}

// inferUnsetFlag conditionally updates the value of the given pointer with the
// return value of the given function. If the flag with the given name was supplied
// by the user, then this function no-ops. An argumentInferenceError is returned if
// the given function returns an error.
//
// Note: This function must not be called before lsifUploadFlagSet.Parse.
func inferUnsetFlag(name string, target *string, f func() (string, error)) *argumentInferenceError {
	if isFlagSet(lsifUploadFlagSet, name) {
		return nil
	}

	value, err := f()
	if err != nil {
		return &argumentInferenceError{name, err}
	}

	*target = value
	return nil
}

// isFlagSet returns true if the flag with the given name was supplied by the user.
// This lets us distinguish between zero-values (empty strings) and void values without
// requiring pointers and adding a layer of indirection deeper in the program.
func isFlagSet(fs *flag.FlagSet, name string) (found bool) {
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

// inferIndexRoot returns the root directory based on the configured index file path.
//
// Note: This function must not be called before lsifUploadFlagSet.Parse.
func inferIndexRoot() (string, error) {
	return codeintel.InferRoot(lsifUploadFlags.file)
}

// readIndexerName returns the indexer name read from the configured index file.
//
// Note: This function must not be called before lsifUploadFlagSet.Parse.
func readIndexerName() (string, error) {
	file, err := os.Open(lsifUploadFlags.file)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return upload.ReadIndexerName(file)
}

// validateLSIFUploadFlags returns an error if any of the parsed flag values are illegal.
//
// Note: This function must not be called before lsifUploadFlagSet.Parse.
func validateLSIFUploadFlags() error {
	lsifUploadFlags.root = codeintel.SanitizeRoot(lsifUploadFlags.root)

	if strings.HasPrefix(lsifUploadFlags.root, "..") {
		return errors.New("root must not be outside of repository")
	}

	if lsifUploadFlags.maxPayloadSizeMb < 25 {
		return errors.New("max-payload-size must be at least 25 (MB)")
	}

	return nil
}
