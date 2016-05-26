package gendata

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rogpeppe/rog-go/parallel"

	"sync"

	"sourcegraph.com/sourcegraph/srclib/dep"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// URefsRepoCmd generates a synthetic repository with the specified
// unit and file hierarchy. In each file, it emits NDefs defs. For
// each file of defs, for each source unit, it emits a file of
// references to the defs. This results in a repository with a high
// number of cross-unit references.
type URefsRepoCmd struct {
	GenDataOpt

	NUnits []int `short:"u" long:"units" description:"number of units to generate; uses same input structure as --files" required:"yes"`
	NFiles []int `short:"f" long:"files" description:"number of files at each level" required:"yes"`
	NDefs  int   `long:"ndefs" description:"number of defs to generate per file" required:"yes"`
}

func (c *URefsRepoCmd) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := removeGlob(".srclib-*"); err != nil {
		return err
	}

	units := make([]*unit.SourceUnit, 0)
	unitNames := hierarchicalNames("u", "unit", "", c.NUnits)
	for _, unitName := range unitNames {
		ut := &unit.SourceUnit{
			Key: unit.Key{
				Name:     unitName,
				Type:     "GoPackage",
				Repo:     c.Repo,
				CommitID: c.CommitID,
			},
			Info: unit.Info{
				Files: []string{},
				Dir:   unitName,
			},
		}
		units = append(units, ut)
	}

	if c.GenSource {
		if err := resetSource(); err != nil {
			return err
		}

		// generate source files
		par := parallel.NewRun(runtime.GOMAXPROCS(0))
		for _, ut := range units {
			ut := ut
			par.Do(func() error {
				_, _, _, err := c.genUnit(ut, units)
				return err
			})
		}
		if err := par.Wait(); err != nil {
			return err
		}

		// get commit ID
		commitID, err := getGitCommitID()
		if err != nil {
			return err
		}

		// update command to generate graph data
		c.CommitID = commitID
		c.GenSource = false
	}

	// generate graph data
	gr := make(map[string]*graph.Output)
	for _, ut := range units {
		gr[ut.Name] = &graph.Output{}
	}
	var grmu sync.Mutex
	par := parallel.NewRun(runtime.GOMAXPROCS(0))
	for _, ut := range units {
		ut := ut
		ut.CommitID = c.CommitID
		par.Do(func() error {
			defs, refs, reffiles, err := c.genUnit(ut, units)
			if err != nil {
				return err
			}

			grmu.Lock()
			defer grmu.Unlock()

			gr[ut.Name].Defs = append(gr[ut.Name].Defs, defs...)
			for utName, utRefs := range refs {
				gr[utName].Refs = append(gr[utName].Refs, utRefs...)
			}
			for _, ut2 := range units {
				ut2.Files = append(ut2.Files, reffiles[ut2.Name]...)
			}
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return err
	}
	for _, ut := range units {
		utgraph := gr[ut.Name]
		utgraph.Docs = make([]*graph.Doc, 0)
		if err := writeSrclibCache(ut, utgraph, make([]*dep.Resolution, 0)); err != nil {
			return err
		}
	}

	return nil
}

// genUnit must NOT mutate any unit except ut, because it is
// called on multiple units in parallel.
func (c *URefsRepoCmd) genUnit(ut *unit.SourceUnit, allUt []*unit.SourceUnit) ([]*graph.Def, map[string][]*graph.Ref, map[string][]string, error) {
	var defs []*graph.Def
	refs := make(map[string][]*graph.Ref)
	reffiles := make(map[string][]string)

	for _, filename := range hierarchicalNames("dir", "file", "", c.NFiles) {
		defs_, defRefs, err := c.genDefsFile(ut, filename)
		if err != nil {
			return nil, nil, nil, err
		}
		refs[ut.Name] = append(refs[ut.Name], defRefs...)
		defs = append(defs, defs_...)

		refs_, reffiles_, err := c.genRefsFiles(ut, defs_, filename, allUt)
		if err != nil {
			return nil, nil, nil, err
		}
		for utname, utrefs := range refs_ {
			refs[utname] = append(refs[utname], utrefs...)
		}
		for utname, utreffiles := range reffiles_ {
			reffiles[utname] = append(reffiles[utname], utreffiles...)
		}
	}
	return defs, refs, reffiles, nil
}

func (c *URefsRepoCmd) genDefsFile(ut *unit.SourceUnit, filename string) (defs []*graph.Def, defRefs []*graph.Ref, err error) {
	filename = filepath.Join(ut.Name, filename)
	ut.Files = append(ut.Files, filename)

	offset := 0
	defName := "foo"

	var sourceFile *os.File
	if c.GenSource {
		err := os.MkdirAll(filepath.Dir(filename), 0700)
		if err != nil {
			return nil, nil, err
		}
		file, err := os.Create(filename)
		if err != nil {
			return nil, nil, err
		}
		sourceFile = file
	}

	for i := 0; i < c.NDefs; i++ {
		def := &graph.Def{
			DefKey: graph.DefKey{
				Repo:     ut.Repo,
				CommitID: ut.CommitID,
				UnitType: ut.Type,
				Unit:     ut.Name,
				Path:     filepath.ToSlash(filepath.Join(filename, fmt.Sprintf("method_%d", i))),
			},
			Name:     defName,
			Exported: true,
			File:     filename,
			DefStart: uint32(offset),
			DefEnd:   uint32(offset + len(defName) - 1),
		}
		if sourceFile != nil {
			_, err := sourceFile.WriteString(def.Name + "\n")
			if err != nil {
				return nil, nil, err
			}
		}
		offset += len(defName) + 1
		defs = append(defs, def)

		defRef := &graph.Ref{
			DefRepo:     def.Repo,
			DefUnitType: def.UnitType,
			DefUnit:     def.Unit,
			DefPath:     def.Path,
			Repo:        def.Repo,
			CommitID:    def.CommitID,
			UnitType:    def.UnitType,
			Unit:        def.Unit,
			Def:         true,
			File:        def.File,
			Start:       def.DefStart,
			End:         def.DefEnd,
		}
		defRefs = append(defRefs, defRef)
	}

	// Close source file
	if sourceFile != nil {
		sourceFile.WriteString("\n")
		sourceFile.Close()
	}

	return defs, defRefs, nil
}

func (c *URefsRepoCmd) genRefsFiles(ut *unit.SourceUnit, defs []*graph.Def, defsFilename string, allUt []*unit.SourceUnit) (refs map[string][]*graph.Ref, reffiles map[string][]string, err error) {
	refs = make(map[string][]*graph.Ref)
	reffiles = make(map[string][]string)

	for _, refUt := range allUt {
		offset := 0

		refsFilename := filepath.Join(refUt.Name, ut.Name, defsFilename+"_refs")
		reffiles[refUt.Name] = append(reffiles[refUt.Name], refsFilename)

		var sourcefile *os.File
		if c.GenSource {
			err := os.MkdirAll(filepath.Dir(refsFilename), 0700)
			if err != nil {
				return nil, nil, err
			}
			file, err := os.Create(refsFilename)
			if err != nil {
				return nil, nil, err
			}
			sourcefile = file
		}

		for _, def := range defs {
			ref := &graph.Ref{
				DefRepo:     def.Repo,
				DefUnitType: def.UnitType,
				DefUnit:     def.Unit,
				DefPath:     def.Path,
				Repo:        def.Repo,
				CommitID:    def.CommitID,
				UnitType:    refUt.Type,
				Unit:        refUt.Name,
				Def:         false,
				File:        refsFilename,
				Start:       uint32(offset),
				End:         uint32(offset + len(def.Name) - 1),
			}

			if sourcefile != nil {
				_, err := sourcefile.WriteString(def.Name + "\n")
				if err != nil {
					return nil, nil, err
				}
			}
			offset += len(def.Name) + 1
			refs[refUt.Name] = append(refs[refUt.Name], ref)
		}

		// Close source file
		if sourcefile != nil {
			sourcefile.WriteString("\n")
			sourcefile.Close()
		}
	}

	return refs, reffiles, nil
}
