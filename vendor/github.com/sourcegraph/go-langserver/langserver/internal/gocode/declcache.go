package gocode

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//-------------------------------------------------------------------------
// []package_import
//-------------------------------------------------------------------------

type package_import struct {
	alias   string
	abspath string
	path    string
}

// Parses import declarations until the first non-import declaration and fills
// `packages` array with import information.
func collect_package_imports(filename string, decls []ast.Decl, context *package_lookup_context) []package_import {
	pi := make([]package_import, 0, 16)
	for _, decl := range decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			for _, spec := range gd.Specs {
				imp := spec.(*ast.ImportSpec)
				path, alias := path_and_alias(imp)
				abspath, ok := abs_path_for_package(filename, path, context)
				if ok && alias != "_" {
					pi = append(pi, package_import{alias, abspath, path})
				}
			}
		} else {
			break
		}
	}
	return pi
}

//-------------------------------------------------------------------------
// decl_file_cache
//
// Contains cache for top-level declarations of a file as well as its
// contents, AST and import information.
//-------------------------------------------------------------------------

type decl_file_cache struct {
	name  string // file name
	mtime int64  // last modification time

	decls     map[string]*decl // top-level declarations
	error     error            // last error
	packages  []package_import // import information
	filescope *scope

	fset    *token.FileSet
	context *package_lookup_context
}

func new_decl_file_cache(name string, context *package_lookup_context) *decl_file_cache {
	return &decl_file_cache{
		name:    name,
		context: context,
	}
}

func (f *decl_file_cache) update() {
	stat, err := os.Stat(f.name)
	if err != nil {
		f.decls = nil
		f.error = err
		f.fset = nil
		return
	}

	statmtime := stat.ModTime().UnixNano()
	if f.mtime == statmtime {
		return
	}

	f.mtime = statmtime
	f.read_file()
}

func (f *decl_file_cache) read_file() {
	var data []byte
	data, f.error = file_reader.read_file(f.name)
	if f.error != nil {
		return
	}
	data, _ = filter_out_shebang(data)

	f.process_data(data)
}

func (f *decl_file_cache) process_data(data []byte) {
	var file *ast.File
	f.fset = token.NewFileSet()
	file, f.error = parser.ParseFile(f.fset, "", data, 0)
	f.filescope = new_scope(nil)
	for _, d := range file.Decls {
		anonymify_ast(d, 0, f.filescope)
	}
	f.packages = collect_package_imports(f.name, file.Decls, f.context)
	f.decls = make(map[string]*decl, len(file.Decls))
	for _, decl := range file.Decls {
		append_to_top_decls(f.decls, decl, f.filescope)
	}
}

func append_to_top_decls(decls map[string]*decl, decl ast.Decl, scope *scope) {
	foreach_decl(decl, func(data *foreach_decl_struct) {
		class := ast_decl_class(data.decl)
		for i, name := range data.names {
			typ, v, vi := data.type_value_index(i)

			d := new_decl_full(name.Name, class, ast_decl_flags(data.decl), typ, v, vi, scope)
			if d == nil {
				return
			}

			methodof := method_of(decl)
			if methodof != "" {
				decl, ok := decls[methodof]
				if ok {
					decl.add_child(d)
				} else {
					decl = new_decl(methodof, decl_methods_stub, scope)
					decls[methodof] = decl
					decl.add_child(d)
				}
			} else {
				decl, ok := decls[d.name]
				if ok {
					decl.expand_or_replace(d)
				} else {
					decls[d.name] = d
				}
			}
		}
	})
}

func abs_path_for_package(filename, p string, context *package_lookup_context) (string, bool) {
	dir, _ := filepath.Split(filename)
	if len(p) == 0 {
		return "", false
	}
	if p[0] == '.' {
		return fmt.Sprintf("%s.a", filepath.Join(dir, p)), true
	}
	pkg, ok := find_go_dag_package(p, dir)
	if ok {
		return pkg, true
	}
	return find_global_file(p, context)
}

