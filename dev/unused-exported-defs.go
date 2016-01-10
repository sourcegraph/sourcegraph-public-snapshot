// +build ignore

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/oracle"
)

var (
	skipFuncs = flag.Bool("skipfuncs", false, "skip funcs (which are slower to query for callers of)")
	matchStr  = flag.String("match", "", "only process defs whose names match this regexp")

	match *regexp.Regexp

	conf = loader.Config{
		TypeCheckFuncBodies: func(path string) bool { return false },
		Build:               &build.Default,
	}
)

func main() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", buildutil.TagsFlagDoc)
	flag.Parse()

	log.SetFlags(0)

	var err error
	match, err = regexp.Compile(*matchStr)
	if err != nil {
		log.Fatal(err)
	}

	rest, err := conf.FromArgs(flag.Args(), false)
	if err != nil {
		log.Fatal(err)
	}
	if len(rest) > 0 {
		log.Fatal("error: additional args provided")
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	for _, pkg := range prog.InitialPackages() {
		if err := do(pkg); err != nil {
			log.Fatalf("%s: %s", pkg.Pkg.Path(), err)
		}
	}
}

func do(pkg *loader.PackageInfo) error {
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			if err := doDecl(pkg, decl); err != nil {
				return err
			}
		}
	}
	return nil
}

func doDecl(pkg *loader.PackageInfo, decl ast.Decl) error {
	switch decl := decl.(type) {
	case *ast.FuncDecl:
		return doDef(pkg, decl, true, decl.Name.Name, decl.Name.Pos())

	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			switch spec := spec.(type) {
			case *ast.ValueSpec:
				for _, name := range spec.Names {
					if err := doDef(pkg, decl, false, name.Name, name.NamePos); err != nil {
						return err
					}
				}

			case *ast.TypeSpec:
				return doDef(pkg, decl, false, spec.Name.Name, spec.Name.NamePos)

			case *ast.ImportSpec:
				// do nothing

			default:
				return fmt.Errorf("unrecognized spec %T", spec)
			}
		}

	default:
		return fmt.Errorf("unrecognized %T", decl)
	}
	return nil
}

func doDef(pkg *loader.PackageInfo, decl ast.Decl, isCallable bool, name string, p token.Pos) error {
	if !ast.IsExported(name) {
		return nil
	}
	if *skipFuncs && isCallable {
		return nil
	}
	if !match.MatchString(name) {
		return nil
	}

	pos := conf.Fset.Position(p)

	//log.Printf(gray("%s: %s"), pkg.Pkg.Path(), name)

	var (
		allRefs []string
		extRefs []string
	)
	addRefs := func(refs ...string) {
		allRefs = append(allRefs, refs...)
		for _, ref := range refs {
			filename := ref[:strings.Index(ref, ":")]
			bpkg, err := buildutil.ContainingPackage(conf.Build, filepath.Dir(filename), filename)
			if err != nil {
				log.Fatal(err)
			}
			if bpkg.ImportPath != pkg.Pkg.Path() {
				extRefs = append(extRefs, ref)
			}
		}
	}

	q := oracle.Query{
		Mode:  "referrers",
		Pos:   fmt.Sprintf("%s:#%d", pos.Filename, pos.Offset),
		Build: conf.Build,
		Scope: scope,
		Fset:  conf.Fset,
	}
	if err := oracle.Run(&q); err != nil {
		return fmt.Errorf("%s at %s: %s", name, q.Pos, err)
	}
	desc := q.Serial().Referrers.Desc
	addRefs(q.Serial().Referrers.Refs...)

	if isCallable {
		q.Mode = "callers"
		if err := oracle.Run(&q); err != nil {
			return fmt.Errorf("%s at %s: %s", name, q.Pos, err)
		}
		for _, c := range q.Serial().Callers {
			addRefs(c.Pos)
		}
	}

	if len(extRefs) > 0 {
		log.Printf(gray("%s: %s"), pkg.Pkg.Path(), name)
		return nil
	}

	var err error
	pos.Filename, err = filepath.Rel(cwd, pos.Filename)
	if err != nil {
		return err
	}
	if len(allRefs) == 0 {
		log.Printf(red("%s: not used: %s"), pos, desc)
	} else {
		log.Printf(yellow("%s: not used externally: %s"), pos, desc)
	}
	return nil
}

var cwd string

