package entities

// EdgeType is available type of edge on phabricator.
type EdgeType string

var (
	EdgeCommitRevision EdgeType = "commit.revision"
	EdgeCommitTask     EdgeType = "commit.task"
	EdgeMention        EdgeType = "mention"
	EdgeMentionedIn    EdgeType = "mentioned-in"
	EdgeRevisionChild  EdgeType = "revision.child"
	EdgeRevisionCommit EdgeType = "revision.commit"
	EdgeRevisionParent EdgeType = "revision.parent"
	EdgeRevisionTask   EdgeType = "revision.task"
	EdgeTaskCommit     EdgeType = "task.commit"
	EdgeTaskDuplicate  EdgeType = "task.duplicate"
	EdgeTaskMergedIn   EdgeType = "task.merged-in"
	EdgeTaskParent     EdgeType = "task.parent"
	EdgeTaskRevision   EdgeType = "task.revision"
	EdgeTaskSubtask    EdgeType = "task.subtask"
)

// Edge is a relation between two objects on Phabricator. EdgeType defines the
// type of such relation (it can be parent, child, mention, etc.).
type Edge struct {
	SourcePHID      string   `json:"sourcePHID"`
	DestinationPHID string   `json:"destinationPHID"`
	EdgeType        EdgeType `json:"edgeType"`
}
