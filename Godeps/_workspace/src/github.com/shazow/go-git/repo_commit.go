package git

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var refRexp = regexp.MustCompile("ref: (.*)\n")

var ErrNoPackedRefs = errors.New("no packed-refs")

// RefNotFound error returned when a commit is fetched by ref that is not found.
type RefNotFound string

func (err RefNotFound) Error() string {
	return fmt.Sprintf("ref not found: %s", string(err))
}

// get branch's last commit or a special commit by id string
func (repo *Repository) GetCommitOfBranch(branchName string) (*Commit, error) {
	commitId, err := repo.GetCommitIdOfBranch(branchName)
	if err != nil {
		return nil, err
	}

	return repo.GetCommit(commitId)
}

func (repo *Repository) GetCommitIdOfBranch(branchName string) (string, error) {
	return repo.getCommitIdOfRef("refs/heads/" + branchName)
}

func (repo *Repository) GetCommitOfTag(tagName string) (*Commit, error) {
	commitId, err := repo.GetCommitIdOfTag(tagName)
	if err != nil {
		return nil, err
	}

	return repo.GetCommit(commitId)
}

func (repo *Repository) GetCommitIdOfTag(tagName string) (string, error) {
	return repo.getCommitIdOfRef("refs/tags/" + tagName)
}

func (repo *Repository) getCommitIdOfRef(refpath string) (string, error) {
start:
	path := filepath.Join(repo.Path, refpath)
	f, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		// Check for packed ref
		var packedErr error
		f, packedErr = repo.getCommitIdOfPackedRef(refpath)
		if packedErr == ErrNoPackedRefs {
			return "", RefNotFound(refpath)
		} else if packedErr != nil {
			return "", packedErr
		}
	} else if err != nil {
		return "", err
	}

	allMatches := refRexp.FindAllStringSubmatch(string(f), 1)
	if allMatches == nil {
		// let's assume this is a sha1
		if len(f) < 40 {
			return "", errors.New("sha1 hash too short")
		}
		sha1 := string(f[:40])
		if !IsSha1(sha1) {
			return "", fmt.Errorf("heads file wrong sha1 string %s", sha1)
		}
		return sha1, nil
	}
	// yes, it's "ref: something". Now let's lookup "something"
	refpath = allMatches[0][1]
	goto start
}

func (repo *Repository) getCommitIdOfPackedRef(refpath string) ([]byte, error) {
	path := filepath.Join(repo.Path, "packed-refs")
	f, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return nil, ErrNoPackedRefs
	}
	defer f.Close()

	scan := bufio.NewScanner(f)

	for scan.Scan() {
		if strings.Contains(scan.Text(), refpath) {
			return scan.Bytes(), nil
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return nil, RefNotFound(refpath)
}

// Find the commit object in the repository.
func (repo *Repository) GetCommit(commitId string) (*Commit, error) {
	id, err := NewIdFromString(commitId)
	if err != nil {
		return nil, err
	}

	return repo.getCommit(id)
}

func (repo *Repository) getCommit(id sha1) (*Commit, error) {
	if repo.commitCache != nil {
		if c, ok := repo.commitCache[id]; ok {
			return c, nil
		}
	} else {
		repo.commitCache = make(map[sha1]*Commit, 10)
	}

	_, _, dataRc, err := repo.getRawObject(id, false)
	if err != nil {
		return nil, err
	}

	defer func() {
		dataRc.Close()
	}()

	// TODO reader
	data, err := ioutil.ReadAll(dataRc)
	if err != nil {
		return nil, err
	}

	commit, err := parseCommitData(data)
	if err != nil {
		return nil, err
	}
	commit.repo = repo
	commit.Id = id

	repo.commitCache[id] = commit

	return commit, nil
}

// used only for single tree, (]
func (repo *Repository) CommitsBetween(last *Commit, before *Commit) (*list.List, error) {
	l := list.New()
	if last == nil || last.ParentCount() == 0 {
		return l, nil
	}

	var err error
	cur := last
	for {
		if cur.Id.Equal(before.Id) {
			break
		}
		l.PushBack(cur)
		if cur.ParentCount() == 0 {
			break
		}
		cur, err = cur.Parent(0)
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (repo *Repository) CommitsBefore(commitId string) (*list.List, error) {
	id, err := NewIdFromString(commitId)
	if err != nil {
		return nil, err
	}

	return repo.getCommitsBefore(id)
}

func (repo *Repository) getCommitsBefore(id sha1) (*list.List, error) {
	l := list.New()
	lock := new(sync.Mutex)
	err := repo.commitsBefore(lock, l, nil, id, 0)
	return l, err
}

func (repo *Repository) commitsBefore(lock *sync.Mutex, l *list.List, parent *list.Element, id sha1, limit int) error {
	commit, err := repo.getCommit(id)
	if err != nil {
		return err
	}

	var e *list.Element
	if parent == nil {
		e = l.PushBack(commit)
	} else {
		var in = parent
		//lock.Lock()
		for {
			if in == nil {
				break
			} else if in.Value.(*Commit).Id.Equal(commit.Id) {
				//lock.Unlock()
				return nil
			} else {
				if in.Next() == nil {
					break
				}
				if in.Value.(*Commit).Committer.When.Equal(commit.Committer.When) {
					break
				}

				if in.Value.(*Commit).Committer.When.After(commit.Committer.When) &&
					in.Next().Value.(*Commit).Committer.When.Before(commit.Committer.When) {
					break
				}
			}
			in = in.Next()
		}

		e = l.InsertAfter(commit, in)
		//lock.Unlock()
	}

	var pr = parent
	if commit.ParentCount() > 1 {
		pr = e
	}

	for i := 0; i < commit.ParentCount(); i++ {
		id, err := commit.ParentId(i)
		if err != nil {
			return err
		}
		err = repo.commitsBefore(lock, l, pr, id, 0)
		if err != nil {
			return err
		}
	}

	return nil
}
