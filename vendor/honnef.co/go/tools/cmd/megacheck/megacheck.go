// megacheck runs staticcheck, gosimple and unused.
package main // import "honnef.co/go/tools/cmd/megacheck"

import (
	"os"

	"honnef.co/go/tools/lint"
	"honnef.co/go/tools/lint/lintutil"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/unused"
)

type Checker struct {
	Checkers []lint.Checker
}

func (c *Checker) Init(prog *lint.Program) {
	for _, cc := range c.Checkers {
		cc.Init(prog)
	}
}

func (c *Checker) Funcs() map[string]lint.Func {
	fns := map[string]lint.Func{}
	for _, cc := range c.Checkers {
		for k, v := range cc.Funcs() {
			fns[k] = v
		}
	}
	return fns
}

func main() {
	var flags struct {
		staticcheck struct {
			enabled   bool
			generated bool
		}
		gosimple struct {
			enabled   bool
			generated bool
		}
		unused struct {
			enabled      bool
			constants    bool
			fields       bool
			functions    bool
			types        bool
			variables    bool
			debug        string
			wholeProgram bool
			reflection   bool
		}
	}
	fs := lintutil.FlagSet("megacheck")
	fs.BoolVar(&flags.gosimple.enabled,
		"simple.enabled", true, "Run gosimple")
	fs.BoolVar(&flags.gosimple.generated,
		"simple.generated", false, "Check generated code")

	fs.BoolVar(&flags.staticcheck.enabled,
		"staticcheck.enabled", true, "Run staticcheck")
	fs.BoolVar(&flags.staticcheck.generated,
		"staticcheck.generated", false, "Check generated code (only applies to a subset of checks)")

	fs.BoolVar(&flags.unused.enabled,
		"unused.enabled", true, "Run unused")
	fs.BoolVar(&flags.unused.constants,
		"unused.consts", true, "Report unused constants")
	fs.BoolVar(&flags.unused.fields,
		"unused.fields", true, "Report unused fields")
	fs.BoolVar(&flags.unused.functions,
		"unused.funcs", true, "Report unused functions and methods")
	fs.BoolVar(&flags.unused.types,
		"unused.types", true, "Report unused types")
	fs.BoolVar(&flags.unused.variables,
		"unused.vars", true, "Report unused variables")
	fs.BoolVar(&flags.unused.wholeProgram,
		"unused.exported", false, "Treat arguments as a program and report unused exported identifiers")
	fs.BoolVar(&flags.unused.reflection, "unused.reflect", true, "Consider identifiers as used when it's likely they'll be accessed via reflection")

	fs.Parse(os.Args[1:])

	c := &Checker{}

	if flags.staticcheck.enabled {
		sac := staticcheck.NewChecker()
		sac.CheckGenerated = flags.staticcheck.generated
		c.Checkers = append(c.Checkers, sac)
	}

	if flags.gosimple.enabled {
		sc := simple.NewChecker()
		sc.CheckGenerated = flags.gosimple.generated
		c.Checkers = append(c.Checkers, sc)
	}

	if flags.unused.enabled {
		var mode unused.CheckMode
		if flags.unused.constants {
			mode |= unused.CheckConstants
		}
		if flags.unused.fields {
			mode |= unused.CheckFields
		}
		if flags.unused.functions {
			mode |= unused.CheckFunctions
		}
		if flags.unused.types {
			mode |= unused.CheckTypes
		}
		if flags.unused.variables {
			mode |= unused.CheckVariables
		}
		uc := unused.NewChecker(mode)
		uc.WholeProgram = flags.unused.wholeProgram
		uc.ConsiderReflection = flags.unused.reflection
		c.Checkers = append(c.Checkers, unused.NewLintChecker(uc))
	}

	lintutil.ProcessFlagSet(c, fs)
}
