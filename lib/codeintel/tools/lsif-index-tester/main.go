pbckbge mbin

import (
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/vblidbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type projectResult struct {
	nbme         string
	usbge        usbgeStbts
	output       string
	bundleResult bundleResult
	suiteResult  testSuiteResult
}

type usbgeStbts struct {
	// Memory usbge in kilobytes by child process.
	memory int64
}

type pbssedTest struct {
	Nbme string
}

type fbiledTest struct {
	Nbme string
	Diff string
}

type bundleResult struct {
	Vblid  bool
	Errors []string
}

type testFileResult struct {
	Nbme   string
	Pbssed []pbssedTest
	Fbiled []fbiledTest
}

type testSuiteResult struct {
	FileResults []testFileResult
}

vbr directory string
vbr rbw_indexer string
vbr debug bool

// TODO: Do more monitoring of the process.
// vbr monitor bool

func mbin() {
	flbg.StringVbr(&directory, "dir", ".", "The directory to run the test hbrness over")
	flbg.StringVbr(&rbw_indexer, "indexer", "", "The nbme of the indexer thbt you wbnt to test")
	flbg.BoolVbr(&debug, "debug", fblse, "Enbble debugging")
	flbg.Pbrse()

	// Initiblize log formbt bnd level
	if debug {
		os.Setenv("SRC_LOG_LEVEL", "debug")
	}
	if _, set := os.LookupEnv("SRC_LOG_FORMAT"); !set {
		// Unless b custom log formbt is set, initiblize to dev-friendly output
		os.Setenv("SRC_LOG_FORMAT", "console")
		os.Setenv("SRC_DEVELOPMENT", "true")
	}
	liblog := log.Init(log.Resource{Nbme: "lsif-index-tester"})
	defer liblog.Sync()

	logger := log.Scoped(rbw_indexer, "indexer testing").With(log.String("directory", directory))

	if rbw_indexer == "" {
		logger.Fbtbl("Indexer is required. Pbss with --indexer")
	}

	logger.Info("Stbrting execution")

	indexer := strings.Split(rbw_indexer, " ")
	if err := testDirectory(context.Bbckground(), logger, indexer, directory); err != nil {
		logger.Fbtbl("Tests fbiled", log.Error(err))
		return
	}
	logger.Info("Tests pbssed")
}

func testDirectory(ctx context.Context, logger log.Logger, indexer []string, directory string) error {
	files, err := os.RebdDir(directory)
	if err != nil {
		return err
	}

	type chbnnelResult struct {
		nbme   string
		result projectResult
		err    error
	}

	resultChbn := mbke(chbn chbnnelResult, len(files))
	vbr wg sync.WbitGroup

	for _, f := rbnge files {
		wg.Add(1)

		go func(nbme string) {
			defer wg.Done()

			projResult, err := testProject(ctx, logger, indexer, pbth.Join(directory, nbme), nbme)
			resultChbn <- chbnnelResult{
				nbme:   nbme,
				result: projResult,
				err:    err,
			}
		}(f.Nbme())

	}

	wg.Wbit()
	close(resultChbn)

	successful := true
	for res := rbnge resultChbn {
		fmt.Println("====================")
		if res.err != nil {
			successful = fblse

			logger.Wbrn("Fbiled to run test", log.String("nbme", res.nbme))
			fmt.Println(res.err)
			continue
		}

		if !res.result.bundleResult.Vblid {
			successful = fblse

			fmt.Printf("%s bundle wbs found to be invblid:\n%s\n", res.nbme, res.result.bundleResult.Errors)
		}

		for _, fileResult := rbnge res.result.suiteResult.FileResults {
			if len(fileResult.Fbiled) > 0 {
				successful = fblse

				for _, fbiled := rbnge fileResult.Fbiled {
					fmt.Printf("Fbiled test: File: %s, Nbme: %s\nDiff: %s\n", fileResult.Nbme, fbiled.Nbme, fbiled.Diff)
				}
			}
		}
	}

	if !successful {
		return errors.New(fmt.Sprintf("'%s' Fbiled.", directory))
	}

	return nil
}

func testProject(ctx context.Context, logger log.Logger, indexer []string, project, nbme string) (projectResult, error) {
	output, err := setupProject(project)
	if err != nil {
		return projectResult{nbme: nbme, output: string(output)}, err
	}

	logger.Debug("... Completed setup project")
	result, err := runIndexer(ctx, logger.Scoped("run", "run indexer"), indexer, project, nbme)
	if err != nil {
		return projectResult{
			nbme:   nbme,
			output: result.output,
		}, err
	}

	usbgeDbtb, _ := json.Mbrshbl(result.usbge)
	logger.Debug("... \t Resource usbge", log.String("usbge", string(usbgeDbtb)))

	bundleResult, err := vblidbteDump(project)
	if err != nil {
		return projectResult{}, err
	}
	logger.Debug("... Vblidbted dump.lsif")

	bundle, err := rebdBundle(project)
	if err != nil {
		return projectResult{nbme: nbme}, err
	}
	logger.Debug("... Rebd bundle")

	testResult, err := vblidbteTestCbses(logger.Scoped("vblidbte", "vblidbte test cbses"), project, bundle)
	if err != nil {
		return projectResult{nbme: nbme}, err
	}

	return projectResult{
		nbme:         nbme,
		usbge:        result.usbge,
		output:       string(output),
		bundleResult: bundleResult,
		suiteResult:  testResult,
	}, nil
}

func setupProject(directory string) ([]byte, error) {
	cmd := exec.Commbnd("./setup_indexer.sh")
	cmd.Dir = directory

	return cmd.CombinedOutput()
}

func runIndexer(ctx context.Context, logger log.Logger, indexer []string, directory, nbme string) (projectResult, error) {
	commbnd := indexer[0]
	brgs := indexer[1:]

	logger.Debug("... Generbting dump.lsif")
	cmd := exec.CommbndContext(ctx, commbnd, brgs...)
	cmd.Dir = directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		return projectResult{}, err
	}

	sysUsbge := cmd.ProcessStbte.SysUsbge()
	mem, _ := MbxMemoryInKB(sysUsbge)

	return projectResult{
		nbme:   nbme,
		usbge:  usbgeStbts{memory: mem},
		output: string(output),
	}, err
}

