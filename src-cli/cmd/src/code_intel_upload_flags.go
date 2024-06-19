package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/scip/bindings/go/scip"
	libscip "github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/scip"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/upload"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/codeintel"
)

var codeintelUploadFlags struct {
	file string

	// UploadRecordOptions
	repo              string
	commit            string
	root              string
	indexer           string
	indexerVersion    string
	associatedIndexID int

	// SourcegraphInstanceOptions
	uploadRoute      string
	maxPayloadSizeMb int64
	maxConcurrency   int

	// Codehost authorization secrets
	gitHubToken string
	gitLabToken string

	// Output and error behavior
	ignoreUploadFailures bool
	noProgress           bool
	verbosity            int
	json                 bool
	open                 bool
	apiFlags             *api.Flags
}

var (
	codeintelUploadFlagSet = flag.NewFlagSet("upload", flag.ExitOnError)
	apiClientFlagSet       = flag.NewFlagSet("upload client", flag.ExitOnError)
	// Used to include the insecure-skip-verify flag in the help output, as we don't use any of the
	// other api.Client methods, so only the insecureSkipVerify flag is relevant here.
	dummyflag bool

	// Used to skip the LSIF -> SCIP conversion during the migration. Not expected to be used outside
	// of codeintel-qa pipelines and not expected to last much longer than a few releases while we
	// deprecate all LSIF code paths.
	skipConversionToSCIP bool
)

func init() {
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.file, "file", "", `The path to the LSIF dump file.`)

	// UploadRecordOptions
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.repo, "repo", "", `The name of the repository (e.g. github.com/gorilla/mux). By default, derived from the origin remote.`)
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.commit, "commit", "", `The 40-character hash of the commit. Defaults to the currently checked-out commit.`)
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.root, "root", "", `The path in the repository that matches the LSIF projectRoot (e.g. cmd/project1). Defaults to the directory where the dump file is located.`)
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.indexer, "indexer", "", `The name of the indexer that generated the dump. This will override the 'toolInfo.name' field in the metadata vertex of the LSIF dump file. This must be supplied if the indexer does not set this field (in which case the upload will fail with an explicit message).`)
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.indexerVersion, "indexerVersion", "", `The version of the indexer that generated the dump. This will override the 'toolInfo.version' field in the metadata vertex of the LSIF dump file. This must be supplied if the indexer does not set this field (in which case the upload will fail with an explicit message).`)
	codeintelUploadFlagSet.IntVar(&codeintelUploadFlags.associatedIndexID, "associated-index-id", -1, "ID of the associated index record for this upload. For internal use only.")

	// SourcegraphInstanceOptions
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.uploadRoute, "upload-route", "/.api/lsif/upload", "The path of the upload route. For internal use only.")
	codeintelUploadFlagSet.Int64Var(&codeintelUploadFlags.maxPayloadSizeMb, "max-payload-size", 100, `The maximum upload size (in megabytes). Indexes exceeding this limit will be uploaded over multiple HTTP requests.`)
	codeintelUploadFlagSet.IntVar(&codeintelUploadFlags.maxConcurrency, "max-concurrency", -1, "The maximum number of concurrent uploads. Only relevant for multipart uploads. Defaults to all parts concurrently.")

	// Codehost authorization secrets
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.gitHubToken, "github-token", "", `A GitHub access token with 'public_repo' scope that Sourcegraph uses to verify you have access to the repository.`)
	codeintelUploadFlagSet.StringVar(&codeintelUploadFlags.gitLabToken, "gitlab-token", "", `A GitLab access token with 'read_api' scope that Sourcegraph uses to verify you have access to the repository.`)

	// Output and error behavior
	codeintelUploadFlagSet.BoolVar(&codeintelUploadFlags.ignoreUploadFailures, "ignore-upload-failure", false, `Exit with status code zero on upload failure.`)
	codeintelUploadFlagSet.BoolVar(&codeintelUploadFlags.noProgress, "no-progress", false, `Do not display progress updates.`)
	codeintelUploadFlagSet.IntVar(&codeintelUploadFlags.verbosity, "trace", 0, "-trace=0 shows no logs; -trace=1 shows requests and response metadata; -trace=2 shows headers, -trace=3 shows response body")
	codeintelUploadFlagSet.BoolVar(&codeintelUploadFlags.json, "json", false, `Output relevant state in JSON on success.`)
	codeintelUploadFlagSet.BoolVar(&codeintelUploadFlags.open, "open", false, `Open the LSIF upload page in your browser.`)
	codeintelUploadFlagSet.BoolVar(&dummyflag, "insecure-skip-verify", false, "Skip validation of TLS certificates against trusted chains")

	// Testing flags
	codeintelUploadFlagSet.BoolVar(&skipConversionToSCIP, "skip-scip", false, "Skip converting LSIF index to SCIP if the instance supports it; this option should only used for debugging")
}

