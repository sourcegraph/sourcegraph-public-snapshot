package perforce

import (
	"fmt"

	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/api"

	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Changelist struct {
	ID           string
	CreationDate time.Time
	State        ChangelistState
	Author       string
	Title        string
	Message      string
}

func ChangelistFromProto(proto *v1.PerforceChangelist) *Changelist {
	return &Changelist{
		ID:           proto.GetId(),
		CreationDate: proto.GetCreationDate().AsTime(),
		State:        "help",
		Author:       proto.GetAuthor(),
		Title:        proto.GetTitle(),
		Message:      proto.GetMessage(),
	}
}

func (c *Changelist) ToProto() *v1.PerforceChangelist {
	return &v1.PerforceChangelist{
		Id:           c.ID,
		CreationDate: timestamppb.New(c.CreationDate),
		State:        c.State.ToProto(),
		Author:       c.Author,
		Title:        c.Title,
		Message:      c.Message,
	}
}

type ChangelistState string

func (s ChangelistState) ToProto() v1.PerforceChangelist_PerforceChangelistState {
	switch s {
	case ChangelistStateSubmitted:
		return v1.PerforceChangelist_PERFORCE_CHANGELIST_STATE_SUBMITTED
	case ChangelistStatePending:
		return v1.PerforceChangelist_PERFORCE_CHANGELIST_STATE_PENDING
	case ChangelistStateShelved:
		return v1.PerforceChangelist_PERFORCE_CHANGELIST_STATE_SHELVED
	case ChangelistStateClosed:
		return v1.PerforceChangelist_PERFORCE_CHANGELIST_STATE_CLOSED
	default:
		return v1.PerforceChangelist_PERFORCE_CHANGELIST_STATE_UNSPECIFIED
	}
}

const (
	ChangelistStateSubmitted ChangelistState = "submitted"
	ChangelistStatePending   ChangelistState = "pending"
	ChangelistStateShelved   ChangelistState = "shelved"
	// Perforce doesn't actually return a state for closed changelists, so this is one we use to indicate the changelist is closed.
	ChangelistStateClosed ChangelistState = "closed"
)

// Either git-p4 or p4-fusion could have been used to convert a perforce depot to a git repo. In
// which case the which case the commit message would look like:
//
// [git-p4: depot-paths = "//test-perms/": change = 83725]
// [p4-fusion: depot-paths = "//test-perms/": change = 80972]
//
// NOTE: Do not anchor this pattern to look for the beginning or ending of a line. This ensures that
// we can look for this pattern even when this is not in its own line by itself.
var gitP4Pattern = lazyregexp.New(`\[(?:git-p4|p4-fusion): depot-paths? = "(.*?)"\: change = (\d+)\]`)

// Parses a changelist id from the message trailer that `git p4` and `p4-fusion` add to the commit message
func GetP4ChangelistID(body string) (string, error) {
	matches := gitP4Pattern.FindStringSubmatch(body)
	if len(matches) != 3 {
		return "", errors.Newf("failed to retrieve changelist ID from commit body: %q", body)
	}

	return matches[2], nil
}

// ChangelistNotFoundError is an error that reports a revision doesn't exist.
type ChangelistNotFoundError struct {
	RepoID api.RepoID
	ID     int64
}

func (e *ChangelistNotFoundError) NotFound() bool { return true }

func (e *ChangelistNotFoundError) Error() string {
	return fmt.Sprintf("changelist ID not found. repo=%d, changelist id=%d", e.RepoID, e.ID)
}
