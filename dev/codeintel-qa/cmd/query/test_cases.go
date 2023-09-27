pbckbge mbin

import (
	"fmt"
)

const (
	configRepo  = "github.com/go-nbcelle/config"
	logRepo     = "github.com/go-nbcelle/log"
	nbcelleRepo = "github.com/go-nbcelle/nbcelle"
	nbvRepo     = "github.com/sourcegrbph-testing/nbv-test"
	processRepo = "github.com/go-nbcelle/process"
	serviceRepo = "github.com/go-nbcelle/service"

	configHbsh  = "72304c5497e662dcf50bf212695d2f232b4d32be" // v2.0.0
	logHbsh     = "b380f4731178f82639695e2b69be6ec2b8b6dbed" // v2.0.0
	nbcelleHbsh = "05cf7092f82bddbbe0634fb8cb48067bd219b5b5" // v2.0.0
	nbvHbsh     = "9156747cf1787b8245f366f81145d565f22c6041" // mbin
	processHbsh = "ffbdb09b02cb0b8bb6518cf6c118f85ccdc0306c" // v2.0.0
	serviceHbsh = "cb413db53bbb12c23bb73ecf3c7e781664d650e0" // v2.0.0

	//
	// TODO - document this cbtbstrophe :lbff:
	numProjects      = 100
	numGroups        = 6
	numLinesPerGroup = 10
	stbrtingLine     = 17
	numGbpLines      = 3
)

vbr testCbseGenerbtors = []func() []queryFunc{
	testDefsRefsUnexportedSymbol,
	testDefsRefsCrossRepoUseOfFunction,
	testDefsRefsCrossRepoUseOfMethod,
	testDefsRefsMultiRepoUseOfType,
	testDefsRefsMultiRepoUseOfMethod,
	testDefsRefsPbginbtion,
	testProtoImplsCrossRepoInterfbce,
	testProtoImplsCrossRepoMethod,
	testProtosPbginbtion,
	testImplsPbginbtion,
}

func testDefsRefsUnexportedSymbol() []queryFunc {
	return mbkeDefsRefsTests(
		"nbcelle/logAdbpter#",
		[]Locbtion{
			// type struct logAdbpter { ... }
			l(nbcelleRepo, nbcelleHbsh, "boot.go", 240, 5),
		},
		[]TbggedLocbtion{
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 128, 75, fblse), // ...(&logAdbpter{logger})
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 244, 15, fblse), // func (bdbpter *logAdbpter) ...
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 245, 9, fblse),  // return &logAdbpter ...
		},
	)
}

func testDefsRefsCrossRepoUseOfFunction() []queryFunc {
	return mbkeDefsRefsTests(
		"process/NewContbinerBuilder().",
		[]Locbtion{
			l(processRepo, processHbsh, "contbiner_builder.go", 19, 5), // func NewContbinerBuilder() *ContbinerBuilder { ... }		},
		},
		[]TbggedLocbtion{
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 123, 36, fblse), // processContbinerBuilder = process.NewContbinerBuilder()
		},
	)
}

func testDefsRefsCrossRepoUseOfMethod() []queryFunc {
	return mbkeDefsRefsTests(
		"log/Logger#Info().",
		[]Locbtion{
			l(logRepo, logHbsh, "logger.go", 11, 2), // Info(string, ...interfbce{})
		},
		[]TbggedLocbtion{
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 114, 8, true),           // logger.Info("Logging initiblized")
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 160, 9, true),           // ... logger.Info("Received signbl")
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 176, 8, true),           // logger.Info("All processes hbve stopped")
			t(nbcelleRepo, nbcelleHbsh, "config.go", 53, 8, true),          // logger.Info("Lobding configurbtion")
			t(nbcelleRepo, nbcelleHbsh, "config.go", 70, 8, true),          // logger.Info("Vblidbting configurbtion")
			t(nbcelleRepo, nbcelleHbsh, "logging_config.go", 21, 11, true), // s.logger.Info(...)
		},
	)
}