func path_and_alias(imp *ast.ImportSpec) (string, string) {
	path := ""
	if imp.Path != nil && len(imp.Path.Value) > 0 {
		path = string(imp.Path.Value)
		path = path[1 : len(path)-1]
	}
	alias := ""
	if imp.Name != nil {
		alias = imp.Name.Name
	}
	return path, alias
}

func find_go_dag_package(imp, filedir string) (string, bool) {
	// Support godag directory structure
	dir, pkg := filepath.Split(imp)
	godag_pkg := filepath.Join(filedir, "..", dir, "_obj", pkg+".a")
	if file_exists(godag_pkg) {
		return godag_pkg, true
	}
	return "", false
}

// autobuild compares the mod time of the source files of the package, and if any of them is newer
// than the package object file will rebuild it.
func autobuild(p *build.Package) error {
	if p.Dir == "" {
		return fmt.Errorf("no files to build")
	}
	ps, err := os.Stat(p.PkgObj)
	if err != nil {
		// Assume package file does not exist and build for the first time.
		return build_package(p)
	}
	pt := ps.ModTime()
	fs, err := readdir_lstat(p.Dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		if f.ModTime().After(pt) {
			// Source file is newer than package file; rebuild.
			return build_package(p)
		}
	}
	return nil
}

// build_package builds the package by calling `go install package/import`. If everything compiles
// correctly, the newly compiled package should then be in the usual place in the `$GOPATH/pkg`
// directory, and gocode will pick it up from there.
func build_package(p *build.Package) error {
	if *g_debug {
		log.Printf("-------------------")
		log.Printf("rebuilding package %s", p.Name)
		log.Printf("package import: %s", p.ImportPath)
		log.Printf("package object: %s", p.PkgObj)
		log.Printf("package source dir: %s", p.Dir)
		log.Printf("package source files: %v", p.GoFiles)
		log.Printf("GOPATH: %v", g_daemon.context.GOPATH)
		log.Printf("GOROOT: %v", g_daemon.context.GOROOT)
	}
	env := os.Environ()
	for i, v := range env {
		if strings.HasPrefix(v, "GOPATH=") {
			env[i] = "GOPATH=" + g_daemon.context.GOPATH
		} else if strings.HasPrefix(v, "GOROOT=") {
			env[i] = "GOROOT=" + g_daemon.context.GOROOT
		}
	}

	cmd := exec.Command("go", "install", p.ImportPath)
	cmd.Env = env

	// TODO: Should read STDERR rather than STDOUT.
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if *g_debug {
		log.Printf("build out: %s\n", string(out))
	}
	return nil
}

// executes autobuild function if autobuild option is enabled, logs error and
// ignores it
func try_autobuild(p *build.Package) {
	if g_config.Autobuild {
		err := autobuild(p)
		if err != nil && *g_debug {
			log.Printf("Autobuild error: %s\n", err)
		}
	}
}

func log_found_package_maybe(imp, pkgpath string) {
	if *g_debug {
		log.Printf("Found %q at %q\n", imp, pkgpath)
	}
}

func log_build_context(context *package_lookup_context) {
	log.Printf(" GOROOT: %s\n", context.GOROOT)
	log.Printf(" GOPATH: %s\n", context.GOPATH)
	log.Printf(" GOOS: %s\n", context.GOOS)
	log.Printf(" GOARCH: %s\n", context.GOARCH)
	log.Printf(" BzlProjectRoot: %q\n", context.BzlProjectRoot)
	log.Printf(" GBProjectRoot: %q\n", context.GBProjectRoot)
	log.Printf(" lib-path: %q\n", g_config.LibPath)
}

