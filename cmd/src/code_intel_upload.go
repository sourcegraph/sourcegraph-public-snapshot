package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/upload"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:
  Before running any of these, first use src auth to authenticate.
  Alternately, use the SRC_ACCESS_TOKEN environment variable for
  individual src-cli invocations. 

  If run from within the project itself, src-cli will infer various
  flags based on git metadata.

        $ src code-intel upload # uploads ./index.scip

  If src-cli is invoked outside the project root, or if you're using
  a version control system other than git, specify flags explicitly:

    	$ src code-intel upload -root='' -repo=FOO -commit=BAR -file=index.scip

  Upload a SCIP index for a subproject:

    	$ src code-intel upload -root=cmd/

  Upload a SCIP index when lsif.enforceAuth is enabled in site settings:

    	$ src code-intel upload -github-token=BAZ, or
    	$ src code-intel upload -gitlab-token=BAZ

  For any of these commands, an LSIF index (default name: dump.lsif) can be
  used instead of a SCIP index (default name: index.scip).
`
	codeintelCommands = append(codeintelCommands, &command{
		flagSet: codeintelUploadFlagSet,
		handler: handleCodeIntelUpload,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src code-intel %s':\n", codeintelUploadFlagSet.Name())
			codeintelUploadFlagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})

	// Make 'upload' available under 'src lsif' for backwards compatibility.
	lsifCommands = append(lsifCommands, &command{
		flagSet: codeintelUploadFlagSet,
		handler: handleCodeIntelUpload,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src lsif %s':\n", codeintelUploadFlagSet.Name())
			codeintelUploadFlagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

// handleCodeIntelUpload is the handler for `src code-intel upload`.
func handleCodeIntelUpload(args []string) error {
	ctx := context.Background()

	out, isSCIPAvailable, err := parseAndValidateCodeIntelUploadFlags(args)
	if !codeintelUploadFlags.json {
		if out != nil {
			printInferredArguments(out)
		} else {
			// Always display inferred arguments except when -json is set
			printInferredArguments(emergencyOutput())
		}
	}
	if err != nil {
		return handleUploadError(nil, err)
	}

	client := api.NewClient(api.ClientOpts{
		Out:   io.Discard,
		Flags: codeintelUploadFlags.apiFlags,
	})

	uploadID, err := upload.UploadIndex(ctx, codeintelUploadFlags.file, client, codeintelUploadOptions(out, isSCIPAvailable))
	if err != nil {
		return handleUploadError(out, err)
	}

	uploadURL, err := makeCodeIntelUploadURL(uploadID)
	if err != nil {
		return err
	}

	if codeintelUploadFlags.json {
		serialized, err := json.Marshal(map[string]interface{}{
			"repo":           codeintelUploadFlags.repo,
			"commit":         codeintelUploadFlags.commit,
			"root":           codeintelUploadFlags.root,
			"file":           codeintelUploadFlags.file,
			"indexer":        codeintelUploadFlags.indexer,
			"indexerVersion": codeintelUploadFlags.indexerVersion,
			"uploadId":       uploadID,
			"uploadUrl":      uploadURL,
		})
		if err != nil {
			return err
		}

		fmt.Println(string(serialized))
	} else {
		if out == nil {
			out = emergencyOutput()
		}

		out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleItalic, "View processing status at %s", uploadURL))
	}

	if codeintelUploadFlags.open {
		if err := browser.OpenURL(uploadURL); err != nil {
			return err
		}
	}

	return nil
}

// codeintelUploadOptions creates a set of upload options given the values in the flags.
func codeintelUploadOptions(out *output.Output, isSCIPAvailable bool) upload.UploadOptions {
	var associatedIndexID *int
	if codeintelUploadFlags.associatedIndexID != -1 {
		associatedIndexID = &codeintelUploadFlags.associatedIndexID
	}

	cfg.AdditionalHeaders["Content-Type"] = "application/x-ndjson+lsif"
	path := codeintelUploadFlags.uploadRoute
	if isSCIPAvailable && filepath.Ext(codeintelUploadFlags.file) == ".scip" {
		cfg.AdditionalHeaders["Content-Type"] = "application/x-protobuf+scip"
		path = strings.ReplaceAll(path, "lsif", "scip")
	}

	logger := upload.NewRequestLogger(
		os.Stdout,
		// Don't need to check upper bounds as we only compare verbosity ranges
		// It's fine if someone supplies -trace=42, but it will just behave the
		// same as if they supplied the highest verbosity level we define
		// internally.
		upload.RequestLoggerVerbosity(codeintelUploadFlags.verbosity),
	)

	return upload.UploadOptions{
		UploadRecordOptions: upload.UploadRecordOptions{
			Repo:              codeintelUploadFlags.repo,
			Commit:            codeintelUploadFlags.commit,
			Root:              codeintelUploadFlags.root,
			Indexer:           codeintelUploadFlags.indexer,
			IndexerVersion:    codeintelUploadFlags.indexerVersion,
			AssociatedIndexID: associatedIndexID,
		},
		SourcegraphInstanceOptions: upload.SourcegraphInstanceOptions{
			SourcegraphURL:      cfg.Endpoint,
			AccessToken:         cfg.AccessToken,
			AdditionalHeaders:   cfg.AdditionalHeaders,
			MaxRetries:          5,
			RetryInterval:       time.Second,
			Path:                path,
			MaxPayloadSizeBytes: codeintelUploadFlags.maxPayloadSizeMb * 1000 * 1000,
			MaxConcurrency:      codeintelUploadFlags.maxConcurrency,
			GitHubToken:         codeintelUploadFlags.gitHubToken,
			GitLabToken:         codeintelUploadFlags.gitLabToken,
		},
		OutputOptions: upload.OutputOptions{
			Output: out,
			Logger: logger,
		},
	}
}

// printInferredArguments prints a block showing the effective values of flags that are
// inferrably defined. This function is called on all paths except for -json uploads. This
// function no-ops if the given output object is nil.
func printInferredArguments(out *output.Output) {
	if out == nil {
		return
	}

	block := out.Block(output.Line(output.EmojiLightbulb, output.StyleItalic, "Inferred arguments"))
	block.Writef("repo: %s", codeintelUploadFlags.repo)
	block.Writef("commit: %s", codeintelUploadFlags.commit)
	block.Writef("root: %s", codeintelUploadFlags.root)
	block.Writef("file: %s", codeintelUploadFlags.file)
	block.Writef("indexer: %s", codeintelUploadFlags.indexer)
	block.Writef("indexerVersion: %s", codeintelUploadFlags.indexerVersion)
	block.Close()
}

// makeCodeIntelUploadURL constructs a URL to the upload with the given internal identifier.
// The base of the URL is constructed from the configured Sourcegraph instance.
func makeCodeIntelUploadURL(uploadID int) (string, error) {
	url, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return "", err
	}

	graphqlID := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(`LSIFUpload:%d`, uploadID)))
	url.Path = codeintelUploadFlags.repo + "/-/code-intelligence/uploads/" + graphqlID
	url.User = nil
	return url.String(), nil
}

type errorWithHint struct {
	err  error
	hint string
}

func (e errorWithHint) Error() string {
	return fmt.Sprintf("%s\n\n%s\n", e.err, e.hint)
}

// handleUploadError writes the given error to the given output. If the
// given output object is nil then the error will be written to standard out.
//
// This method returns the error that should be passed back up to the runner.
func handleUploadError(out *output.Output, err error) error {
	if err == upload.ErrUnauthorized {
		err = filterLSIFUnauthorizedError(out, err)
	}

	if codeintelUploadFlags.ignoreUploadFailures {
		// Report but don't return the error
		fmt.Println(err.Error())
		return nil
	}

	return err
}

func filterLSIFUnauthorizedError(out *output.Output, err error) error {
	var actionableHints []string
	needsGitHubToken := strings.HasPrefix(codeintelUploadFlags.repo, "github.com")
	needsGitLabToken := strings.HasPrefix(codeintelUploadFlags.repo, "gitlab.com")

	if needsGitHubToken {
		if codeintelUploadFlags.gitHubToken != "" {
			actionableHints = append(actionableHints,
				fmt.Sprintf("The supplied -github-token does not indicate that you have collaborator access to %s.", codeintelUploadFlags.repo),
				"Please check the value of the supplied token and its permissions on the code host and try again.",
			)
		} else {
			actionableHints = append(actionableHints,
				fmt.Sprintf("Please retry your request with a -github-token=XXX with collaborator access to %s.", codeintelUploadFlags.repo),
				"This token will be used to check with the code host that the uploading user has write access to the target repository.",
			)
		}
	} else if needsGitLabToken {
		if codeintelUploadFlags.gitLabToken != "" {
			actionableHints = append(actionableHints,
				fmt.Sprintf("The supplied -gitlab-token does not indicate that you have write access to %s.", codeintelUploadFlags.repo),
				"Please check the value of the supplied token and its permissions on the code host and try again.",
			)
		} else {
			actionableHints = append(actionableHints,
				fmt.Sprintf("Please retry your request with a -gitlab-token=XXX with write access to %s.", codeintelUploadFlags.repo),
				"This token will be used to check with the code host that the uploading user has write access to the target repository.",
			)
		}
	} else {
		actionableHints = append(actionableHints,
			"Verification is supported for the following code hosts: github.com, gitlab.com.",
			"Please request support for additional code host verification at https://github.com/sourcegraph/sourcegraph/issues/4967.",
		)
	}

	return errorWithHint{err: err, hint: strings.Join(mergeStringSlices(
		[]string{"This Sourcegraph instance has enforced auth for SCIP/LSIF uploads."},
		actionableHints,
		[]string{"For more details, see https://docs.sourcegraph.com/cli/references/code-intel/upload."},
	), "\n")}
}

// emergencyOutput creates a default Output object writing to standard out.
func emergencyOutput() *output.Output {
	return output.NewOutput(os.Stdout, output.OutputOpts{})
}

func mergeStringSlices(ss ...[]string) []string {
	var combined []string
	for _, s := range ss {
		combined = append(combined, s...)
	}

	return combined
}