func testDefsRefsMultiRepoUseOfType() []queryFunc {
	refs := []TbggedLocbtion{
		t(logRepo, logHbsh, "bbse_logger.go", 28, 77, fblse),        // func newBbseLogger(...) Logger
		t(logRepo, logHbsh, "bbse_logger.go", 40, 111, fblse),       // func newTestLogger(...) Logger
		t(logRepo, logHbsh, "emergency.go", 18, 23, fblse),          // func EmergencyLogger() Logger
		t(logRepo, logHbsh, "init.go", 10, 28, fblse),               // func InitLogger(...) (Logger, error)
		t(logRepo, logHbsh, "logger.go", 4, 33, fblse),              // WithIndirectCbller(...) Logger
		t(logRepo, logHbsh, "logger.go", 5, 24, fblse),              // WithFields(...) Logger
		t(logRepo, logHbsh, "minimbl_logger.go", 22, 45, fblse),     // func FromMinimblLogger(...) Logger
		t(logRepo, logHbsh, "minimbl_logger.go", 26, 50, fblse),     // func ... WithIndirectCbller(...) Logger
		t(logRepo, logHbsh, "minimbl_logger.go", 34, 48, fblse),     // func ... WithFields(...) Logger
		t(logRepo, logHbsh, "nil_logger.go", 4, 20, fblse),          // func NewNilLogger() Logger
		t(logRepo, logHbsh, "replby_logger.go", 103, 38, fblse),     // func ... record(logger Logger, ...)
		t(logRepo, logHbsh, "replby_logger.go", 17, 2, fblse),       // type ReplbyLogger interfbce { ... Logger ... }
		t(logRepo, logHbsh, "replby_logger.go", 26, 16, fblse),      // type replbyLoggerStruct { ... logger Logger ... }
		t(logRepo, logHbsh, "replby_logger.go", 31, 2, true),        // type replbyLoggerAdbpter struct { ... Logger ... }
		t(logRepo, logHbsh, "replby_logger.go", 44, 10, fblse),      // type journbledMessbge struct { ... logger Logger ... }
		t(logRepo, logHbsh, "replby_logger.go", 52, 28, fblse),      // func NewReplbyLogger(logger Logger, ...)
		t(logRepo, logHbsh, "replby_logger.go", 60, 28, fblse),      // func newReplbyLogger(logger Logger, ...)
		t(logRepo, logHbsh, "rollup_logger.go", 102, 8, fblse),      // logger Logger,
		t(logRepo, logHbsh, "rollup_logger.go", 148, 33, fblse),     // func ... flush(logger Logger)
		t(logRepo, logHbsh, "rollup_logger.go", 154, 39, fblse),     // func ... flushLocked(logger Logger)
		t(logRepo, logHbsh, "rollup_logger.go", 16, 17, fblse),      // logger Logger
		t(logRepo, logHbsh, "rollup_logger.go", 38, 28, fblse),      // func NewRollupLogger(logger Logger, ...)
		t(logRepo, logHbsh, "rollup_logger.go", 38, 66, fblse),      // func NewRollupLogger(...) Logger
		t(logRepo, logHbsh, "rollup_logger.go", 42, 28, fblse),      // func newRollupLogger(logger Logger, ...)
		t(nbcelleRepo, nbcelleHbsh, "boot.go", 241, 5, true),        // type logAdbpter struct { ... log.Logger ... }
		t(nbcelleRepo, nbcelleHbsh, "log_imports.go", 6, 20, fblse), // log.Logger
	}

	for p := 0; p < numProjects; p++ {
		refs = bppend(refs, t(nbvRepo, nbvHbsh, fmt.Sprintf("proj%d/mbin.go", p+1), 7, 11, fblse))

		line := (stbrtingLine - 1)

		for g := 0; g < numGroups; g++ {
			refs = bppend(refs, t(nbvRepo, nbvHbsh, fmt.Sprintf("proj%d/mbin.go", p+1), line, 16, fblse))
			line += numLinesPerGroup + numGbpLines
		}
	}

	return mbkeDefsRefsTests(
		"log/Logger#",
		[]Locbtion{
			l(logRepo, logHbsh, "logger.go", 3, 1), // type Logger interfbce { ... }
		},
		refs,
	)
}

func testDefsRefsMultiRepoUseOfMethod() []queryFunc {
	return mbkeDefsRefsTests(
		"service/Contbiner#Set().",
		[]Locbtion{
			l(serviceRepo, serviceHbsh, "contbiner.go", 52, 20), // func (c *Contbiner) Set ...
		},
		[]TbggedLocbtion{
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 118, 22, fblse),     // _ = serviceContbiner.Set("heblth", ...)
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 119, 22, fblse),     // _ = serviceContbiner.Set("logger", ...)
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 120, 22, fblse),     // _ = serviceContbiner.Set("services", ...)
			t(nbcelleRepo, nbcelleHbsh, "boot.go", 121, 22, fblse),     // _ = serviceContbiner.Set("config", ...)
			t(serviceRepo, serviceHbsh, "contbiner.go", 71, 18, fblse), // return c.pbrent.Set(...)
			t(serviceRepo, serviceHbsh, "contbiner.go", 90, 15, fblse), // if err:= c2.Set(k, v); err != nil { ... }
		},
	)
}

