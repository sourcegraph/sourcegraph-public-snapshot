package python

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type GraphContext struct {
	Unit *unit.SourceUnit
	Reqs []*requirement
}

func NewGraphContext(unit *unit.SourceUnit) *GraphContext {
	var g GraphContext
	g.Unit = unit
	for _, dep := range unit.Dependencies {
		if req, err := asRequirement(dep); err == nil {
			g.Reqs = append(g.Reqs, req)
		}
	}
	return &g
}

// Graphs the Python source unit. If run outside of a Docker container, this assumes that the source unit has already
// been installed (via pip or `python setup.py install`).
func (c *GraphContext) Graph() (*graph.Output, error) {
	programMode := os.Getenv("IN_DOCKER_CONTAINER") == ""
	pipBin := "pip"
	pythonBin := "python"

	if programMode {
		tc, err := toolchain.Lookup("sourcegraph.com/sourcegraph/srclib-python")
		if err != nil {
			return nil, err
		}

		tempDir, err := ioutil.TempDir("", "srclib-python-graph")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tempDir)
		envName := fmt.Sprintf("%s-%s-env", getHash(c.Unit.Dir), url.QueryEscape(c.Unit.Name))
		envDir := filepath.Join(tempDir, envName)

		// Use binaries from our virutal env.
		pipBin = filepath.Join(envDir, "bin", "pip")
		pythonBin = filepath.Join(envDir, "bin", "python")

		if _, err := os.Stat(filepath.Join(envDir)); os.IsNotExist(err) {
			// We don't have virtual env for this SourceUnit, create one.
			tcVENVBinPath := filepath.Join(tc.Dir, ".env", "bin")
			cmd := exec.Command(filepath.Join(tcVENVBinPath, "virtualenv"), envDir)
			if err := runCmdStderr(cmd); err != nil {
				return nil, err
			}
			// Install our dependencies.
			// Todo(MaikuMori): Use symlinks from toolchains virtualenv to project virtual env.
			// NOTE: If SourceUnit requirements overwrite our requirements, things will fail.
			// 			 We could install them last, but then we would have to do this before each
			//			 graphing which noticably increases graphing time (since our deps are always
			//       downloaded by pip due to dependency on git commit not actual package version).
			requirementFile := filepath.Join(tc.Dir, "requirements.txt")
			if err := runCmdStderr(exec.Command(pipBin, "install", "-r", requirementFile)); err != nil {
				return nil, err
			}
			if err := runCmdStderr(exec.Command(pipBin, "install", "-e", tc.Dir)); err != nil {
				return nil, err
			}
		}
	}

	// NOTE: this may cause an error when graphing any source unit that depends
	// on jedi (or any other dependency of the graph code)
	requirementFiles, err := filepath.Glob(filepath.Join(c.Unit.Dir, "*requirements*.txt"))
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(c.Unit.Dir, "setup.py")); !os.IsNotExist(err) {
		runCmdLogError(exec.Command(pipBin, "install", "-I", c.Unit.Dir))
	}
	installPipRequirements(pipBin, requirementFiles)

	cmd := exec.Command(pythonBin, "-m", "grapher.graph", "--verbose", "--dir", c.Unit.Dir, "--files")
	cmd.Args = append(cmd.Args, c.Unit.Files...)
	cmd.Stderr = os.Stderr
	log.Printf("Running %v", cmd.Args)
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var raw RawOutput
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	out := c.transform(&raw, c.Unit)
	return out, nil
}

func (c *GraphContext) transform(raw *RawOutput, unit *unit.SourceUnit) *graph.Output {
	var out graph.Output

	for _, def := range raw.Defs {
		out.Defs = append(out.Defs, c.transformDef(def))
		if doc := c.transformDefDoc(def); doc != nil {
			out.Docs = append(out.Docs, doc)
		}
	}
	for _, ref := range raw.Refs {
		if outRef, err := c.transformRef(ref); err == nil {
			out.Refs = append(out.Refs, outRef)
		} else {
			log.Printf("Could not transform ref %v: %s", ref, err)
		}
	}

	return &out
}

var jediKindToDefKind = map[string]string{
	"statement":        "var",
	"statementelement": "var",
	"param":            "var",
	"module":           "module",
	"submodule":        "module",
	"class":            "type",
	"function":         "func",
	"lambda":           "func",
	"import":           "var",
}

func (c *GraphContext) transformDef(rawDef *RawDef) *graph.Def {
	return &graph.Def{
		DefKey: graph.DefKey{
			Repo:     c.Unit.Repo,
			Unit:     c.Unit.Name,
			UnitType: c.Unit.Type,
			Path:     string(rawDef.Path),
		},
		TreePath: string(rawDef.Path), // TODO: make this consistent w/ old way
		Kind:     jediKindToDefKind[rawDef.Kind],
		Name:     rawDef.Name,
		File:     rawDef.File,
		DefStart: rawDef.DefStart,
		DefEnd:   rawDef.DefEnd,
		Exported: rawDef.Exported,
		Data:     nil, // TODO
	}
}

