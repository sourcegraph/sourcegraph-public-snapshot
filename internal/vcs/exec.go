package vcs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

var DeniedCommands = map[string]any{
	"show":      nil,
	"rev-parse": nil,
	"log":       nil,
	"diff":      nil,
	"ls-tree":   nil,
}

const TTL = time.Hour * 24 * 7

type FilterFunc func(cmd *exec.Cmd) bool

func NewGitExec(cmd *exec.Cmd, filters ...FilterFunc) *GitExec {
	return &GitExec{
		filters: append(filters, checkDeniedCommand()),
		Cmd:     cmd,
	}
}

type GitExec struct {
	filters []FilterFunc
	*exec.Cmd
}

func RecordRepo(name string) FilterFunc {
	return func(cmd *exec.Cmd) bool {
		return strings.Contains(cmd.Dir, name)
	}
}

func (g GitExec) CombinedOutput() ([]byte, error) {
	if err := g.record(false); err != nil {
		return nil, err
	}
	return g.Cmd.CombinedOutput()
}

func (g GitExec) Run() error {
	if err := g.record(false); err != nil {
		return err
	}
	return g.Cmd.Run()
}

func (g GitExec) Wait() error {
	if err := g.record(true); err != nil {
		return err
	}
	return g.Cmd.Wait()
}

func checkDeniedCommand() FilterFunc {
	return func(cmd *exec.Cmd) bool {
		if len(cmd.Args) < 2 {
			return false
		}
		command := cmd.Args[1]
		_, ok := DeniedCommands[command]
		return ok
	}
}

func all(cmd *exec.Cmd, filters ...FilterFunc) bool {
	for _, fn := range filters {
		if !fn(cmd) {
			return false
		}
	}
	return true
}

func (g GitExec) record(imprecise bool) error {
	if !all(g.Cmd, g.filters...) {
		return nil
	}
	// record this command in redis
	r := rcache.New("gitexec")
	val := struct {
		Start     time.Time `json:"Start"`
		Args      []string  `json:"Args"`
		Dir       string    `json:"Dir"`
		Imprecise bool      `json:"Imprecise"`
	}{
		Start:     time.Now(),
		Args:      g.Args,
		Dir:       g.Dir,
		Imprecise: imprecise,
	}
	data, err := json.Marshal(&val)
	if err != nil {
		return err
	}

	r.SetWithTTL(fmt.Sprintf("git-%v", time.Now().Unix()), data, int(TTL.Seconds()))
	return nil
}
