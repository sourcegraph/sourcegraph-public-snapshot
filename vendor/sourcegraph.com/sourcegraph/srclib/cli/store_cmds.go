package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alexsaveliev/go-colorable-wrapper"
	"github.com/neelance/parallel"

	"golang.org/x/tools/godoc/vfs"

	"sort"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/grapher"
	"sourcegraph.com/sourcegraph/srclib/plan"
	"sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		storeC, err := cli.AddCommand("store",
			"graph store commands",
			"",
			&storeCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
		lrepo, _ := OpenLocalRepo()
		if lrepo != nil && lrepo.RootDir != "" {
			absDir, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			relDir, err := filepath.Rel(absDir, lrepo.RootDir)
			if err == nil {
				SetOptionDefaultValue(storeC.Group, "root", filepath.Join(relDir, store.SrclibStoreDir))
			}
		}

		InitStoreCmds(storeC)
	})
}

func InitStoreCmds(c *flags.Command) {
	importC, err := c.AddCommand("import",
		"import data",
		`The import command imports data (from .srclib-cache) into the store.`,
		&storeImportCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
	SetDefaultCommitIDOpt(importC)

	_, err = c.AddCommand("indexes",
		"list indexes",
		"The indexes command lists all of a store's indexes that match the specified criteria.",
		&storeIndexesCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("index",
		"build indexes",
		"The index command builds indexes that match the specified index criteria. Built indexes are printed to stdout.",
		&storeIndexCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("repos",
		"list repos",
		"The repos command lists all repos that match a filter.",
		&storeReposCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("versions",
		"list versions",
		"The versions command lists all versions that match a filter.",
		&storeVersionsCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("units",
		"list units",
		"The units command lists all units that match a filter.",
		&storeUnitsCmd,
	)
	if err != nil {
		log.Fatal(err)
	}

	defsC, err := c.AddCommand("defs",
		"list defs",
		"The defs command lists all defs that match a filter.",
		&storeDefsCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
	defsC.Aliases = []string{"def"}

	_, err = c.AddCommand("refs",
		"list refs",
		"The refs command lists all refs that match a filter.",
		&storeRefsCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

// OpenStore is called by all of the store subcommands to open the
// store.
var OpenStore func() (interface{}, error) = storeCmd.store

type StoreCmd struct {
	Type   string `short:"t" long:"type" description:"the (multi-)repo store type to use (RepoStore, MultiRepoStore, etc.)" default:"RepoStore"`
	Root   string `short:"r" long:"root" description:"the root of the store (repo clone dir for RepoStore, global path for MultiRepoStore, etc.)" default:".srclib-store"`
	Config string `long:"config" description:"(rarely used) JSON-encoded config for extra config, specific to each store type"`
}

var storeCmd StoreCmd

func (c *StoreCmd) Execute(args []string) error { return nil }

// store returns the store specified by StoreCmd's Type and Root
// options.
func (c *StoreCmd) store() (interface{}, error) {
	fs := rwvfs.OS(c.Root)

	type createParents interface {
		CreateParentDirs(bool)
	}
	if fs, ok := fs.(createParents); ok {
		fs.CreateParentDirs(true)
	}

	switch c.Type {
	case "RepoStore":
		return store.NewFSRepoStore(rwvfs.Walkable(fs)), nil
	case "MultiRepoStore":
		return store.NewFSMultiRepoStore(rwvfs.Walkable(fs), nil), nil
	default:
		return nil, fmt.Errorf("unrecognized store --type value: %q (valid values are RepoStore, MultiRepoStore)", c.Type)
	}
}

type StoreImportCmd struct {
	ImportOpt

	Quiet bool `short:"q" long:"quiet" description:"silence all output"`

	Sample           bool `long:"sample" description:"(sample data) import sample data, not .srclib-cache data"`
	SampleDefs       int  `long:"sample-defs" description:"(sample data) number of sample defs to import" default:"100"`
	SampleRefs       int  `long:"sample-refs" description:"(sample data) number of sample refs to import" default:"100"`
	SampleImportOnly bool `long:"sample-import-only" description:"(sample data) only import, don't demonstrate listing data"`
}

var storeImportCmd StoreImportCmd

func (c *StoreImportCmd) Execute(args []string) error {
	start := time.Now()

	s, err := OpenStore()
	if err != nil {
		return err
	}

	if c.Sample {
		return c.sample(s)
	}

	bdfs, err := GetBuildDataFS(c.CommitID)
	if err != nil {
		return err
	}
	if GlobalOpt.Verbose {
		log.Printf("# Importing build data for %s (commit %s)", c.Repo, c.CommitID)
	}

	if err := Import(bdfs, s, c.ImportOpt); err != nil {
		return err
	}
	if !c.Quiet {
		log.Printf("# Import completed in %s.", time.Since(start))
	}
	return nil
}

type ImportOpt struct {
	DryRun  bool `short:"n" long:"dry-run" description:"print what would be done but don't do anything"`
	NoIndex bool `long:"no-index" description:"don't build indexes (indexes inside a single source unit are always built)"`

	Repo     string `long:"repo" description:"only import for this repo"`
	Unit     string `long:"unit" description:"only import source units with this name"`
	UnitType string `long:"unit-type" description:"only import source units with this type"`
	CommitID string `long:"commit" description:"commit ID of commit whose data to import"`

	Verbose bool
}

// Import imports build data into a RepoStore or MultiRepoStore.
func Import(buildDataFS vfs.FileSystem, stor interface{}, opt ImportOpt) error {
	// Traverse the build data directory for this repo and commit to
	// create the makefile that lists the targets (which are the data
	// files we will import).
	treeConfig, err := config.ReadCached(buildDataFS)
	if err != nil {
		return fmt.Errorf("error calling config.ReadCached: %s", err)
	}
	mf, err := plan.CreateMakefile(".", nil, "", treeConfig)
	if err != nil {
		return fmt.Errorf("error calling plan.Makefile: %s", err)
	}

	// hasIndexableData is set if at least one source unit's graph data is
	// successfully imported to the graph store.
	//
	// This flag is set concurrently in calls to importGraphData, but it doesn't
	// need to be protected by a mutex since its value is only modified in one
	// direction (false to true), and it is only read in the sequential section
	// after parallel.NewRun completes.
	//
	// However, we still protect it with a mutex to avoid data race errors from the
	// Go race detector.
	var (
		mu               sync.Mutex
		hasIndexableData bool
	)

	importGraphData := func(graphFile string, sourceUnit *unit.SourceUnit) error {
		var data graph.Output
		if err := readJSONFileFS(buildDataFS, graphFile, &data); err != nil {
			if err == errEmptyJSONFile {
				log.Printf("Warning: the JSON file is empty for unit %s %s.", sourceUnit.Type, sourceUnit.Name)
				return nil
			}
			if os.IsNotExist(err) {
				log.Printf("Warning: no build data for unit %s %s.", sourceUnit.Type, sourceUnit.Name)
				return nil
			}
			return fmt.Errorf("error reading JSON file %s for unit %s %s: %s", graphFile, sourceUnit.Type, sourceUnit.Name, err)
		}
		if opt.DryRun || GlobalOpt.Verbose {
			log.Printf("# Importing graph data (%d defs, %d refs, %d docs, %d anns) for unit %s %s", len(data.Defs), len(data.Refs), len(data.Docs), len(data.Anns), sourceUnit.Type, sourceUnit.Name)
			if opt.DryRun {
				return nil
			}
		}

		// HACK: Transfer docs to [def].Docs.
		docsByPath := make(map[string]*graph.Doc, len(data.Docs))
		for _, doc := range data.Docs {
			docsByPath[doc.Path] = doc
		}
		for _, def := range data.Defs {
			if doc, present := docsByPath[def.Path]; present {
				def.Docs = append(def.Docs, &graph.DefDoc{Format: doc.Format, Data: doc.Data})
			}
		}

		switch imp := stor.(type) {
		case store.RepoImporter:
			if err := imp.Import(opt.CommitID, sourceUnit, data); err != nil {
				return fmt.Errorf("error running store.RepoImporter.Import: %s", err)
			}
		case store.MultiRepoImporter:
			if err := imp.Import(opt.Repo, opt.CommitID, sourceUnit, data); err != nil {
				return fmt.Errorf("error running store.MultiRepoImporter.Import: %s", err)
			}
		default:
			return fmt.Errorf("store (type %T) does not implement importing", stor)
		}

		mu.Lock()
		hasIndexableData = true
		mu.Unlock()

		return nil
	}

	par := parallel.NewRun(10)
	for _, rule_ := range mf.Rules {
		rule := rule_
		switch rule := rule.(type) {
		case *grapher.GraphUnitRule:
			if (opt.Unit != "" && rule.Unit.Name != opt.Unit) || (opt.UnitType != "" && rule.Unit.Type != opt.UnitType) {
				continue
			}
			par.Acquire()
			go func() {
				defer par.Release()
				if err := importGraphData(rule.Target(), rule.Unit); err != nil {
					par.Error(err)
				}
			}()
		case *grapher.GraphMultiUnitsRule:
			for target_, sourceUnit_ := range rule.Targets() {
				target, sourceUnit := target_, sourceUnit_
				if (opt.Unit != "" && sourceUnit.Name != opt.Unit) || (opt.UnitType != "" && sourceUnit.Type != opt.UnitType) {
					continue
				}
				par.Acquire()
				go func() {
					defer par.Release()
					if err := importGraphData(target, sourceUnit); err != nil {
						par.Error(err)
					}
				}()
			}
		}
	}
	if err := par.Wait(); err != nil {
		return err
	}

	if hasIndexableData && !opt.NoIndex {
		if GlobalOpt.Verbose {
			log.Printf("# Building indexes")
		}
		switch s := stor.(type) {
		case store.RepoIndexer:
			if err := s.Index(opt.CommitID); err != nil {
				return fmt.Errorf("Error indexing commit %s: %s", opt.CommitID, err)
			}
		case store.MultiRepoIndexer:
			if err := s.Index(opt.Repo, opt.CommitID); err != nil {
				return fmt.Errorf("error indexing %s@%s: %s", opt.Repo, opt.CommitID, err)
			}
		}
	}

	switch imp := stor.(type) {
	case store.RepoImporter:
		if err := imp.CreateVersion(opt.CommitID); err != nil {
			return fmt.Errorf("error running store.RepoImporter.CreateVersion: %s", err)
		}
	case store.MultiRepoImporter:
		if err := imp.CreateVersion(opt.Repo, opt.CommitID); err != nil {
			return fmt.Errorf("error running store.MultiRepoImporter.CreateVersion: %s", err)
		}
	}

	return nil
}

// sample imports sample data (when the --sample option is given).
func (c *StoreImportCmd) sample(s interface{}) error {
	dataString := []byte(`"abcdabcdabcdabcdabcdcdabcdabcdabcdabcdabcdabcdabcdabcdcdabcdabcdabcdabcdabcdabcdabcdabcdcdabcdabcdabcdabcdabcdabcdabcdabcdcdabcdabcdabcdabcdabcdabcdabcdabcdcdabcdabcdabcdabcdabcdabcdabcdabcdcdabcdabcdabcdabcdabcdabcdabcdabcdcdabcdabcdabcd"`)
	makeGraphData := func(numDefs, numRefs int) *graph.Output {
		defs := make([]graph.Def, numDefs)
		refs := make([]graph.Ref, numRefs)

		data := graph.Output{
			Defs: make([]*graph.Def, numDefs),
			Refs: make([]*graph.Ref, numRefs),
		}

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numDefs; i++ {
				defs[i] = graph.Def{
					DefKey:   graph.DefKey{Path: fmt.Sprintf("def-path-%d", i)},
					Name:     fmt.Sprintf("def-name-%d", i),
					Kind:     "mykind",
					DefStart: uint32((i % 53) * 37),
					DefEnd:   uint32((i%53)*37 + (i % 20)),
					File:     fmt.Sprintf("dir%d/subdir%d/subsubdir%d/file-%d.foo", i%5, i%3, i%7, i%11),
					Exported: i%5 == 0,
					Local:    i%3 == 0,
					Data:     dataString,
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numRefs; i++ {
				refs[i] = graph.Ref{
					DefPath: fmt.Sprintf("ref-path-%d", i%179),
					Def:     i%5 == 0,
					Start:   uint32((i % 51) * 39),
					End:     uint32((i%51)*37 + (int(i) % 18)),
					File:    fmt.Sprintf("dir%d/subdir%d/subsubdir%d/file-%d.foo", i%3, i%5, i%7, i%11),
				}
				if i%3 == 0 {
					refs[i].DefUnit = fmt.Sprintf("def-unit-%d", i%5)
					refs[i].DefUnitType = fmt.Sprintf("def-unit-type-%d", i%4)
					if i%7 == 0 {
						refs[i].DefRepo = fmt.Sprintf("def-repo-%d", i%7)
					}
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range defs {
				data.Defs[i] = &defs[i]
			}
			for i := range refs {
				data.Refs[i] = &refs[i]
			}
		}()

		wg.Wait()
		return &data
	}

	start := time.Now()
	log.Printf("Making sample data (%d defs, %d refs)", c.SampleDefs, c.SampleRefs)
	data := makeGraphData(c.SampleDefs, c.SampleRefs)
	unit := &unit.SourceUnit{Key: unit.Key{Type: "MyUnitType", Name: "MyUnit"}}
	files := map[string]struct{}{}
	for _, def := range data.Defs {
		files[def.File] = struct{}{}
	}
	for _, ref := range data.Refs {
		files[ref.File] = struct{}{}
	}
	for f := range files {
		unit.Files = append(unit.Files, f)
	}
	if d := time.Since(start); d > time.Millisecond*250 {
		log.Printf("Done making sample data (took %s).", d)
	}

	// Find some sample objects to query for down below.
	ref2 := data.Refs[len(data.Refs)/2+1]
	ref3 := data.Refs[len(data.Refs)/3+1]
	def2 := data.Defs[len(data.Defs)/2+1]
	def3 := data.Defs[len(data.Defs)/3+1]
	// Find a ref that has the Unit/UnitType filled in.
	var (
		refDef4      graph.RefDefKey
		foundRefDef4 bool
	)
	for _, ref := range data.Refs[len(data.Refs)/2:] {
		if ref.DefUnit != "" && ref.DefUnitType != "" {
			refDef4 = ref.RefDefKey()
			foundRefDef4 = true
			break
		}
	}

	size, err := store.Codec.NewEncoder(ioutil.Discard).Encode(data)
	if err != nil {
		return err
	}
	log.Printf("Encoded data is %s", bytesString(size))

	if c.CommitID == "" {
		c.CommitID = strings.Repeat("f", 40)
	}

	log.Printf("Importing %d defs and %d refs into the source unit %+v at commit %s", len(data.Defs), len(data.Refs), unit.ID2(), c.CommitID)
	start = time.Now()
	switch imp := s.(type) {
	case store.RepoImporter:
		c.Repo = ""
		if err := imp.Import(c.CommitID, unit, *data); err != nil {
			return err
		}
	case store.MultiRepoImporter:
		if c.Repo == "" {
			c.Repo = "example.com/my/repo"
		}
		log.Printf(" - repo %s", c.Repo)
		if err := imp.Import(c.Repo, c.CommitID, unit, *data); err != nil {
			return err
		}
	default:
		return fmt.Errorf("store (type %T) does not implement importing", s)
	}
	log.Printf("Import took %s (~%s per def/ref)", time.Since(start), time.Duration(int64(time.Since(start))/int64(len(data.Defs)+len(data.Refs))))

	start = time.Now()
	if !c.NoIndex {
		switch s := s.(type) {
		case store.RepoIndexer:
			if err := s.Index(c.CommitID); err != nil {
				return err
			}
		case store.MultiRepoIndexer:
			if err := s.Index(c.Repo, c.CommitID); err != nil {
				return err
			}
		}
	}
	log.Printf("Index took %s (~%s per def/ref)", time.Since(start), time.Duration(int64(time.Since(start))/int64(len(data.Defs)+len(data.Refs))))

	switch imp := s.(type) {
	case store.RepoImporter:
		if err := imp.CreateVersion(c.CommitID); err != nil {
			return err
		}
	case store.MultiRepoImporter:
		if err := imp.CreateVersion(c.Repo, c.CommitID); err != nil {
			return err
		}
	}

	if c.SampleImportOnly {
		return nil
	}

	log.Println()
	log.Printf("Running some commands to list sample data...")

	runCmd := func(args ...string) error {
		start := time.Now()
		var b bytes.Buffer
		c := exec.Command(args[0], args[1:]...)
		c.Stdout = &b
		c.Stderr = &b
		log.Println()
		log.Println(strings.Join(c.Args, " "))
		if err := c.Run(); err != nil {
			return fmt.Errorf("command %v failed\n\noutput was:\n%s", c.Args, b.String())
		}
		if GlobalOpt.Verbose {
			log.Println(b.String())
		} else {
			log.Printf("-> printed %d lines of output (run with `%s -v` to view)", bytes.Count(b.Bytes(), []byte{'\n'}), os.Args[0])
		}
		log.Printf("-> took %s", time.Since(start))
		return nil
	}

	prog := os.Args[0]
	storeCmdArgs := func(subcmd, repo string, args ...string) []string {
		x := []string{prog, "store", subcmd}
		if repo != "" {
			x = append(x, "--repo", repo)
		}
		x = append(x, "--commit", c.CommitID)
		x = append(x, args...)
		return x
	}

	if err := runCmd(storeCmdArgs("versions", c.Repo)...); err != nil {
		return err
	}
	if err := runCmd(storeCmdArgs("units", c.Repo)...); err != nil {
		return err
	}
	if err := runCmd(storeCmdArgs("units", c.Repo, "--file", def2.File)...); err != nil {
		return err
	}
	if err := runCmd(storeCmdArgs("units", c.Repo, "--file", ref2.File)...); err != nil {
		return err
	}
	if err := runCmd(storeCmdArgs("defs", c.Repo, "--file", def3.File)...); err != nil {
		return err
	}
	if err := runCmd(storeCmdArgs("defs", c.Repo, "--query", def3.Name[:len(def3.Name)-2])...); err != nil {
		return err
	}
	if err := runCmd(storeCmdArgs("refs", c.Repo, "--file", ref3.File)...); err != nil {
		return err
	}
	if foundRefDef4 {
		if err := runCmd(storeCmdArgs("refs", c.Repo, "--def-unit-type", refDef4.DefUnitType, "--def-unit", refDef4.DefUnit, "--def-path", refDef4.DefPath)...); err != nil {
			return err
		}
	}

	return nil
}

// countingWriter wraps an io.Writer, counting the number of bytes
// written.

type storeIndexCriteria struct {
	Repo     string `long:"repo" description:"only indexes for this repo"`
	CommitID string `long:"commit" description:"only indexes for this commit ID"`
	UnitType string `long:"unit-type" description:"only indexes for this source unit type"`
	Unit     string `long:"unit" description:"only indexes for this source unit name"`
	Name     string `long:"name" description:"only indexes whose name contains this substring"`
	Type     string `long:"type" description:"only indexes whose Go type contains this substring"`

	NoSourceUnit bool `long:"no-unit" description:"skip indexes for specific source units"`

	Stale    bool `long:"stale" description:"only stale indexes"`
	NotStale bool `long:"not-stale" description:"only non-stale indexes"`

	ReposLimit  int `long:"repos-limit" description:"only indexes from this many repos (0=all)"`
	ReposOffset int `long:"repos-offset" description:"only indexes after skipping this many repos"`
}

func (c storeIndexCriteria) IndexCriteria() store.IndexCriteria {
	crit := store.IndexCriteria{
		Repo:     c.Repo,
		CommitID: c.CommitID,
		Name:     c.Name,
		Type:     c.Type,
	}
	if c.Stale && c.NotStale {
		log.Fatal("must specify exactly one of --stale and --not-stale")
	}
	if c.Stale {
		t := true
		crit.Stale = &t
	}
	if c.NotStale {
		f := false
		crit.Stale = &f
	}
	if c.NoSourceUnit {
		crit.Unit = store.NoSourceUnit
	}
	if c.UnitType != "" || c.Unit != "" {
		crit.Unit = &unit.ID2{Type: c.UnitType, Name: c.Unit}
		if crit.Unit.Type == "" || crit.Unit.Name == "" {
			log.Fatal("must specify either both or neither of --unit-type and --unit (to filter by source unit)")
		}
	}
	crit.ReposLimit = c.ReposLimit
	crit.ReposOffset = c.ReposOffset
	return crit
}

// doStoreIndexesCmd is invoked by both StoreIndexesCmd.Execute and
// StoreBuildIndexesCmd.Execute.
func doStoreIndexesCmd(crit store.IndexCriteria, opt storeIndexOptions, f func(interface{}, store.IndexCriteria, chan<- store.IndexStatus) ([]store.IndexStatus, error)) error {
	if opt.Parallel != 1 {
		log.Printf("NOTE: Index parallelism is %d. Output will printed as it is available, not necessarily ordered and grouped by repo, source unit, etc.", opt.Parallel)
	}
	store.MaxIndexParallel = opt.Parallel

	printIndex := func(x store.IndexStatus) error {
		if !opt.Print {
			return nil
		}
		return x.Fprint(os.Stderr)
	}

	s, err := OpenStore()
	if err != nil {
		return err
	}

	hasError := false
	done := make(chan struct{})
	indexChan := make(chan store.IndexStatus)
	switch opt.Output {
	case "json":
		go func() {
			for x := range indexChan {
				PrintJSON(x, "")
				if err := printIndex(x); err != nil {
					log.Fatal(err)
				}
			}
			done <- struct{}{}
		}()
	case "text":
		_, isMultiRepo := s.(store.MultiRepoStore)
		var repoTab string
		if isMultiRepo {
			repoTab = "\t"
		}

		go func() {
			var lastRepo, lastCommitID string
			var lastUnit *unit.ID2
			for x := range indexChan {
				if isMultiRepo {
					if x.Repo != lastRepo {
						if lastRepo != "" {
							colorable.Println()
						}
						colorable.Println(x.Repo)
					}
				}
				if x.CommitID != lastCommitID {
					colorable.Print(repoTab, x.CommitID, "\n")
				}
				if x.Unit != lastUnit && x.Unit != nil {
					if x.Repo == lastRepo && x.CommitID == lastCommitID {
						colorable.Println()
					}
					colorable.Print(repoTab, "\t", x.Unit.Name, " ", x.Unit.Type, "\n")
				}

				if x.Unit != nil {
					colorable.Print("\t")
				}

				colorable.Print(repoTab, "\t")
				colorable.Printf("%s (%s) ", x.Name, x.Type)
				if x.Stale {
					colorable.Print("STALE ")
				}
				if x.Size != 0 {
					colorable.Print(bytesString(uint64(x.Size)), " ")
				}
				if x.Error != "" {
					colorable.Printf("(ERROR: %s) ", x.Error)
					hasError = true
				}
				if x.BuildError != "" {
					colorable.Printf("(BUILD ERROR: %s) ", x.BuildError)
					hasError = true
				}
				if x.BuildDuration != 0 {
					colorable.Printf("- build took %s ", x.BuildDuration)
				}
				colorable.Println()

				if err := printIndex(x); err != nil {
					log.Fatal(err)
				}

				lastRepo = x.Repo
				lastCommitID = x.CommitID
				lastUnit = x.Unit
			}
			done <- struct{}{}
		}()
	default:
		return fmt.Errorf("unexpected --output value: %q", opt.Output)
	}

	_, err = f(s, crit, indexChan)
	defer func() {
		close(indexChan)
		<-done
	}()
	if err != nil {
		return err
	}
	if hasError {
		return errors.New("\nindex listing or index building errors occurred (see above)")
	}
	return nil
}

type storeIndexOptions struct {
	Output   string `short:"o" long:"output" description:"output format (text|json)" default:"text"`
	Parallel int    `short:"p" long:"parallel" description:"parallelism (may produce out-of-order output)" default:"1"`

	Print bool `long:"print" description:"(debug) print representation of index"`
}

type StoreIndexesCmd struct {
	storeIndexCriteria
	storeIndexOptions
}

var storeIndexesCmd StoreIndexesCmd

func (c *StoreIndexesCmd) Execute(args []string) error {
	return doStoreIndexesCmd(c.IndexCriteria(), c.storeIndexOptions, store.Indexes)
}

type StoreIndexCmd struct {
	storeIndexCriteria
	storeIndexOptions
}

var storeIndexCmd StoreIndexCmd

func (c *StoreIndexCmd) Execute(args []string) error {
	return doStoreIndexesCmd(c.IndexCriteria(), c.storeIndexOptions, store.BuildIndexes)
}

type StoreReposCmd struct {
	IDContains string `short:"i" long:"id-contains" description:"filter to repos whose ID contains this substring"`
}

func (c *StoreReposCmd) filters() []store.RepoFilter {
	var fs []store.RepoFilter
	if c.IDContains != "" {
		fs = append(fs, store.RepoFilterFunc(func(repo string) bool { return strings.Contains(repo, c.IDContains) }))
	}
	return fs
}

var storeReposCmd StoreReposCmd

func (c *StoreReposCmd) Execute(args []string) error {
	s, err := OpenStore()
	if err != nil {
		return err
	}

	mrs, ok := s.(store.MultiRepoStore)
	if !ok {
		return fmt.Errorf("store (type %T) does not implement listing repositories", s)
	}

	repos, err := mrs.Repos(c.filters()...)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		colorable.Println(repo)
	}
	return nil
}

type StoreVersionsCmd struct {
	Repo           string `long:"repo"`
	CommitIDPrefix string `long:"commit" description:"commit ID prefix"`

	RepoCommitIDs string `long:"repo-commits" description:"comma-separated list of repo@commitID specifiers"`
}

func (c *StoreVersionsCmd) filters() []store.VersionFilter {
	var fs []store.VersionFilter
	if c.Repo != "" {
		fs = append(fs, store.ByRepos(c.Repo))
	}
	if c.CommitIDPrefix != "" {
		fs = append(fs, store.VersionFilterFunc(func(version *store.Version) bool {
			return strings.HasPrefix(version.CommitID, c.CommitIDPrefix)
		}))
	}
	if c.RepoCommitIDs != "" {
		fs = append(fs, makeRepoCommitIDsFilter(c.RepoCommitIDs))
	}
	return fs
}

var storeVersionsCmd StoreVersionsCmd

func (c *StoreVersionsCmd) Execute(args []string) error {
	s, err := OpenStore()
	if err != nil {
		return err
	}

	rs, ok := s.(store.RepoStore)
	if !ok {
		return fmt.Errorf("store (type %T) does not implement listing versions", s)
	}

	versions, err := rs.Versions(c.filters()...)
	if err != nil {
		return err
	}
	for _, version := range versions {
		if version.Repo != "" {
			colorable.Print(version.Repo, "\t")
		}
		colorable.Println(version.CommitID)
	}
	return nil
}

type StoreUnitsCmd struct {
	Type     string `long:"type" `
	Name     string `long:"name"`
	CommitID string `long:"commit"`
	Repo     string `long:"repo"`

	RepoCommitIDs string `long:"repo-commits" description:"comma-separated list of repo@commitID specifiers"`

	File string `long:"file" description:"filter by units whose Files list contains this file"`
}

func (c *StoreUnitsCmd) filters() []store.UnitFilter {
	var fs []store.UnitFilter
	if c.Type != "" && c.Name != "" {
		fs = append(fs, store.ByUnits(unit.ID2{Type: c.Type, Name: c.Name}))
	}
	if (c.Type != "" && c.Name == "") || (c.Type == "" && c.Name != "") {
		log.Fatal("must specify either both or neither of --type and --name (to filter by source unit)")
	}
	if c.CommitID != "" {
		fs = append(fs, store.ByCommitIDs(c.CommitID))
	}
	if c.Repo != "" {
		fs = append(fs, store.ByRepos(c.Repo))
	}
	if c.RepoCommitIDs != "" {
		fs = append(fs, makeRepoCommitIDsFilter(c.RepoCommitIDs))
	}
	if c.File != "" {
		fs = append(fs, store.ByFiles(false, path.Clean(c.File)))
	}
	return fs
}

var storeUnitsCmd StoreUnitsCmd

func (c *StoreUnitsCmd) Execute(args []string) error {
	s, err := OpenStore()
	if err != nil {
		return err
	}

	ts, ok := s.(store.TreeStore)
	if !ok {
		return fmt.Errorf("store (type %T) does not implement listing source units", s)
	}

	units, err := ts.Units(c.filters()...)
	if err != nil {
		return err
	}
	PrintJSON(units, "  ")
	return nil
}

type StoreDefsCmd struct {
	Repo     string `long:"repo"`
	Path     string `long:"path"`
	UnitType string `long:"unit-type" `
	Unit     string `long:"unit"`
	File     string `long:"file"`
	CommitID string `long:"commit"`

	RepoCommitIDs string `long:"repo-commits" description:"comma-separated list of repo@commitID specifiers"`

	Query string `long:"query"`

	Limit  int `short:"n" long:"limit" description:"max results to return (0 for all)"`
	Offset int `long:"offset" description:"results offset (0 to start with first results)"`

	// If Filter is non-nil, it is applied along with the above
	// filters.
	Filter store.DefFilter
}

func (c *StoreDefsCmd) filters() []store.DefFilter {
	var fs []store.DefFilter
	if c.UnitType != "" && c.Unit != "" {
		fs = append(fs, store.ByUnits(unit.ID2{Type: c.UnitType, Name: c.Unit}))
	}
	if (c.UnitType != "" && c.Unit == "") || (c.UnitType == "" && c.Unit != "") {
		log.Fatal("must specify either both or neither of --unit-type and --unit (to filter by source unit)")
	}
	if c.CommitID != "" {
		fs = append(fs, store.ByCommitIDs(c.CommitID))
	}
	if c.Repo != "" {
		fs = append(fs, store.ByRepos(c.Repo))
	}
	if c.RepoCommitIDs != "" {
		fs = append(fs, makeRepoCommitIDsFilter(c.RepoCommitIDs))
	}
	if c.Path != "" {
		fs = append(fs, store.ByDefPath(c.Path))
	}
	if c.File != "" {
		fs = append(fs, store.ByFiles(false, path.Clean(c.File)))
	}
	if c.Query != "" {
		fs = append(fs, store.ByDefQuery(c.Query))
	}
	if c.Filter != nil {
		fs = append(fs, c.Filter)
	}
	if c.Limit != 0 || c.Offset != 0 {
		fs = append(fs, store.Limit(c.Limit, c.Offset))
	}
	return fs
}

var storeDefsCmd StoreDefsCmd

func (c *StoreDefsCmd) Execute(args []string) error {
	defs, err := c.Get()
	if err != nil {
		return err
	}
	PrintJSON(defs, "  ")
	return nil
}

func (c *StoreDefsCmd) Get() ([]*graph.Def, error) {
	s, err := OpenStore()
	if err != nil {
		return nil, err
	}

	us, ok := s.(store.UnitStore)
	if !ok {
		return nil, fmt.Errorf("store (type %T) does not implement listing defs", s)
	}

	defs, err := us.Defs(c.filters()...)
	if err != nil {
		return nil, err
	}
	return defs, nil
}

type StoreRefsCmd struct {
	Repo     string `long:"repo"`
	UnitType string `long:"unit-type" `
	Unit     string `long:"unit"`
	File     string `long:"file"`
	CommitID string `long:"commit"`

	RepoCommitIDs string `long:"repo-commits" description:"comma-separated list of repo@commitID specifiers"`

	Start uint32 `long:"start"`
	End   uint32 `long:"end"`

	DefRepo     string `long:"def-repo"`
	DefUnitType string `long:"def-unit-type" `
	DefUnit     string `long:"def-unit"`
	DefPath     string `long:"def-path"`

	Broken   bool `long:"broken" description:"only show refs that point to nonexistent defs"`
	Coverage bool `long:"coverage" description:"print a coverage summary (resolved refs, broken refs, total refs)"`

	Format string `long:"format" description:"output format ('json' or 'none')" default:"json"`

	Limit  int `short:"n" long:"limit" description:"max results to return (0 for all)"`
	Offset int `long:"offset" description:"results offset (0 to start with first results)"`
}

func (c *StoreRefsCmd) filters() []store.RefFilter {
	var fs []store.RefFilter
	if c.UnitType != "" && c.Unit != "" {
		fs = append(fs, store.ByUnits(unit.ID2{Type: c.UnitType, Name: c.Unit}))
	}
	if (c.UnitType != "" && c.Unit == "") || (c.UnitType == "" && c.Unit != "") {
		log.Fatal("must specify either both or neither of --unit-type and --unit (to filter by source unit)")
	}
	if c.CommitID != "" {
		fs = append(fs, store.ByCommitIDs(c.CommitID))
	}
	if c.Repo != "" {
		fs = append(fs, store.ByRepos(c.Repo))
	}
	if c.RepoCommitIDs != "" {
		fs = append(fs, makeRepoCommitIDsFilter(c.RepoCommitIDs))
	}
	if c.File != "" {
		fs = append(fs, store.ByFiles(false, path.Clean(c.File)))
	}
	if c.Start != 0 {
		fs = append(fs, store.RefFilterFunc(func(ref *graph.Ref) bool {
			return ref.Start >= c.Start
		}))
	}
	if c.End != 0 {
		fs = append(fs, store.RefFilterFunc(func(ref *graph.Ref) bool {
			return ref.End <= c.End
		}))
	}
	if c.DefPath != "" {
		fs = append(fs, store.ByRefDef(graph.RefDefKey{
			DefRepo:     c.DefRepo,
			DefUnitType: c.DefUnitType,
			DefUnit:     c.DefUnit,
			DefPath:     c.DefPath,
		}))
	} else {
		// Slower filters since they don't use an index.
		if c.DefRepo != "" {
			fs = append(fs, store.AbsRefFilterFunc(store.RefFilterFunc(func(ref *graph.Ref) bool {
				return ref.DefRepo == c.DefRepo
			})))
		}
		if c.DefUnitType != "" {
			fs = append(fs, store.AbsRefFilterFunc(store.RefFilterFunc(func(ref *graph.Ref) bool {
				return ref.DefUnitType == c.DefUnitType
			})))
		}
		if c.DefUnit != "" {
			fs = append(fs, store.AbsRefFilterFunc(store.RefFilterFunc(func(ref *graph.Ref) bool {
				return ref.DefUnit == c.DefUnit
			})))
		}
	}
	if c.Limit != 0 || c.Offset != 0 {
		fs = append(fs, store.Limit(c.Limit, c.Offset))
	}
	return fs
}

var storeRefsCmd StoreRefsCmd

func (c *StoreRefsCmd) Execute(args []string) error {
	refs, err := c.Get()
	if err != nil {
		return err
	}
	switch c.Format {
	case "json":
		PrintJSON(refs, "  ")
	}
	return nil
}

func (c *StoreRefsCmd) Get() ([]*graph.Ref, error) {
	s, err := OpenStore()
	if err != nil {
		return nil, err
	}

	us, ok := s.(store.UnitStore)
	if !ok {
		return nil, fmt.Errorf("store (type %T) does not implement listing refs", s)
	}

	refs, err := us.Refs(c.filters()...)
	if err != nil {
		return nil, err
	}

	allRefs := refs
	var brokenRefs []*graph.Ref
	if c.Broken || c.Coverage {
		brokenRefs, err = brokenRefsOnly(refs, s)
		if err != nil {
			return nil, err
		}
		if c.Broken {
			refs = brokenRefs
		}
	}
	if c.Coverage {
		log.Printf("# Coverage summary:")
		log.Printf("#  - %d total refs", len(allRefs))
		resolvedRefs := len(allRefs) - len(brokenRefs)
		log.Printf("#  - %d resolved refs (%.1f)", resolvedRefs, percent(resolvedRefs, len(allRefs)))
		log.Printf("#  - %d broken refs (%.1f)", len(brokenRefs), percent(len(brokenRefs), len(allRefs)))
	}

	return refs, nil
}

func brokenRefsOnly(refs []*graph.Ref, s interface{}) ([]*graph.Ref, error) {
	uniqRefDefs := map[graph.DefKey][]*graph.Ref{}
	loggedDefRepos := map[string]struct{}{}
	for _, ref := range refs {
		if ref.Repo != ref.DefRepo {
			if _, logged := loggedDefRepos[ref.DefRepo]; !logged {
				// TODO(sqs): need to skip these because we don't know the
				// "DefCommitID" in the def's repo, and ByDefKey requires
				// the key to have a CommitID.
				log.Printf("WARNING: Can't check resolution of cross-repo ref (ref.Repo=%q != ref.DefRepo=%q) - cross-repo ref checking is not yet implemented. (This log message will not be repeated.)", ref.Repo, ref.DefRepo)
				loggedDefRepos[ref.DefRepo] = struct{}{}
			}
			continue
		}
		def := ref.DefKey()
		def.CommitID = ref.CommitID
		uniqRefDefs[def] = append(uniqRefDefs[def], ref)
	}

	var (
		brokenRefs  []*graph.Ref
		brokenRefMu sync.Mutex
		par         = parallel.NewRun(runtime.GOMAXPROCS(0))
	)
	for def_, refs_ := range uniqRefDefs {
		def, refs := def_, refs_
		par.Acquire()
		go func() {
			defer par.Release()
			defs, err := s.(store.RepoStore).Defs(store.ByDefKey(def))
			if err != nil {
				par.Error(err)
				return
			}
			if len(defs) == 0 {
				brokenRefMu.Lock()
				brokenRefs = append(brokenRefs, refs...)
				brokenRefMu.Unlock()
			}
		}()
	}
	err := par.Wait()
	sort.Sort(graph.Refs(brokenRefs))
	return brokenRefs, err
}

func makeRepoCommitIDsFilter(repoCommitIDs string) interface {
	store.ByRepoCommitIDsFilter
	store.VersionFilter
	store.DefFilter
	store.UnitFilter
	store.RefFilter
} {
	if repoCommitIDs == "" {
		panic("empty repoCommitIDs")
	}
	rcs := strings.Split(repoCommitIDs, ",")
	vs := make([]store.Version, len(rcs))
	for i, rc := range rcs {
		rc = strings.TrimSpace(rc)
		repo, commitID := parseRepoAndCommitID(rc)
		if len(commitID) != 40 {
			log.Printf("WARNING: --repo-commits entry #%d (%q) has no commit ID or a non-absolute commit ID. Nothing will match it.", i, rc)
		}
		vs[i] = store.Version{Repo: repo, CommitID: commitID}
	}
	return store.ByRepoCommitIDs(vs...)
}
