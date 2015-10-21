package hgcmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/internal"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/util"
	"sourcegraph.com/sqs/pbtypes"

	"golang.org/x/tools/godoc/vfs"
)

func init() {
	vcs.RegisterOpener("hg", func(dir string) (vcs.Repository, error) {
		return Open(dir)
	})
	vcs.RegisterCloner("hg", func(url, dir string, opt vcs.CloneOpt) (vcs.Repository, error) {
		return CloneHgRepository(url, dir, opt)
	})
}

func CloneHgRepository(url, dir string, opt vcs.CloneOpt) (*Repository, error) {
	args := []string{"clone"}
	if opt.Bare {
		args = append(args, "--noupdate")
	}
	args = append(args, "--", url, dir)
	cmd := exec.Command("hg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec `hg clone` failed: %s. Output was:\n\n%s", err, out)
	}
	return Open(dir)
}

func Open(dir string) (*Repository, error) {
	if _, err := os.Stat(filepath.Join(dir, ".hg")); err != nil {
		// All hg directories have a ".hg" directory; hg does not have
		// the concept of bare repos (a "bare repo" in mercurial is
		// simply a directory with only a ".hg"). So this is the only
		// check we need.
		return nil, &os.PathError{
			Op:   "Open",
			Path: filepath.Join(dir, ".hg"),
			Err:  errors.New("Mercurial repository not found."),
		}
	}
	return &Repository{dir}, nil
}

type Repository struct {
	Dir string
}

func (r *Repository) RepoDir() string {
	return r.Dir
}

func (r *Repository) ResolveRevision(spec string) (vcs.CommitID, error) {
	cmd := exec.Command("hg", "identify", "--debug", "-i", "--rev="+spec)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		out = bytes.TrimSpace(out)
		if isUnknownRevisionError(string(out), spec) {
			return "", vcs.ErrRevisionNotFound
		}
		return "", fmt.Errorf("exec `hg identify` failed: %s. Output was:\n\n%s", err, out)
	}
	return vcs.CommitID(bytes.TrimSpace(out)), nil
}

func (r *Repository) ResolveTag(name string) (vcs.CommitID, error) {
	commitID, err := r.ResolveRevision(name)
	if err == vcs.ErrRevisionNotFound {
		return "", vcs.ErrTagNotFound
	}
	return commitID, nil
}

func (r *Repository) ResolveBranch(name string) (vcs.CommitID, error) {
	commitID, err := r.ResolveRevision(name)
	if err == vcs.ErrRevisionNotFound {
		return "", vcs.ErrBranchNotFound
	}
	return commitID, nil
}

func (r *Repository) Branches(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	if opt.ContainsCommit != "" {
		return nil, fmt.Errorf("vcs.BranchesOptions.ContainsCommit option not implemented")
	}

	refs, err := r.execAndParseCols("branches")
	if err != nil {
		return nil, err
	}

	branches := make([]*vcs.Branch, len(refs))
	for i, ref := range refs {
		branches[i] = &vcs.Branch{
			Name: ref[1],
			Head: vcs.CommitID(ref[0]),
		}
	}
	return branches, nil
}

func (r *Repository) Tags() ([]*vcs.Tag, error) {
	refs, err := r.execAndParseCols("tags")
	if err != nil {
		return nil, err
	}

	tags := make([]*vcs.Tag, len(refs))
	for i, ref := range refs {
		tags[i] = &vcs.Tag{
			Name:     ref[1],
			CommitID: vcs.CommitID(ref[0]),
		}
	}
	return tags, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p byteSlices) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (r *Repository) execAndParseCols(subcmd string) ([][2]string, error) {
	cmd := exec.Command("hg", "-v", "--debug", subcmd)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec `hg -v --debug %s` failed: %s. Output was:\n\n%s", subcmd, err, out)
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	lines := bytes.Split(out, []byte("\n"))
	sort.Sort(byteSlices(lines)) // sort for consistency
	refs := make([][2]string, len(lines))
	for i, line := range lines {
		line = bytes.TrimSuffix(line, []byte(" (inactive)"))

		// format: "NAME      SEQUENCE:ID" (arbitrary amount of whitespace between NAME and SEQUENCE)
		if len(line) <= 41 {
			return nil, fmt.Errorf("unexpectedly short (<=41 bytes) line in `hg -v --debug %s` output", subcmd)
		}
		id := line[len(line)-40:]

		// find where the SEQUENCE begins
		seqIdx := bytes.LastIndex(line, []byte(" "))
		if seqIdx == -1 {
			return nil, fmt.Errorf("unexpectedly no whitespace in line in `hg -v --debug %s` output", subcmd)
		}
		name := bytes.TrimRight(line[:seqIdx], " ")
		refs[i] = [2]string{string(id), string(name)}
	}
	return refs, nil
}