// parseAndValidateCodeIntelUploadFlags calls codeintelUploadFlagset.Parse, then infers values for
// missing flags, normalizes supplied values, and validates the state of the codeintelUploadFlags
// object.
//
// On success, the global codeintelUploadFlags object will be populated with valid values. An
// error is returned on failure.
func parseAndValidateCodeIntelUploadFlags(args []string) (*output.Output, bool, error) {
	if err := codeintelUploadFlagSet.Parse(args); err != nil {
		return nil, false, err
	}

	out := codeintelUploadOutput()

	// extract only the -insecure-skip-verify flag so we dont get 'flag provided but not defined'
	var insecureSkipVerifyFlag []string
	for _, s := range args {
		if strings.HasPrefix(s, "-insecure-skip-verify") {
			insecureSkipVerifyFlag = append(insecureSkipVerifyFlag, s)
		}
	}

	// parse the api client flags separately and then populate the codeintelUploadFlags struct with the result
	// we could just use insecureSkipVerify but I'm including everything here because it costs nothing
	// and maybe we'll use some in the future
	codeintelUploadFlags.apiFlags = api.NewFlags(apiClientFlagSet)
	if err := apiClientFlagSet.Parse(insecureSkipVerifyFlag); err != nil {
		return nil, false, err
	}

	if !isFlagSet(codeintelUploadFlagSet, "file") {
		defaultFile, err := inferDefaultFile()
		if err != nil {
			return nil, false, err
		}
		codeintelUploadFlags.file = defaultFile
	}

	// Check to see if input file exists
	if _, err := os.Stat(codeintelUploadFlags.file); os.IsNotExist(err) {
		if !isFlagSet(codeintelUploadFlagSet, "file") {
			return nil, false, formatInferenceError(argumentInferenceError{"file", err})
		}

		return nil, false, errors.Newf("file %q does not exist", codeintelUploadFlags.file)
	}

	isSCIPAvailable, err := isSCIPAvailable()
	if err != nil {
		return nil, false, err
	}

	if !isSCIPAvailable {
		if err := handleSCIP(out); err != nil {
			return nil, false, err
		}
	} else {
		if err := handleLSIF(out); err != nil {
			return nil, false, err
		}
	}

	// Check for new file existence after transformation
	if _, err := os.Stat(codeintelUploadFlags.file); os.IsNotExist(err) {
		return nil, false, errors.Newf("file %q does not exist", codeintelUploadFlags.file)
	}

	// Infer the remaining default arguments (may require reading from new file)
	if inferenceErrors := inferMissingCodeIntelUploadFlags(); len(inferenceErrors) > 0 {
		return nil, false, formatInferenceError(inferenceErrors[0])
	}

	if err := validateCodeIntelUploadFlags(); err != nil {
		return nil, false, err
	}

	return out, isSCIPAvailable, nil
}

// codeintelUploadOutput returns an output object that should be used to print the progres
// of requests made during this upload. If -json, -no-progress, or -trace>0 is given,
// then no output object is defined.
//
// For -no-progress and -trace>0 conditions, emergency loggers will be used to display
// inferred arguments and the URL at which processing status is shown.
func codeintelUploadOutput() (out *output.Output) {
	if codeintelUploadFlags.json || codeintelUploadFlags.noProgress || codeintelUploadFlags.verbosity > 0 {
		return nil
	}

	return output.NewOutput(flag.CommandLine.Output(), output.OutputOpts{
		Verbose: true,
	})
}

func isSCIPAvailable() (bool, error) {
	client := cfg.apiClient(codeintelUploadFlags.apiFlags, codeintelUploadFlagSet.Output())
	req, err := client.NewHTTPRequest(context.Background(), "HEAD", strings.ReplaceAll(codeintelUploadFlags.uploadRoute, "lsif", "scip"), nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return false, upload.ErrUnauthorized
	}

	return resp.StatusCode == http.StatusOK, nil
}

type argumentInferenceError struct {
	argument string
	err      error
}

func replaceExtension(oldPath string, newExtension string) string {
	oldExtLen := len(path.Ext(oldPath))
	if oldExtLen == 0 {
		panic(fmt.Sprintf("Expected path %s to have an extension", oldPath))
	}
	return oldPath[:len(oldPath)-oldExtLen] + newExtension
}

func replaceBaseName(oldPath string, newBaseName string) string {
	if filepath.Dir(newBaseName) != "." {
		panic(fmt.Sprintf("Expected bare file name but found %s", newBaseName))
	}
	return filepath.Join(filepath.Dir(oldPath), newBaseName)
}

