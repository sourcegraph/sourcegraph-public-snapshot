package perforce

import (
	"fmt"

	"encoding/json"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

type BadChangelistError struct {
	CID  string
	Repo api.RepoName
}

func (e *BadChangelistError) Error() string {
	return fmt.Sprintf("invalid changelist ID %q for repo %q", e.Repo, e.CID)
}

// Example changelist info output in "long" format
// (from `p4 changes -l ...`)
// Change 1188 on 2023/06/09 by admin@yet-moar-lines *pending*
//
//	Append still another line to all SECOND.md files
//
// "admin@yet-moar-lines" is the username @ the client spec name, which in this case is the branch name from the batch change
// the final field - "*pending*" in this example - is optional and not present when the changelist has been submitted ("merged", in Git parlance)
// Example changelist info in json format
// (from `p4 -ztags -Mj changes -l ...`)
// {"data":"Change 1178 on 2023/06/01 by admin@hello-third-world *pending*\n\n\tAppend Hello World to all THIRD.md files\n","level":0}
var changelistInfoPattern = lazyregexp.New(`^Change (\d+) on (\d{4}/\d{2}/\d{2}) by ([^ ]+)@([^ ]+)(?: [*](pending|submitted|shelved)[*])?(?: '(.+)')?$`)

type changelistJson struct {
	Data  string `json:"data"`
	Level int    `json:"level"`
}

// Parses the output of `p4 changes`
// Handles one changelist only
// Accepts any format: standard, long, json standard, json long
func ParseChangelistOutput(output string) (*protocol.PerforceChangelist, error) {
	// output will be whitespace-trimmed and not empty

	// if the given text is json format, extract the Data portion
	// so that it will have the same format as the standard output
	cidj := new(changelistJson)
	err := json.Unmarshal([]byte(output), cidj)
	if err == nil {
		output = strings.TrimSpace(cidj.Data)
	}

	lines := strings.Split(output, "\n")

	// the first line contains the changelist information
	matches := changelistInfoPattern.FindStringSubmatch(lines[0])

	if matches == nil || len(matches) < 5 {
		return nil, errors.New("invalid changelist output")
	}

	pcl := new(protocol.PerforceChangelist)
	pcl.ID = matches[1]
	time, err := time.Parse("2006/01/02", matches[2])
	if err != nil {
		return nil, errors.Wrap(err, "invalid date: "+matches[2])
	}
	pcl.CreationDate = time
	pcl.Author = matches[3]
	pcl.Title = matches[4]
	status := "submitted"
	if len(matches) > 5 && matches[5] != "" {
		status = matches[5]
	}
	cls, err := protocol.ParsePerforceChangelistState(status)
	if err != nil {
		return nil, err
	}
	pcl.State = cls

	if len(matches) > 6 && matches[6] != "" {
		// the commit message is inline with the info
		pcl.Message = strings.TrimSpace(matches[6])
	} else {
		// the commit message is in subsequent lines of the output
		var builder strings.Builder
		for i := 2; i < len(lines); i++ {
			if i > 2 {
				builder.WriteString("\n")
			}
			builder.WriteString(strings.TrimSpace(lines[i]))
		}
		pcl.Message = builder.String()
	}
	return pcl, nil
}
