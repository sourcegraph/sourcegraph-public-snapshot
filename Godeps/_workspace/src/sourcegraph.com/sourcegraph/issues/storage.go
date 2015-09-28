package issues

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

var pluginPath string = "issues"

func sgpath() string {
	sgpath := filepath.SplitList(os.Getenv("SGPATH"))
	if len(sgpath) == 0 {
		curUser, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(curUser.HomeDir, ".sourcegraph")
	} else {
		return sgpath[0]
	}
}

func readIssues(path string) (*IssueList, error) {
	dataPath := filepath.Join(sgpath(), pluginPath, path)
	err := os.MkdirAll(dataPath, 0755)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dataPath)
	var ret []*Issue
	for _, file := range files {
		i := new(Issue)
		p := filepath.Join(dataPath, file.Name())
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(content, i)
		ret = append(ret, i)
	}
	il := &IssueList{ret}
	sort.Sort(il)
	return il, nil
}

// Sort most recently updated first
func (il IssueList) Len() int           { return len(il.Issues) }
func (il IssueList) Swap(i, j int)      { il.Issues[i], il.Issues[j] = il.Issues[j], il.Issues[i] }
func (il IssueList) Less(i, j int) bool { return il.Issues[i].Updated.After(il.Issues[j].Updated) }

func issueNumber(path string) (int64, error) {
	issues, err := readIssues(path)
	if err != nil {
		return 0, err
	}
	var max int64 = 0
	for _, issue := range issues.Issues {
		if issue.UID > max {
			max = issue.UID
		}
	}
	return max + int64(1), nil
}

func writeIssue(path string, i *issueInternal) (int64, error) {
	// TODO mutex.Lock()
	if i.UID == 0 {
		n, err := issueNumber(path)
		i.UID = n
		if err != nil {
			return 0, err
		}
	}
	i.Updated = time.Now()
	j, err := json.Marshal(i)
	if err != nil {
		return 0, err
	}
	num := strconv.FormatInt(int64(i.UID), 10)
	issuePath := filepath.Join(sgpath(), pluginPath, path, num)
	return i.UID, ioutil.WriteFile(issuePath, j, 0644)
}