func handleSCIP(out *output.Output) error {
	fileExt := path.Ext(codeintelUploadFlags.file)
	if len(fileExt) == 0 {
		return errors.Newf("missing file extension for %s; expected .scip or .lsif", codeintelUploadFlags.file)
	}
	inputFile := codeintelUploadFlags.file
	if fileExt == ".scip" || fileExt == ".lsif-typed" {
		// The user explicitly passed in a -file flag that points to an SCIP index.
		outputFile := replaceExtension(inputFile, ".lsif")
		if filepath.Base(inputFile) == "index.scip" {
			outputFile = replaceBaseName(inputFile, "dump.lsif")
		}
		// HACK: Modify the flags to point to the output file, because
		// that field of the flags is read when performing the upload.
		codeintelUploadFlags.file = outputFile
		return convertSCIPToLSIFGraph(out, inputFile, outputFile)
	}

	if _, err := os.Stat(inputFile); err == nil {
		// Do nothing, the provided -flag flag points to an existing
		// file that does not have the file extension `.lsif-typed` or `.scip`.
		return nil
	}

	scipFile := replaceExtension(inputFile, ".scip")
	if _, err := os.Stat(scipFile); os.IsNotExist(err) {
		// The input may be named 'dump.lsif', but the default name for SCIP
		// indexes is 'index.scip', not 'dump.scip'.
		scipFile = replaceBaseName(inputFile, "index.scip")
		if _, err := os.Stat(scipFile); os.IsNotExist(err) {
			lsifTypedFile := replaceExtension(inputFile, ".lsif-typed")
			if _, err := os.Stat(lsifTypedFile); os.IsNotExist(err) {
				// There is no `*.scip` or `*.lsif-typed` file for the inferred path.
				return nil
			}
			scipFile = lsifTypedFile
		}
	}

	// The provided -file flag points to an `*.lsif` file that doesn't exist
	// so we convert the sibling file (which we confirmed exists).
	return convertSCIPToLSIFGraph(out, scipFile, codeintelUploadFlags.file)
}

// Reads the SCIP encoded input file and writes the corresponding LSIF
// Graph encoded output file.
func convertSCIPToLSIFGraph(out *output.Output, inputFile, outputFile string) error {
	if out != nil {
		out.Writef("%s  Converting %s into %s", output.EmojiInfo, inputFile, outputFile)
	}
	tmp, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer tmp.Close()

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read SCIP index '%s'", inputFile)
	}
	index := scip.Index{}
	err = proto.Unmarshal(data, &index)
	if err != nil {
		return errors.Wrapf(err, "failed to parse protobuf file '%s'", inputFile)
	}
	els, err := scip.ConvertSCIPToLSIF(&index)
	if err != nil {
		return errors.Wrapf(err, "failed to convert SCIP index at '%s' to LSIF", inputFile)
	}
	err = scip.WriteNDJSON(scip.ElementsToJsonElements(els), tmp)
	if err != nil {
		return errors.Wrapf(err, "failed to write LSIF JSON output to '%s'", tmp.Name())
	}
	err = tmp.Close()
	if err != nil {
		return err
	}
	return nil
}

func inferDefaultFile() (string, error) {
	hasSCIP := true
	const scipFilename = "index.scip"
	if _, err := os.Stat(scipFilename); err != nil {
		if os.IsNotExist(err) {
			hasSCIP = false
		} else {
			return "", err
		}
	}

	hasLSIF := true
	const lsifFilename = "dump.lsif"
	if _, err := os.Stat(lsifFilename); err != nil {
		if os.IsNotExist(err) {
			hasLSIF = false
		} else {
			return "", err
		}
	}

	if hasSCIP && hasLSIF {
		return "", errors.Newf("both %s and %s exists - cannot determine unambiguous choice", scipFilename, lsifFilename)
	}
	if !(hasSCIP || hasLSIF) {
		return "", formatInferenceError(argumentInferenceError{"file", errors.Newf("neither %s nor %s exists", scipFilename, lsifFilename)})
	}

	if hasSCIP {
		return scipFilename, nil
	}

	return lsifFilename, nil
}

func formatInferenceError(inferenceErr argumentInferenceError) error {
	return errorWithHint{
		err: inferenceErr.err, hint: strings.Join([]string{
			fmt.Sprintf(
				"Unable to determine %s from environment. Check your working directory or supply -%s={value} explicitly",
				inferenceErr.argument,
				inferenceErr.argument,
			),
		}, "\n"),
	}
}

