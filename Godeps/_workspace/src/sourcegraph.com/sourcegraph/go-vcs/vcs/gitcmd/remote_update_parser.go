package gitcmd

import (
	"fmt"
	"strings"
	"unicode"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

// parseRemoteUpdate parses stderr output from running `git remote update`,
// and returns a vcs.UpdateResult.
func parseRemoteUpdate(stderr []byte) (vcs.UpdateResult, error) {
	var result vcs.UpdateResult

	lines := strings.Split(string(stderr), "\n")
	if len(lines) < 3 { // Minimum input should have 3 lines (1 for header, 1 for change, and last empty line).
		return result, nil
	}
	for _, line := range lines[1 : len(lines)-1] {
		change, err := parseRemoteUpdateLine(line)
		if err != nil {
			return result, err
		}
		result.Changes = append(result.Changes, change)
	}

	return result, nil
}

// parseRemoteUpdateLine parses a line like `   e8569f7..de0ad17  master     -> master`.
func parseRemoteUpdateLine(line string) (vcs.Change, error) {
	var change vcs.Change

	// Shortest valid input is len(`   d6d0813..e8569f7  m -> m`) = 27 characters.
	if len(line) < 27 {
		return change, fmt.Errorf("line too short")
	}

	// Parse operation.
	switch line[:3] {
	case " * ":
		change.Op = vcs.NewOp
	case "   ":
		change.Op = vcs.FFUpdatedOp
	case " + ":
		const suffix = " (forced update)"
		if !strings.HasSuffix(line, suffix) {
			return change, fmt.Errorf(`unsupported " + " format`)
		}
		line = line[:len(line)-len(suffix)]
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		change.Op = vcs.ForceUpdatedOp
	case " x ":
		change.Op = vcs.DeletedOp
	default:
		return change, fmt.Errorf("unsupported format")
	}

	// Parse branch name.
	branch, err := parseBranchArrowBranch(line[21:])
	if err != nil {
		return change, fmt.Errorf("failed to parse branch name")
	}
	change.Branch = branch

	return change, nil
}

// parseBranchArrowBranch parses a `master     -> master` segment to extract
// relevant branch name. Currently always using 2nd branch name.
func parseBranchArrowBranch(bab string) (branch string, err error) {
	branches := strings.SplitN(bab, " -> ", 2)
	if len(branches) != 2 {
		return "", fmt.Errorf("failed to parse `branch -> branch` segment")
	}
	// Note, if we wanted to use branches[0], we should trim whitespace on its right.
	// Return second branch name since it's always valid.
	return branches[1], nil
}