func (r *Repository) GetCommit(id vcs.CommitID) (*vcs.Commit, error) {
	commits, _, err := r.commitLog(vcs.CommitsOptions{Head: id, N: 1, NoTotal: true})
	if err != nil {
		return nil, err
	}

	if len(commits) != 1 {
		return nil, fmt.Errorf("hg log: expected 1 commit, got %d", len(commits))
	}

	return commits[0], nil
}

func (r *Repository) Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	return r.commitLog(opt)
}

var hgNullParentNodeID = []byte("0000000000000000000000000000000000000000")

func isUnknownRevisionError(output, revSpec string) bool {
	return output == "abort: unknown revision '"+string(revSpec)+"'!"
}

func (r *Repository) commitLog(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	revSpec := string(opt.Head)
	if opt.Skip != 0 {
		revSpec += "~" + strconv.FormatUint(uint64(opt.N), 10)
	}

	args := []string{"log", `--template={node}\x00{author|person}\x00{author|email}\x00{date|rfc3339date}\x00{desc}\x00{p1node}\x00{p2node}\x00`}
	if opt.N != 0 {
		args = append(args, "--limit", strconv.FormatUint(uint64(opt.N), 10))
	}
	args = append(args, "--rev="+revSpec+":0")

	cmd := exec.Command("hg", args...)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		out = bytes.TrimSpace(out)
		if isUnknownRevisionError(string(out), revSpec) {
			return nil, 0, vcs.ErrCommitNotFound
		}
		return nil, 0, fmt.Errorf("exec `hg log` failed: %s. Output was:\n\n%s", err, out)
	}

	const partsPerCommit = 7 // number of \x00-separated fields per commit
	allParts := bytes.Split(out, []byte{'\x00'})
	numCommits := len(allParts) / partsPerCommit
	commits := make([]*vcs.Commit, numCommits)
	for i := 0; i < numCommits; i++ {
		parts := allParts[partsPerCommit*i : partsPerCommit*(i+1)]
		id := vcs.CommitID(parts[0])

		authorTime, err := time.Parse(time.RFC3339, string(parts[3]))
		if err != nil {
			log.Println(err)
			//return nil, 0, err
		}

		parents, err := r.getParents(id)
		if err != nil {
			return nil, 0, fmt.Errorf("r.GetParents failed: %s. Output was:\n\n%s", err, out)
		}

		commits[i] = &vcs.Commit{
			ID:      id,
			Author:  vcs.Signature{string(parts[1]), string(parts[2]), pbtypes.NewTimestamp(authorTime)},
			Message: string(parts[4]),
			Parents: parents,
		}
	}

	// Count commits.
	var total uint
	if !opt.NoTotal {
		cmd = exec.Command("hg", "id", "--num", "--rev="+revSpec)
		cmd.Dir = r.Dir
		out, err = cmd.CombinedOutput()
		if err != nil {
			return nil, 0, fmt.Errorf("exec `hg id --num` failed: %s. Output was:\n\n%s", err, out)
		}
		out = bytes.TrimSpace(out)
		total, err = parseUint(string(out))
		if err != nil {
			return nil, 0, err
		}
		total++ // sequence number is 1 less than total number of commits

		// Add back however many we skipped.
		total += opt.Skip
	}

	return commits, total, nil
}

func parseUint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	return uint(n), err
}

