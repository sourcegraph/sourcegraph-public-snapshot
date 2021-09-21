package search

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func FormatDiff(rawDiff []*diff.FileDiff, highlights map[int]protocol.FileDiffHighlight) (string, protocol.Ranges) {
	var buf strings.Builder
	var loc protocol.Location
	var ranges protocol.Ranges

	for fileIdx, fileDiff := range rawDiff {
		fdh, ok := highlights[fileIdx]
		if !ok {
			continue
		}

		ranges = append(ranges, fdh.OldFile.Add(loc)...)
		buf.WriteString(fileDiff.OrigName)
		buf.WriteByte(' ')
		loc = loc.Add(protocol.Location{
			Offset: len(fileDiff.OrigName) + len(" "),
		})

		ranges = append(ranges, fdh.NewFile.Add(loc)...)
		buf.WriteString(fileDiff.NewName)
		buf.WriteByte('\n')
		loc = loc.Add(protocol.Location{
			Line:   1,
			Offset: len(fileDiff.NewName) + len("\n"),
		})
		loc.Column = 0

		// TODO extract consts
		filteredHunks, filteredHighlights := splitHunkMatches(fileDiff.Hunks, fdh.HunkHighlights, 1, 5)

		for hunkIdx, hunk := range filteredHunks {
			hmh, ok := filteredHighlights[hunkIdx]
			if !ok {
				continue
			}

			n, _ := fmt.Fprintf(&buf,
				"@@ -%d,%d +%d,%d @@ %s\n",
				hunk.OrigStartLine,
				hunk.OrigLines,
				hunk.NewStartLine,
				hunk.NewLines,
				hunk.Section,
			)
			loc = loc.Add(protocol.Location{
				Offset: n,
				Line:   1,
			})
			loc.Column = 0

			lines := bytes.Split(hunk.Body, []byte("\n"))
			for lineIdx, line := range lines {
				if len(line) == 0 {
					continue
				}
				loc = loc.Add(protocol.Location{Offset: 1, Column: 1})
				if lineHighlights, ok := hmh.LineHighlights[lineIdx]; ok {
					ranges = append(ranges, lineHighlights.Add(loc)...)
				}

				buf.Write(line)
				buf.WriteByte('\n')
				loc = loc.Add(protocol.Location{Line: 1, Offset: len(line)})
				loc.Column = 0
			}
		}
	}

	return buf.String(), ranges
}

// splitHunkMatches returns a list of hunks that are a subset of the input hunks,
// filtered down to only hunks that match the query. Non-matching context lines
// and non-matching changed lines are eliminated, and the hunk header (start/end
// lines) are adjusted accordingly.
func splitHunkMatches(hunks []*diff.Hunk, hunkHighlights map[int]protocol.HunkHighlight, matchContextLines, maxLinesPerHunk int) (results []*diff.Hunk, newHighlights map[int]protocol.HunkHighlight) {
	addExtraHunkMatchesSection := func(hunk *diff.Hunk, extraHunkMatches int) {
		if extraHunkMatches > 0 {
			if hunk.Section != "" {
				hunk.Section += " "
			}
			hunk.Section += fmt.Sprintf("... +%d", extraHunkMatches)
		}
	}

	newHighlights = make(map[int]protocol.HunkHighlight, len(hunkHighlights))

	for i, hunk := range hunks {
		var cur *diff.Hunk
		var curLines [][]byte
		origHighlights := hunkHighlights[i].LineHighlights
		curHighlights := make(map[int]protocol.Ranges, len(hunkHighlights))

		lines := bytes.SplitAfter(hunk.Body, []byte("\n"))
		lineInfo := computeDiffHunkInfo(lines, hunkHighlights[i].LineHighlights, matchContextLines)

		extraHunkMatches := 0
		var origLineOffset, newLineOffset int32
		for i, line := range lines {
			if len(line) == 0 {
				break
			}
			lineInfo := lineInfo[i]
			if lineInfo.matching || lineInfo.context {
				if maxLinesPerHunk > 0 && len(curLines) == matchContextLines*2+maxLinesPerHunk {
					if lineInfo.matching {
						extraHunkMatches++
					}
					continue
				}

				if cur == nil {
					cur = &diff.Hunk{
						OrigStartLine: hunk.OrigStartLine + origLineOffset,
						NewStartLine:  hunk.NewStartLine + newLineOffset,
						Section:       hunk.Section,
					}
				}

				if !lineInfo.added {
					cur.OrigLines++
				}
				if !lineInfo.removed {
					cur.NewLines++
				}
				curHighlights[len(curLines)] = origHighlights[i]
				curLines = append(curLines, line)
			} else if cur != nil {
				addExtraHunkMatchesSection(cur, extraHunkMatches)
				cur.Body = bytes.Join(curLines, nil)
				newHighlights[len(results)] = protocol.HunkHighlight{LineHighlights: curHighlights}
				curHighlights = make(map[int]protocol.Ranges)
				results = append(results, cur)
				cur = nil
				curLines = nil
				extraHunkMatches = 0
			}

			if !lineInfo.added {
				origLineOffset++
			}
			if !lineInfo.removed {
				newLineOffset++
			}
		}

		if cur != nil {
			addExtraHunkMatchesSection(cur, extraHunkMatches)
			cur.Body = bytes.Join(curLines, nil)
			results = append(results, cur)
		}
	}

	return results, newHighlights
}

