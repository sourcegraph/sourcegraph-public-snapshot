package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/campaigns"
)

const defaultTimeout = 60 * time.Minute

func init() {
	usage := `
Execute an action on code in repositories. The output of an action is a set of patches that can be used to create a campaign to open changesets and perform large-scale code changes.

Examples:

  Execute an action defined in ~/run-gofmt.json and save the patches it produced to 'patches.json'

	$ src actions exec -f ~/run-gofmt.json

  Execute an action and create a patch set from the produced patches:

	$ src actions exec -f ~/run-gofmt.json -create-patchset

  Verbosely execute an action and keep the logs available for debugging:

	$ src -v actions exec -keep-logs -f ~/run-gofmt.json

  Execute an action and pipe the patches it produced to 'src campaign patchset create-from-patches':

	$ src actions exec -f ~/run-gofmt.json | src campaign patchset create-from-patches

  Execute an action and save the patches it produced to 'patches.json'

	$ src actions exec -f ~/run-gofmt.json -o patches.json 

  Read and execute an action definition from standard input:

	$ cat ~/my-action.json | src actions exec -f -


Format of the action JSON files:

	An action JSON needs to specify:

	- "scopeQuery" - a Sourcegraph search query to generate a list of repositories over which to run the action. Use 'src actions scope-query' to see which repositories are matched by the query
	- "steps" - a list of action steps to execute in each repository

	A single "step" can either be a of type "command", which means the step is executed on the machine on which 'src actions exec' is executed, or it can be of type "docker" which then (optionally builds) and runs a container in which the repository is mounted.

	This action has a single step that produces a README.md file in repositories whose name starts with "go-" and that doesn't have a README.md file yet:

		{
		  "scopeQuery": "repo:go-* -repohasfile:README.md",
		  "steps": [
		    {
		      "type": "command",
		      "args": ["sh", "-c", "echo '# README' > README.md"]
		    }
		  ]
		}

	This action runs a single step over repositories whose name contains "github", building and starting a Docker container based on the image defined through the "dockerfile". In the container the word 'this' is replaced with 'that' in all text files.

		{
		  "scopeQuery": "repo:github",
		  "steps": [
		    {
		      "type": "docker",
		      "image": "alpine:3",
			  "args": ["sh", "-c", "find /work -iname '*.txt' -type f | xargs -n 1 sed -i s/this/that/g"]
		    }
		  ]
		}

`

	flagSet := flag.NewFlagSet("exec", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src actions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}

	cacheDir, _ := campaigns.UserCacheDir()
	if cacheDir != "" {
		cacheDir = filepath.Join(cacheDir, "action-exec")
	}

	displayUserCacheDir := strings.Replace(cacheDir, os.Getenv("HOME"), "$HOME", 1)

	var (
		fileFlag        = flagSet.String("f", "-", "The action file. If not given or '-' standard input is used. (Required)")
		outputFlag      = flagSet.String("o", "patches.json", "The output file. Will be used as the destination for patches unless the command is being piped in which case patches are piped to stdout")
		parallelismFlag = flagSet.Int("j", runtime.GOMAXPROCS(0), "The number of parallel jobs.")

		cacheDirFlag   = flagSet.String("cache", displayUserCacheDir, "Directory for caching results.")
		clearCacheFlag = flagSet.Bool("clear-cache", false, "Remove possibly cached results for an action before executing it.")

		keepLogsFlag = flagSet.Bool("keep-logs", false, "Do not remove execution log files when done.")
		timeoutFlag  = flagSet.Duration("timeout", defaultTimeout, "The maximum duration a single action run can take.")

		createPatchSetFlag      = flagSet.Bool("create-patchset", false, "Create a patch set from the produced set of patches. When the execution of the action fails in a single repository a prompt will ask to confirm or reject the patch set creation.")
		forceCreatePatchSetFlag = flagSet.Bool("force-create-patchset", false, "Force creation of patch set from the produced set of patches, without asking for confirmation even when the execution of the action failed for a subset of repositories.")

		includeUnsupportedFlag = flagSet.Bool("include-unsupported", false, "When specified, also repos from unsupported codehosts are processed. Those can be created once the integration is done.")

		apiFlags = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		if !isGitAvailable() {
			return errors.New("Could not find git in $PATH. 'src actions exec' requires git to be available.")
		}

		if *cacheDirFlag == displayUserCacheDir {
			*cacheDirFlag = cacheDir
		}

		if *cacheDirFlag == "" {
			// This can only happen if `userCacheDir()` fails or the user
			// specifies a blank string.
			return errors.New("cache is not a valid path")
		}

		// Read action file content.
		var actionFile []byte
		if *fileFlag == "-" {
			actionFile, err = ioutil.ReadAll(os.Stdin)
		} else {
			actionFile, err = ioutil.ReadFile(*fileFlag)
		}
		if err != nil {
			return err
		}

		var outputWriter io.Writer
		if !*createPatchSetFlag && !*forceCreatePatchSetFlag {
			// If stdout is a pipe, write to pipe, otherwise
			// write to output file
			fi, err := os.Stdout.Stat()
			if err != nil {
				return err
			}
			isPipe := fi.Mode()&os.ModeCharDevice == 0

			if isPipe {
				outputWriter = os.Stdout
			} else {
				f, err := os.Create(*outputFlag)
				if err != nil {
					return errors.Wrap(err, "creating output file")
				}
				defer f.Close()
				outputWriter = f
			}
		}

		// Convert action file to JSON.
		jsonActionFile, err := yaml.YAMLToJSONStrict(actionFile)
		if err != nil {
			return errors.Wrap(err, "unable to parse action file")
		}

		err = campaigns.ValidateActionDefinition(jsonActionFile)
		if err != nil {
			return err
		}

		var action campaigns.Action
		if err := jsonxUnmarshal(string(jsonActionFile), &action); err != nil {
			return errors.Wrap(err, "invalid JSON action file")
		}

		ctx, cancel := context.WithCancel(context.Background())
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		defer func() {
			signal.Stop(c)
			cancel()
		}()
		go func() {
			select {
			case <-c:
				cancel()
			case <-ctx.Done():
			}
			<-c // If user hits Ctrl-C second time, we do a hard exit
			os.Exit(2)
		}()

		logger := campaigns.NewActionLogger(*verbose, *keepLogsFlag)

		// Fetch Docker images etc.
		err = campaigns.PrepareAction(ctx, action, logger)
		if err != nil {
			return errors.Wrap(err, "Failed to prepare action")
		}

		opts := campaigns.ExecutorOpts{
			Endpoint:          cfg.Endpoint,
			AccessToken:       cfg.AccessToken,
			AdditionalHeaders: cfg.AdditionalHeaders,
			Timeout:           *timeoutFlag,
			KeepLogs:          *keepLogsFlag,
			ClearCache:        *clearCacheFlag,
			Cache:             campaigns.ExecutionDiskCache{Dir: *cacheDirFlag},
		}

		// Query repos over which to run action
		logger.Infof("Querying %s for repositories matching '%s'...\n", cfg.Endpoint, action.ScopeQuery)
		repos, err := actionRepos(ctx, action.ScopeQuery, *includeUnsupportedFlag, logger)
		if err != nil {
			return err
		}
		logger.Infof("Use 'src actions scope-query' for help with scoping.\n\n")

		totalSteps := len(repos) * len(action.Steps)
		logger.Start(totalSteps)

		executor := campaigns.NewExecutor(action, *parallelismFlag, logger, opts)
		for _, repo := range repos {
			executor.EnqueueRepo(repo)
		}

		go executor.Start(ctx)
		err = executor.Wait()

		patches := executor.AllPatches()
		if len(patches) == 0 {
			// We call os.Exit because we don't want to return the error
			// and have it printed.
			logger.ActionFailed(err, patches)
			os.Exit(1)
		}

		if !*createPatchSetFlag && !*forceCreatePatchSetFlag {
			if err != nil {
				logger.ActionFailed(err, patches)
				os.Exit(1)
			}

			err = json.NewEncoder(outputWriter).Encode(patches)
			if err != nil {
				return errors.Wrap(err, "writing patches")
			}

			logger.ActionSuccess(patches)

			if out, ok := outputWriter.(*os.File); ok && out == os.Stdout {
				// Don't print instructions when piping
				return nil
			}

			// Print instructions when we've written patches to a file, even when not in verbose mode
			fmt.Fprintf(os.Stderr, "\n\nPatches saved to %s, to create a patch set on your Sourcegraph instance please do the following:\n", *outputFlag)
			fmt.Fprintln(os.Stderr, "\n ", color.HiCyanString("▶"), fmt.Sprintf("src campaign patchset create-from-patches < %s", *outputFlag))
			fmt.Fprintln(os.Stderr)

			return nil
		}

		if err != nil {
			logger.ActionFailed(err, patches)

			if len(patches) == 0 {
				os.Exit(1)
			}

			if !*forceCreatePatchSetFlag {
				canInput := isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
				if !canInput {
					return err
				}

				c, _ := askForConfirmation("Create a patch set for the produced patches anyway?")
				if !c {
					return err
				}
			}
		} else {
			logger.ActionSuccess(patches)
		}

		tmpl, err := parseTemplate("{{friendlyPatchSetCreatedMessage .}}")
		if err != nil {
			return err
		}

		return createPatchSetFromPatches(apiFlags, patches, tmpl, 100)
	}

	// Register the command.
	actionsCommands = append(actionsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

func actionRepos(ctx context.Context, scopeQuery string, includeUnsupported bool, logger *campaigns.ActionLogger) ([]campaigns.ActionRepo, error) {
	hasCount, err := regexp.MatchString(`count:\d+`, scopeQuery)
	if err != nil {
		return nil, err
	}

	if !hasCount {
		scopeQuery = scopeQuery + " count:999999"
	}

	query := `
query ActionRepos($query: String!) {
	search(query: $query, version: V2) {
		results {
			results {
				__typename
				... on Repository {
					...repositoryFields
				}
				... on FileMatch {
					repository {
						...repositoryFields
					}
				}
			}
			...SearchResultsAlertFields
		}
	}
}

fragment repositoryFields on Repository {
	id
	name
	externalRepository {
		serviceType
	}
	defaultBranch {
		name
		target {
			oid
		}
	}
}
` + searchResultsAlertFragment

	type Repository struct {
		ID, Name           string
		ExternalRepository struct {
			ServiceType string
		}
		DefaultBranch *struct {
			Name   string
			Target struct{ OID string }
		}
	}
	var result struct {
		Data struct {
			Search struct {
				Results struct {
					Results []struct {
						Typename           string `json:"__typename"`
						ID, Name           string
						ExternalRepository struct {
							ServiceType string
						}
						DefaultBranch *struct {
							Name   string
							Target struct{ OID string }
						}
						Repository Repository `json:"repository"`
					}
					Alert searchResultsAlert
				}
			}
		} `json:"data,omitempty"`

		Errors []struct {
			Message string
			Path    []interface{}
		} `json:"errors,omitempty"`
	}

	if err := (&apiRequest{
		query: query,
		vars: map[string]interface{}{
			"query": scopeQuery,
		},
		// Do not unpack errors and return error. Instead we want to go through
		// the results and check whether they're complete.
		// If we don't do this and the query returns an error for _one_
		// repository because that is still cloning, we don't get any repositories.
		// Instead we simply want to skip those repositories that are still
		// being cloned.
		dontUnpackErrors: true,
		result:           &result,
	}).do(); err != nil {

		// Ignore exitCodeError with error == nil, because we explicitly set
		// dontUnpackErrors, which can lead to an empty exitCodeErr being
		// returned.
		exitCodeErr, ok := err.(*exitCodeError)
		if !ok {
			return nil, err
		}
		if exitCodeErr.error != nil {
			return nil, exitCodeErr
		}
	}

	skipped := []string{}
	unsupported := []string{}
	reposByID := map[string]campaigns.ActionRepo{}
	for _, searchResult := range result.Data.Search.Results.Results {

		var repo Repository
		if searchResult.Repository.ID != "" {
			repo = searchResult.Repository
		} else {
			repo = Repository{
				ID:                 searchResult.ID,
				Name:               searchResult.Name,
				ExternalRepository: searchResult.ExternalRepository,
				DefaultBranch:      searchResult.DefaultBranch,
			}
		}

		// Skip repos from unsupported code hosts but don't report them explicitly.
		if !includeUnsupported {
			ok, err := isCodeHostSupportedForCampaigns(repo.ExternalRepository.ServiceType)
			if err != nil {
				return nil, errors.Wrap(err, "failed code host check")
			}
			if !ok {
				unsupported = append(unsupported, repo.Name)
				continue
			}
		}

		if repo.DefaultBranch == nil || repo.DefaultBranch.Name == "" {
			skipped = append(skipped, repo.Name)
			continue
		}

		if repo.DefaultBranch.Target.OID == "" {
			skipped = append(skipped, repo.Name)
			continue
		}

		if _, ok := reposByID[repo.ID]; !ok {
			reposByID[repo.ID] = campaigns.ActionRepo{
				ID:      repo.ID,
				Name:    repo.Name,
				Rev:     repo.DefaultBranch.Target.OID,
				BaseRef: repo.DefaultBranch.Name,
			}
		}
	}

	repos := make([]campaigns.ActionRepo, 0, len(reposByID))
	for _, repo := range reposByID {
		repos = append(repos, repo)
	}
	logger.RepoMatches(len(repos), skipped, unsupported)

	if content, err := result.Data.Search.Results.Alert.Render(); err != nil {
		yellow.Fprint(os.Stderr, err)
	} else {
		os.Stderr.WriteString(content)
	}

	return repos, nil
}

var yellow = color.New(color.FgYellow)

func isGitAvailable() bool {
	cmd := exec.Command("git", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// askForConfirmation asks the user for confirmation. A user must type in "yes"
// and press enter to confirm. It has fuzzy matching, so "y", "Y", "yes",
// "YES", and "Yes" all count as confirmations. Everything else counts as "no".
func askForConfirmation(s string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/n]: ", s)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return true, nil
	}

	return false, nil
}

type minimumVersionDate struct {
	version string
	date    string
}

// codeHostCampaignVersions contains the minimum Sourcegraph version and build
// date required for the given code host kind. If a code host is present with a
// value of nil, this means that any Sourcegraph version will pass the check.
var codeHostCampaignVersions = map[string]*minimumVersionDate{
	"github":          nil,
	"bitbucketserver": nil,
	"gitlab": {
		version: "3.18.0",
		date:    "2020-07-14",
	},
}

func isCodeHostSupportedForCampaigns(kind string) (bool, error) {
	// TODO(LawnGnome): this is a temporary hack; I intend to improve our
	// testing story including mocking requests to Sourcegraph as part of
	// https://github.com/sourcegraph/sourcegraph/issues/12333
	return isCodeHostSupportedForCampaignsImpl(kind, getSourcegraphVersion)
}

func isCodeHostSupportedForCampaignsImpl(kind string, getVersion func() (string, error)) (bool, error) {
	mvd, ok := codeHostCampaignVersions[strings.ToLower(kind)]
	if !ok {
		return false, nil
	}
	if mvd == nil {
		return true, nil
	}

	ver, err := getVersion()
	if err != nil {
		return false, errors.Wrap(err, "getting Sourcegraph version")
	}

	return sourcegraphVersionCheck(ver, fmt.Sprintf(">= %s", mvd.version), mvd.date)
}