func (c *GraphContext) transformRef(rawRef *RawRef) (*graph.Ref, error) {
	defUnit, err := c.inferSourceUnit(rawRef, c.Reqs)
	if err != nil {
		return nil, err
	}

	defPath := string(rawRef.DefPath)
	if defPath == "" {
		defPath = "."
	}

	return &graph.Ref{
		DefRepo:     defUnit.Repo,
		DefUnitType: defUnit.Type,
		DefUnit:     defUnit.Name,
		DefPath:     defPath,

		Repo:     c.Unit.Repo,
		Unit:     c.Unit.Name,
		UnitType: c.Unit.Type,

		File:  rawRef.File,
		Start: rawRef.Start,
		End:   rawRef.End,
		Def:   rawRef.Def,
	}, nil
}

func (c *GraphContext) transformDefDoc(rawDef *RawDef) *graph.Doc {
	if rawDef.Docstring != "" {
		return &graph.Doc{
			DefKey: graph.DefKey{
				Repo:     c.Unit.Repo,
				Unit:     c.Unit.Name,
				UnitType: c.Unit.Type,
				Path:     string(rawDef.Path),
			},
			Data: rawDef.Docstring,
		}
	} else {
		return nil
	}
}

func (c *GraphContext) inferSourceUnit(rawRef *RawRef, reqs []*requirement) (*unit.SourceUnit, error) {
	if rawRef.ToBuiltin {
		return stdLibPkg.SourceUnit(), nil
	}
	return c.inferSourceUnitFromFile(rawRef.DefFile, reqs)
}

// Note: file is expected to be an absolute path
func (c *GraphContext) inferSourceUnitFromFile(file string, reqs []*requirement) (*unit.SourceUnit, error) {
	// Case: in current source unit (u)
	pwd, _ := os.Getwd()
	if isSubPath(pwd, file) {
		return c.Unit, nil
	}

	// Case: in dependent source unit(depUnits)
	fileCmps := strings.Split(file, string(filepath.Separator))
	pkgsDirIdx := -1
	for i, cmp := range fileCmps {
		if cmp == "site-packages" || cmp == "dist-packages" {
			pkgsDirIdx = i
			break
		}
	}
	if pkgsDirIdx != -1 {
		fileSubCmps := fileCmps[pkgsDirIdx+1:]
		fileSubPath := filepath.Join(fileSubCmps...)

		var foundReq *requirement
	FindReq:
		for _, req := range reqs {
			for _, pkg := range req.Packages {
				if isSubPath(moduleToFilepath(pkg, true), fileSubPath) {
					foundReq = req
					break FindReq
				}
			}
			for _, mod := range req.Modules {
				if moduleToFilepath(mod, false) == fileSubPath {
					foundReq = req
					break FindReq
				}
			}
		}

		if foundReq == nil {
			var formattedCanditates []string
			end := ""
			candiates := reqs

			if len(reqs) > 7 {
				candiates = reqs[:7]
				end = ", ..."
			}

			for _, candidate := range candiates {
				formattedCanditates = append(formattedCanditates, fmt.Sprintf("%v", *candidate))
			}

			return nil, fmt.Errorf("Could not find requirement that contains file %s. Candidates were: %s",
				file, strings.Join(formattedCanditates, ", ")+end)
		}

		return foundReq.SourceUnit(), nil
	}

	// Case 3: in std lib
	pythonDirIdx := -1
	for i, cmp := range fileCmps {
		if strings.HasPrefix(cmp, "python") {
			pythonDirIdx = i
			break
		}
	}
	if pythonDirIdx != -1 {
		return stdLibPkg.SourceUnit(), nil
	}

	return nil, fmt.Errorf("Cannot infer source unit for file %s", file)
}

func isSubPath(parent, child string) bool {
	relpath, err := filepath.Rel(parent, child)
	return err == nil && !strings.HasPrefix(relpath, "..")
}

func moduleToFilepath(moduleName string, isPackage bool) string {
	moduleName = strings.Replace(moduleName, ".", "/", -1)
	if !isPackage {
		moduleName += ".py"
	}
	return moduleName
}

type RawOutput struct {
	Defs []*RawDef
	Refs []*RawRef
}

type RawDef struct {
	Path      string
	Kind      string
	Name      string
	File      string // relative path (to source unit directory)
	DefStart  uint32
	DefEnd    uint32
	Exported  bool
	Docstring string
	Data      interface{}
}

type RawRef struct {
	DefPath   string
	Def       bool
	DefFile   string // absolute path
	File      string // relative path (to source unit directory)
	Start     uint32
	End       uint32
	ToBuiltin bool
}

func installPipRequirements(pipBin string, requirementFiles []string) {
	for _, requirementFile := range requirementFiles {
		err := runCmdStderr(exec.Command(pipBin, "install", "-r", requirementFile))
		if err != nil {
			log.Printf("Error installing dependencies in %s. Trying piecemeal install", requirementFile)
			if b, err := ioutil.ReadFile(requirementFile); err == nil {
				for _, req := range strings.Split(string(b), "\n") {
					runCmdLogError(exec.Command(pipBin, "install", req))
				}
			} else {
				log.Printf("Could not read %s: %s", requirementFile, err)
			}
		}
	}
}
