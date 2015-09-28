package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/dep"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/grapher"
	"sourcegraph.com/sourcegraph/srclib/unit"

	"github.com/kr/fs"
)

func init() {
	c, err := CLI.AddCommand("lint",
		"detect common issues in srclib output data",
		`The lint command checks srclib output files (*.graph.json, *.unit.json, *.depresolve.json, etc.) for common data integrity and correctness issues:

* Refs that point to nonexistent defs in the same source unit

* Refs, defs, and source units whose 'Files' and/or 'Dir' fields do not exist in the repository

Note that the lint command operates on single files at a time, so it can't detect cross-source-unit or cross-repo ref resolution errors (only those on refs to defs in the same source unit).

If no PATHs are specified, the current directory is used. If a PATH is a directory, it is traversed recursively for files named with any of the above suffixes.

To suppress specific kinds of warnings, or to include only specific kinds of warnings, pipe the output of the lint command through grep.
`,
		&lintCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	SetDefaultRepoOpt(c)
}

type LintCmd struct {
	Repo string `short:"r" long:"repo" description:"repository URI (defaults to VCS 'srclib' or 'origin' remote URL)"`

	NoCheckFiles   bool `long:"no-check-files" description:"don't check that file/dir fields refer to actual files"`
	NoCheckResolve bool `long:"no-check-resolve" description:"don't check that internal refs resolve to existing defs"`

	Args struct {
		Paths []string `name:"PATH" description:"path to srclib JSON output file, or a directory tree of such"`
	} `positional-args:"YES"`
}

var lintCmd LintCmd

func (c *LintCmd) Execute(args []string) error {
	if len(c.Args.Paths) == 0 {
		c.Args.Paths = []string{"."}
	}

	issuec := make(chan string)
	quitc := make(chan struct{})
	go func() {
		for {
			select {
			case issue := <-issuec:
				fmt.Println(issue)
			case <-quitc:
				return
			}
		}
	}()

	lrepo, lrepoErr := OpenLocalRepo()
	if lrepoErr != nil {
		log.Printf("warning: while opening current dir's repo: %s", lrepoErr)
	}

	var wg sync.WaitGroup
	for _, path := range c.Args.Paths {
		w := fs.Walk(path)
		for w.Step() {
			if err := w.Err(); err != nil {
				return err
			}
			if fi := w.Stat(); fi.Mode().IsRegular() {
				suffix, typ := buildstore.DataType(fi.Name())
				if suffix != "" {
					absPath, err := filepath.Abs(w.Path())
					if err != nil {
						return err
					}

					var unitType, unitName string
					if !c.NoCheckResolve {
						if !strings.Contains(absPath, buildstore.BuildDataDirName) {
							return fmt.Errorf("couldn't infer which source unit %s corresponds to, because its absolute path is not under any %s dir; either run with --no-check-resolve to skip checking that internal refs resolve to valid defs (which requires knowing what source unit each output file is from), or run 'srclib lint' against .srclib-cache or subdirectories of it", w.Path(), buildstore.BuildDataDirName)
						}
						unitType = strings.TrimSuffix(fi.Name(), "."+suffix+".json")
						// Infer source unit name from file path (the
						// path components after .srclib-cache until
						// the basename).
						pcs := strings.Split(absPath, string(os.PathSeparator))
						for i, pc := range pcs {
							if pc == buildstore.BuildDataDirName && len(pcs) > i+2 {
								unitName = filepath.Clean(strings.Join(pcs[i+2:len(pcs)-1], "/"))
								break
							}
						}
					}

					var commitID string
					if !c.NoCheckFiles {
						// Infer commit ID from file path (the path component after .srclib-cache).
						pcs := strings.Split(absPath, string(os.PathSeparator))
						for i, pc := range pcs {
							if pc == buildstore.BuildDataDirName && len(pcs) > i+1 {
								commitID = pcs[i+1]
								break
							}
						}
					}
					if commitID == "" && !c.NoCheckFiles {
						return fmt.Errorf("couldn't infer which commit ID %s was built from, which is necessary to check that file/dir fields refer to actual files; either run with --no-check-files to skip the file/dir check or pass paths that contain '.../.srclib-cache/COMMITID/...' (which allows this command to infer the commit ID)", w.Path())
					}

					// Ensure that commitID matches the local repo's commit ID.
					if commitID != "" && !c.NoCheckFiles {
						if lrepo != nil && lrepo.CommitID != commitID {
							return fmt.Errorf("%s was built from commit %s, but the current repo working tree HEAD is commit %s; these must be the same to check that file/dir fields refer to actual files in the repo that their specific commit, so you must either (1) only run lint against build data files for commit %s; (2) run with --no-check-files to skip the file/dir check; or (3) check out commit %s in this repo", w.Path(), commitID, lrepo.CommitID, lrepo.CommitID, commitID)
						}
					}

					checkFilesExist := !c.NoCheckFiles

					wg.Add(1)
					go func(path string) {
						defer wg.Done()

						var issues []string
						var err error
						switch typ.(type) {
						case unit.SourceUnit:
							issues, err = lintSourceUnit(lrepo.RootDir, path, checkFilesExist)
						case *graph.Output:
							issues, err = lintGraphOutput(lrepo.RootDir, c.Repo, unitType, unitName, path, checkFilesExist)
						case []*dep.ResolvedDep:
							issues, err = lintDepresolveOutput(lrepo.RootDir, path, checkFilesExist)
						}
						for _, issue := range prependLabelToStrings(path, issues) {
							issuec <- issue
						}
						if err != nil {
							log.Fatalf(redbg("ERR")+" %s: %s", path, err)
						}
					}(w.Path())
				}
			}
		}
	}

	wg.Wait()
	close(quitc)

	return nil
}

func prependLabelToStrings(prefix string, ss []string) []string {
	ps := make([]string, len(ss))
	for i, s := range ss {
		ps[i] = prefix + ": " + s
	}
	return ps
}

func lintSourceUnit(baseDir, path string, checkFilesExist bool) (issues []string, err error) {
	var u unit.SourceUnit
	if err := readJSONFile(path, &u); err != nil {
		return nil, err
	}

	if u.Name == "" {
		issues = append(issues, `Name: must be non-empty (consider using "." to mean "root directory"; remember that unit names are opaque to srclib except for string equality)`)
	}

	if u.Type == "" {
		issues = append(issues, "Type: must be non-empty (e.g., 'GoPackage', 'RubyGem'; remember that unit types are opaque to srclib except for string equality)")
	}

	if len(u.Ops) == 0 {
		issues = append(issues, `Ops: it rarely makes sense for this to be empty unless no further analysis (graph/depresolve) should be performed on this source unit; a typical value is {"graph": null, "depresolve": null}, which means that the source unit should be graphed and depresolved (you can specify specific other toolchains to perform those steps if you wish, instead of saying 'null')`)
	}

	if len(u.Files) > 100 {
		issues = append(issues, fmt.Sprintf(`Files: list contains a large number of files (%d files), which could lead to slow builds and inefficient incrmental rebuilds; if the language and build system allow, consider devising a different definition of source unit that leads to smaller source units`, len(u.Files)))
	}

	if u.Repo != "" {
		issues = append(issues, "Repo: can be left blank by scanner (will be filled in at import time)")
	}

	issues0, err := lintCheckFiles(baseDir, checkFilesExist, nil, u.Files...)
	issues = append(issues, prependLabelToStrings("SourceUnit.Files", issues0)...)
	if err != nil {
		return issues, err
	}

	isDir := func(fi os.FileInfo) error {
		if !fi.Mode().IsDir() {
			return errors.New("not a dir")
		}
		return nil
	}
	issues0, err = lintCheckFiles(baseDir, checkFilesExist, isDir, u.Dir)
	issues = append(issues, prependLabelToStrings("SourceUnit.Dir", issues0)...)
	if err != nil {
		return issues, err
	}

	return issues, nil
}

func lintGraphOutput(baseDir, repoURI, unitType, unitName, path string, checkFilesExist bool) (issues []string, err error) {
	var o graph.Output
	if err := readJSONFile(path, &o); err != nil {
		return nil, err
	}

	checkOrigin := func(label, repo, commitID, unitType, unit string) {
		if repo != "" {
			issues = append(issues, fmt.Sprintf("%s: %s", label, "Repo field can be left blank by grapher (the repository from which it was built is implied; will be filled in at import time)"))
		}
		if commitID != "" {
			issues = append(issues, fmt.Sprintf("%s: %s", label, "CommitID can be left blank by grapher (the commit from which it was built is implied; will be filled in at import time)"))
		}
		if unitType != "" {
			issues = append(issues, fmt.Sprintf("%s: %s", label, "UnitType can be left blank by grapher (containing source unit's unit type is implied; will be filled in at import time)"))
		}
		if unit != "" {
			issues = append(issues, fmt.Sprintf("%s: %s", label, "Unit can be left blank by grapher (containing unit is implied; will be filled in at import time)"))
		}
	}
	checkFile := func(label, file string) error {
		issues0, err := lintCheckFiles(baseDir, checkFilesExist, nil, file)
		issues = append(issues, prependLabelToStrings(label+": File", issues0)...)
		return err
	}

	for _, def := range o.Defs {
		label := "Def " + def.DefKey.String()
		checkOrigin(label, def.Repo, def.CommitID, def.UnitType, def.Unit)
		if def.DefStart == 0 && def.DefEnd == 0 {
			issues = append(issues, fmt.Sprintf("Def %s: DefStart and DefEnd are both 0; if possible, set these values to the byte offet range of the definition (but not all language constructs have meaningful ranges)", def.DefKey))
		}
		if def.Path == "" {
			issues = append(issues, fmt.Sprintf(`Def %s: Path must not be empty (if you want to designate a def as 'top-level', consider using ".", but remember that def paths are opaque to srclib and it's all up to whatever convention you establish`, def.DefKey))
		}
		if err := checkFile(label, def.File); err != nil {
			return issues, err
		}
	}
	for _, ref := range o.Refs {
		label := fmt.Sprintf("Ref %+v", ref)
		checkOrigin(label, ref.Repo, ref.CommitID, ref.UnitType, ref.Unit)
		if err := checkFile(label, ref.File); err != nil {
			return issues, err
		}
	}
	for _, ann := range o.Anns {
		label := fmt.Sprintf("Ann %+v", ann)
		checkOrigin(label, ann.Repo, ann.CommitID, ann.UnitType, ann.Unit)
		if err := checkFile(label, ann.File); err != nil {
			return issues, err
		}
	}
	for _, doc := range o.Docs {
		label := fmt.Sprintf("Doc %+v", doc.DefKey)
		checkOrigin(label, doc.Repo, doc.CommitID, doc.UnitType, doc.Unit)
		if doc.File != "" {
			// It's OK for Doc.File to be empty because they are attached to defs.
			if err := checkFile(label, doc.File); err != nil {
				return issues, err
			}
		}
	}

	addMultiErrorAsIssues := func(errs grapher.MultiError) {
		for _, issue := range errs {
			issues = append(issues, issue.Error())
		}
	}

	// Fill in implied fields.
	grapher.PopulateImpliedFields(repoURI, "", unitType, unitName, &o)

	// Check that defs and refs are unique.
	addMultiErrorAsIssues(grapher.ValidateDefs(o.Defs))
	addMultiErrorAsIssues(grapher.ValidateRefs(o.Refs))
	addMultiErrorAsIssues(grapher.ValidateDocs(o.Docs))

	// TODO(sqs): check that docs point to valid defs in the same source unit

	unresolvedInternalRefsByDefKey := grapher.UnresolvedInternalRefs(repoURI, o.Refs, o.Defs)
	for defKey, unresolvedIRefs := range unresolvedInternalRefsByDefKey {
		var refLines []string
		const max = 20
		if len(unresolvedIRefs) > max {
			refLines = append(refLines, fmt.Sprintf("NOTE: only showing first %d/%d unresolved refs", max, len(unresolvedIRefs)))
		}
		for i, ref := range unresolvedIRefs {
			if i >= max {
				break
			}
			refLines = append(refLines, fmt.Sprintf("%s bytes %d-%d", ref.File, ref.Start, ref.End))
		}
		refsList := strings.Join(refLines, "; ")

		issues = append(issues, fmt.Sprintf("%d intra-source-unit refs point to nonexistent def %s: refs are at: %s", len(unresolvedIRefs), defKey, refsList))
	}

	return issues, nil
}

func lintDepresolveOutput(baseDir, path string, checkFilesExist bool) (issues []string, err error) {
	// TODO(sqs): lint depresolve output
	return nil, nil
}

func lintCheckFiles(baseDir string, checkExist bool, fn func(os.FileInfo) error, paths ...string) (issues []string, err error) {
	if fn == nil {
		fn = func(fi os.FileInfo) error {
			if !fi.Mode().IsRegular() {
				return errors.New("not a file")
			}
			return nil
		}
	}

	seenOrig := map[string]int{}
	seenClean := map[string]int{}
	for _, path := range paths {
		cpath := filepath.Clean(path)
		seenOrig[path]++
		seenClean[cpath]++
		if seenOrig[path] > 1 {
			issues = append(issues, fmt.Sprintf("path %q appears %d other times in list", path, seenOrig[path]))
			continue
		} else if seenClean[cpath] > 1 {
			issues = append(issues, fmt.Sprintf("path %q and equivalent paths (after cleaning path) appear %d other times in list", path, seenClean[cpath]))
			continue
		}
		if path == "" {
			issues = append(issues, fmt.Sprintf(`path is empty string (use "." if that's what you intended)`))
			continue
		}
		if filepath.IsAbs(cpath) {
			issues = append(issues, fmt.Sprintf("path %s is absolute (should be relative to the repository root dir)", cpath))
		}

		if checkExist {
			absPath := filepath.Join(baseDir, cpath)
			fi, err := os.Stat(absPath)
			if os.IsNotExist(err) {
				issues = append(issues, fmt.Sprintf("path %s does not exist (note: paths must be relative to the repository root, not their source unit's dir)", cpath))
				continue
			} else if err != nil {
				return nil, err
			}
			if err := fn(fi); err != nil {
				issues = append(issues, fmt.Sprintf("path %s is %s", cpath, err))
			}
		}
	}
	return issues, nil
}