func (r *Repository) getParents(revSpec vcs.CommitID) ([]vcs.CommitID, error) {
	var parents []vcs.CommitID

	cmd := exec.Command("hg", "parents", "-r", string(revSpec), "--template",
		`{node}\x00{author|person}\x00{author|email}\x00{date|rfc3339date}\x00{desc}\x00{p1node}\x00{p2node}\x00`)
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec `hg parents` failed: %s. Output was:\n\n%s", err, out)
	}

	const partsPerCommit = 7 // number of \x00-separated fields per commit
	allParts := bytes.Split(out, []byte{'\x00'})
	numCommits := len(allParts) / partsPerCommit
	for i := 0; i < numCommits; i++ {
		parts := allParts[partsPerCommit*i : partsPerCommit*(i+1)]

		if p1 := parts[0]; len(p1) > 0 && !bytes.Equal(p1, hgNullParentNodeID) {
			parents = append(parents, vcs.CommitID(p1))
		}
		if p2 := parts[5]; len(p2) > 0 && !bytes.Equal(p2, hgNullParentNodeID) {
			parents = append(parents, vcs.CommitID(p2))
		}
		if p3 := parts[6]; len(p3) > 0 && !bytes.Equal(p3, hgNullParentNodeID) {
			parents = append(parents, vcs.CommitID(p3))
		}
	}

	return parents, nil
}

func (r *Repository) Diff(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	cmd := exec.Command("hg", "-v", "diff", "-p", "--git", "--rev="+string(base), "--rev="+string(head), "--")
	if opt != nil {
		cmd.Args = append(cmd.Args, opt.Paths...)
	}
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		out = bytes.TrimSpace(out)
		if isUnknownRevisionError(string(out), string(base)) || isUnknownRevisionError(string(out), string(head)) {
			return nil, vcs.ErrCommitNotFound
		}
		return nil, fmt.Errorf("exec `hg diff` failed: %s. Output was:\n\n%s", err, out)
	}

	if opt == nil {
		opt = &vcs.DiffOptions{}
	}

	// Hackily apply OrigPrefix and NewPrefix.
	fdiffs, err := diff.ParseMultiFileDiff(out)
	if err != nil {
		return nil, err
	}
	for _, f := range fdiffs {
		for i, x := range f.Extended {
			f.Extended[i] = strings.Replace(strings.Replace(x, "b/", opt.NewPrefix, 1), "a/", opt.OrigPrefix, 1)
		}
		f.OrigName = filepath.Join(opt.OrigPrefix, strings.TrimPrefix(f.OrigName, "a/"))
		f.NewName = filepath.Join(opt.NewPrefix, strings.TrimPrefix(f.NewName, "b/"))
	}
	out, err = diff.PrintMultiFileDiff(fdiffs)
	if err != nil {
		return nil, err
	}

	return &vcs.Diff{
		Raw: string(out),
	}, nil
}

func (r *Repository) UpdateEverything(opt vcs.RemoteOpts) error {
	if opt.SSH != nil {
		return fmt.Errorf("hgcmd: ssh remote not supported")
	}
	cmd := exec.Command("hg", "pull")
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec `hg pull` failed: %s. Output was:\n\n%s", err, out)
	}
	return nil
}

func (r *Repository) BlameFile(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	if opt == nil {
		opt = &vcs.BlameOptions{}
	}

	// TODO(sqs): implement OldestCommit
	cmd := exec.Command("python", "-", r.Dir, string(opt.NewestCommit), path)
	cmd.Dir = r.Dir
	cmd.Stdin = strings.NewReader(hgRepoAnnotatePy)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	in := bufio.NewReader(stdout)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var data struct {
		Commits map[string]struct {
			Author     struct{ Name, Email string }
			AuthorDate time.Time
		}
		Hunks map[string][]struct {
			CommitID           string
			StartLine, EndLine int
			StartByte, EndByte int
		}
	}
	jsonErr := json.NewDecoder(in).Decode(&data)
	errOut, _ := ioutil.ReadAll(stderr)
	if jsonErr != nil {
		cmd.Wait()
		return nil, fmt.Errorf("%s (stderr: %s)", jsonErr, errOut)
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s (stderr: %s)", err, errOut)
	}

	hunks := make([]*vcs.Hunk, len(data.Hunks[path]))
	for i, hunk := range data.Hunks[path] {
		c := data.Commits[hunk.CommitID]
		hunks[i] = &vcs.Hunk{
			StartLine: hunk.StartLine,
			EndLine:   hunk.EndLine,
			StartByte: hunk.StartByte,
			EndByte:   hunk.EndByte,
			CommitID:  vcs.CommitID(hunk.CommitID),
			Author: vcs.Signature{
				Name:  c.Author.Name,
				Email: c.Author.Email,
				Date:  pbtypes.NewTimestamp(c.AuthorDate.In(time.UTC)),
			},
		}
	}
	return hunks, nil
}

