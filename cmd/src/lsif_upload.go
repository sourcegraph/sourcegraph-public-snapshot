package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/efritz/pentimento"
	"github.com/mattn/go-isatty"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/sourcegraph/codeintelutils"
	"github.com/sourcegraph/src-cli/internal/codeintel"
)

func init() {
	usage := `
Examples:

  Upload an LSIF dump with explicit repo, commit, and upload files:

    	$ src lsif upload -repo=FOO -commit=BAR -file=dump.lsif

  Upload an LSIF dump for a subproject:

    	$ src lsif upload -root=cmd/

  Upload an LSIF dump when lsifEnforceAuth is enabled:

    	$ src lsif upload -github-token=BAZ

  Upload an LSIF dump when the LSIF indexer does not not declare a tool name.

    	$ src lsif upload -indexer=lsif-elixir
`

	var flags struct {
		repo                 *string
		commit               *string
		file                 *string
		root                 *string
		indexer              *string
		gitHubToken          *string
		open                 *bool
		json                 *bool
		noProgress           *bool
		maxPayloadSizeMb     *int
		ignoreUploadFailures *bool
		uploadRoute          *string
	}

	flagSet := flag.NewFlagSet("upload", flag.ExitOnError)
	flags.repo = flagSet.String("repo", "", `The name of the repository (e.g. github.com/gorilla/mux). By default, derived from the origin remote.`)
	flags.commit = flagSet.String("commit", "", `The 40-character hash of the commit. Defaults to the currently checked-out commit.`)
	flags.root = flagSet.String("root", "", `The path in the repository that matches the LSIF projectRoot (e.g. cmd/project1). Defaults to the directory where the dump file is located.`)
	flags.file = flagSet.String("file", "./dump.lsif", `The path to the LSIF dump file.`)
	flags.indexer = flagSet.String("indexer", "", `The name of the indexer that generated the dump. This will override the 'toolInfo.name' field in the metadata vertex of the LSIF dump file. This must be supplied if the indexer does not set this field (in which case the upload will fail with an explicit message).`)
	flags.gitHubToken = flagSet.String("github-token", "", `A GitHub access token with 'public_repo' scope that Sourcegraph uses to verify you have access to the repository.`)
	flags.open = flagSet.Bool("open", false, `Open the LSIF upload page in your browser.`)
	flags.json = flagSet.Bool("json", false, `Output relevant state in JSON on success.`)
	flags.noProgress = flagSet.Bool("no-progress", false, `Do not display a progress bar.`)
	flags.maxPayloadSizeMb = flagSet.Int("max-payload-size", 100, `The maximum upload size (in megabytes). Indexes exceeding this limit will be uploaded over multiple HTTP requests.`)
	flags.ignoreUploadFailures = flagSet.Bool("ignore-upload-failure", false, `Exit with status code zero on upload failure.`)
	flags.uploadRoute = flagSet.String("upload-route", "/.api/lsif/upload", "The path of the upload route.")

	parseAndValidateFlags := func(args []string) error {
		flagSet.Parse(args)

		type inferError struct {
			argument string
			err      error
		}
		var inferErrors []inferError

		if _, err := os.Stat(*flags.file); os.IsNotExist(err) {
			inferErrors = append(inferErrors, inferError{"file", err})
		}

		if *flags.repo == "" {
			if repo, err := codeintel.InferRepo(); err != nil {
				inferErrors = append(inferErrors, inferError{"repo", err})
			} else {
				flags.repo = &repo
			}
		}

		if *flags.commit == "" {
			if commit, err := codeintel.InferCommit(); err != nil {
				inferErrors = append(inferErrors, inferError{"commit", err})
			} else {
				flags.commit = &commit
			}
		}

		if !isFlagSet(flagSet, "root") {
			if root, err := codeintel.InferRoot(*flags.file); err != nil {
				inferErrors = append(inferErrors, inferError{"root", err})
			} else {
				flags.root = &root
			}
		}
		*flags.root = codeintel.SanitizeRoot(*flags.root)

		if *flags.indexer == "" {
			file, err := os.Open(*flags.file)
			if err != nil {
				inferErrors = append(inferErrors, inferError{"indexer", err})
			}
			defer file.Close()

			if indexer, err := codeintelutils.ReadIndexerName(file); err != nil {
				inferErrors = append(inferErrors, inferError{"indexer", err})
			} else {
				flags.indexer = &indexer
			}
		}

		argsString := strings.Join([]string{
			"Inferred arguments:",
			fmt.Sprintf("  -repo=%s", *flags.repo),
			fmt.Sprintf("  -commit=%s", *flags.commit),
			fmt.Sprintf("  -root=%s", *flags.root),
			fmt.Sprintf("  -file=%s", *flags.file),
			fmt.Sprintf("  -indexer=%s", *flags.indexer),
			"",
		}, "\n")

		for _, v := range inferErrors {
			return errors.New(strings.Join([]string{
				fmt.Sprintf("error: %s", v.err),
				fmt.Sprintf("Unable to determine %s from environment. Either cd into a git repository or set -%s explicitly.", v.argument, v.argument),
				argsString,
			}, "\n\n"))
		}

		if strings.HasPrefix(*flags.root, "..") {
			return errors.New("root must not be outside of repository")
		}

		if *flags.maxPayloadSizeMb <= 0 {
			return errors.New("max-payload-size must be positive")
		}

		if !*flags.json {
			fmt.Println(argsString)
		}

		return nil
	}

	handler := func(args []string) error {
		if err := parseAndValidateFlags(args); err != nil {
			return &usageError{err}
		}

		opts := codeintel.UploadIndexOpts{
			Endpoint:             cfg.Endpoint,
			AccessToken:          cfg.AccessToken,
			AdditionalHeaders:    cfg.AdditionalHeaders,
			Path:                 *flags.uploadRoute,
			Repo:                 *flags.repo,
			Commit:               *flags.commit,
			Root:                 *flags.root,
			Indexer:              *flags.indexer,
			GitHubToken:          *flags.gitHubToken,
			File:                 *flags.file,
			MaxPayloadSizeBytes:  *flags.maxPayloadSizeMb * 1000 * 1000,
			MaxRetries:           10,
			RetryInterval:        time.Millisecond * 250,
			UploadProgressEvents: make(chan codeintelutils.UploadProgressEvent),
		}

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			if *flags.json || *flags.noProgress {
				return
			}

			pentimento.PrintProgress(func(p *pentimento.Printer) error {
				for event := range opts.UploadProgressEvents {
					content := pentimento.NewContent()
					content.AddLine(formatProgressBar(event.TotalProgress, fmt.Sprintf("%d/%d", event.Part, event.NumParts)))
					p.WriteContent(content)
				}

				_ = p.Reset()
				return nil
			})
		}()

		uploadID, err := codeintel.UploadIndex(opts)
		close(opts.UploadProgressEvents) // Stop progress bar updates
		wg.Wait()                        // Wait for progress bar goroutine to clear screen
		if err != nil {
			if err == codeintelutils.ErrUnauthorized {
				if *flags.gitHubToken == "" {
					return fmt.Errorf("you must provide -github-token=TOKEN, where TOKEN is a GitHub personal access token with 'repo' or 'public_repo' scope")
				}

				if isatty.IsTerminal(os.Stdout.Fd()) {
					fmt.Println("You may need to specify or update your GitHub access token to use this endpoint.")
					fmt.Println("See https://github.com/sourcegraph/src-cli#authentication.")
				}
			}

			if *flags.ignoreUploadFailures {
				fmt.Printf("error: %s\n", err)
				return nil
			}

			return err
		}

		uploadURL := fmt.Sprintf("%s/%s/-/settings/code-intelligence/lsif-uploads/%s", cfg.Endpoint, *flags.repo, uploadID)

		if *flags.json {
			serialized, err := json.Marshal(map[string]interface{}{
				"repo":      *flags.repo,
				"commit":    *flags.commit,
				"root":      *flags.root,
				"file":      *flags.file,
				"indexer":   *flags.indexer,
				"uploadId":  uploadID,
				"uploadUrl": uploadURL,
			})
			if err != nil {
				return err
			}

			fmt.Println(string(serialized))
		} else {
			fmt.Printf("LSIF dump successfully uploaded for processing.\n")
			fmt.Printf("View processing status at %s.\n", uploadURL)
		}

		if *flags.open {
			if err := browser.OpenURL(uploadURL); err != nil {
				return err
			}
		}

		return nil
	}

	lsifCommands = append(lsifCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src lsif %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

func isFlagSet(fs *flag.FlagSet, name string) (found bool) {
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

// maxDisplayWidth is the number of columns that can be used to draw a progress bar.
const maxDisplayWidth = 80

// formatProgressBar draws a progress bar with the given percentage complete and the
// given literal suffix.
func formatProgressBar(progress float64, suffix string) string {
	if len(suffix) > 0 {
		suffix = " " + suffix
	}

	maxWidth := maxDisplayWidth - 3 - len(suffix)
	width := int(float64(maxWidth) * float64(progress))

	var arrow string
	if width < maxWidth {
		arrow = ">"
	}

	return fmt.Sprintf(
		"[%s%s%s]%s",
		strings.Repeat("=", width),
		arrow,
		strings.Repeat(" ", maxWidth-width-len(arrow)),
		suffix,
	)
}

// digits returns the number of digits of n.
func digits(n int) int {
	if n >= 10 {
		return 1 + digits(n/10)
	}
	return 1
}
