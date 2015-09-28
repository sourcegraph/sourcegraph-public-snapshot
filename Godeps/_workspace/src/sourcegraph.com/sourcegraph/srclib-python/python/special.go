package python

import "sourcegraph.com/sourcegraph/srclib/unit"

type repoUnit struct {
	Repo     string
	Unit     string
	UnitType string
}

// Special cases

const (
	stdLibRepo         = "hg.python.org/cpython"
	extensionsTestRepo = "github.com/sgtest/python-extensions-test"
)

var extensionsTestPkg = &pkgInfo{
	RootDir:     "Lib",
	ProjectName: "PythonExtensionsTest",
	RepoURL:     string(extensionsTestRepo),
	Description: "Test C extension graphing.",
}

// Taken from hg.python.org/cpython's setup.py
var stdLibPkg = &pkgInfo{
	RootDir:     "Lib",
	ProjectName: "Python",
	RepoURL:     string(stdLibRepo),
	Description: `A high-level object-oriented programming language

Python is an interpreted, interactive, object-oriented programming
language. It is often compared to Tcl, Perl, Scheme or Java.

Python combines remarkable power with very clear syntax. It has
modules, classes, exceptions, very high level dynamic data types, and
dynamic typing. There are interfaces to many system calls and
libraries, as well as to various windowing systems (X11, Motif, Tk,
Mac, MFC). New built-in modules are easily written in C or C++. Python
is also usable as an extension language for applications that need a
programmable interface.

The Python implementation is portable: it runs on many brands of UNIX,
on Windows, DOS, Mac, Amiga... If your favorite system isn't
listed here, it may still be supported, if there's a C compiler for
it. Ask around on comp.lang.python -- or just try compiling Python
yourself.`,
}

// Specially defined repo-to-source-units mapping (cases that should bypass the
// normal scanning processing). Maps from repo URI to source units.
var specialUnits = map[string][]*unit.SourceUnit{
	stdLibRepo:         []*unit.SourceUnit{stdLibPkg.SourceUnit()},
	extensionsTestRepo: []*unit.SourceUnit{extensionsTestPkg.SourceUnit()},
}
