package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/consts"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/generation"
	"github.com/derision-test/go-mockgen/v2/internal/mockgen/paths"
	"gopkg.in/yaml.v3"
)

func parseAndValidateOptions() ([]*generation.Options, error) {
	allOptions, err := parseOptions()
	if err != nil {
		return nil, err
	}

	validators := []func(opts *generation.Options) (bool, error){
		validateOutputPaths,
		validateOptions,
	}

	for _, opts := range allOptions {
		for _, f := range validators {
			if fatal, err := f(opts); err != nil {
				if !fatal {
					kingpin.Fatalf("%s, try --help", err.Error())
				}

				return nil, err
			}
		}
	}

	return allOptions, nil
}

func parseOptions() ([]*generation.Options, error) {
	if len(os.Args) == 1 {
		return parseManifest("")
	}

	opts, err := parseFlags()
	if err != nil {
		return nil, err
	}

	if opts.ManifestDir != "" {
		return parseManifest(opts.ManifestDir)
	}

	return []*generation.Options{opts}, nil
}

func parseFlags() (*generation.Options, error) {
	opts := &generation.Options{
		PackageOptions: []generation.PackageOptions{
			{
				ImportPaths: []string{},
				Interfaces:  []string{},
			},
		},
	}

	app := kingpin.New(consts.Name, consts.Description).Version(consts.Version)
	app.UsageWriter(os.Stdout)

	app.Arg("path", "The import paths used to search for eligible interfaces").StringsVar(&opts.PackageOptions[0].ImportPaths)
	app.Flag("package", "The name of the generated package. It will be inferred from the output options by default.").Short('p').StringVar(&opts.ContentOptions.PkgName)
	app.Flag("interfaces", "A list of target interfaces to generate defined in the given the import paths.").Short('i').StringsVar(&opts.PackageOptions[0].Interfaces)
	app.Flag("exclude", "A list of interfaces to exclude from generation. Mocks for all other exported interfaces defined in the given import paths are generated.").Short('e').StringsVar(&opts.PackageOptions[0].Exclude)
	app.Flag("dirname", "The target output directory. Each mock will be written to a unique file.").Short('d').StringVar(&opts.OutputOptions.OutputDir)
	app.Flag("filename", "The target output file. All mocks are written to this file.").Short('o').StringVar(&opts.OutputOptions.OutputFilename)
	app.Flag("import-path", "The import path of the generated package. It will be inferred from the target directory by default.").StringVar(&opts.ContentOptions.OutputImportPath)
	app.Flag("prefix", "A prefix used in the name of each mock struct. Should be TitleCase by convention.").StringVar(&opts.ContentOptions.Prefix)
	app.Flag("constructor-prefix", "A prefix used in the name of each mock constructor function (after the initial `New`/`NewStrict` prefixes). Should be TitleCase by convention.").StringVar(&opts.ContentOptions.ConstructorPrefix)
	app.Flag("force", "Do not abort if a write to disk would overwrite an existing file.").Short('f').BoolVar(&opts.OutputOptions.Force)
	app.Flag("disable-formatting", "Do not run goimports over the rendered files.").BoolVar(&opts.OutputOptions.DisableFormatting)
	app.Flag("goimports", "Path to the goimports binary.").Default("goimports").StringVar(&opts.OutputOptions.GoImportsBinary)
	app.Flag("for-test", "Append _test suffix to generated package names and file names.").Default("false").BoolVar(&opts.OutputOptions.ForTest)
	app.Flag("file-prefix", "Content that is written at the top of each generated file.").StringVar(&opts.ContentOptions.FilePrefix)
	app.Flag("build-constraints", "Build constraints that are added to each generated file.").StringVar(&opts.ContentOptions.BuildConstraints)
	app.Flag("manifest-dir", "Dir in which to search for the root mockgen.yaml file in. All other flags are ignored if this is set, and config is taken from the manifest file(s).").StringVar(&opts.ManifestDir)

	if _, err := app.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	return opts, nil
}