func (r *Repository) Committers(opt vcs.CommittersOptions) ([]*vcs.Committer, error) {
	return nil, fmt.Errorf("Committers() not implemented for vcs type: hg")
}

func (r *Repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	return &hgFSCmd{
		dir: r.Dir,
		at:  at,
	}, nil
}

type hgFSCmd struct {
	dir string
	at  vcs.CommitID
}

func (fs *hgFSCmd) Open(name string) (vfs.ReadSeekCloser, error) {
	name = internal.Rel(name)
	cmd := exec.Command("hg", "cat", "--rev="+string(fs.at), "--", name)
	cmd.Dir = fs.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("no such file in rev")) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("exec `hg cat` failed: %s. Output was:\n\n%s", err, out)
	}
	return util.NopCloser{bytes.NewReader(out)}, nil
}

func (fs *hgFSCmd) Lstat(path string) (os.FileInfo, error) {
	return fs.Stat(internal.Rel(path))
}

func (fs *hgFSCmd) Stat(path string) (os.FileInfo, error) {
	// TODO(sqs): follow symlinks (as Stat is required to do)

	path = internal.Rel(path)
	var mtime time.Time

	cmd := exec.Command("hg", "log", "-l1", `--template={date|date}`,
		"-r "+string(fs.at)+":0", "--", path)
	cmd.Dir = fs.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	mtime, err = time.Parse("Mon Jan 02 15:04:05 2006 -0700",
		strings.Trim(string(out), "\n"))
	if err != nil {
		log.Println(err)
		// return nil, err
	}

	// this just determines if the file exists.
	cmd = exec.Command("hg", "locate", "--rev="+string(fs.at), "--", path)
	cmd.Dir = fs.dir
	err = cmd.Run()
	if err != nil {
		// hg doesn't track dirs, so use a workaround to see if path is a dir.
		if _, err := fs.ReadDir(path); err == nil {
			return &util.FileInfo{Name_: filepath.Base(path), Mode_: os.ModeDir,
				ModTime_: mtime}, nil
		}
		return nil, os.ErrNotExist
	}

	// read file to determine file size
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)

	return &util.FileInfo{Name_: filepath.Base(path), Size_: int64(len(data)),
		ModTime_: mtime}, nil
}

func (fs *hgFSCmd) ReadDir(path string) ([]os.FileInfo, error) {
	path = filepath.Clean(internal.Rel(path))
	// This combination of --include and --exclude opts gets all the files in
	// the dir specified by path, plus all files one level deeper (but no
	// deeper). This lets us list the files *and* subdirs in the dir without
	// needlessly listing recursively.
	cmd := exec.Command("hg", "locate", "--rev="+string(fs.at), "--include="+path, "--exclude="+filepath.Clean(path)+"/*/*/*")
	cmd.Dir = fs.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("exec `hg cat` failed: %s. Output was:\n\n%s", err, out)
	}

	subdirs := make(map[string]struct{})
	prefix := []byte(path + "/")
	files := bytes.Split(out, []byte{'\n'})
	var fis []os.FileInfo
	for _, nameb := range files {
		nameb = bytes.TrimPrefix(nameb, prefix)
		if len(nameb) == 0 {
			continue
		}
		if bytes.Contains(nameb, []byte{'/'}) {
			subdir := strings.SplitN(string(nameb), "/", 2)[0]
			if _, seen := subdirs[subdir]; !seen {
				fis = append(fis, &util.FileInfo{Name_: subdir, Mode_: os.ModeDir})
				subdirs[subdir] = struct{}{}
			}
			continue
		}
		fis = append(fis, &util.FileInfo{Name_: filepath.Base(string(nameb))})
	}

	return fis, nil
}

func (fs *hgFSCmd) String() string {
	return fmt.Sprintf("hg repository %s commit %s (cmd)", fs.dir, fs.at)
}