type diffHunkLineInfo struct {
	added    bool // line starts with '+'
	removed  bool // line starts with '-'
	matching bool // line matches query (only computed for changed lines)
	context  bool // include because it's context for a matching changed line
}

func (info diffHunkLineInfo) changed() bool { return info.added || info.removed }

func computeDiffHunkInfo(lines [][]byte, lineHighlights map[int]protocol.Ranges, matchContextLines int) []diffHunkLineInfo {
	// Return context line numbers for a given line number.
	contextLines := func(line int) (start, end int) {
		start = line - matchContextLines
		if start < 0 {
			start = 0
		}
		end = line + matchContextLines
		if end > len(lines)-1 {
			end = len(lines) - 1
		}
		return
	}

	lineInfo := make([]diffHunkLineInfo, len(lines))
	for i, line := range lines {
		lineInfo[i].added, lineInfo[i].removed = diffHunkLineStatus(line)
		if lineInfo[i].changed() {
			_, ok := lineHighlights[i]
			lineInfo[i].matching = ok
			if lineInfo[i].matching {
				// Mark context lines before/after matching lines.
				start, end := contextLines(i)
				for j := start; j <= end; j++ {
					if i == j {
						continue
					}
					lineInfo[j].context = true
				}
			}
		}
	}
	return lineInfo
}

func diffHunkLineStatus(line []byte) (added, removed bool) {
	if len(line) >= 1 {
		added = line[0] == '+'
		removed = line[0] == '-'
	}
	return
}

// DiffFetcher is a handle to the stdin and stdout of a git diff-tree subprocess
// started with StartDiffFetcher
type DiffFetcher struct {
	stdin   io.Writer
	stderr  bytes.Buffer
	scanner *bufio.Scanner
}

// StartDiffFetcher starts a git diff-tree subprocess that waits, listening on stdin
// for comimt hashes to generate patches for.
func StartDiffFetcher(ctx context.Context, dir string) (*DiffFetcher, error) {
	cmd := exec.CommandContext(ctx, "git", "diff-tree", "--stdin", "--no-prefix", "-p", "--format=format:")
	cmd.Dir = dir

	stdoutReader, stdoutWriter := io.Pipe()
	cmd.Stdout = stdoutWriter

	stdinReader, stdinWriter := io.Pipe()
	cmd.Stdin = stdinReader

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdoutReader)
	scanner.Buffer(make([]byte, 1024), 1<<30)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Note that this only works when we write to stdin, then read from stdout before writing
		// anything else to stdin, since we are using `HasSuffix` and not `Contains`.
		if bytes.HasSuffix(data, []byte("ENDOFPATCH\n")) {
			if bytes.Equal(data, []byte("ENDOFPATCH\n")) {
				// Empty patch
				return len(data), data[:0], nil
			}
			return len(data), data[:len(data)-len("ENDOFPATCH\n")], nil
		}

		return 0, nil, nil
	})

	return &DiffFetcher{
		stdin:   stdinWriter,
		scanner: scanner,
		stderr:  stderrBuf,
	}, nil
}

// FetchDiff fetches a diff from the git diff-tree subprocess, writing to its stdin
// and waiting for its response on stdout. Note that this is not safe to call concurrently.
func (d *DiffFetcher) FetchDiff(hash []byte) ([]byte, error) {
	// HACK: There is no way (as far as I can tell) to make `git diff-tree --stdin` to
	// write a trailing null byte or tell us how much to read in advance, and since we're
	// using a long-running process, the stream doesn't close at the end, and we can't use the
	// start of a new patch to signify end of patch since we want to be able to do each round-trip
	// serially. We resort to sending the subprocess a bogus commit hash named "ENDOFPATCH", which it
	// will fail to read as a tree, and print back to stdout literally. We use this as a signal
	// that the subprocess is done outputting for this commit.
	d.stdin.Write(append(hash, []byte("\nENDOFPATCH\n")...))

	if d.scanner.Scan() {
		return d.scanner.Bytes(), nil
	} else if err := d.scanner.Err(); err != nil {
		return nil, err
	} else if d.stderr.String() != "" {
		return nil, errors.Errorf("git subprocess stderr: %s", d.stderr.String())
	}
	return nil, errors.New("expected scan to succeed")
}
