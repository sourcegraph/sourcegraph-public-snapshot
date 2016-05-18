package cli

import (
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

	"github.com/alexsaveliev/go-colorable-wrapper"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/toolchain"
	"sourcegraph.com/sourcegraph/srclib/util"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		c, err := cli.AddCommand("toolchain",
			"manage toolchains",
			"Manage srclib toolchains.",
			&toolchainCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
		c.Aliases = []string{"tc"}

		_, err = c.AddCommand("list",
			"list available toolchains",
			"List available toolchains.",
			&toolchainListCmd,
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("list-tools",
			"list tools in toolchains",
			"List available tools in all toolchains.",
			&toolchainListToolsCmd,
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("bundle",
			"bundle a toolchain",
			"The bundle subcommand builds and archives toolchain bundles (.tar.gz files, one per toolchain variant). Bundles contain prebuilt toolchains and allow people to use srclib toolchains without needing to compile them on their own system.",
			&toolchainBundleCmd,
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("unbundle",
			"unbundle a toolchain",
			"The unbundle subcommand unarchives a toolchain bundle (previously created with the 'bundle' subcommand). It allows people to download and use prebuilt toolchains without needing to compile them on their system.",
			&toolchainUnbundleCmd,
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("install",
			"install toolchains",
			"Download and install toolchains",
			&toolchainInstallCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
	})
}

type ToolchainPath string

func (t ToolchainPath) Complete(match string) []flags.Completion {
	toolchains, err := toolchain.List()
	if err != nil {
		log.Println(err)
		return nil
	}
	var comps []flags.Completion
	for _, tc := range toolchains {
		if strings.HasPrefix(tc.Path, match) {
			comps = append(comps, flags.Completion{Item: tc.Path})
		}
	}
	return comps
}

type ToolchainCmd struct{}

var toolchainCmd ToolchainCmd

func (c *ToolchainCmd) Execute(args []string) error { return nil }

type ToolchainListCmd struct {
}

var toolchainListCmd ToolchainListCmd

func (c *ToolchainListCmd) Execute(args []string) error {
	toolchains, err := toolchain.List()
	if err != nil {
		return err
	}
	for _, t := range toolchains {
		fmt.Println(t.Path)
	}
	return nil
}

type ToolchainListToolsCmd struct {
	Op             string `short:"p" long:"op" description:"only list tools that perform these operations only" value-name:"OP"`
	SourceUnitType string `short:"u" long:"source-unit-type" description:"only list tools that operate on this source unit type" value-name:"TYPE"`
	Args           struct {
		Toolchains []ToolchainPath `name:"TOOLCHAINS" description:"only list tools in these toolchains"`
	} `positional-args:"yes" required:"yes"`
}

var toolchainListToolsCmd ToolchainListToolsCmd

func (c *ToolchainListToolsCmd) Execute(args []string) error {
	tcs, err := toolchain.List()
	if err != nil {
		log.Fatal(err)
	}

	fmtStr := "%-40s  %-18s  %-15s  %-25s\n"
	colorable.Printf(fmtStr, "TOOLCHAIN", "TOOL", "OP", "SOURCE UNIT TYPES")
	for _, tc := range tcs {
		if len(c.Args.Toolchains) > 0 {
			found := false
			for _, tc2 := range c.Args.Toolchains {
				if string(tc2) == tc.Path {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		cfg, err := tc.ReadConfig()
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range cfg.Tools {
			if c.Op != "" && c.Op != t.Op {
				continue
			}
			if c.SourceUnitType != "" {
				found := false
				for _, u := range t.SourceUnitTypes {
					if c.SourceUnitType == u {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			colorable.Printf(fmtStr, tc.Path, t.Subcmd, t.Op, strings.Join(t.SourceUnitTypes, " "))
		}
	}
	return nil
}

type ToolchainBundleCmd struct {
	Variant string `long:"variant" description:"only produce a bundle for the given variant (default is all variants)"`
	DryRun  bool   `short:"n" long:"dry-run" description:"don't do anything, but print what would be done"`

	Args struct {
		Toolchain ToolchainPath `name:"TOOLCHAIN" description:"toolchain to bundle" required:"yes"`
		Dir       string        `name:"TOOLCHAIN-DIR" description:"dir containing toolchain files (default: look up TOOLCHAIN in SRCLIBPATH)"`
	} `positional-args:"yes"`
}

var toolchainBundleCmd ToolchainBundleCmd

func (c *ToolchainBundleCmd) Execute(args []string) error {
	log.Printf("Bundling toolchain %s...", c.Args.Toolchain)

	tmpDir, err := ioutil.TempDir("", path.Base(string(c.Args.Toolchain))+"toolchain-bundle")
	if err != nil {
		return err
	}
	log.Printf(" - output dir: %s", tmpDir)

	var variants []toolchain.Variant
	if c.Variant != "" {
		variants = append(variants, toolchain.ParseVariant(c.Variant))
	}

	if c.Args.Dir == "" {
		info, err := toolchain.Lookup(string(c.Args.Toolchain))
		if err != nil {
			return err
		}
		c.Args.Dir = info.Dir
	}

	bundles, err := toolchain.Bundle(c.Args.Dir, tmpDir, variants, c.DryRun, GlobalOpt.Verbose)
	if err != nil {
		return err
	}

	log.Println()
	log.Println("Bundles ready:", tmpDir)
	for _, b := range bundles {
		log.Println("   ", b)
	}

	return nil
}

type ToolchainUnbundleCmd struct {
	Args struct {
		Toolchain  string `name:"TOOLCHAIN" description:"toolchain path to unbundle to"`
		BundleFile string `name:"BUNDLE-FILE" description:"bundle file containing toolchain dir contents (.tar.gz, .tar, etc.)"`
	} `positional-args:"yes" required:"yes"`
}

var toolchainUnbundleCmd ToolchainUnbundleCmd

func (c *ToolchainUnbundleCmd) Execute(args []string) error {
	if GlobalOpt.Verbose {
		log.Printf("Unarchiving from bundle file %s to toolchain %s", c.Args.BundleFile, c.Args.Toolchain)
	}

	f, err := os.Open(c.Args.BundleFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return toolchain.Unbundle(c.Args.Toolchain, c.Args.BundleFile, f)
}

type toolchainInstaller struct {
	name string
	fn   func() error
}

type toolchainMap map[string]toolchainInstaller

var stdToolchains = toolchainMap{
	"go":         toolchainInstaller{"Go (sourcegraph.com/sourcegraph/srclib-go)", installGoToolchain},
	"python":     toolchainInstaller{"Python (sourcegraph.com/sourcegraph/srclib-python)", installPythonToolchain},
	"ruby":       toolchainInstaller{"Ruby (sourcegraph.com/sourcegraph/srclib-ruby)", installRubyToolchain},
	"javascript": toolchainInstaller{"JavaScript (sourcegraph.com/sourcegraph/srclib-javascript)", installJavaScriptToolchain},
	"typescript": toolchainInstaller{"TypeScript (sourcegraph.com/sourcegraph/srclib-typescript)", installTypeScriptToolchain},
	"java":       toolchainInstaller{"Java (sourcegraph.com/sourcegraph/srclib-java)", installJavaToolchain},
	"basic":      toolchainInstaller{"PHP, Objective-C (sourcegraph.com/sourcegraph/srclib-basic)", installBasicToolchain},
	"csharp":     toolchainInstaller{"C# (sourcegraph.com/sourcegraph/srclib-csharp)", installCSharpToolchain},
}

func (m toolchainMap) listKeys() string {
	var langs string
	for i := range m {
		langs += i + ", "
	}
	// Remove the last comma from langs before returning it.
	return strings.TrimSuffix(langs, ", ")
}

type ToolchainInstallCmd struct {
	// Args are not required so we can print out a more detailed
	// error message inside (*ToolchainInstallCmd).Execute.
	Args struct {
		Languages []string `value-name:"LANG" description:"language toolchains to install"`
	} `positional-args:"yes"`
}

var toolchainInstallCmd ToolchainInstallCmd

func (c *ToolchainInstallCmd) Execute(args []string) error {
	if len(c.Args.Languages) == 0 {
		return errors.New(colorable.Red(fmt.Sprintf("No languages specified. Standard languages include: %s", stdToolchains.listKeys())))
	}
	var is []toolchainInstaller
	for _, l := range c.Args.Languages {
		i, ok := stdToolchains[l]
		if !ok {
			return errors.New(colorable.Red(fmt.Sprintf("Language %s unrecognized. Standard languages include: %s", l, stdToolchains.listKeys())))
		}
		is = append(is, i)
	}
	return installToolchains(is)
}

func installToolchains(langs []toolchainInstaller) error {
	for _, l := range langs {
		colorable.Println(colorable.Cyan(l.name + " " + strings.Repeat("=", 78-len(l.name))))
		if err := l.fn(); err != nil {
			return fmt.Errorf("%s\n", colorable.Red(fmt.Sprintf("failed to install/upgrade %s toolchain: %s", l.name, err)))
		}

		colorable.Println(colorable.Green("OK! Installed/upgraded " + l.name + " toolchain"))
		colorable.Println(colorable.Cyan(strings.Repeat("=", 80)))
		colorable.Println()
	}
	return nil
}

func installGoToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-go"

	// Identify if Go is installed already or not.
	if _, err := exec.LookPath("go"); isExecErrNotFound(err) {
		return errors.New(`
Refusing to install Go toolchain because Go is not installed or is not on the
system path.

-> Please install the latest version of Go (https://golang.org/doc/install) and
run this command again.`)
	} else if err != nil {
		return err
	}

	// retrieve or create GOPATH
	gopathDir := getGoPath()

	// Go-based toolchains should be cloned into GOPATH/src/TOOLCHAIN
	// otherwise govendor refuses to work if source code is located outside of GOPATH
	gopathDir = filepath.Join(gopathDir, "src", toolchain)
	if err := os.MkdirAll(filepath.Dir(gopathDir), 0700); err != nil {
		return err
	}

	log.Println("Downloading Go toolchain")
	if err := cloneToolchain(gopathDir, toolchain); err != nil {
		return err
	}

	// making parent directory of toolchain in SRCLIBPATH
	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH
	if err := os.MkdirAll(filepath.Dir(srclibpathDir), 0700); err != nil {
		return err
	}

	// Adding symlink SRCLIBPATH/TOOLCHAIN that points to GOPATH/src/TOOLCHAIN
	err := symlink(gopathDir, srclibpathDir)
	if err != nil {
		return err
	}

	log.Println("Building Go toolchain program")
	if err := execCmdInDir(gopathDir, "make"); err != nil {
		return err
	}

	return nil
}

func installRubyToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-ruby"

	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH
	if err := os.MkdirAll(filepath.Dir(srclibpathDir), 0700); err != nil {
		return err
	}

	if _, err := exec.LookPath("ruby"); isExecErrNotFound(err) {
		return errors.New("no `ruby` in PATH (do you have Ruby installed properly?)")
	} else if err != nil {
		return err
	}
	if _, err := exec.LookPath("bundle"); isExecErrNotFound(err) {
		return fmt.Errorf("found `ruby` in PATH but did not find `bundle` in PATH; Ruby toolchain requires bundler (run `gem install bundler` to install it)")
	} else if err != nil {
		return err
	}

	log.Println("Downloading Ruby toolchain in", srclibpathDir)
	if err := cloneToolchain(srclibpathDir, toolchain); err != nil {
		return err
	}

	log.Println("Installing deps for Ruby toolchain in", srclibpathDir)
	if err := execCmdInDir(srclibpathDir, "make"); err != nil {
		return fmt.Errorf("%s\n\nTip: If you are using a version of Ruby other than 2.1.2 (the default for srclib), or if you are using your system Ruby, try using a Ruby version manager (such as https://rvm.io) to install a more standard Ruby, and try Ruby 2.1.2.\n\nIf you are still having problems, post an issue at https://github.com/sourcegraph/srclib-ruby/issues with the full log output and information about your OS and Ruby version.\n\n`.", err)
	}

	return nil
}

func installJavaScriptToolchain() error {
	const toolchain = "github.com/sourcegraph/srclib-javascript"

	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH

	if _, err := exec.LookPath("node"); isExecErrNotFound(err) {
		return errors.New("no `node` in PATH (do you have Node.js installed properly?)")
	}
	if _, err := exec.LookPath("npm"); isExecErrNotFound(err) {
		return fmt.Errorf("no `npm` in PATH; JavaScript toolchain requires npm")
	}

	log.Println("Downloading JavaScript toolchain in", srclibpathDir)
	if err := cloneToolchain(srclibpathDir, toolchain); err != nil {
		return err
	}

	log.Println("Building JavaScript toolchain program")
	if err := execCmdInDir(srclibpathDir, "npm", "install"); err != nil {
		return err
	}

	return nil
}

func installTypeScriptToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-typescript"

	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH

	if _, err := exec.LookPath("node"); isExecErrNotFound(err) {
		return errors.New("no `node` in PATH (do you have Node.js installed properly?)")
	}
	if _, err := exec.LookPath("npm"); isExecErrNotFound(err) {
		return fmt.Errorf("no `npm` in PATH; TypeScript toolchain requires npm")
	}

	if _, err := exec.LookPath("tsc"); isExecErrNotFound(err) {
		return fmt.Errorf("no `tsc` in PATH; TypeScript toolchain requires tsc")
	}

	log.Println("Downloading TypeScript toolchain in", srclibpathDir)
	if err := cloneToolchain(srclibpathDir, toolchain); err != nil {
		return err
	}

	log.Println("Building TypeScript toolchain program")
	if err := execCmdInDir(srclibpathDir, "make"); err != nil {
		return err
	}

	return nil
}

func installPythonToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-python"

	requiredCmds := map[string]string{
		"go":         "visit https://golang.org/doc/install",
		"python":     "visit https://www.python.org/downloads/",
		"pip":        "visit http://pip.readthedocs.org/en/latest/installing.html",
		"virtualenv": "run `[sudo] pip install virtualenv`",
	}
	for requiredCmd, instructions := range requiredCmds {
		if _, err := exec.LookPath(requiredCmd); isExecErrNotFound(err) {
			return fmt.Errorf("no `%s` found in PATH; to install, %s", requiredCmd, instructions)
		}
	}

	// retrieve or create GOPATH
	gopathDir := getGoPath()

	// Go-based toolchains should be cloned into GOPATH/src/TOOLCHAIN
	// otherwise govendor refuses to work if source code is located outside of GOPATH
	gopathDir = filepath.Join(gopathDir, "src", toolchain)
	if err := os.MkdirAll(filepath.Dir(gopathDir), 0700); err != nil {
		return err
	}

	log.Println("Downloading Python toolchain")
	if err := cloneToolchain(gopathDir, toolchain); err != nil {
		return err
	}

	// making parent directory of toolchain in SRCLIBPATH
	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH
	if err := os.MkdirAll(filepath.Dir(srclibpathDir), 0700); err != nil {
		return err
	}

	// Adding symlink SRCLIBPATH/TOOLCHAIN that points to GOPATH/src/TOOLCHAIN
	err := symlink(gopathDir, srclibpathDir)
	if err != nil {
		return err
	}

	log.Println("Building Python toolchain program")
	if err := execCmdInDir(gopathDir, "make"); err != nil {
		return err
	}

	return nil
}

func installJavaToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-java"

	reqCmds := []string{"java", "gradle"}
	for _, cmd := range reqCmds {
		if _, err := exec.LookPath(cmd); isExecErrNotFound(err) {
			return fmt.Errorf(`
Refusing to install Java toolchain because %s is not installed or is not on the system path.

-> Please install %s and run this command again`, cmd, cmd)
		} else if err != nil {
			return err
		}
	}

	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH
	if err := os.MkdirAll(filepath.Dir(srclibpathDir), 0700); err != nil {
		return err
	}

	log.Println("Downloading Java toolchain in", srclibpathDir)
	if err := cloneToolchain(srclibpathDir, toolchain); err != nil {
		return err
	}

	log.Println("Building Java toolchain program")
	if err := execCmdInDir(srclibpathDir, "make"); err != nil {
		return err
	}

	return nil
}

func installCSharpToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-csharp"

	requiredCmds := map[string]string{
		"dnx": "see http://docs.asp.net/en/latest/getting-started/installing-on-linux.html for details",
		"dnu": "see http://docs.asp.net/en/latest/getting-started/installing-on-linux.html for details",
	}
	for requiredCmd, instructions := range requiredCmds {
		if _, err := exec.LookPath(requiredCmd); isExecErrNotFound(err) {
			return fmt.Errorf("no `%s` found in PATH; to install, %s", requiredCmd, instructions)
		}
	}

	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH
	if err := os.MkdirAll(filepath.Dir(srclibpathDir), 0700); err != nil {
		return err
	}

	log.Println("Downloading C# toolchain in", srclibpathDir)
	if err := cloneToolchain(srclibpathDir, toolchain); err != nil {
		return err
	}

	nugetdir := filepath.Join(srclibpathDir, "Srclib.Nuget")
	log.Println("Downloading toolchain dependencies in", nugetdir)
	if err := execCmdInDir(nugetdir, "dnu", "restore"); err != nil {
		return err
	}

	return nil
}

func installBasicToolchain() error {
	const toolchain = "sourcegraph.com/sourcegraph/srclib-basic"

	reqCmds := []string{"java"}
	for _, cmd := range reqCmds {
		if _, err := exec.LookPath(cmd); isExecErrNotFound(err) {
			return fmt.Errorf(`
Refusing to install Basic toolchain because %s is not installed or is not on the system path.

-> Please install %s and run this command again`, cmd, cmd)
		} else if err != nil {
			return err
		}
	}

	srclibpathDir := filepath.Join(filepath.SplitList(srclib.Path)[0], toolchain) // toolchain dir under SRCLIBPATH
	if err := os.MkdirAll(filepath.Dir(srclibpathDir), 0700); err != nil {
		return err
	}

	log.Println("Downloading Basic toolchain in", srclibpathDir)
	if err := cloneToolchain(srclibpathDir, toolchain); err != nil {
		return err
	}

	log.Println("Building Basic toolchain program")
	if err := execCmdInDir(srclibpathDir, "make"); err != nil {
		return err
	}

	return nil
}

func cloneToolchain(dest, toolchain string) error {
	if fi, err := os.Stat(dest); os.IsNotExist(err) {
		// Clone
		if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			return err
		}

		cmd := exec.Command("git", "clone", "https://"+toolchain)
		cmd.Dir = filepath.Dir(dest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else if err != nil {
		return err
	} else if !fi.Mode().IsDir() {
		return fmt.Errorf("not a directory: %s", dest)
	}
	log.Printf("Toolchain directory %q already exists, using existing version.", dest)
	return nil
}

func isExecErrNotFound(err error) bool {
	if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
		return true
	}
	return false
}

// getGoPath returns first item in the GOPATH list
// If there is no GOPATH set function sets GOPATH to ~/.srclib-gopath and returns ~/.srclib-gopath
func getGoPath() string {
	list := os.Getenv("GOPATH")
	if list == "" {
		goPath := path.Join(util.CurrentUserHomeDir(), ".srclib-gopath")
		os.Setenv("GOPATH", goPath)
		return goPath
	}
	return filepath.SplitList(list)[0]
}

// symlink makes symlink "target" that points to "source"
func symlink(source, target string) error {
	if _, err := os.Lstat(target); os.IsNotExist(err) {
		log.Printf("mkdir -p %s", filepath.Dir(target))
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
		if runtime.GOOS != "windows" {
			log.Printf("ln -s %s %s", source, target)
			if err := os.Symlink(source, target); err != nil {
				return err
			}
		} else {
			// os.Symlink makes "file symbolic link" on Windows making impossible to install Go toolchain
			// because `cd foo && make` requires "foo" to be either a directory or so-called "directory symbolic link".
			// That's why we had to use `mklink /D bar foo`
			if err := execCmdInDir(source, "cmd", "/c", "mklink", "/D", target, source); err != nil {
				return err
			}
		}
		return nil
	} else {
		return err
	}
}