func parseManifest(manifestDir string) ([]*generation.Options, error) {
	payload, err := readManifest(manifestDir)
	if err != nil {
		return nil, err
	}

	allOptions := make([]*generation.Options, 0, len(payload.Mocks))
	for _, opts := range payload.Mocks {
		// Mix
		opts.Exclude = append(opts.Exclude, payload.Exclude...)

		// Set if not overwritten in this entry
		if opts.Prefix == "" {
			opts.Prefix = payload.Prefix
		}
		if opts.ConstructorPrefix == "" {
			opts.ConstructorPrefix = payload.ConstructorPrefix
		}
		if opts.Goimports == "" {
			opts.Goimports = payload.Goimports
		}
		if opts.FilePrefix == "" {
			opts.FilePrefix = payload.FilePrefix
		}

		// Overwrite
		if payload.Force {
			opts.Force = true
		}
		if payload.DisableFormatting {
			opts.DisableFormatting = true
		}
		if payload.ForTest {
			opts.ForTest = true
		}

		if len(opts.Paths) > 0 && len(opts.Archives) > 0 {
			return nil, fmt.Errorf("multiple import paths and archives are mutually exclusive")
		}

		// Canonicalization
		paths := opts.Paths
		if opts.Path != "" {
			paths = append(paths, opts.Path)
		}

		// Defaults
		if opts.Goimports == "" {
			opts.Goimports = "goimports"
		}

		var packageOptions []generation.PackageOptions
		if len(opts.Sources) > 0 {
			if len(opts.Paths) > 0 || len(opts.Interfaces) > 0 || opts.Path != "" {
				return nil, fmt.Errorf("sources and path/paths/interfaces are mutually exclusive")
			}

			for _, source := range opts.Sources {
				if len(source.Paths) > 0 && len(opts.Archives) > 0 {
					return nil, fmt.Errorf("multiple import paths and archives are mutually exclusive")
				}

				// Canonicalization
				paths := source.Paths
				if source.Path != "" {
					paths = append(paths, source.Path)
				}

				packageOptions = append(packageOptions, generation.PackageOptions{
					ImportPaths: paths,
					Interfaces:  source.Interfaces,
					Exclude:     source.Exclude,
					Prefix:      source.Prefix,
					Archives:    opts.Archives,
					SourceFiles: source.SourceFiles,
					StdlibRoot:  payload.StdlibRoot,
				})
			}
		} else {
			packageOptions = append(packageOptions, generation.PackageOptions{
				ImportPaths: paths,
				Interfaces:  opts.Interfaces,
				Exclude:     opts.Exclude,
				Prefix:      opts.Prefix,
				Archives:    opts.Archives,
				SourceFiles: opts.SourceFiles,
				StdlibRoot:  payload.StdlibRoot,
			})
		}

		allOptions = append(allOptions, &generation.Options{
			PackageOptions: packageOptions,
			OutputOptions: generation.OutputOptions{
				OutputDir:         opts.Dirname,
				OutputFilename:    opts.Filename,
				Force:             opts.Force,
				DisableFormatting: opts.DisableFormatting,
				GoImportsBinary:   opts.Goimports,
				ForTest:           opts.ForTest,
			},
			ContentOptions: generation.ContentOptions{
				PkgName:           opts.Package,
				OutputImportPath:  opts.ImportPath,
				Prefix:            opts.Prefix,
				ConstructorPrefix: opts.ConstructorPrefix,
				FilePrefix:        opts.FilePrefix,
			},
		})
	}

	return allOptions, nil
}

type yamlPayload struct {
	// Meta options
	IncludeConfigPaths []string `yaml:"include-config-paths"`

	// Global options
	Exclude           []string `yaml:"exclude"`
	Prefix            string   `yaml:"prefix"`
	ConstructorPrefix string   `yaml:"constructor-prefix"`
	Force             bool     `yaml:"force"`
	DisableFormatting bool     `yaml:"disable-formatting"`
	Goimports         string   `yaml:"goimports"`
	ForTest           bool     `yaml:"for-test"`
	FilePrefix        string   `yaml:"file-prefix"`

	StdlibRoot string `yaml:"stdlib-root"`

	Mocks []yamlMock `yaml:"mocks"`
}

type yamlMock struct {
	Path              string       `yaml:"path"`
	Paths             []string     `yaml:"paths"`
	Sources           []yamlSource `yaml:"sources"`
	SourceFiles       []string     `yaml:"source-files"`
	Archives          []string     `yaml:"archives"`
	Package           string       `yaml:"package"`
	Interfaces        []string     `yaml:"interfaces"`
	Exclude           []string     `yaml:"exclude"`
	Dirname           string       `yaml:"dirname"`
	Filename          string       `yaml:"filename"`
	ImportPath        string       `yaml:"import-path"`
	Prefix            string       `yaml:"prefix"`
	ConstructorPrefix string       `yaml:"constructor-prefix"`
	Force             bool         `yaml:"force"`
	DisableFormatting bool         `yaml:"disable-formatting"`
	Goimports         string       `yaml:"goimports"`
	ForTest           bool         `yaml:"for-test"`
	FilePrefix        string       `yaml:"file-prefix"`
}