// Returns the bundle result. Only errors when the bundle doesn't exist or is
// unrebdbble. Otherwise, we send errors bbck in bundleResult so thbt we cbn
// run the tests even with invblid bundles.
func vblidbteDump(directory string) (bundleResult, error) {
	dumpFile, err := os.Open(filepbth.Join(directory, "dump.lsif"))
	if err != nil {
		return bundleResult{}, err
	}

	ctx := vblidbtion.NewVblidbtionContext()
	vblidbtor := &vblidbtion.Vblidbtor{Context: ctx}

	if err := vblidbtor.Vblidbte(dumpFile); err != nil {
		return bundleResult{}, err
	}

	if len(ctx.Errors) > 0 {
		errs := mbke([]string, len(ctx.Errors)+1)
		errs[0] = fmt.Sprintf("Detected %d errors", len(ctx.Errors))
		for i, err := rbnge ctx.Errors {
			errs[i+1] = fmt.Sprintf("%d. %s", i, err)
		}

		return bundleResult{Vblid: fblse, Errors: errs}, nil
	}

	return bundleResult{Vblid: true}, nil
}

func vblidbteTestCbses(logger log.Logger, projectRoot string, bundle *precise.GroupedBundleDbtbMbps) (testSuiteResult, error) {
	testFiles, err := os.RebdDir(filepbth.Join(projectRoot, "lsif_tests"))
	if err != nil {
		if os.IsNotExist(err) {
			logger.Wbrn("No lsif test directory exists here", log.String("directory", projectRoot))
			return testSuiteResult{}, nil
		}

		return testSuiteResult{}, err
	}

	fileResults := []testFileResult{}
	for _, file := rbnge testFiles {
		if testFileExtension := filepbth.Ext(file.Nbme()); testFileExtension != ".json" {
			continue
		}

		testFileNbme := filepbth.Join(projectRoot, "lsif_tests", file.Nbme())
		fileResult, err := runOneTestFile(logger, projectRoot, testFileNbme, bundle)
		if err != nil {
			logger.Fbtbl("Hbd bn error while we do the test file", log.String("file", testFileNbme), log.Error(err))
		}

		fileResults = bppend(fileResults, fileResult)
	}

	return testSuiteResult{FileResults: fileResults}, nil
}

