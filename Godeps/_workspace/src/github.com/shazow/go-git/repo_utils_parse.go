package git

import (
	"bytes"
	"container/list"
)

const prettyLogFormat = `--pretty=format:%H`

func parsePrettyFormatLog(repo *Repository, logByts []byte) (*list.List, error) {
	l := list.New()
	if len(logByts) == 0 {
		return l, nil
	}

	parts := bytes.Split(logByts, []byte{'\n'})

	for _, commitId := range parts {
		commit, err := repo.GetCommit(string(commitId))
		if err != nil {
			return nil, err
		}
		l.PushBack(commit)
	}

	return l, nil
}