type yamlSource struct {
	Path        string   `yaml:"path"`
	Paths       []string `yaml:"paths"`
	Interfaces  []string `yaml:"interfaces"`
	Exclude     []string `yaml:"exclude"`
	Prefix      string   `yaml:"prefix"`
	SourceFiles []string `yaml:"source-files"`
}

func readManifest(manifestDir string) (yamlPayload, error) {
	contents, err := os.ReadFile(filepath.Join(manifestDir, "mockgen.yaml"))
	if err != nil {
		return yamlPayload{}, err
	}

	var payload yamlPayload
	if err := yaml.Unmarshal(contents, &payload); err != nil {
		return yamlPayload{}, err
	}

	for _, path := range payload.IncludeConfigPaths {
		payload, err = readIncludeConfig(payload, filepath.Join(manifestDir, path))
		if err != nil {
			return yamlPayload{}, err
		}
	}

	return payload, nil
}

func readIncludeConfig(payload yamlPayload, path string) (yamlPayload, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return yamlPayload{}, err
	}

	var mocks []yamlMock
	if err := yaml.Unmarshal(contents, &mocks); err != nil {
		return yamlPayload{}, err
	}

	payload.Mocks = append(payload.Mocks, mocks...)
	return payload, nil
}

func validateOutputPaths(opts *generation.Options) (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return true, fmt.Errorf("failed to get current directory")
	}

	if opts.OutputOptions.OutputFilename == "" && opts.OutputOptions.OutputDir == "" {
		opts.OutputOptions.OutputDir = wd
	}

	if opts.OutputOptions.OutputFilename != "" && opts.OutputOptions.OutputDir != "" {
		return false, fmt.Errorf("dirname and filename are mutually exclusive")
	}

	if opts.OutputOptions.OutputFilename != "" {
		opts.OutputOptions.OutputDir = path.Dir(opts.OutputOptions.OutputFilename)
		opts.OutputOptions.OutputFilename = path.Base(opts.OutputOptions.OutputFilename)
	}

	if err := paths.EnsureDirExists(opts.OutputOptions.OutputDir); err != nil {
		return true, fmt.Errorf(
			"failed to make output directory %s: %s",
			opts.OutputOptions.OutputDir,
			err.Error(),
		)
	}

	if opts.OutputOptions.OutputDir, err = cleanPath(opts.OutputOptions.OutputDir); err != nil {
		return true, err
	}

	return false, nil
}

var goIdentifierPattern = regexp.MustCompile("^[A-Za-z]([A-Za-z0-9_]*)?$")

func validateOptions(opts *generation.Options) (bool, error) {
	for _, packageOpts := range opts.PackageOptions {
		if len(packageOpts.ImportPaths) == 0 {
			return false, fmt.Errorf("missing interface source import paths")
		}

		if len(packageOpts.Interfaces) != 0 && len(packageOpts.Exclude) != 0 {
			return false, fmt.Errorf("interface lists and exclude lists are mutually exclusive")
		}

		if packageOpts.Prefix != "" && !goIdentifierPattern.Match([]byte(packageOpts.Prefix)) {
			return false, fmt.Errorf("prefix `%s` is illegal", packageOpts.Prefix)
		}
	}

	if opts.ContentOptions.OutputImportPath == "" {
		path, ok := paths.InferImportPath(opts.OutputOptions.OutputDir)
		if !ok {
			return false, fmt.Errorf("could not infer output import path")
		}

		opts.ContentOptions.OutputImportPath = path
	}

	if opts.ContentOptions.PkgName == "" {
		opts.ContentOptions.PkgName = opts.ContentOptions.OutputImportPath[strings.LastIndex(opts.ContentOptions.OutputImportPath, string(os.PathSeparator))+1:]
	}

	if !goIdentifierPattern.Match([]byte(opts.ContentOptions.PkgName)) {
		return false, fmt.Errorf("package name `%s` is illegal", opts.ContentOptions.PkgName)
	}

	if opts.ContentOptions.Prefix != "" && !goIdentifierPattern.Match([]byte(opts.ContentOptions.Prefix)) {
		return false, fmt.Errorf("prefix `%s` is illegal", opts.ContentOptions.Prefix)
	}

	if opts.ContentOptions.ConstructorPrefix != "" && !goIdentifierPattern.Match([]byte(opts.ContentOptions.ConstructorPrefix)) {
		return false, fmt.Errorf("constructor-`prefix `%s` is illegal", opts.ContentOptions.ConstructorPrefix)
	}

	return false, nil
}

func cleanPath(path string) (cleaned string, err error) {
	if path, err = filepath.Abs(path); err != nil {
		return "", err
	}

	if path, err = filepath.EvalSymlinks(path); err != nil {
		return "", err
	}

	return path, nil
}
