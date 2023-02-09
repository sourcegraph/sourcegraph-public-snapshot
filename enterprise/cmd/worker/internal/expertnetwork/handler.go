package expertnetwork

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	gitserverClient gitserver.Client
	db              database.DB
}

var _ workerutil.Handler[*ExpertNetworkJob] = &handler{}

var allFiles = regexp.MustCompile("")

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *ExpertNetworkJob) error {
	repo, err := h.db.Repos().Get(ctx, api.RepoID(record.RepositoryID))
	if err != nil {
		return err
	}

	revision, err := h.gitserverClient.ResolveRevision(ctx, repo.Name, "", gitserver.ResolveRevisionOptions{})
	if err != nil {
		return err
	}

	files, err := h.gitserverClient.ListFiles(ctx, nil, repo.Name, revision, allFiles)
	if err != nil {
		return err
	}

	fmt.Println("files", len(files))

	return h.getFileChangelogs(ctx, repo.Name, revision, files, "/tmp/changelog.json", time.Now(), 4)
}

var botRegexp = regexp.MustCompile(`\bbot\b`)

type gitLogEntry struct {
	AuthorName      string
	AuthorEmail     string
	NumAddedLines   int
	NumRemovedLines int
	Timestamp       time.Time
}

type commitChange struct {
	NumAddedLines   int       `json:"numAddedLines"`
	NumRemovedLines int       `json:"numRemovedLines"`
	DaysSince       int       `json:"daysSince"`
	Timestamp       time.Time `json:"timestamp"`
}

type authorChangelog struct {
	Name    string         `json:"name"`
	Changes []commitChange `json:"changes"`
}

type fileChangelog struct {
	Path    string            `json:"path"`
	Authors []authorChangelog `json:"authors"`
}

func (h *handler) writer(outputPath string, changelogChannel <-chan *fileChangelog) {
	fileChangelogs := []*fileChangelog{}
	for changelog := range changelogChannel {
		fileChangelogs = append(fileChangelogs, changelog)
		fmt.Println("Finished:", len(fileChangelogs))
	}

	data, err := json.Marshal(fileChangelogs)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(outputPath, data, 0600)
	if err != nil {
		panic(err)
	}
}

func (h *handler) worker(ctx context.Context, repoName api.RepoName, revision api.CommitID, now time.Time, filePathsChannel <-chan string, changelogChannel chan<- *fileChangelog) {
	for filePath := range filePathsChannel {
		fileChangelog, err := h.getFileChangelog(ctx, repoName, revision, filePath, now)
		if err != nil {
			// Skip file
			continue
		}
		changelogChannel <- fileChangelog
	}
}

func (h *handler) getFileChangelogs(ctx context.Context, repoName api.RepoName, revision api.CommitID, files []string, outputPath string, now time.Time, nWorkers int) error {
	filePathsChannel := make(chan string, nWorkers*256)
	changelogChannel := make(chan *fileChangelog, nWorkers*256)

	workersWaitGroup := &sync.WaitGroup{}
	for w := 0; w < nWorkers; w++ {
		workersWaitGroup.Add(1)
		go func() {
			defer workersWaitGroup.Done()
			h.worker(ctx, repoName, revision, now, filePathsChannel, changelogChannel)
		}()
	}

	writersWaitGroup := &sync.WaitGroup{}
	writersWaitGroup.Add(1)
	go func() {
		defer writersWaitGroup.Done()
		h.writer(outputPath, changelogChannel)
	}()

	for _, file := range files {
		filePathsChannel <- file
	}

	close(filePathsChannel)

	workersWaitGroup.Wait()

	close(changelogChannel)

	writersWaitGroup.Wait()

	return nil
}

func (h *handler) getFileChangelog(ctx context.Context, repoName api.RepoName, revision api.CommitID, filePath string, now time.Time) (*fileChangelog, error) {
	fileLog, err := h.getFileGitLog(ctx, repoName, revision, filePath)
	if err != nil {
		return nil, err
	}

	lines := []string{}
	for _, line := range strings.Split(fileLog, "\n") {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		lines = append(lines, line)
	}

	logEntries := make([]gitLogEntry, 0, len(lines)/2)
	for i := 0; i < len(lines); i += 2 {
		authorWithTimestampLine, numChangesLine := lines[i], lines[i+1]

		authorWithTimestampLineColumns := strings.Split(authorWithTimestampLine, ",")
		numChangesLineColumns := strings.Split(numChangesLine, "\t")

		authorName, authorEmail, commitTimestamp := authorWithTimestampLineColumns[0], authorWithTimestampLineColumns[1], authorWithTimestampLineColumns[2]

		// Filter out bots (renovate, etc.)
		if len(botRegexp.FindString(strings.ToLower(authorName))) > 0 {
			continue
		}

		numAddedLines, err := strconv.Atoi(numChangesLineColumns[0])
		if err != nil {
			continue
		}

		numDeletedLines, err := strconv.Atoi(numChangesLineColumns[1])
		if err != nil {
			continue
		}

		timestamp, err := time.Parse("2006-01-02T15:04:05-07:00", commitTimestamp)
		if err != nil {
			return nil, err
		}

		logEntries = append(logEntries, gitLogEntry{authorName, authorEmail, numAddedLines, numDeletedLines, timestamp})
	}

	// Group by author name
	logEntriesByAuthor := map[string][]gitLogEntry{}
	for _, entry := range logEntries {
		logEntriesByAuthor[entry.AuthorName] = append(logEntriesByAuthor[entry.AuthorName], entry)
	}

	authors := make([]authorChangelog, 0, len(logEntriesByAuthor))
	for author, logEntries := range logEntriesByAuthor {
		commits := make([]commitChange, 0, len(logEntries))
		for _, log := range logEntries {
			commits = append(commits, commitChange{log.NumAddedLines, log.NumRemovedLines, int(now.Sub(log.Timestamp).Hours()) / 24, log.Timestamp})
		}
		authors = append(authors, authorChangelog{author, commits})
	}

	return &fileChangelog{filePath, authors}, nil
}

func (h *handler) getFileGitLog(ctx context.Context, repoName api.RepoName, revision api.CommitID, filePath string) (string, error) {
	return h.gitserverClient.GetFileLog(ctx, repoName, revision, filePath)
}