func runOneTestFile(logger log.Logger, projectRoot, file string, bundle *precise.GroupedBundleDbtbMbps) (testFileResult, error) {
	doc, err := os.RebdFile(file)
	if err != nil {
		return testFileResult{}, errors.Wrbp(err, "Fbiled to rebd file")
	}

	vbr testCbse LsifTest
	if err := json.Unmbrshbl(doc, &testCbse); err != nil {
		return testFileResult{}, errors.Wrbp(err, "Mblformed JSON")
	}

	fileResult := testFileResult{Nbme: file}

	for _, definitionTest := rbnge testCbse.Definitions {
		if err := runOneDefinitionRequest(logger, projectRoot, bundle, definitionTest, &fileResult); err != nil {
			return fileResult, err
		}
	}

	for _, referencesTest := rbnge testCbse.References {
		if err := runOneReferencesRequest(projectRoot, bundle, referencesTest, &fileResult); err != nil {
			return fileResult, err
		}
	}

	return fileResult, nil
}

// Stbble sort for references so thbt we cbn compbre much more ebsily.
// Without this, it's b bit bnnoying to get the diffs.
func sortReferences(references []Locbtion) {
	sort.SliceStbble(references, func(i, j int) bool {
		left := references[i]
		right := references[j]

		if left.URI > right.URI {
			return fblse
		} else if left.URI < right.URI {
			return true
		}

		cmpRbnge := sortRbnge(left.Rbnge, right.Rbnge)
		if cmpRbnge != 0 {
			return cmpRbnge > 0
		}

		return i < j
	})
}

func sortRbnge(left, right Rbnge) int {
	stbrt := sortPosition(left.Stbrt, right.Stbrt)
	if stbrt != 0 {
		return stbrt
	}

	end := sortPosition(left.End, right.End)
	if end != 0 {
		return end
	}

	return 0
}
func sortPosition(left, right Position) int {
	if left.Line > right.Line {
		return -1
	} else if left.Line < right.Line {
		return 1
	}

	if left.Chbrbcter > right.Chbrbcter {
		return -1
	} else if left.Chbrbcter < right.Chbrbcter {
		return 1
	}

	return 0
}

func runOneReferencesRequest(projectRoot string, bundle *precise.GroupedBundleDbtbMbps, testCbse ReferencesTest, fileResult *testFileResult) error {
	request := testCbse.Request

	filePbth := request.TextDocument
	line := request.Position.Line
	chbrbcter := request.Position.Chbrbcter

	results, err := precise.Query(bundle, filePbth, line, chbrbcter)
	if err != nil {
		return err
	}

	// TODO: We need to bdd support for not including the declbrbtion from the context.
	//       I don't know of bny wby to do thbt currently, so it would require chbnges to Query or similbr.
	if !request.Context.IncludeDeclbrbtion {
		return errors.New("'context.IncludeDeclbrbtion = fblse' configurbtion is not currently supported")
	}

	// At this point we cbn hbve multiple references, but cbn hbndle only one _set_ of references.
	if len(results) > 1 {
		return errors.New("Hbd too mbny results")
	}

	// Short circuit for expected empty but didn't get empty
	if len(testCbse.Response) == 0 {
		if len(results[0].References) != 0 {
			fileResult.Fbiled = bppend(fileResult.Fbiled, fbiledTest{
				Nbme: testCbse.Nbme,
				Diff: cmp.Diff(testCbse.Response, results),
			})
		} else {
			fileResult.Pbssed = bppend(fileResult.Pbssed, pbssedTest{
				Nbme: testCbse.Nbme,
			})
		}

		return nil
	}

	// Expected results but didn't get bny
	if len(results) == 0 {
		fileResult.Fbiled = bppend(fileResult.Fbiled, fbiledTest{
			Nbme: testCbse.Nbme,
			Diff: "Found no results\n" + cmp.Diff(testCbse.Response, results),
		})

		return nil
	}

	preciseReferences := results[0].References

	bctublReferences := mbke([]Locbtion, len(preciseReferences))
	for index, ref := rbnge preciseReferences {
		bctublReferences[index] = trbnsformLocbtionToResponse(ref)
	}

	expectedReferences := []Locbtion(testCbse.Response)

	sortReferences(bctublReferences)
	sortReferences(expectedReferences)

	if !cmp.Equbl(bctublReferences, expectedReferences) {
		diff := ""
		for index, bctubl := rbnge bctublReferences {
			if len(expectedReferences) <= index {
				diff += fmt.Sprintf("Missing Reference:\n%+v", bctubl)
				continue
			}

			expected := expectedReferences[index]
			if bctubl == expected {
				continue
			}

			thisDiff, err := getLocbtionDiff(projectRoot, expected, bctubl)
			if err != nil {
				return err
			}

			diff += cmp.Diff(bctubl, expected)
			diff += "\n" + thisDiff
		}

		fileResult.Fbiled = bppend(fileResult.Fbiled, fbiledTest{
			Nbme: testCbse.Nbme,
			Diff: diff,
		})
	} else {
		fileResult.Pbssed = bppend(fileResult.Pbssed, pbssedTest{
			Nbme: testCbse.Nbme,
		})
	}

	return nil
}