// find_global_file returns the file path of the compiled package corresponding to the specified
// import, and a boolean stating whether such path is valid.
// TODO: Return only one value, possibly empty string if not found.
func find_global_file(imp string, context *package_lookup_context) (string, bool) {
	// gocode synthetically generates the builtin package
	// "unsafe", since the "unsafe.a" package doesn't really exist.
	// Thus, when the user request for the package "unsafe" we
	// would return synthetic global file that would be used
	// just as a key name to find this synthetic package
	if imp == "unsafe" {
		return "unsafe", true
	}

	pkgfile := fmt.Sprintf("%s.a", imp)

	// if lib-path is defined, use it
	if g_config.LibPath != "" {
		for _, p := range filepath.SplitList(g_config.LibPath) {
			pkg_path := filepath.Join(p, pkgfile)
			if file_exists(pkg_path) {
				log_found_package_maybe(imp, pkg_path)
				return pkg_path, true
			}
			// Also check the relevant pkg/OS_ARCH dir for the libpath, if provided.
			pkgdir := fmt.Sprintf("%s_%s", context.GOOS, context.GOARCH)
			pkg_path = filepath.Join(p, "pkg", pkgdir, pkgfile)
			if file_exists(pkg_path) {
				log_found_package_maybe(imp, pkg_path)
				return pkg_path, true
			}
		}
	}

	// gb-specific lookup mode, only if the root dir was found
	if g_config.PackageLookupMode == "gb" && context.GBProjectRoot != "" {
		root := context.GBProjectRoot
		pkgdir := filepath.Join(root, "pkg", context.GOOS+"-"+context.GOARCH)
		if !is_dir(pkgdir) {
			pkgdir = filepath.Join(root, "pkg", context.GOOS+"-"+context.GOARCH+"-race")
		}
		pkg_path := filepath.Join(pkgdir, pkgfile)
		if file_exists(pkg_path) {
			log_found_package_maybe(imp, pkg_path)
			return pkg_path, true
		}
	}

	// bzl-specific lookup mode, only if the root dir was found
	if g_config.PackageLookupMode == "bzl" && context.BzlProjectRoot != "" {
		var root, impath string
		if strings.HasPrefix(imp, g_config.CustomPkgPrefix+"/") {
			root = filepath.Join(context.BzlProjectRoot, "bazel-bin")
			impath = imp[len(g_config.CustomPkgPrefix)+1:]
		} else if g_config.CustomVendorDir != "" {
			// Try custom vendor dir.
			root = filepath.Join(context.BzlProjectRoot, "bazel-bin", g_config.CustomVendorDir)
			impath = imp
		}

		if root != "" && impath != "" {
			// There might be more than one ".a" files in the pkg path with bazel.
			// But the best practice is to keep one go_library build target in each
			// pakcage directory so that it follows the standard Go package
			// structure. Thus here we assume there is at most one ".a" file existing
			// in the pkg path.
			if d, err := os.Open(filepath.Join(root, impath)); err == nil {
				defer d.Close()

				if fis, err := d.Readdir(-1); err == nil {
					for _, fi := range fis {
						if !fi.IsDir() && filepath.Ext(fi.Name()) == ".a" {
							pkg_path := filepath.Join(root, impath, fi.Name())
							log_found_package_maybe(imp, pkg_path)
							return pkg_path, true
						}
					}
				}
			}
		}
	}

	if context.CurrentPackagePath != "" {
		// Try vendor path first, see GO15VENDOREXPERIMENT.
		// We don't check this environment variable however, seems like there is
		// almost no harm in doing so (well.. if you experiment with vendoring,
		// gocode will fail after enabling/disabling the flag, and you'll be
		// forced to get rid of vendor binaries). But asking users to set this
		// env var is up will bring more trouble. Because we also need to pass
		// it from client to server, make sure their editors set it, etc.
		// So, whatever, let's just pretend it's always on.
		package_path := context.CurrentPackagePath
		for {
			limp := filepath.Join(package_path, "vendor", imp)
			if p, err := context.Import(limp, "", build.AllowBinary|build.FindOnly); err == nil {
				try_autobuild(p)
				if file_exists(p.PkgObj) {
					log_found_package_maybe(imp, p.PkgObj)
					return p.PkgObj, true
				}
			}
			if package_path == "" {
				break
			}
			next_path := filepath.Dir(package_path)
			// let's protect ourselves from inf recursion here
			if next_path == package_path {
				break
			}
			package_path = next_path
		}
	}

	if p, err := context.Import(imp, "", build.AllowBinary|build.FindOnly); err == nil {
		try_autobuild(p)
		if file_exists(p.PkgObj) {
			log_found_package_maybe(imp, p.PkgObj)
			return p.PkgObj, true
		}
	}

	if *g_debug {
		log.Printf("Import path %q was not resolved\n", imp)
		log.Println("Gocode's build context is:")
		log_build_context(context)
	}
	return "", false
}

