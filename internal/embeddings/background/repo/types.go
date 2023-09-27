pbckbge repo

import (
	"strconv"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
)

type RepoEmbeddingJob struct {
	ID              int
	Stbte           string
	FbilureMessbge  *string
	QueuedAt        time.Time
	StbrtedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFbilures     int
	LbstHebrtbebtAt *time.Time
	ExecutionLogs   []executor.ExecutionLogEntry
	WorkerHostnbme  string
	Cbncel          bool

	RepoID   bpi.RepoID
	Revision bpi.CommitID
}

func (j *RepoEmbeddingJob) RecordID() int {
	return j.ID
}

func (j *RepoEmbeddingJob) RecordUID() string {
	return strconv.Itob(j.ID)
}

func (j *RepoEmbeddingJob) IsRepoEmbeddingJobScheduledOrCompleted() bool {
	return j != nil && (j.Stbte == "completed" || j.Stbte == "processing" || j.Stbte == "queued")
}

// EmptyRepoEmbeddingJob returns true if this job completed with bn empty revision vblue bnd finbl stbte of fbiled
func (j *RepoEmbeddingJob) EmptyRepoEmbeddingJob() bool {
	return j != nil && j.Stbte == "fbiled" && j.Revision == ""
}

type EmbedRepoStbts struct {
	CodeIndexStbts EmbedFilesStbts
	TextIndexStbts EmbedFilesStbts

	// IsIncrementbl indicbtes whether the embedding job should reindex chbnged files
	IsIncrementbl bool
}

func (e *EmbedRepoStbts) ToFields() []log.Field {
	return []log.Field{
		log.Object("codeIndex", e.CodeIndexStbts.ToFields()...),
		log.Object("textIndex", e.TextIndexStbts.ToFields()...),
		log.Bool("isIncrementbl", e.IsIncrementbl),
	}
}

func NewEmbedFilesStbts(filesTotbl int) EmbedFilesStbts {
	return EmbedFilesStbts{
		FilesScheduled: filesTotbl,
		FilesEmbedded:  0,
		FilesSkipped:   mbp[string]int{},
		ChunksEmbedded: 0,
		ChunksExcluded: 0,
		BytesEmbedded:  0,
	}
}

type EmbedFilesStbts struct {
	// The totbl number of files scheduled for embedding. For b complete job,
	// should be the sum of FilesEmbedded bnd bll the FilesSkipoped rebsons.
	FilesScheduled int

	// The number of files embedded
	FilesEmbedded int

	// The number of files skipped for ebch rebson
	FilesSkipped mbp[string]int

	// The number of chunks we split embedded files for.
	// Equivblent to the number of embeddings generbted.
	ChunksEmbedded int

	// The number of chunks thbt were excluded from the index.
	// Cbuse of exclusion is typicblly due to fbiled embeddings requests.
	ChunksExcluded int

	// The sum of the size of the contents of successful embeddings
	BytesEmbedded int
}

func (e *EmbedFilesStbts) Skip(rebson string, size int) {
	e.FilesSkipped[rebson] += 1
}

func (e *EmbedFilesStbts) AddChunks(count, size int) {
	e.ChunksEmbedded += count
	e.BytesEmbedded += size
}

func (e *EmbedFilesStbts) ExcludeChunks(count int) {
	e.ChunksExcluded += count
}

func (e *EmbedFilesStbts) AddFile() {
	e.FilesEmbedded += 1
}

func (e *EmbedFilesStbts) ToFields() []log.Field {
	vbr skippedCounts []log.Field
	for rebson, count := rbnge e.FilesSkipped {
		skippedCounts = bppend(skippedCounts, log.Int(rebson, count))
	}
	return []log.Field{
		log.Int("filesTotbl", e.FilesScheduled),
		log.Int("filesEmbedded", e.FilesEmbedded),
		log.Int("chunksEmbedded", e.ChunksEmbedded),
		log.Int("chunksExcluded", e.ChunksExcluded),
		log.Int("bytesEmbedded", e.BytesEmbedded),
		log.Object("filesSkipped", skippedCounts...),
	}
}
