package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
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
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
)

type Action struct {
	ScopeQuery string        `json:"scopeQuery,omitempty"`
	Steps      []*ActionStep `json:"steps"`
}

type ActionStep struct {
	Type      string   `json:"type"`            // "command"
	Image     string   `json:"image,omitempty"` // Docker image
	CacheDirs []string `json:"cacheDirs,omitempty"`
	Args      []string `json:"args,omitempty"`

	// ImageContentDigest is an internal field that should not be set by users.
	ImageContentDigest string
}

type PatchInput struct {
	Repository   string `json:"repository"`
	BaseRevision string `json:"baseRevision"`
	BaseRef      string `json:"baseRef"`
	Patch        string `json:"patch"`
}

func userCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userCacheDir, "sourcegraph-src"), nil
}

const defaultTimeout = 60 * time.Minute

func init() {
	usage := `
Execute an action on code in repositories. The output of an action is a set of patches that can be used to create a campaign to open changesets and perform large-scale code changes.

Examples:

  Execute an action defined in ~/run-gofmt.json:

	$ src actions exec -f ~/run-gofmt.json

  Execute an action and create a patch set from the produced patches:

	$ src actions exec -f ~/run-gofmt.json -create-patchset

  Verbosely execute an action and keep the logs available for debugging:

	$ src -v actions exec -keep-logs -f ~/run-gofmt.json

  Execute an action and pipe the patches it produced to 'src campaign patchset create-from-patches':

	$ src actions exec -f ~/run-gofmt.json | src campaign patchset create-from-patches

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

	cacheDir, _ := userCacheDir()
	if cacheDir != "" {
		cacheDir = filepath.Join(cacheDir, "action-exec")
	}

	displayUserCacheDir := strings.Replace(cacheDir, os.Getenv("HOME"), "$HOME", 1)

	var (
		fileFlag        = flagSet.String("f", "-", "The action file. If not given or '-' standard input is used. (Required)")
		parallelismFlag = flagSet.Int("j", runtime.GOMAXPROCS(0), "The number of parallel jobs.")

		cacheDirFlag   = flagSet.String("cache", displayUserCacheDir, "Directory for caching results.")
		clearCacheFlag = flagSet.Bool("clear-cache", false, "Remove possibly cached results for an action before executing it.")

		keepLogsFlag = flagSet.Bool("keep-logs", false, "Do not remove execution log files when done.")
		timeoutFlag  = flagSet.Duration("timeout", defaultTimeout, "The maximum duration a single action run can take.")

		createPatchSetFlag      = flagSet.Bool("create-patchset", false, "Create a patch set from the produced set of patches. When the execution of the action fails in a single repository a prompt will ask to confirm or reject the patch set creation.")
		forceCreatePatchSetFlag = flagSet.Bool("force-create-patchset", false, "Force creation of patch set from the produced set of patches, without asking for confirmation even when the execution of the action failed for a subset of repositories.")

		apiFlags = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

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

		var (
			actionFile []byte
			err        error
		)

		if *fileFlag == "-" {
			actionFile, err = ioutil.ReadAll(os.Stdin)
		} else {
			actionFile, err = ioutil.ReadFile(*fileFlag)
		}
		if err != nil {
			return err
		}

		var action Action
		if err := jsonxUnmarshal(string(actionFile), &action); err != nil {
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

		logger := newActionLogger(*verbose, *keepLogsFlag)

		err = validateAction(ctx, action)
		if err != nil {
			return errors.Wrap(err, "Validation of action failed")
		}

		// Build Docker images etc.
		err = prepareAction(ctx, action)
		if err != nil {
			return errors.Wrap(err, "Failed to prepare action")
		}

		opts := actionExecutorOptions{
			timeout:    *timeoutFlag,
			keepLogs:   *keepLogsFlag,
			clearCache: *clearCacheFlag,
			cache:      actionExecutionDiskCache{dir: *cacheDirFlag},
		}
		if !*verbose {
			opts.onUpdate = newTerminalUI(*keepLogsFlag)
		}

		// Query repos over which to run action
		logger.Infof("Querying %s for repositories matching '%s'...\n", cfg.Endpoint, action.ScopeQuery)
		repos, skipped, err := actionRepos(ctx, action.ScopeQuery)
		if err != nil {
			return err
		}
		for _, r := range skipped {
			logger.Infof("Skipping repository %s because we couldn't determine default branch.\n", r)
		}
		logger.Infof("%d repositories match. Use 'src actions scope-query' for help with scoping.\n", len(repos))

		logger.Start()

		executor := newActionExecutor(action, *parallelismFlag, logger, opts)
		for _, repo := range repos {
			executor.enqueueRepo(repo)
		}

		// Execute actions
		if opts.onUpdate != nil {
			opts.onUpdate(executor.repos)
		}

		go executor.start(ctx)
		err = executor.wait()

		patches := executor.allPatches()
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

			logger.ActionSuccess(patches, true)

			return json.NewEncoder(os.Stdout).Encode(patches)
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

				c, _ := askForConfirmation(fmt.Sprintf("Create a patch set for the produced patches anyway?"))
				if !c {
					return err
				}
			}
		} else {
			logger.ActionSuccess(patches, false)
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

func validateAction(ctx context.Context, action Action) error {
	for _, step := range action.Steps {
		if step.Type == "docker" {
			if step.Image == "" {
				return fmt.Errorf("docker run step has to specify 'image'")
			}

			if step.ImageContentDigest != "" {
				return errors.New("setting the ImageContentDigest field of a docker run step is not allowed")
			}
		}

		if step.Type == "command" && len(step.Args) < 1 {
			return errors.New("command run step has to specify 'args'")
		}
	}

	return nil
}

func prepareAction(ctx context.Context, action Action) error {
	// Build any Docker images.
	for _, step := range action.Steps {
		if step.Type == "docker" {
			// Set digests for Docker images so we don't cache action runs in 2 different images with
			// the same tag.
			if step.Image != "" {
				var err error
				step.ImageContentDigest, err = getDockerImageContentDigest(ctx, step.Image)
				if err != nil {
					return errors.Wrap(err, "Failed to get Docker image content digest")
				}
			}
		}
	}

	return nil
}

// getDockerImageContentDigest gets the content digest for the image. Note that this
// is different from the "distribution digest" (which is what you can use to specify
// an image to `docker run`, as in `my/image@sha256:xxx`). We need to use the
// content digest because the distribution digest is only computed for images that
// have been pulled from or pushed to a registry. See
// https://windsock.io/explaining-docker-image-ids/ under "A Final Twist" for a good
// explanation.
func getDockerImageContentDigest(ctx context.Context, image string) (string, error) {
	// TODO!(sqs): is image id the right thing to use here? it is NOT the
	// digest. but the digest is not calculated for all images (unless they are
	// pulled/pushed from/to a registry), see
	// https://github.com/moby/moby/issues/32016.
	out, err := exec.CommandContext(ctx, "docker", "image", "inspect", "--format", "{{.Id}}", "--", image).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error inspecting docker image (try `docker pull %q` to fix this): %s", image, bytes.TrimSpace(out))
	}
	id := string(bytes.TrimSpace(out))
	if id == "" {
		return "", fmt.Errorf("unexpected empty docker image content ID for %q", image)
	}
	return id, nil
}

type ActionRepo struct {
	ID      string
	Name    string
	Rev     string
	BaseRef string
}

func actionRepos(ctx context.Context, scopeQuery string) ([]ActionRepo, []string, error) {
	hasCount, err := regexp.MatchString(`count:\d+`, scopeQuery)
	if err != nil {
		return nil, nil, err
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
					id
					name
					defaultBranch {
						name
						target { oid }
					}
				}
				... on FileMatch {
					repository {
						id
						name
						defaultBranch {
							name
							target { oid }
						}
					}
				}
			}
		}
	}
}
`
	type Repository struct {
		ID, Name      string
		DefaultBranch struct {
			Name   string
			Target struct{ OID string }
		}
	}
	var result struct {
		Search struct {
			Results struct {
				Results []struct {
					Typename      string `json:"__typename"`
					ID, Name      string
					DefaultBranch struct {
						Name   string
						Target struct{ OID string }
					}
					Repository Repository `json:"repository"`
				}
			}
		}
	}

	if err := (&apiRequest{
		query: query,
		vars: map[string]interface{}{
			"query": scopeQuery,
		},
		result: &result,
	}).do(); err != nil {
		return nil, nil, err
	}

	skipped := []string{}
	reposByID := map[string]ActionRepo{}
	for _, searchResult := range result.Search.Results.Results {
		var repo Repository
		if searchResult.Repository.ID != "" {
			repo = searchResult.Repository
		} else {
			repo = Repository{
				ID:            searchResult.ID,
				Name:          searchResult.Name,
				DefaultBranch: searchResult.DefaultBranch,
			}
		}

		if repo.DefaultBranch.Name == "" {
			skipped = append(skipped, repo.Name)
			continue
		}

		if repo.DefaultBranch.Target.OID == "" {
			skipped = append(skipped, repo.Name)
			continue
		}

		if _, ok := reposByID[repo.ID]; !ok {
			reposByID[repo.ID] = ActionRepo{
				ID:      repo.ID,
				Name:    repo.Name,
				Rev:     repo.DefaultBranch.Target.OID,
				BaseRef: repo.DefaultBranch.Name,
			}
		}
	}

	repos := make([]ActionRepo, 0, len(reposByID))
	for _, repo := range reposByID {
		repos = append(repos, repo)
	}
	return repos, skipped, nil
}

func sumDiffStats(fileDiffs []*diff.FileDiff) diff.Stat {
	sum := diff.Stat{}
	for _, fileDiff := range fileDiffs {
		stat := fileDiff.Stat()
		sum.Added += stat.Added
		sum.Changed += stat.Changed
		sum.Deleted += stat.Deleted
	}
	return sum
}

func diffStatDescription(fileDiffs []*diff.FileDiff) string {
	var plural string
	if len(fileDiffs) > 1 {
		plural = "s"
	}

	return fmt.Sprintf("%d file%s changed", len(fileDiffs), plural)
}

func diffStatSummary(stat diff.Stat) string {
	return fmt.Sprintf("%d insertions(+), %d deletions(-)", stat.Added+stat.Changed, stat.Deleted+stat.Changed)
}

func diffStatDiagram(stat diff.Stat) string {
	const maxWidth = 20
	added := float64(stat.Added + stat.Changed)
	deleted := float64(stat.Deleted + stat.Changed)
	if total := added + deleted; total > maxWidth {
		x := float64(20) / total
		added *= x
		deleted *= x
	}
	return color.GreenString(strings.Repeat("+", int(added))) + color.RedString(strings.Repeat("-", int(deleted)))
}

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
