package apitest

import (
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

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

func (ds DiffStat) ToDiffStat() *diff.Stat {
	return &diff.Stat{Added: ds.Added, Deleted: ds.Deleted, Changed: ds.Changed}
}

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
	Publishable         bool
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
	ID                      string
	Name                    string
	Description             string
	Branch                  string
	Author                  User
	ViewerCanAdminister     bool
	Namespace               UserOrg
	CreatedAt               string
	UpdatedAt               string
	Patches                 PatchConnection
	HasUnpublishedPatches   bool
	Changesets              ChangesetConnection
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
	Typename      string `json:"__typename"`
	ID            string
	Repository    Repository
	Campaigns     CampaignConnection
	CreatedAt     string
	UpdatedAt     string
	NextSyncAt    string
	Title         string
	Body          string
	State         string
	ExternalState string
	ExternalURL   struct {
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

type CampaignSpec struct {
	Typename string `json:"__typename"`
	ID       string

	OriginalInput string
	ParsedInput   graphqlbackend.JSONValue

	PreviewURL string

	Namespace UserOrg
	Creator   User

	ChangesetSpecs ChangesetSpecConnection

	ViewerCanAdminister bool

	CreatedAt graphqlbackend.DateTime
	ExpiresAt *graphqlbackend.DateTime
}

type ChangesetSpec struct {
	Typename string `json:"__typename"`
	ID       string

	Description ChangesetSpecDescription

	ExpiresAt *graphqlbackend.DateTime
}

type ChangesetSpecConnection struct {
	Nodes      []ChangesetSpec
	TotalCount int
	PageInfo   struct {
		HasNextPage bool
		EndCursor   *string
	}
}

type ChangesetSpecDescription struct {
	Typename string `json:"__typename"`

	BaseRepository Repository
	ExternalID     string
	BaseRef        string

	HeadRepository Repository
	HeadRef        string

	Title string
	Body  string

	Commits []GitCommitDescription

	Published bool

	Diff struct {
		FileDiffs FileDiffs
	}
}

type GitCommitDescription struct {
	Message string
	Diff    string
}
