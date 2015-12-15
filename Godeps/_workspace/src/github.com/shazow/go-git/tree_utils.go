package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
)

type TreeScanner struct {
	parent *Tree

	*bufio.Scanner
	closer io.Closer

	treeEntry *TreeEntry
	err       error
}

func NewTreeScanner(parent *Tree, rc io.ReadCloser) *TreeScanner {
	ts := &TreeScanner{
		parent:  parent,
		Scanner: bufio.NewScanner(rc),
		closer:  rc,
	}
	ts.Split(ScanTreeEntry)
	return ts
}

var TreeEntryRe = regexp.MustCompile("^([0-9]+) ([^\x00]+)\x00")

func (t *TreeScanner) parse() error {
	t.treeEntry = nil

	data := t.Bytes()

	match := TreeEntryRe.FindSubmatch(data)
	if match == nil {
		return fmt.Errorf("failed to parse tree entry: %q", data)
	}

	modeString, name := string(match[1]), string(match[2])
	id, err := NewId(data[len(match[0]):])
	if err != nil {
		return err
	}

	entryMode, objectType, err := ParseModeType(modeString)
	if err != nil {
		return err
	}

	t.treeEntry = &TreeEntry{
		name:  name,
		mode:  entryMode,
		Id:    id,
		Type:  objectType,
		ptree: t.parent,
	}

	return nil
}

func (t *TreeScanner) Scan() bool {
	if !t.Scanner.Scan() {
		if t.closer != nil {
			// Upon hitting any error, close the input.
			t.closer.Close()
			t.closer = nil
		}
		return false
	}

	t.err = t.parse()
	if t.err != nil {
		return false
	}

	return true
}

func (t *TreeScanner) Err() error {
	// Underlying IO errs take priority
	if err := t.Scanner.Err(); err != nil {
		return err
	}
	return t.err
}

func ScanTreeEntry(
	data []byte,
	atEOF bool,
) (
	advance int, token []byte, err error,
) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	const shaLen = 20

	nullIndex := bytes.IndexByte(data, '\x00')
	recordLength := nullIndex + 1 + shaLen
	if recordLength <= len(data) {
		// We found 20 bytes after a null, we're done.
		return recordLength, data[:recordLength], nil
	}

	if atEOF {
		// atEOF but don't have a complete record
		return 0, nil, fmt.Errorf("malformed record %q", data)
	}

	return 0, nil, nil // Request more data.
}

func (t *TreeScanner) TreeEntry() *TreeEntry {
	return t.treeEntry
}

func ParseModeType(modeString string) (EntryMode, ObjectType, error) {
	switch modeString {
	case "100644":
		return ModeBlob, ObjectBlob, nil
	case "100755":
		return ModeExec, ObjectBlob, nil
	case "120000":
		return ModeSymlink, ObjectBlob, nil
	case "160000":
		return ModeCommit, ObjectCommit, nil
	case "40000":
		return ModeTree, ObjectTree, nil
	default:
	}
	return 0, 0, fmt.Errorf("unknown type: %q", modeString)
}
