package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type CommitData struct {
	parent       string
	pathStatuses []PathStatus
}

type MockGit struct {
	commitToCommitData map[string]CommitData
}

func NewMockGit() MockGit {
	return MockGit{
		commitToCommitData: map[string]CommitData{},
	}
}

func (git MockGit) LogReverse(repo string, commit string, n int) ([]LogEntry, error) {
	logEntries := []LogEntry{}
	for commit != "" && n > 0 {
		data, ok := git.commitToCommitData[commit]
		if !ok {
			break
		}
		logEntries = append(logEntries, LogEntry{
			Commit:       commit,
			PathStatuses: data.pathStatuses,
		})
		commit = data.parent
		n -= 1
	}

	// Reverse
	for i, j := 0, len(logEntries)-1; i < j; i, j = i+1, j-1 {
		logEntries[i], logEntries[j] = logEntries[j], logEntries[i]
	}

	return logEntries, nil
}

func (git MockGit) RevList(repo string, commit string) ([]string, error) {
	commits := []string{}
	for commit != "" {
		commits = append(commits, commit)
		commit = git.commitToCommitData[commit].parent
	}
	return commits, nil
}

func RandomCommit() string {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func (git MockGit) AddCommit(commit string, data CommitData) {
	git.commitToCommitData[commit] = data
}

type MockDB struct {
	commitToHeight        map[string]int
	commitToAncestor      map[string]string
	pathToHopToStatusToId map[string]map[string]map[StatusAD]int
	blobs                 map[int]*Blob
}

func NewMockDB() MockDB {
	return MockDB{
		commitToHeight:        map[string]int{},
		commitToAncestor:      map[string]string{},
		pathToHopToStatusToId: map[string]map[string]map[StatusAD]int{},
		blobs:                 map[int]*Blob{},
	}
}

func (db MockDB) GetCommit(givenCommit string) (commit string, height int, present bool, err error) {
	height, ok := db.commitToHeight[givenCommit]
	if !ok {
		return "", 0, false, nil
	}
	ancestor, ok := db.commitToAncestor[givenCommit]
	if !ok {
		return "", 0, false, nil
	}

	return ancestor, height, true, nil
}

func (db MockDB) InsertCommit(commit string, height int, ancestor string) error {
	db.commitToHeight[commit] = height
	db.commitToAncestor[commit] = ancestor
	return nil
}

func (db MockDB) GetBlob(hop string, status StatusAD, path string) (id int, found bool, err error) {
	hopToStatusToId, ok := db.pathToHopToStatusToId[path]
	if !ok {
		return 0, false, nil
	}
	statusToId, ok := hopToStatusToId[hop]
	if !ok {
		return 0, false, nil
	}
	foundId, ok := statusToId[status]
	if !ok {
		return 0, false, nil
	}

	return foundId, true, nil
}

func (db MockDB) UpdateBlobHops(id int, status StatusAD, hop string) error {
	if status == AddedAD {
		db.blobs[id].added = append(db.blobs[id].added, hop)
	}
	if status == DeletedAD {
		db.blobs[id].deleted = append(db.blobs[id].deleted, hop)
	}

	if _, ok := db.pathToHopToStatusToId[db.blobs[id].path]; !ok {
		db.pathToHopToStatusToId[db.blobs[id].path] = map[string]map[StatusAD]int{}
	}
	if _, ok := db.pathToHopToStatusToId[db.blobs[id].path][hop]; !ok {
		db.pathToHopToStatusToId[db.blobs[id].path][hop] = map[StatusAD]int{}
	}
	db.pathToHopToStatusToId[db.blobs[id].path][hop][status] = id

	return nil
}

func (db MockDB) InsertBlob(blob Blob) error {
	id := len(db.blobs)
	db.blobs[id] = &blob
	if _, ok := db.pathToHopToStatusToId[blob.path]; !ok {
		db.pathToHopToStatusToId[blob.path] = map[string]map[StatusAD]int{}
	}
	for _, hop := range blob.added {
		if _, ok := db.pathToHopToStatusToId[blob.path][hop]; !ok {
			db.pathToHopToStatusToId[blob.path][hop] = map[StatusAD]int{}
		}
		db.pathToHopToStatusToId[blob.path][hop][AddedAD] = id
	}
	for _, hop := range blob.deleted {
		if _, ok := db.pathToHopToStatusToId[blob.path][hop]; !ok {
			db.pathToHopToStatusToId[blob.path][hop] = map[StatusAD]int{}
		}
		db.pathToHopToStatusToId[blob.path][hop][DeletedAD] = id
	}
	return nil
}

func (db MockDB) AppendHop(hops []string, givenStatus StatusAD, newHop string) error {
	for _, hop := range hops {
		for _, hopToStatusToId := range db.pathToHopToStatusToId {
			for status, id := range hopToStatusToId[hop] {
				if status == givenStatus {
					db.UpdateBlobHops(id, status, newHop)
				}
			}
		}
	}

	return nil
}

func (db MockDB) Search(hops []string) ([]Blob, error) {
	blobs := []Blob{}
NextPath:
	for _, hopToStatusToId := range db.pathToHopToStatusToId {
		for _, hop := range hops {
			if _, ok := hopToStatusToId[hop][DeletedAD]; ok {
				continue NextPath
			}
			blobs = append(blobs, *db.blobs[hopToStatusToId[hop][AddedAD]])
		}
	}
	return blobs, nil
}

func (db MockDB) PrintInternals() {
	fmt.Println("Commit ancestry:")
	fmt.Println()

	heights := []int{}
	for _, height := range db.commitToHeight {
		heights = append(heights, height)
	}
	sort.Ints(heights)

	heightToCommits := map[int][]string{}
	for commit, height := range db.commitToHeight {
		if _, ok := heightToCommits[height]; !ok {
			heightToCommits[height] = []string{}
		}
		heightToCommits[height] = append(heightToCommits[height], commit)
	}

	for _, height := range heights {
		for _, commit := range heightToCommits[height] {
			ancestor := db.commitToAncestor[commit]
			ancestorHeight := db.commitToHeight[ancestor]
			fmt.Printf("  - height %d commit %-40s ancestor %-40s (height %d)\n", height, commit, ancestor, ancestorHeight)
		}
	}

	fmt.Println()
	fmt.Println("Blobs:")
	fmt.Println()

	ids := []int{}
	for id := range db.blobs {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		blob := db.blobs[id]
		fmt.Printf("  id %d path %-10s\n", id, blob.path)
		for _, added := range blob.added {
			height := db.commitToHeight[added]
			fmt.Printf("    + %-40s (height %d)\n", added, height)
		}
		fmt.Println()
		for _, deleted := range blob.deleted {
			height := db.commitToHeight[deleted]
			fmt.Printf("    - %-40s (height %d)\n", deleted, height)
		}
		fmt.Println()
	}

	fmt.Println()
}

func TestIndex(t *testing.T) {
	git := NewMockGit()
	db := NewMockDB()

	c1 := RandomCommit()
	c2 := RandomCommit()
	c3 := RandomCommit()
	c4 := RandomCommit()
	git.AddCommit(c1, CommitData{parent: NULL, pathStatuses: []PathStatus{{Path: "foo.go", Status: AddedAMD}}})
	git.AddCommit(c2, CommitData{parent: c1, pathStatuses: []PathStatus{{Path: "bar.go", Status: AddedAMD}}})
	git.AddCommit(c3, CommitData{parent: c2, pathStatuses: []PathStatus{{Path: "foo.go", Status: DeletedAMD}}})
	git.AddCommit(c4, CommitData{parent: c3, pathStatuses: []PathStatus{{Path: "foo.go", Status: AddedAMD}}})

	err := Index(git, db, "github.com/foo/bar", c4)
	if err != nil {
		t.Fatalf("🚨 Index: %s", err)
	}

	blobs, err := Search(db, c4)
	if err != nil {
		t.Fatalf("🚨 PathsAtCommit: %s", err)
	}
	paths := []string{}
	for _, blob := range blobs {
		paths = append(paths, blob.path)
	}

	expected := []string{"foo.go", "bar.go"}

	sort.Strings(paths)
	sort.Strings(expected)

	if diff := cmp.Diff(paths, expected); diff != "" {
		fmt.Println("🚨 PathsAtCommit: unexpected paths (-got +want)", diff)
		db.PrintInternals()
		t.Fail()
	}
}
