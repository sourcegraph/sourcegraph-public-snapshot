package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kr/fs"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/dep"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/plan"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	c, err := CLI.AddCommand("api",
		"API",
		"",
		&apiCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	/* START APIDescribeCmdDoc OMIT
	This command is used by editor plugins to retrieve information about
	the identifier at a specific position in a file.

	It will hit Sourcegraph's API to get a definition's examples. With the
	flag `--no-examples`, this command does not hit Sourcegraph's API.
		END APIDescribeCmdDoc OMIT */
	_, err = c.AddCommand("describe",
		"display documentation for the def under the cursor",
		"Returns information about the definition referred to by the cursor's current position in a file.",
		&apiDescribeCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	/* START APIListCmdDoc OMIT
	This command will return a list of all the definitions,
	references, and docs in a file. It can be used for finding all
	uses of a reference in a file.
	END APIListCmdDoc OMIT */
	_, err = c.AddCommand("list",
		"list all defs, refs, and docs in a given file",
		"Return a list of all definitions, references, and docs that are in the current file.",
		&apiListCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	/* START APIDepsCmdDoc OMIT
	This command returns a list of all resolved and unresolved
	dependencies for the current repository.
		END APIDepsCmdDoc OMIT */
	_, err = c.AddCommand("deps",
		"list all resolved and unresolved dependencies",
		`Return a list of all resolved and unresolved dependencies that are in the current repository.`,
		&apiDepsCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	/* START APIUnitsCmdDoc OMIT
	This command returns a list of all of the source units in the current
	repository.
		END APIUnitsCmdDoc OMIT */
	_, err = c.AddCommand("units",
		"list all source unit information",
		"Return a list of all source units that are in the current repository.",
		&apiUnitsCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type APICmd struct{}

var apiCmd APICmd

func (c *APICmd) Execute(args []string) error { return nil }

type APIDescribeCmd struct {
	File      string `long:"file" required:"yes" value-name:"FILE"`
	StartByte uint32 `long:"start-byte" required:"yes" value-name:"BYTE"`
}

type APIListCmd struct {
	File   string `long:"file" required:"yes" value-name:"FILE"`
	NoRefs bool   `long:"no-refs"`
	NoDefs bool   `long:"no-defs"`
	NoDocs bool   `long:"no-docs"`
}

type APIDepsCmd struct {
	Args struct {
		Dir Directory `name:"DIR" default:"." description:"root directory of target project"`
	} `positional-args:"yes"`
}

type APIUnitsCmd struct {
	Args struct {
		Dir Directory `name:"DIR" default:"." description:"root directory of target project"`
	} `positional-args:"yes"`
}

var apiDescribeCmd APIDescribeCmd
var apiListCmd APIListCmd
var apiDepsCmd APIDepsCmd
var apiUnitsCmd APIUnitsCmd

type commandContext struct {
	repo         *Repo
	relativeFile string
	buildStore   buildstore.RepoBuildStore
	commitFS     rwvfs.WalkableFileSystem
}

// prepareCommandContext prepare the context for the the "src"
// command. If file is "." or ends in a "/", then it is a directory. It is safe
// for "file" to have multiple trailing slashes. prepareCommandContext
// creates commandContext, changes the process' working directory to
// file's directory and ensures that a build has been made. It is
// meant to be used by user-facing commands.
func prepareCommandContext(file string) (commandContext, error) {
	var (
		err   error
		c     commandContext
		isDir bool
	)
	if file == "" {
		return commandContext{}, errors.New("prepareCommandContext: file cannot be empty")
	}
	if file == "." || file[len(file)-1] == os.PathSeparator {
		isDir = true
	}
	file, err = filepath.Abs(file)
	if err != nil {
		return commandContext{}, err
	}
	// filepath.Abs returns a cleaned file path, so we need to add
	// the path separator to it to presrve filepath.Dir's
	// semantics.
	if isDir {
		file += string(os.PathSeparator)
	}

	repo, err := OpenRepo(filepath.Dir(file))
	if err != nil {
		return commandContext{}, err
	}
	c.repo = repo

	rel, err := filepath.Rel(repo.RootDir, file)
	if err != nil {
		return commandContext{}, err
	}
	c.relativeFile = rel

	if err := os.Chdir(repo.RootDir); err != nil {
		return commandContext{}, err
	}

	buildStore, err := buildstore.LocalRepo(repo.RootDir)
	if err != nil {
		return commandContext{}, err
	}
	c.buildStore = buildStore
	c.commitFS = buildStore.Commit(repo.CommitID)

	if err := ensureBuild(buildStore, repo); err != nil {
		if err := buildstore.RemoveAllDataForCommit(buildStore, repo.CommitID); err != nil {
			log.Println(err)
		}
		return commandContext{}, err
	}
	return c, nil
}

// ensureBuild invokes the build process on the given repository
func ensureBuild(buildStore buildstore.RepoBuildStore, repo *Repo) error {
	configOpt := config.Options{
		Repo:   repo.URI(),
		Subdir: ".",
	}
	toolchainExecOpt := ToolchainExecOpt{ExeMethods: "program"}

	// Config repository if not yet built.
	exists, err := buildstore.BuildDataExistsForCommit(buildStore, repo.CommitID)
	if err != nil {
		return err
	}
	if !exists {
		configCmd := &ConfigCmd{
			Options:          configOpt,
			ToolchainExecOpt: toolchainExecOpt,
			Quiet:            true,
		}
		if err := configCmd.Execute(nil); err != nil {
			return err
		}
	}

	// Always re-make.
	//
	// TODO(sqs): optimize this
	makeCmd := &MakeCmd{
		Options:          configOpt,
		ToolchainExecOpt: toolchainExecOpt,
		Quiet:            true,
	}
	if err := makeCmd.Execute(nil); err != nil {
		return err
	}

	// Always re-import.
	i := &StoreImportCmd{
		ImportOpt: ImportOpt{
			Repo:     repo.CloneURL,
			CommitID: repo.CommitID,
		},
		Quiet: true,
	}
	if err := i.Execute(nil); err != nil {
		return err
	}
	return nil
}

func getSourceUnits(commitFS rwvfs.WalkableFileSystem, repo *Repo) []string {
	var unitFiles []string
	unitSuffix := buildstore.DataTypeSuffix(unit.SourceUnit{})
	w := fs.WalkFS(".", commitFS)
	for w.Step() {
		if strings.HasSuffix(w.Path(), unitSuffix) {
			unitFiles = append(unitFiles, w.Path())
		}
	}
	return unitFiles
}

// getSourceUnitsWithFile gets a list of all source units that contain
// the given file.
func getSourceUnitsWithFile(buildStore buildstore.RepoBuildStore, repo *Repo, filename string) ([]*unit.SourceUnit, error) {
	filename = filepath.Clean(filename)

	// TODO(sqs): This whole lookup is totally inefficient. The storage format
	// is not optimized for lookups.
	commitFS := buildStore.Commit(repo.CommitID)
	unitFiles := getSourceUnits(commitFS, repo)

	// Find all source unit definition files.
	unitSuffix := buildstore.DataTypeSuffix(unit.SourceUnit{})
	w := fs.WalkFS(".", commitFS)
	for w.Step() {
		if strings.HasSuffix(w.Path(), unitSuffix) {
			unitFiles = append(unitFiles, w.Path())
		}
	}

	// Find which source units the file belongs to.
	var units []*unit.SourceUnit
	for _, unitFile := range unitFiles {
		var u *unit.SourceUnit
		f, err := commitFS.Open(unitFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&u); err != nil {
			return nil, fmt.Errorf("%s: %s", unitFile, err)
		}
		for _, f2 := range u.Files {
			if filepath.Clean(f2) == filename {
				units = append(units, u)
				break
			}
		}
	}

	return units, nil
}

// START APIListCmdOutput OMIT
type apiListCmdOutput struct {
	Defs []*graph.Def `json:",omitempty"`
	Refs []*graph.Ref `json:",omitempty"`
	Docs []*graph.Doc `json:",omitempty"`
}

// END APIListCmdOutput OMIT

func (c *APIListCmd) Execute(args []string) error {
	context, err := prepareCommandContext(c.File)
	if err != nil {
		return err
	}

	file := context.relativeFile
	units, err := getSourceUnitsWithFile(context.buildStore, context.repo, file)
	if err != nil {
		return err
	}

	if GlobalOpt.Verbose {
		if len(units) > 0 {
			ids := make([]string, len(units))
			for i, u := range units {
				ids[i] = string(u.ID())
			}
			log.Printf("File %s is in %d source units %v.", file, len(units), ids)
		} else {
			log.Printf("File %s is not in any source units.", file)
		}
	}

	// Grab all the data for the file.
	var output apiListCmdOutput
	for _, u := range units {
		var g graph.Output
		graphFile := plan.SourceUnitDataFilename("graph", u)
		f, err := context.commitFS.Open(graphFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&g); err != nil {
			return fmt.Errorf("%s: %s", graphFile, err)
		}
		if !c.NoRefs {
			for _, ref := range g.Refs {
				if file == ref.File {
					output.Refs = append(output.Refs, ref)
				}
			}
		}
		if !c.NoDefs {
			for _, def := range g.Defs {
				if file == def.File {
					output.Defs = append(output.Defs, def)
				}
			}
		}
		if !c.NoDocs {
			for _, doc := range g.Docs {
				if file == doc.File {
					output.Docs = append(output.Docs, doc)
				}
			}
		}

	}

	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
		return err
	}
	return nil
}

/* START APIDescribeCmdOutput OMIT

The output is defined in
[api_cmds.go](https://github.com/sourcegraph/srclib/blob/e5295dfcd719535ff9cbb37a2771337d44fe5953/src/api_cmds.go#L190-L193),
as the JSON representation of the following struct.

The Def and Example structs are defined as follows in the Sourcegraph API.

[[.code "src/api_cmds.go" "APIDescribeCmdOutputQuickHack"]]

[[.code "https://raw.githubusercontent.com/sourcegraph/go-sourcegraph/6937daba84bf2d0f919191fd74e5193171b4f5d5/sourcegraph/defs.go" 105 113]]

[[.code "graph/def.pb.go" "Def "]]

[[.code "https://raw.githubusercontent.com/sourcegraph/go-sourcegraph/6937daba84bf2d0f919191fd74e5193171b4f5d5/sourcegraph/defs.go" 236 252]]

[[.code "graph/ref.pb.go" "Ref"]]

END APIDescribeCmdOutput OMIT */
// START APIDescribeCmdOutputQuickHack OMIT
type apiDescribeCmdOutput struct {
	Def *graph.Def
}

// END APIDescribeCmdOutputQuickHack OMIT

func (c *APIDescribeCmd) Execute(args []string) error {
	context, err := prepareCommandContext(c.File)
	if err != nil {
		return err
	}
	file := context.relativeFile
	units, err := getSourceUnitsWithFile(context.buildStore, context.repo, file)
	if err != nil {
		return err
	}

	if GlobalOpt.Verbose {
		if len(units) > 0 {
			ids := make([]string, len(units))
			for i, u := range units {
				ids[i] = string(u.ID())
			}
			log.Printf("Position %s:%d is in %d source units %v.", file, c.StartByte, len(units), ids)
		} else {
			log.Printf("Position %s:%d is not in any source units.", file, c.StartByte)
		}
	}

	// Find the ref(s) at the character position.
	var ref *graph.Ref
	var nearbyRefs []*graph.Ref // Find nearby refs to help with debugging.
OuterLoop:
	for _, u := range units {
		var g graph.Output
		graphFile := plan.SourceUnitDataFilename("graph", u)
		f, err := context.commitFS.Open(graphFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&g); err != nil {
			return fmt.Errorf("%s: %s", graphFile, err)
		}
		for _, ref2 := range g.Refs {
			if file == ref2.File {
				if c.StartByte >= ref2.Start && c.StartByte <= ref2.End {
					ref = ref2
					if ref.DefUnit == "" {
						ref.DefUnit = u.Name
					}
					if ref.DefUnitType == "" {
						ref.DefUnitType = u.Type
					}
					break OuterLoop
				} else if GlobalOpt.Verbose && abs(int(ref2.Start)-int(c.StartByte)) < 25 {
					nearbyRefs = append(nearbyRefs, ref2)
				}
			}
		}
	}

	if ref == nil {
		if GlobalOpt.Verbose {
			log.Printf("No ref found at %s:%d.", file, c.StartByte)

			if len(nearbyRefs) > 0 {
				log.Printf("However, nearby refs were found in the same file:")
				for _, nref := range nearbyRefs {
					log.Printf("Ref at bytes %d-%d to %v", nref.Start, nref.End, nref.DefKey())
				}
			}

			f, err := os.Open(file)
			if err == nil {
				defer f.Close()
				b, err := ioutil.ReadAll(f)
				if err != nil {
					log.Fatalf("Error reading source file: %s.", err)
				}
				start := c.StartByte
				if start < 0 || int(start) > len(b)-1 {
					log.Fatalf("Start byte %d is out of file bounds.", c.StartByte)
				}
				end := c.StartByte + 50
				if int(end) > len(b)-1 {
					end = uint32(len(b) - 1)
				}
				log.Printf("Surrounding source is:\n\n%s", b[start:end])
			} else {
				log.Printf("Error opening source file to show surrounding source: %s.", err)
			}
		}
		fmt.Println(`{}`)
		return nil
	}

	// ref.DefRepo is *not* guaranteed to be non-empty, as
	// repo.URI() will return the empty string if the repo's
	// CloneURL is empty or malformed.
	if ref.DefRepo == "" {
		ref.DefRepo = context.repo.URI()
	}

	var resp apiDescribeCmdOutput
	// Now find the def for this ref.

	defInCurrentRepo := ref.DefRepo == context.repo.URI()
	if defInCurrentRepo {
		// Def is in the current repo.
		var g graph.Output
		graphFile := plan.SourceUnitDataFilename("graph", &unit.SourceUnit{Name: ref.DefUnit, Type: ref.DefUnitType})
		f, err := context.commitFS.Open(graphFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&g); err != nil {
			return fmt.Errorf("%s: %s", graphFile, err)
		}
		for _, def := range g.Defs {
			if def.Path == ref.DefPath {
				resp.Def = def
				break
			}
		}
		if resp.Def != nil {
			// If Def is in the current Repo, transform that path to be an absolute path
			resp.Def.File = filepath.Join(context.repo.RootDir, resp.Def.File)
		}
		if resp.Def == nil && GlobalOpt.Verbose {
			log.Printf("No definition found with path %q in unit %q type %q.", ref.DefPath, ref.DefUnit, ref.DefUnitType)
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		return err
	}
	return nil
}

func abs(n int) int {
	if n < 0 {
		return -1 * n
	}
	return n
}

/* START APIDepsCmdOutput OMIT
This command returns a dep.Resolution slice.

[[.code "dep/resolve.go" "Resolution"]]
END APIDepsCmdOutput OMIT */

func (c *APIDepsCmd) Execute(args []string) error {
	// HACK(samertm): append a backslash to Dir to assure that it's parsed
	// as a directory, but Directory should have an unmarshalling
	// method that does this.
	context, err := prepareCommandContext(c.Args.Dir.String())
	if err != nil {
		return err
	}

	var depSlice []*dep.Resolution
	// TODO: Make DataTypeSuffix work with type of depSlice
	depSuffix := buildstore.DataTypeSuffix([]*dep.ResolvedDep{})
	depCache := make(map[string]struct{})
	foundDepresolve := false
	w := fs.WalkFS(".", context.commitFS)
	for w.Step() {
		depfile := w.Path()
		if strings.HasSuffix(depfile, depSuffix) {
			foundDepresolve = true
			var deps []*dep.Resolution
			f, err := context.commitFS.Open(depfile)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := json.NewDecoder(f).Decode(&deps); err != nil {
				return fmt.Errorf("%s: %s", depfile, err)
			}
			for _, d := range deps {
				key, err := d.RawKeyId()
				if err != nil {
					return err
				}
				if _, ok := depCache[key]; !ok {
					depCache[key] = struct{}{}
					depSlice = append(depSlice, d)
				}
			}
		}
	}

	if foundDepresolve == false {
		return fmt.Errorf("No dependency information found. Try running `%s config` first.", srclib.CommandName)
	}

	return json.NewEncoder(os.Stdout).Encode(depSlice)
}

/* START APIUnitsCmdOutput OMIT
This command returns a unit.SourceUnit slice.

[[.code "unit/source_unit.go" "SourceUnit"]]
END APIUnitsCmdOutput OMIT */

func (c *APIUnitsCmd) Execute(args []string) error {
	context, err := prepareCommandContext(c.Args.Dir.String())
	if err != nil {
		return err
	}

	var unitSlice []unit.SourceUnit
	unitSuffix := buildstore.DataTypeSuffix(unit.SourceUnit{})
	foundUnit := false
	w := fs.WalkFS(".", context.commitFS)
	for w.Step() {
		unitFile := w.Path()
		if strings.HasSuffix(unitFile, unitSuffix) {
			var unit unit.SourceUnit
			foundUnit = true
			f, err := context.commitFS.Open(unitFile)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := json.NewDecoder(f).Decode(&unit); err != nil {
				return fmt.Errorf("%s: %s", unitFile, err)
			}
			unitSlice = append(unitSlice, unit)
		}
	}

	if foundUnit == false {
		return fmt.Errorf("No source units found. Try running `%s config` first.", srclib.CommandName)
	}

	return json.NewEncoder(os.Stdout).Encode(unitSlice)
}