func handleLSIF(out *output.Output) error {
	if skipConversionToSCIP {
		return nil
	}

	fileExt := path.Ext(codeintelUploadFlags.file)
	if len(fileExt) == 0 {
		return errors.Newf("missing file extension for %s; expected .scip or .lsif", codeintelUploadFlags.file)
	}

	if fileExt == ".lsif" || fileExt == ".dump" {
		inputFile := codeintelUploadFlags.file
		outputFile := replaceExtension(inputFile, ".scip")
		codeintelUploadFlags.file = outputFile
		return convertLSIFToSCIP(out, inputFile, outputFile)
	}

	return nil
}

// Reads the LSIF encoded input file and writes the corresponding SCIP encoded output file.
func convertLSIFToSCIP(out *output.Output, inputFile, outputFile string) error {
	if out != nil {
		out.Writef("%s  Converting %s into %s", output.EmojiInfo, inputFile, outputFile)
	}

	ctx := context.Background()
	uploadID := -time.Now().Nanosecond()
	root := codeintelUploadFlags.root

	if !isFlagSet(codeintelUploadFlagSet, "root") {
		// Best-effort infer the root; we have a strange cyclic init order where we're
		// currently trying to determine the filename that determines the root, but we
		// need the root when converting from LSIF to SCIP.
		root, _ = inferIndexRoot()
	}

	rc, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer rc.Close()

	index, err := libscip.ConvertLSIF(ctx, uploadID, rc, root)
	if err != nil {
		return err
	}
	serialized, err := proto.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, serialized, os.ModePerm)
}

// inferMissingCodeIntelUploadFlags updates the flags values which were not explicitly
// supplied by the user with default values inferred from the current git state and
// filesystem.
//
// Note: This function must not be called before codeintelUploadFlagset.Parse.
func inferMissingCodeIntelUploadFlags() (inferErrors []argumentInferenceError) {
	indexerName, indexerVersion, readIndexerNameAndVersionErr := readIndexerNameAndVersion()
	getIndexerName := func() (string, error) { return indexerName, readIndexerNameAndVersionErr }
	getIndexerVersion := func() (string, error) { return indexerVersion, readIndexerNameAndVersionErr }

	if err := inferUnsetFlag("repo", &codeintelUploadFlags.repo, codeintel.InferRepo); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("commit", &codeintelUploadFlags.commit, codeintel.InferCommit); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("root", &codeintelUploadFlags.root, inferIndexRoot); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("indexer", &codeintelUploadFlags.indexer, getIndexerName); err != nil {
		inferErrors = append(inferErrors, *err)
	}
	if err := inferUnsetFlag("indexerVersion", &codeintelUploadFlags.indexerVersion, getIndexerVersion); err != nil {
		inferErrors = append(inferErrors, *err)
	}

	return inferErrors
}

// inferUnsetFlag conditionally updates the value of the given pointer with the
// return value of the given function. If the flag with the given name was supplied
// by the user, then this function no-ops. An argumentInferenceError is returned if
// the given function returns an error.
//
// Note: This function must not be called before codeintelUploadFlagset.Parse.
func inferUnsetFlag(name string, target *string, f func() (string, error)) *argumentInferenceError {
	if isFlagSet(codeintelUploadFlagSet, name) {
		return nil
	}

	value, err := f()
	if err != nil {
		return &argumentInferenceError{name, err}
	}

	*target = value
	return nil
}

// isFlagSet returns true if the flag with the given name was supplied by the user.
// This lets us distinguish between zero-values (empty strings) and void values without
// requiring pointers and adding a layer of indirection deeper in the program.
func isFlagSet(fs *flag.FlagSet, name string) (found bool) {
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

// inferIndexRoot returns the root directory based on the configured index file path.
//
// Note: This function must not be called before codeintelUploadFlagset.Parse.
func inferIndexRoot() (string, error) {
	return codeintel.InferRoot(codeintelUploadFlags.file)
}

// readIndexerNameAndVersion returns the indexer name and version values read from the
// toolInfo value in the configured index file.
//
// Note: This function must not be called before codeintelUploadFlagset.Parse.
func readIndexerNameAndVersion() (string, string, error) {
	file, err := os.Open(codeintelUploadFlags.file)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	return upload.ReadIndexerNameAndVersion(file)
}

// validateCodeIntelUploadFlags returns an error if any of the parsed flag values are illegal.
//
// Note: This function must not be called before codeintelUploadFlagset.Parse.
func validateCodeIntelUploadFlags() error {
	codeintelUploadFlags.root = codeintel.SanitizeRoot(codeintelUploadFlags.root)

	if strings.HasPrefix(codeintelUploadFlags.root, "..") {
		return errors.New("root must not be outside of repository")
	}

	if codeintelUploadFlags.maxPayloadSizeMb < 25 {
		return errors.New("max-payload-size must be at least 25 (MB)")
	}

	return nil
}
