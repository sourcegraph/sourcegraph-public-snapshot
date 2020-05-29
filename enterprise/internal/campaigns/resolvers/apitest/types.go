package apitest

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

type GitTarget struct {
	OID            string
	AbbreviatedOID string
	TargetType     string `json:"type"`
}

type GitRef struct {
	Name        string
	AbbrevName  string
	DisplayName string
	Prefix      string
	RefType     string `json:"type"`
	Repository  struct{ ID string }
	URL         string
	Target      GitTarget
}

type DiffRange struct{ StartLine, Lines int }

type DiffStat struct{ Added, Deleted, Changed int32 }

type FileDiffHunk struct {
	Body, Section      string
	OldNoNewlineAt     bool
	OldRange, NewRange DiffRange
}

type File struct {
	Name string
	// Ignoring other fields of File2, since that would require gitserver
}

type FileDiff struct {
	OldPath, NewPath string
	Hunks            []FileDiffHunk
	Stat             DiffStat
	OldFile          File
}

type FileDiffs struct {
	RawDiff  string
	DiffStat DiffStat
	PageInfo struct {
		HasNextPage bool
		EndCursor   string
	}
	Nodes []FileDiff
}

type PatchConnection struct {
	Nodes      []Patch
	TotalCount int
	PageInfo   struct {
		HasNextPage bool
	}
}

type Patch struct {
	Typename            string `json:"__typename"`
	ID                  string
	PublicationEnqueued bool
	Repository          struct{ Name, URL string }
	Diff                struct {
		FileDiffs FileDiffs
	}
}

type PatchSet struct {
	ID         string
	Patches    PatchConnection
	PreviewURL string
	DiffStat   DiffStat
}

type User struct {
	ID         string
	DatabaseID int32
	SiteAdmin  bool
}

type Org struct {
	ID   string
	Name string
}

type UserOrg struct {
	ID         string
	DatabaseID int32
	SiteAdmin  bool
	Name       string
}

type Campaign struct {
	ID                  string
	Name                string
	Description         string
	Branch              string
	Author              User
	ViewerCanAdminister bool
	Namespace           UserOrg
	CreatedAt           string
	UpdatedAt           string
	PublishedAt         string
	Status              struct {
		State  string
		Errors []string
	}
	Patches                 PatchConnection
	Changesets              ChangesetConnection
	OpenChangesets          ChangesetConnection
	ChangesetCountsOverTime []ChangesetCounts
	DiffStat                DiffStat
	PatchSet                PatchSet
}

type CampaignConnection struct {
	Nodes      []Campaign
	TotalCount int
	PageInfo   struct {
		HasNextPage bool
	}
}

type ChangesetEventConnection struct {
	TotalCount int
}

type Repository struct {
	ID   string
	Name string
}

type Changeset struct {
	Typename    string `json:"__typename"`
	ID          string
	Repository  Repository
	Campaigns   CampaignConnection
	CreatedAt   string
	UpdatedAt   string
	Title       string
	Body        string
	State       string
	ExternalURL struct {
		URL         string
		ServiceType string
	}
	ReviewState string
	CheckState  string
	Events      ChangesetEventConnection
	Head        GitRef
	Base        GitRef

	Diff struct {
		FileDiffs FileDiffs
	}
}

type ChangesetConnection struct {
	Nodes      []Changeset
	TotalCount int
	PageInfo   struct {
		HasNextPage bool
	}
}

type ChangesetCounts struct {
	Date                 graphqlbackend.DateTime
	Total                int32
	Merged               int32
	Closed               int32
	Open                 int32
	OpenApproved         int32
	OpenChangesRequested int32
	OpenPending          int32
}