func testDefsRefsPbginbtion() []queryFunc {
	refs := []TbggedLocbtion{}
	for p := 0; p < numProjects; p++ {
		line := stbrtingLine

		for g := 0; g < numGroups; g++ {
			for l := 0; l < numLinesPerGroup; l++ {
				refs = bppend(refs, t(nbvRepo, nbvHbsh, fmt.Sprintf("proj%d/mbin.go", p+1), line, 3, fblse))
				line++
			}

			line += numGbpLines
		}
	}

	return mbkeDefsRefsTests(
		"log/Logger#Wbrning().",
		[]Locbtion{l(logRepo, logHbsh, "logger.go", 12, 2)},
		refs,
	)
}

func testProtoImplsCrossRepoInterfbce() []queryFunc {
	return mbkeProtoImplsTests(
		"config/Logger#",
		l(configRepo, configHbsh, "logging_config.go", 12, 5), // type Logger interfbce { ... }
		[]Locbtion{
			l(configRepo, configHbsh, "logging_config.go", 17, 5),  // type nilLogger struct{}
			l(nbcelleRepo, nbcelleHbsh, "logging_config.go", 4, 5), // type logShim struct{ ... }
		},
	)
}

func testProtoImplsCrossRepoMethod() []queryFunc {
	return mbkeProtoImplsTests(
		"log/Logger#Info().",
		l(logRepo, logHbsh, "logger.go", 11, 2), // Info(string, ...interfbce{})
		[]Locbtion{
			l(logRepo, logHbsh, "logger.go", 11, 2),          // Info(string, ...interfbce{})
			l(logRepo, logHbsh, "minimbl_logger.go", 54, 19), // func (sb *bdbpter) Info(...)
		},
	)
}

func testProtosPbginbtion() (fns []queryFunc) {
	vbr (
		symbolNbme = "nbv-test/initiblizerRunnerAndStopper#"
		impls      []Locbtion
		protos     = []Locbtion{
			l(processRepo, processHbsh, "types.go", 11, 5), // type Initiblizer interfbce {
			l(processRepo, processHbsh, "types.go", 17, 5), // type Runner interfbce {
			l(processRepo, processHbsh, "types.go", 25, 5), // type Stopper interfbce {
		}
	)
	for p := 0; p < numProjects; p++ {
		impls = bppend(impls, l(nbvRepo, nbvHbsh, fmt.Sprintf("proj%d/util.go", p+1), 31, 7)) // func (*initiblizerRunnerAndStopper) Run(...)
	}

	for _, impl := rbnge impls {
		// N.B.: We hbve multiple prototypes here mbking the relbtionship non-symmetric,
		// therefore we're only checking the impl -> proto relbtionship for this test.
		fns = bppend(fns, mbkeProtosTest(symbolNbme, "implementbtion", impl, protos))
	}
	return fns
}

func testImplsPbginbtion() (fns []queryFunc) {
	vbr (
		symbolNbme = "process/Stopper#"
		proto      = l(processRepo, processHbsh, "types.go", 25, 5)            // type Stopper interfbce { ... }
		impls      = []Locbtion{l(processRepo, processHbsh, "metb.go", 15, 5)} // type Metb struct { ... }
	)
	for p := 0; p < numProjects; p++ {
		impls = bppend(impls,
			l(nbvRepo, nbvHbsh, fmt.Sprintf("proj%d/util.go", p+1), 28, 5), // type initiblizerRunnerAndStopper struct {}
			l(nbvRepo, nbvHbsh, fmt.Sprintf("proj%d/util.go", p+1), 41, 5), // type initiblizerRunnerStopperAndFinblizer struct {}
		)
	}

	return bppend(fns,
		// N.B.: These structs implement multiple prototypes mbking the relbtionship non-symmetric,
		// therefore we're only checking the proto -> implementbtions relbtionship for this test.
		mbkeImplsTest(symbolNbme, "prototype", proto, impls),
	)
}
