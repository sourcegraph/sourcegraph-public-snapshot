package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
)

// Older versions of GNU diff (< 3.3) do not support all the flags we want, but
// since macOS Mojave and Catalina ship with GNU diff 2.8.1, we try to detect
// missing flags and degrade behavior gracefully instead of failing. check for
// the flags and degrade if they're not available.
var (
	diffSupportsNoDereference = false
	diffSupportsColor         = false
)

type Action struct {
	ScopeQuery string        `json:"scopeQuery,omitempty"`
	Steps      []*ActionStep `json:"steps"`
}

type ActionStep struct {
	Type       string   `json:"type"` // "command"
	Dockerfile string   `json:"dockerfile,omitempty"`
	Image      string   `json:"image,omitempty"` // Docker image
	CacheDirs  []string `json:"cacheDirs,omitempty"`
	Args       []string `json:"args,omitempty"`

	// ImageContentDigest is an internal field that should not be set by users.
	ImageContentDigest string
}

type CampaignPlanPatch struct {
	Repository   string `json:"repository"`
	BaseRevision string `json:"baseRevision"`
	Patch        string `json:"patch"`
}

func userCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userCacheDir, "sourcegraph-src"), nil
}

const defaultTimeout = 2 * time.Minute

func init() {
	usage := `
Execute an action on code in repositories. The output of an action is a set of patches that can be used to create a campaign to open changesets and perform large-scale code changes.

Examples:

  Execute an action defined in ~/run-gofmt-in-dockerfile.json:

    	$ src actions exec -f ~/run-gofmt-in-dockerfile.json

  Verbosely execute an action and keep the logs available for debugging:

		$ src -v actions exec -keep-logs -f ~/run-gofmt-in-dockerfile.json

  Execute an action and create a campaign plan from the patches it produced:

    	$ src actions exec -f ~/run-gofmt-in-dockerfile.json | src campaign plan create-from-patches

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
		      "dockerfile": "FROM alpine:3 \n CMD find /work -iname '*.txt' -type f | xargs -n 1 sed -i s/this/that/g"
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
		cacheDirFlag    = flagSet.String("cache", displayUserCacheDir, "Directory for caching results.")
		keepLogsFlag    = flagSet.Bool("keep-logs", false, "Do not remove execution log files when done.")
		timeoutFlag     = flagSet.Duration("timeout", defaultTimeout, "The maximum duration a single action run can take (excluding the building of Docker images).")
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

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

		ctx := context.Background()

		diffSupportsNoDereference, err = diffSupportsFlag(ctx, "--no-dereference")
		if err != nil {
			return err
		}

		diffSupportsColor, err = diffSupportsFlag(ctx, "--color")
		if err != nil {
			return err
		}

		err = validateAction(ctx, action)
		if err != nil {
			return errors.Wrap(err, "Validation of action failed")
		}

		// Build Docker images etc.
		err = prepareAction(ctx, action)
		if err != nil {
			return errors.Wrap(err, "Failed to prepare action")
		}

		uilive.Out = os.Stderr
		uilive.RefreshInterval = 10 * time.Hour // TODO!(sqs): manually flush
		color.NoColor = false                   // force color even when in a pipe
		var (
			lwMu sync.Mutex
			lw   = uilive.New()
		)
		lw.Start()
		defer lw.Stop()

		spinner := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		spinnerI := 0
		onUpdate := func(reposMap map[ActionRepo]ActionRepoStatus) {
			lwMu.Lock()
			defer lwMu.Unlock()

			spinnerRune := spinner[spinnerI%len(spinner)]
			spinnerI++

			reposSorted := make([]ActionRepo, 0, len(reposMap))
			repoNameLen := 0
			for repo := range reposMap {
				reposSorted = append(reposSorted, repo)
				if n := utf8.RuneCountInString(repo.Name); n > repoNameLen {
					repoNameLen = n
				}
			}
			sort.Slice(reposSorted, func(i, j int) bool { return reposSorted[i].Name < reposSorted[j].Name })

			for i, repo := range reposSorted {
				status := reposMap[repo]

				var (
					timerDuration time.Duration

					statusColor func(string, ...interface{}) string

					statusText  string
					logFileText string
				)
				if *keepLogsFlag && status.LogFile != "" {
					logFileText = color.HiBlackString(status.LogFile)
				}
				switch {
				case !status.Cached && status.StartedAt.IsZero():
					statusColor = color.HiBlackString
					statusText = statusColor(string(spinnerRune))
					timerDuration = time.Since(status.EnqueuedAt)

				case !status.Cached && status.FinishedAt.IsZero():
					statusColor = color.YellowString
					statusText = statusColor(string(spinnerRune))
					timerDuration = time.Since(status.StartedAt)

				case status.Cached || !status.FinishedAt.IsZero():
					if status.Err != nil {
						statusColor = color.RedString
						statusText = "error: see " + status.LogFile
						logFileText = "" // don't show twice
					} else {
						statusColor = color.GreenString
						if status.Patch != (CampaignPlanPatch{}) && status.Patch.Patch != "" {
							fileDiffs, err := diff.ParseMultiFileDiff([]byte(status.Patch.Patch))
							if err != nil {
								panic(err)
								// return errors.Wrapf(err, "invalid patch for repository %q", repo.Name)
							}
							statusText = diffStatDescription(fileDiffs) + " " + diffStatDiagram(sumDiffStats(fileDiffs))
							if status.Cached {
								statusText += " (cached)"
							}
						} else {
							statusText = color.HiBlackString("0 files changed")
						}
					}
					timerDuration = status.FinishedAt.Sub(status.StartedAt)
				}

				var w io.Writer
				if i == 0 {
					w = lw
				} else {
					w = lw.Newline()
				}

				var appendTexts []string
				if statusText != "" {
					appendTexts = append(appendTexts, statusText)
				}
				if logFileText != "" {
					appendTexts = append(appendTexts, logFileText)
				}
				repoText := statusColor(fmt.Sprintf("%-*s", repoNameLen, repo.Name))
				pipe := color.HiBlackString("|")
				fmt.Fprintf(w, "%s %s ", repoText, pipe)
				fmt.Fprintf(w, "%s", strings.Join(appendTexts, " "))
				if timerDuration != 0 {
					fmt.Fprintf(w, color.HiBlackString(" %s"), timerDuration.Round(time.Second))
				}
				fmt.Fprintln(w)
			}
			_ = lw.Flush()
		}
		executor := newActionExecutor(action, *parallelismFlag, actionExecutorOptions{
			timeout:  *timeoutFlag,
			keepLogs: *keepLogsFlag,
			cache:    actionExecutionDiskCache{dir: *cacheDirFlag},
			onUpdate: onUpdate,
		})

		if *verbose {
			log.Printf("# Querying %s for repositories matching %q...", cfg.Endpoint, action.ScopeQuery)
		}

		// Query repos over which to run action
		repos, err := actionRepos(ctx, *verbose, action.ScopeQuery)
		if err != nil {
			return err
		}
		if *verbose {
			log.Printf("# %d repositories match. Use 'src actions scope-query' for help with scoping.", len(repos))
		}
		for _, repo := range repos {
			executor.enqueueRepo(repo)
		}

		// Execute actions
		onUpdate(executor.repos)

		go executor.start(ctx)
		if err := executor.wait(); err != nil {
			return err
		}
		patches := executor.allPatches()
		if *verbose {
			log.Printf("# Action produced %d patches.", len(patches))
		}
		return json.NewEncoder(os.Stdout).Encode(patches)
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
			if step.Dockerfile == "" && step.Image == "" {
				return fmt.Errorf("docker run step has to specify either 'image' or 'dockerfile'")
			}

			if step.Dockerfile != "" && step.Image != "" {
				return fmt.Errorf("docker run step may specify either image (%q) or dockerfile, not both", step.Image)
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
	for i, step := range action.Steps {
		if step.Type == "docker" {
			if step.Dockerfile == "" && step.Image == "" {
				return fmt.Errorf("docker run step has to specify either 'image' or 'dockerfile'")
			}

			if step.Dockerfile != "" && step.Image != "" {
				return fmt.Errorf("docker run step may specify either image (%q) or dockerfile, not both", step.Image)
			}

			if step.Dockerfile != "" {
				iidFile, err := ioutil.TempFile("", "src-actions-exec-image-id")
				if err != nil {
					return err
				}
				defer os.Remove(iidFile.Name())

				if *verbose {
					log.Printf("# Building Docker container for step %d...", i)
				}

				cmd := exec.CommandContext(ctx, "docker", "build", "--iidfile", iidFile.Name(), "-")
				cmd.Stdin = strings.NewReader(step.Dockerfile)
				verboseCmdOutput(cmd)
				if err := cmd.Run(); err != nil {
					return errors.Wrap(err, "build docker image")
				}
				if *verbose {
					log.Printf("# Done building Docker container for step %d.", i)
				}

				iid, err := ioutil.ReadFile(iidFile.Name())
				if err != nil {
					return err
				}
				step.Image = string(iid)
			}

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
	ID   string
	Name string
	Rev  string
}

func actionRepos(ctx context.Context, verbose bool, scopeQuery string) ([]ActionRepo, error) {
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
					id
					name
					defaultBranch { name }
				}
				... on FileMatch {
					repository {
						id
						name
						defaultBranch { name }
					}
				}
			}
		}
	}
}
`
	type Repository struct {
		ID, Name      string
		DefaultBranch struct{ Name string }
	}
	var result struct {
		Search struct {
			Results struct {
				Results []struct {
					Typename      string `json:"__typename"`
					ID, Name      string
					DefaultBranch struct{ Name string }
					Repository    Repository `json:"repository"`
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
		return nil, err
	}

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
			log.Printf("# Skipping repository %s because we couldn't determine default branch.", repo.Name)
			continue
		}

		if _, ok := reposByID[repo.ID]; !ok {
			reposByID[repo.ID] = ActionRepo{
				ID:   repo.ID,
				Name: repo.Name,
				Rev:  repo.DefaultBranch.Name,
			}
		}
	}

	repos := make([]ActionRepo, 0, len(reposByID))
	for _, repo := range reposByID {
		repos = append(repos, repo)
	}
	return repos, nil
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