func package_name(file *ast.File) string {
	if file.Name != nil {
		return file.Name.Name
	}
	return ""
}

//-------------------------------------------------------------------------
// decl_cache
//
// Thread-safe collection of DeclFileCache entities.
//-------------------------------------------------------------------------

type package_lookup_context struct {
	build.Context
	BzlProjectRoot     string
	GBProjectRoot      string
	CurrentPackagePath string
}

// gopath returns the list of Go path directories.
func (ctxt *package_lookup_context) gopath() []string {
	var all []string
	for _, p := range filepath.SplitList(ctxt.GOPATH) {
		if p == "" || p == ctxt.GOROOT {
			// Empty paths are uninteresting.
			// If the path is the GOROOT, ignore it.
			// People sometimes set GOPATH=$GOROOT.
			// Do not get confused by this common mistake.
			continue
		}
		if strings.HasPrefix(p, "~") {
			// Path segments starting with ~ on Unix are almost always
			// users who have incorrectly quoted ~ while setting GOPATH,
			// preventing it from expanding to $HOME.
			// The situation is made more confusing by the fact that
			// bash allows quoted ~ in $PATH (most shells do not).
			// Do not get confused by this, and do not try to use the path.
			// It does not exist, and printing errors about it confuses
			// those users even more, because they think "sure ~ exists!".
			// The go command diagnoses this situation and prints a
			// useful error.
			// On Windows, ~ is used in short names, such as c:\progra~1
			// for c:\program files.
			continue
		}
		all = append(all, p)
	}
	return all
}

func (ctxt *package_lookup_context) pkg_dirs() (string, []string) {
	pkgdir := fmt.Sprintf("%s_%s", ctxt.GOOS, ctxt.GOARCH)

	var currentPackagePath string
	var all []string
	if ctxt.GOROOT != "" {
		dir := filepath.Join(ctxt.GOROOT, "pkg", pkgdir)
		if is_dir(dir) {
			all = append(all, dir)
		}
	}

	switch g_config.PackageLookupMode {
	case "go":
		currentPackagePath = ctxt.CurrentPackagePath
		for _, p := range ctxt.gopath() {
			dir := filepath.Join(p, "pkg", pkgdir)
			if is_dir(dir) {
				all = append(all, dir)
			}
			dir = filepath.Join(dir, currentPackagePath, "vendor")
			if is_dir(dir) {
				all = append(all, dir)
			}
		}
	case "gb":
		if ctxt.GBProjectRoot != "" {
			pkgdir := fmt.Sprintf("%s-%s", ctxt.GOOS, ctxt.GOARCH)
			if !is_dir(pkgdir) {
				pkgdir = fmt.Sprintf("%s-%s-race", ctxt.GOOS, ctxt.GOARCH)
			}
			dir := filepath.Join(ctxt.GBProjectRoot, "pkg", pkgdir)
			if is_dir(dir) {
				all = append(all, dir)
			}
		}
	case "bzl":
		// TODO: Support bazel mode
	}
	return currentPackagePath, all
}

type decl_cache struct {
	cache   map[string]*decl_file_cache
	context *package_lookup_context
	sync.Mutex
}

func new_decl_cache(context *package_lookup_context) *decl_cache {
	return &decl_cache{
		cache:   make(map[string]*decl_file_cache),
		context: context,
	}
}

func (c *decl_cache) get(filename string) *decl_file_cache {
	c.Lock()
	defer c.Unlock()

	f, ok := c.cache[filename]
	if !ok {
		f = new_decl_file_cache(filename, c.context)
		c.cache[filename] = f
	}
	return f
}

func (c *decl_cache) get_and_update(filename string) *decl_file_cache {
	f := c.get(filename)
	f.update()
	return f
}
