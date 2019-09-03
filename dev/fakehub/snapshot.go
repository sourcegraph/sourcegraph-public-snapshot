package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func snapshot(src, dst string) error {
	name := filepath.Base(src)

	dst = filepath.Join(dst, ".git")
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// create bare repo if missing
	if _, err := os.Stat(filepath.Join(dst, "HEAD")); os.IsNotExist(err) {
		if _, err := run(name, exec.Command("git", "init", "--bare", dst)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	env := []string{
		"GIT_COMMITTER_NAME=sourcegraph-committer",
		"GIT_COMMITTER_EMAIL=support@sourcegraph.com",
		"GIT_AUTHOR_NAME=sourcegraph-committer",
		"GIT_AUTHOR_EMAIL=support@sourcegraph.com",
		"GIT_DIR=" + dst,
		"GIT_WORK_TREE=" + src,
	}

	cmd := exec.Command("git", "status", "--porcelain", "--no-renames")
	cmd.Env = env
	cmd.Dir = src
	n, err := run(name, cmd)
	if err != nil {
		return err
	}

	// no lines in output of git status means nothing changed
	if n == 0 {
		log.Printf("%s: nothing changed", name)
		return nil
	}

	args := [][]string{
		// we can't just git add, since if we are tracking files that are part
		// of .gitignore they will continue to be tracked. So we empty the
		// index.
		[]string{"git", "rm", "-r", "-q", "--cached", "--ignore-unmatch", "."},

		// git add -A makes the index reflect the work tree
		[]string{"git", "add", "-A"},

		[]string{"git", "commit", "-m", "Sync at " + time.Now().Format("Mon Jan _2 15:04:05 2006")},
	}
	for _, a := range args {
		cmd := exec.Command(a[0], a[1:]...)
		cmd.Env = env
		cmd.Dir = src
		_, err := run(name, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func run(name string, cmd *exec.Cmd) (int, error) {
	outW := &lineCountWriter{w: os.Stdout, prefix: []byte("> ")}
	errW := &lineCountWriter{w: os.Stdout, prefix: []byte("! ")}

	cmd.Stdout = outW
	cmd.Stderr = errW

	log.Printf("%s> %v", name, strings.Join(cmd.Args, " "))
	err := cmd.Run()

	_ = outW.Close()
	_ = errW.Close()

	return outW.lines, err
}

type lineCountWriter struct {
	w      io.Writer
	prefix []byte

	inline bool
	lines  int
}

func (w *lineCountWriter) Write(b []byte) (int, error) {
	n := 0
	for len(b) > 0 {
		if !w.inline {
			w.lines++
			_, err := w.w.Write(w.prefix)
			if err != nil {
				return n, err
			}
		}

		var off int
		if i := bytes.Index(b, []byte{'\n'}); i < 0 {
			off = len(b)
			w.inline = true
		} else {
			off = i + 1 // include newline
			w.inline = false
		}

		var part []byte
		part, b = b[:off], b[off:]

		m, err := w.w.Write(part)
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (w *lineCountWriter) Close() error {
	// write a newline if there is one missing
	if w.inline {
		w.inline = false
		_, err := w.w.Write([]byte{'\n'})
		return err
	}
	return nil
}

// Snapshot creates a commit of Dir into the bare git repo Destination.
type Snapshot struct {
	// PreCommand if non-empty is run before taking the snapshot.
	PreCommand string `yaml:",omitempty"`

	// Dir is the directory to treat as the git working directory.
	Dir string `yaml:",omitempty"`

	// Destination is the directory containing the bare git repo.
	Destination string `yaml:",omitempty"`
}

// Snapshotter runs several
type Snapshotter struct {
	// Dir is the directory PreCommand is run from. If a Snapshot's Dir is
	// relative, it will be resolved relative to this directory. Defaults to
	// PWD.
	Dir string

	// If a Snapshot's Destination is relative, it will be resolved relative
	// to Destination.
	Destination string

	// PreCommand before any snapshots are taken, PreCommand is run from Dir.
	PreCommand string

	// Snapshots is a list of Snapshosts to take.
	Snapshots []Snapshot
}

func (o *Snapshotter) SetDefaults() error {
	if o.Dir == "" {
		d, err := os.Getwd()
		if err != nil {
			return err
		}
		o.Dir = d
	}

	if o.Destination == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		o.Destination = filepath.Join(h, ".sourcegraph", "snapshots")
	}

	for _, s := range o.Snapshots {
		if s.Destination == "" && !filepath.IsAbs(s.Dir) {
			s.Destination = s.Dir
		}

		d, err := abs(o.Dir, s.Dir)
		if err != nil {
			return err
		}
		s.Dir = d

		d, err = abs(o.Destination, s.Destination)
		if err != nil {
			return err
		}
		s.Destination = d
	}

	return nil
}

func abs(root, dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	return filepath.Abs(dir)
}

func (o *Snapshotter) Run() error {
	if err := o.SetDefaults(); err != nil {
		return err
	}

	if o.PreCommand != "" {
		cmd := exec.Command("sh", "-c", o.PreCommand)
		cmd.Dir = o.Dir
		if _, err := run("root", cmd); err != nil {
			return err
		}
	}

	for _, s := range o.Snapshots {
		if s.PreCommand != "" {
			cmd := exec.Command("sh", "-c", s.PreCommand)
			cmd.Dir = s.Dir
			if _, err := run(filepath.Base(s.Dir), cmd); err != nil {
				return err
			}
		}

		if err := snapshot(s.Dir, s.Destination); err != nil {
			return err
		}
	}

	return nil
}