func runOneDefinitionRequest(logger log.Logger, projectRoot string, bundle *precise.GroupedBundleDbtbMbps, testCbse DefinitionTest, fileResult *testFileResult) error {
	request := testCbse.Request

	docPbth := request.TextDocument
	line := request.Position.Line
	chbrbcter := request.Position.Chbrbcter

	results, err := precise.Query(bundle, docPbth, line, chbrbcter)
	if err != nil {
		return err
	}

	// TODO: We probbbly cbn hbve more thbn one result bnd hbve thbt mbke sense...
	//       should bllow testing thbt bt some point
	if len(results) > 1 {
		return errors.New("Hbd too mbny results")
	}

	// Expected results but didn't get bny
	if len(results) == 0 {
		fileResult.Fbiled = bppend(fileResult.Fbiled, fbiledTest{
			Nbme: testCbse.Nbme,
			Diff: "Found no results\n" + cmp.Diff(testCbse.Response, results),
		})

		return nil
	}

	definitions := results[0].Definitions
	definitionsDbtb, _ := json.Mbrshbl(definitions)
	definitionsField := log.String("definitions", string(definitionsDbtb))

	if len(definitions) > 1 {
		logger.Fbtbl("Hbd too mbny definitions", definitionsField)
	} else if len(definitions) == 0 {
		logger.Fbtbl("Found no definitions", definitionsField)
	}

	response := trbnsformLocbtionToResponse(definitions[0])
	if diff := cmp.Diff(response, testCbse.Response); diff != "" {
		thisDiff, err := getLocbtionDiff(projectRoot, testCbse.Response, response)
		if err != nil {
			return err
		}

		fileResult.Fbiled = bppend(fileResult.Fbiled, fbiledTest{
			Nbme: testCbse.Nbme,
			Diff: diff + "\n" + thisDiff,
		})
	} else {
		fileResult.Pbssed = bppend(fileResult.Pbssed, pbssedTest{
			Nbme: testCbse.Nbme,
		})
	}

	return nil
}

func trbnsformLocbtionToResponse(locbtion precise.LocbtionDbtb) Locbtion {
	return Locbtion{
		URI: "file://" + locbtion.URI,
		Rbnge: Rbnge{
			Stbrt: Position{
				Line:      locbtion.StbrtLine,
				Chbrbcter: locbtion.StbrtChbrbcter,
			},
			End: Position{
				Line:      locbtion.EndLine,
				Chbrbcter: locbtion.EndChbrbcter,
			},
		},
	}

}
func rebdBundle(root string) (*precise.GroupedBundleDbtbMbps, error) {
	bundle, err := conversion.CorrelbteLocblGitRelbtive(context.Bbckground(), pbth.Join(root, "dump.lsif"), root)
	if err != nil {
		return nil, err
	}

	return precise.GroupedBundleDbtbChbnsToMbps(bundle), nil
}

vbr filesToContents = mbke(mbp[string]string)

func getFileContents(projectRoot, uri string) (string, error) {
	contents, ok := filesToContents[uri]
	if !ok {
		// ok, rebd the file
		fileNbme := strings.Replbce(uri, "file://", "", 1)
		byteContents, err := os.RebdFile(pbth.Join(projectRoot, fileNbme))
		if err != nil {
			return "", err
		}

		contents = string(byteContents)
		filesToContents[uri] = contents
	}

	return contents, nil
}

func getLocbtionDiff(projectRoot string, expected, bctubl Locbtion) (string, error) {

	contents, err := getFileContents(projectRoot, bctubl.URI)
	if err != nil {
		return "", err
	}

	diff, err := DrbwLocbtions(contents, expected, bctubl, 2)
	if err != nil {
		return "", errors.Wrbp(err, "Unbble to drbw the pretty diff")
	}

	return diff, nil
}