func init() {
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func red(s string) string {
	return "\x1b[31m" + s + "\x1b[0m"
}

func yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[0m"
}

func gray(s string) string {
	return "\x1b[30m" + s + "\x1b[0m"
}

var scope = []string{
	"src.sourcegraph.com/sourcegraph",
	"src.sourcegraph.com/sourcegraph/Godeps",
	"src.sourcegraph.com/sourcegraph/app",
	"src.sourcegraph.com/sourcegraph/app/appconf",
	"src.sourcegraph.com/sourcegraph/app/assets",
	"src.sourcegraph.com/sourcegraph/app/auth",
	"src.sourcegraph.com/sourcegraph/app/cmd",
	"src.sourcegraph.com/sourcegraph/app/internal",
	"src.sourcegraph.com/sourcegraph/app/internal/apptest",
	"src.sourcegraph.com/sourcegraph/app/internal/authutil",
	"src.sourcegraph.com/sourcegraph/app/internal/canonicalurl",
	"src.sourcegraph.com/sourcegraph/app/internal/form",
	"src.sourcegraph.com/sourcegraph/app/internal/gzipfileserver",
	"src.sourcegraph.com/sourcegraph/app/internal/localauth",
	"src.sourcegraph.com/sourcegraph/app/internal/markdown",
	"src.sourcegraph.com/sourcegraph/app/internal/oauth2server",
	"src.sourcegraph.com/sourcegraph/app/internal/returnto",
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil",
	"src.sourcegraph.com/sourcegraph/app/internal/static",
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl",
	"src.sourcegraph.com/sourcegraph/app/router",
	"src.sourcegraph.com/sourcegraph/app/templates",
	"src.sourcegraph.com/sourcegraph/auth",
	"src.sourcegraph.com/sourcegraph/auth/accesstoken",
	"src.sourcegraph.com/sourcegraph/auth/authutil",
	"src.sourcegraph.com/sourcegraph/auth/idkey",
	"src.sourcegraph.com/sourcegraph/auth/idkeystore",
	"src.sourcegraph.com/sourcegraph/auth/ldap",
	"src.sourcegraph.com/sourcegraph/auth/sharedsecret",
	"src.sourcegraph.com/sourcegraph/auth/userauth",
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client",
	"src.sourcegraph.com/sourcegraph/cmd/src",
	"src.sourcegraph.com/sourcegraph/conf",
	"src.sourcegraph.com/sourcegraph/conf/feature",
	"src.sourcegraph.com/sourcegraph/dev",
	"src.sourcegraph.com/sourcegraph/dev/check-vendored-pkgs",
	"src.sourcegraph.com/sourcegraph/dev/git-wrapper",
	"src.sourcegraph.com/sourcegraph/dev/misc/network-test",
	"src.sourcegraph.com/sourcegraph/dev/release",
	"src.sourcegraph.com/sourcegraph/devdoc",
	"src.sourcegraph.com/sourcegraph/devdoc/assets",
	"src.sourcegraph.com/sourcegraph/devdoc/cmd/developer",
	"src.sourcegraph.com/sourcegraph/devdoc/tmpl",
	"src.sourcegraph.com/sourcegraph/doc",
	"src.sourcegraph.com/sourcegraph/emailaddrs",
	"src.sourcegraph.com/sourcegraph/env",
	"src.sourcegraph.com/sourcegraph/errcode",
	"src.sourcegraph.com/sourcegraph/events",
	"src.sourcegraph.com/sourcegraph/events/listeners",
	"src.sourcegraph.com/sourcegraph/ext",
	"src.sourcegraph.com/sourcegraph/ext/github",
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli",
	"src.sourcegraph.com/sourcegraph/ext/papertrail",
	"src.sourcegraph.com/sourcegraph/ext/slack",
	"src.sourcegraph.com/sourcegraph/extsvc/github",
	"src.sourcegraph.com/sourcegraph/fed",
	"src.sourcegraph.com/sourcegraph/fed/discover",
	"src.sourcegraph.com/sourcegraph/gen",
	"src.sourcegraph.com/sourcegraph/gitserver",
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb",
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb/mock",
	"src.sourcegraph.com/sourcegraph/gitserver/router",
	"src.sourcegraph.com/sourcegraph/gitserver/sshgit",
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/routevar",
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph",
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock",
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/spec",
	"src.sourcegraph.com/sourcegraph/gogeneratedeps",
	"src.sourcegraph.com/sourcegraph/httpapi",
	"src.sourcegraph.com/sourcegraph/httpapi/auth",
	"src.sourcegraph.com/sourcegraph/httpapi/router",
	"src.sourcegraph.com/sourcegraph/notif",
	"src.sourcegraph.com/sourcegraph/pkg/gitproto",
	"src.sourcegraph.com/sourcegraph/pkg/inventory",
	"src.sourcegraph.com/sourcegraph/pkg/inventory/filelang",
	"src.sourcegraph.com/sourcegraph/pkg/oauth2util",
	"src.sourcegraph.com/sourcegraph/pkg/sysreq",
	"src.sourcegraph.com/sourcegraph/pkg/wellknown",
	"src.sourcegraph.com/sourcegraph/platform",
	"src.sourcegraph.com/sourcegraph/platform/apps/changesets",
	"src.sourcegraph.com/sourcegraph/platform/apps/changesets/assets",
	"src.sourcegraph.com/sourcegraph/platform/apps/docs",
	"src.sourcegraph.com/sourcegraph/platform/apps/godoc",
	"src.sourcegraph.com/sourcegraph/platform/apps/godoc/godocsupport",
	"src.sourcegraph.com/sourcegraph/platform/notifications",
	"src.sourcegraph.com/sourcegraph/platform/pctx",
	"src.sourcegraph.com/sourcegraph/platform/putil",
	"src.sourcegraph.com/sourcegraph/platform/storage",
	"src.sourcegraph.com/sourcegraph/platform/storage/config",
	"src.sourcegraph.com/sourcegraph/repoupdater",
	"src.sourcegraph.com/sourcegraph/server",
	"src.sourcegraph.com/sourcegraph/server/accesscontrol",
	"src.sourcegraph.com/sourcegraph/server/cmd",
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/inner",
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/inner/auth",
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/inner/federated",
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/inner/trace",
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/outer",
	"src.sourcegraph.com/sourcegraph/server/internal/oauth2util",
	"src.sourcegraph.com/sourcegraph/server/internal/store",
	"src.sourcegraph.com/sourcegraph/server/internal/store/fs",
	"src.sourcegraph.com/sourcegraph/server/internal/store/pgsql",
	"src.sourcegraph.com/sourcegraph/server/internal/store/pgsql/dbtypes",
	"src.sourcegraph.com/sourcegraph/server/internal/store/shared/storageutil",
	"src.sourcegraph.com/sourcegraph/server/local",
	"src.sourcegraph.com/sourcegraph/server/local/cli",
	"src.sourcegraph.com/sourcegraph/server/serverctx",
	"src.sourcegraph.com/sourcegraph/server/testserver",
	"src.sourcegraph.com/sourcegraph/sgtool",
	"src.sourcegraph.com/sourcegraph/sgx",
	"src.sourcegraph.com/sourcegraph/sgx/buildvar",
	"src.sourcegraph.com/sourcegraph/sgx/cli",
	"src.sourcegraph.com/sourcegraph/sgx/client",
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd",
	"src.sourcegraph.com/sourcegraph/smoke",
	"src.sourcegraph.com/sourcegraph/smoke/basicgit",
	"src.sourcegraph.com/sourcegraph/sourcecode",
	"src.sourcegraph.com/sourcegraph/srclib_support",
	"src.sourcegraph.com/sourcegraph/store",
	"src.sourcegraph.com/sourcegraph/store/authzchecked",
	"src.sourcegraph.com/sourcegraph/store/cli",
	"src.sourcegraph.com/sourcegraph/store/mockstore",
	"src.sourcegraph.com/sourcegraph/store/testsuite",
	"src.sourcegraph.com/sourcegraph/svc",
	"src.sourcegraph.com/sourcegraph/svc/middleware/remote",
	"src.sourcegraph.com/sourcegraph/ui",
	"src.sourcegraph.com/sourcegraph/ui/payloads",
	"src.sourcegraph.com/sourcegraph/ui/router",
	"src.sourcegraph.com/sourcegraph/ui/uiconf",
	"src.sourcegraph.com/sourcegraph/usercontent",
	"src.sourcegraph.com/sourcegraph/util",
	"src.sourcegraph.com/sourcegraph/util/cacheutil",
	"src.sourcegraph.com/sourcegraph/util/dbutil",
	"src.sourcegraph.com/sourcegraph/util/dbutil2",
	"src.sourcegraph.com/sourcegraph/util/envutil",
	"src.sourcegraph.com/sourcegraph/util/executil",
	"src.sourcegraph.com/sourcegraph/util/expvarutil",
	"src.sourcegraph.com/sourcegraph/util/fileutil",
	"src.sourcegraph.com/sourcegraph/util/githubutil",
	"src.sourcegraph.com/sourcegraph/util/graphstoreutil",
	"src.sourcegraph.com/sourcegraph/util/handlerutil",
	"src.sourcegraph.com/sourcegraph/util/htmlutil",
	"src.sourcegraph.com/sourcegraph/util/httptestutil",
	"src.sourcegraph.com/sourcegraph/util/httputil",
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx",
	"src.sourcegraph.com/sourcegraph/util/jsonutil",
	"src.sourcegraph.com/sourcegraph/util/mdutil",
	"src.sourcegraph.com/sourcegraph/util/metricutil",
	"src.sourcegraph.com/sourcegraph/util/randstring",
	"src.sourcegraph.com/sourcegraph/util/router_util",
	"src.sourcegraph.com/sourcegraph/util/statsutil",
	"src.sourcegraph.com/sourcegraph/util/tempedit",
	"src.sourcegraph.com/sourcegraph/util/testdb",
	"src.sourcegraph.com/sourcegraph/util/testutil",
	"src.sourcegraph.com/sourcegraph/util/testutil/srclibtest",
	"src.sourcegraph.com/sourcegraph/util/textutil",
	"src.sourcegraph.com/sourcegraph/util/timeutil",
	"src.sourcegraph.com/sourcegraph/util/traceutil",
	"src.sourcegraph.com/sourcegraph/util/traceutil/appdashctx",
	"src.sourcegraph.com/sourcegraph/util/traceutil/cli",
	"src.sourcegraph.com/sourcegraph/util/vfsutil",
	"src.sourcegraph.com/sourcegraph/worker",
	"src.sourcegraph.com/sourcegraph/worker/builder",
	"src.sourcegraph.com/sourcegraph/worker/plan",
}
