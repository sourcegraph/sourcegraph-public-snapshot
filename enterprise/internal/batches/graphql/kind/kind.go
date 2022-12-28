package kind

const (
	BatchChange = "BatchChange"
	Changeset   = "Changeset"

	// Kinds that aren't defined by Batch Changes, but we also need to have
	// knowledge of.
	Org  = "Org"
	Repo = "Repository"
	User = "User"
)
