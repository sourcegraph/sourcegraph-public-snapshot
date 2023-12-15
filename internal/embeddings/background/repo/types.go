package repo

import (
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

type RepoEmbeddingJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt *time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	RepoID   api.RepoID
	Revision api.CommitID
}

func (j *RepoEmbeddingJob) RecordID() int {
	return j.ID
}

func (j *RepoEmbeddingJob) RecordUID() string {
	return strconv.Itoa(j.ID)
}

func (j *RepoEmbeddingJob) IsRepoEmbeddingJobScheduledOrCompleted() bool {
	return j.State == "completed" || j.State == "processing" || j.State == "queued"
}

type EmbedRepoStats struct {
	CodeIndexStats EmbedFilesStats
	TextIndexStats EmbedFilesStats

	// IsIncremental indicates whether the embedding job should reindex changed files
	IsIncremental bool
}

func (e *EmbedRepoStats) ToFields() []log.Field {
	return []log.Field{
		log.Object("codeIndex", e.CodeIndexStats.ToFields()...),
		log.Object("textIndex", e.TextIndexStats.ToFields()...),
		log.Bool("isIncremental", e.IsIncremental),
	}
}

func NewEmbedFilesStats(filesTotal int) EmbedFilesStats {
	return EmbedFilesStats{
		FilesScheduled: filesTotal,
		FilesEmbedded:  0,
		FilesSkipped:   map[string]int{},
		ChunksEmbedded: 0,
		ChunksExcluded: 0,
		BytesEmbedded:  0,
	}
}

type EmbedFilesStats struct {
	// The total number of files scheduled for embedding. For a complete job,
	// should be the sum of FilesEmbedded and all the FilesSkipoped reasons.
	FilesScheduled int

	// The number of files embedded
	FilesEmbedded int

	// The number of files skipped for each reason
	FilesSkipped map[string]int

	// The number of chunks we split embedded files for.
	// Equivalent to the number of embeddings generated.
	ChunksEmbedded int

	// The number of chunks that were excluded from the index.
	// Cause of exclusion is typically due to failed embeddings requests.
	ChunksExcluded int

	// The sum of the size of the contents of successful embeddings
	BytesEmbedded int
}

func (e *EmbedFilesStats) Skip(reason string, size int) {
	e.FilesSkipped[reason] += 1
}

func (e *EmbedFilesStats) AddChunks(count, size int) {
	e.ChunksEmbedded += count
	e.BytesEmbedded += size
}

func (e *EmbedFilesStats) ExcludeChunks(count int) {
	e.ChunksExcluded += count
}

func (e *EmbedFilesStats) AddFile() {
	e.FilesEmbedded += 1
}

func (e *EmbedFilesStats) ToFields() []log.Field {
	var skippedCounts []log.Field
	for reason, count := range e.FilesSkipped {
		skippedCounts = append(skippedCounts, log.Int(reason, count))
	}
	return []log.Field{
		log.Int("filesTotal", e.FilesScheduled),
		log.Int("filesEmbedded", e.FilesEmbedded),
		log.Int("chunksEmbedded", e.ChunksEmbedded),
		log.Int("chunksExcluded", e.ChunksExcluded),
		log.Int("bytesEmbedded", e.BytesEmbedded),
		log.Object("filesSkipped", skippedCounts...),
	}
}
