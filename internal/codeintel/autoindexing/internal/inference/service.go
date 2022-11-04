package inference

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	otelog "github.com/opentracing/opentracing-go/log"
	baselua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/lua"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	sandboxService                  SandboxService
	gitService                      GitService
	limiter                         *ratelimit.InstrumentedLimiter
	maximumFilesWithContentCount    int
	maximumFileWithContentSizeBytes int
	operations                      *operations
}

type indexJobOrHint struct {
	indexJob     *config.IndexJob
	indexJobHint *config.IndexJobHint
}

type invocationContext struct {
	sandbox    *luasandbox.Sandbox
	gitService GitService
	repo       api.RepoName
	commit     string
	invocationFunctionTable
}

type invocationFunctionTable struct {
	linearize    func(recognizer *luatypes.Recognizer) []*luatypes.Recognizer
	callback     func(recognizer *luatypes.Recognizer) *baselua.LFunction
	scanLuaValue func(value baselua.LValue) ([]indexJobOrHint, error)
}

type LimitError struct {
	description string
}

func (e LimitError) Error() string {
	return e.description
}

func newService(
	sandboxService SandboxService,
	gitService GitService,
	limiter *ratelimit.InstrumentedLimiter,
	maximumFilesWithContentCount int,
	maximumFileWithContentSizeBytes int,
	observationContext *observation.Context,
) *Service {
	return &Service{
		sandboxService:                  sandboxService,
		gitService:                      gitService,
		limiter:                         limiter,
		maximumFilesWithContentCount:    maximumFilesWithContentCount,
		maximumFileWithContentSizeBytes: maximumFileWithContentSizeBytes,
		operations:                      newOperations(observationContext),
	}
}

// InferIndexJobs invokes the given script in a fresh Lua sandbox. The return value of this script
// is assumed to be a table of recognizer instances. Keys conflicting with the default recognizers
// will overwrite them (to disable or change default behavior). Each recognizer's generate function
// is invoked and the resulting index jobs are combined into a flattened list.
func (s *Service) InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) (_ []config.IndexJob, err error) {
	ctx, _, endObservation := s.operations.inferIndexJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	functionTable := invocationFunctionTable{
		linearize: luatypes.LinearizeGenerator,
		callback:  func(recognizer *luatypes.Recognizer) *baselua.LFunction { return recognizer.Generator() },
		scanLuaValue: func(value baselua.LValue) ([]indexJobOrHint, error) {
			jobs, err := luatypes.IndexJobsFromTable(value)
			if err != nil {
				return nil, err
			}

			jobOrHints := make([]indexJobOrHint, 0, len(jobs))
			for _, job := range jobs {
				job := job // prevent loop capture
				jobOrHints = append(jobOrHints, indexJobOrHint{indexJob: &job})
			}

			return jobOrHints, err
		},
	}

	jobOrHints, err := s.inferIndexJobOrHints(ctx, repo, commit, overrideScript, functionTable)
	if err != nil {
		return nil, err
	}

	jobs := make([]config.IndexJob, 0, len(jobOrHints))
	for _, jobOrHint := range jobOrHints {
		if jobOrHint.indexJob == nil {
			return nil, errors.New("unexpected nil index job")
		}

		jobs = append(jobs, *jobOrHint.indexJob)
	}

	return jobs, nil
}

// InferIndexJobHints invokes the given script in a fresh Lua sandbox. The return value of this script
// is assumed to be a table of recognizer instances. Keys conflicting with the default recognizers
// will overwrite them (to disable or change default behavior). Each recognizer's hints function is
// invoked and the resulting index job hints are combined into a flattened list.
func (s *Service) InferIndexJobHints(ctx context.Context, repo api.RepoName, commit, overrideScript string) (_ []config.IndexJobHint, err error) {
	ctx, _, endObservation := s.operations.inferIndexJobHints.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	functionTable := invocationFunctionTable{
		linearize: luatypes.LinearizeHinter,
		callback:  func(recognizer *luatypes.Recognizer) *baselua.LFunction { return recognizer.Hinter() },
		scanLuaValue: func(value baselua.LValue) ([]indexJobOrHint, error) {
			jobHints, err := luatypes.IndexJobHintsFromTable(value)
			if err != nil {
				return nil, err
			}

			jobOrHints := make([]indexJobOrHint, 0, len(jobHints))
			for _, jobHint := range jobHints {
				jobHint := jobHint // prevent loop capture
				jobOrHints = append(jobOrHints, indexJobOrHint{indexJobHint: &jobHint})
			}

			return jobOrHints, err
		},
	}

	jobOrHints, err := s.inferIndexJobOrHints(ctx, repo, commit, overrideScript, functionTable)
	if err != nil {
		return nil, err
	}

	jobHints := make([]config.IndexJobHint, 0, len(jobOrHints))
	for _, jobOrHint := range jobOrHints {
		if jobOrHint.indexJobHint == nil {
			return nil, errors.New("unexpected nil index job hint")
		}

		jobHints = append(jobHints, *jobOrHint.indexJobHint)
	}

	return jobHints, nil
}

// inferIndexJobOrHints invokes the given script in a fresh Lua sandbox. The return value of this script
// is assumed to be a table of recognizer instances. Keys conflicting with the default recognizers will
// overwrite them (to disable or change default behavior). Each recognizer's callback function is invoked
// and the resulting values are combined into a flattened list. See InferIndexJobs and InferIndexJobHints
// for concrete implementations of the given function table.
func (s *Service) inferIndexJobOrHints(
	ctx context.Context,
	repo api.RepoName,
	commit string,
	overrideScript string,
	invocationContextMethods invocationFunctionTable,
) ([]indexJobOrHint, error) {
	sandbox, err := s.createSandbox(ctx)
	if err != nil {
		return nil, err
	}
	defer sandbox.Close()

	recognizers, err := s.setupRecognizers(ctx, sandbox, overrideScript)
	if err != nil || len(recognizers) == 0 {
		return nil, err
	}

	invocationContext := invocationContext{
		sandbox:                 sandbox,
		gitService:              s.gitService,
		repo:                    repo,
		commit:                  commit,
		invocationFunctionTable: invocationContextMethods,
	}
	return s.invokeRecognizers(ctx, invocationContext, recognizers)
}

// createSandbox creates a Lua sandbox wih the modules loaded for use with auto indexing inference.
func (s *Service) createSandbox(ctx context.Context) (_ *luasandbox.Sandbox, err error) {
	ctx, _, endObservation := s.operations.createSandbox.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	opts := luasandbox.CreateOptions{
		Modules: defaultModules,
	}
	sandbox, err := s.sandboxService.CreateSandbox(ctx, opts)
	if err != nil {
		return nil, err
	}

	return sandbox, nil
}

// setupRecognizers runs the given default and override scripts in the given sandbox and converts the
// script return values to a list of recognizer instances.
func (s *Service) setupRecognizers(ctx context.Context, sandbox *luasandbox.Sandbox, overrideScript string) (_ []*luatypes.Recognizer, err error) {
	ctx, _, endObservation := s.operations.setupRecognizers.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	opts := luasandbox.RunOptions{}
	rawRecognizers, err := sandbox.RunScriptNamed(ctx, opts, lua.Scripts, "recognizers.lua")
	if err != nil {
		return nil, err
	}

	recognizerMap, err := luatypes.NamedRecognizersFromUserDataMap(rawRecognizers, false)
	if err != nil {
		return nil, err
	}

	if overrideScript != "" {
		rawRecognizers, err := sandbox.RunScript(ctx, opts, overrideScript)
		if err != nil {
			return nil, err
		}

		// Allow false values here, which will be indicated by a nil recognizer. In the loop below we will
		// add (or replace) any recognizer with the same name. To _unset_ a recognizer, we allow a user to
		// add nil as the table value.

		overrideRecognizerMap, err := luatypes.NamedRecognizersFromUserDataMap(rawRecognizers, true)
		if err != nil {
			return nil, err
		}

		for name, recognizer := range overrideRecognizerMap {
			if recognizer == nil {
				delete(recognizerMap, name)
			} else {
				recognizerMap[name] = recognizer
			}
		}
	}

	recognizers := make([]*luatypes.Recognizer, 0, len(recognizerMap))
	for _, recognizer := range recognizerMap {
		recognizers = append(recognizers, recognizer)
	}

	return recognizers, nil
}

// invokeRecognizers invokes each of the given recognizer's callback function and returns the resulting
// index job or hint values. This function is called iteratively with recognizers registered by a previous
// invocation of a recognizer. Calls to gitserver are made in as few batches across all recognizer invocations
// as possible.
func (s *Service) invokeRecognizers(
	ctx context.Context,
	invocationContext invocationContext,
	recognizers []*luatypes.Recognizer,
) (_ []indexJobOrHint, err error) {
	ctx, _, endObservation := s.operations.invokeRecognizers.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	patternsForPaths := luatypes.FlattenRecognizerPatterns(recognizers, false)
	patternsForContent := luatypes.FlattenRecognizerPatterns(recognizers, true)

	// Find the list of paths that match either of the partitioned pattern sets. We will feed the
	// concrete paths from this call into the archive call that follows, at least for the concrete
	// paths that match the path-for-content patterns.
	paths, err := s.resolvePaths(ctx, invocationContext, append(patternsForPaths, patternsForContent...))
	if err != nil {
		return nil, err
	}

	contentsByPath, err := s.resolveFileContents(ctx, invocationContext, paths, patternsForContent)
	if err != nil {
		return nil, err
	}

	jobOrHints, err := s.invokeRecognizerChains(ctx, invocationContext, recognizers, paths, contentsByPath)
	if err != nil {
		return nil, err
	}

	return jobOrHints, err
}

// resolvePaths requests all paths matching the given combined regular expression from gitserver. This
// list will likely be a superset of any one recognizer's expected set of paths, so we'll need to filter
// the data before each individual recognizer invocation so that we only pass in what matches the set of
// patterns specific to that recognizer instance.
func (s *Service) resolvePaths(
	ctx context.Context,
	invocationContext invocationContext,
	patternsForPaths []*luatypes.PathPattern,
) (_ []string, err error) {
	ctx, traceLogger, endObservation := s.operations.resolvePaths.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	pathPattern, err := flattenPatterns(patternsForPaths, false)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	rateLimitErr := s.limiter.Wait(ctx)
	traceLogger.Log(otelog.Int("wait_duration_ms", int(time.Since(start).Milliseconds())))
	if rateLimitErr != nil {
		return nil, err
	}

	paths, err := invocationContext.gitService.ListFiles(ctx, invocationContext.repo, invocationContext.commit, pathPattern)
	if err != nil {
		return nil, err
	}

	return paths, err
}

// resolveFileContents requests the content of the paths that match the given combined regular expression.
// The contents are fetched via a single git archive call.
func (s *Service) resolveFileContents(
	ctx context.Context,
	invocationContext invocationContext,
	paths []string,
	patternsForContent []*luatypes.PathPattern,
) (_ map[string]string, err error) {
	ctx, traceLogger, endObservation := s.operations.resolveFileContents.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	relevantPaths, err := filterPathsByPatterns(paths, patternsForContent)
	if err != nil {
		return nil, err
	}
	if len(relevantPaths) == 0 {
		return nil, nil
	}

	start := time.Now()
	rateLimitErr := s.limiter.Wait(ctx)
	traceLogger.Log(otelog.Int("wait_duration_ms", int(time.Since(start).Milliseconds())))
	if rateLimitErr != nil {
		return nil, err
	}

	pathspecs := make([]gitdomain.Pathspec, 0, len(relevantPaths))
	for _, p := range relevantPaths {
		pathspecs = append(pathspecs, gitdomain.PathspecLiteral(p))
	}
	opts := gitserver.ArchiveOptions{
		Treeish:   invocationContext.commit,
		Format:    gitserver.ArchiveFormatTar,
		Pathspecs: pathspecs,
	}
	rc, err := invocationContext.gitService.Archive(ctx, invocationContext.repo, opts)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	contentsByPath := map[string]string{}

	tr := tar.NewReader(rc)
	for {
		header, err := tr.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}

			break
		}

		if len(contentsByPath) >= s.maximumFilesWithContentCount {
			return nil, LimitError{
				description: fmt.Sprintf(
					"inference limit: requested content for more than %d (%d) files",
					s.maximumFilesWithContentCount,
					len(contentsByPath),
				),
			}
		}
		if int(header.Size) > s.maximumFileWithContentSizeBytes {
			return nil, LimitError{
				description: fmt.Sprintf(
					"inference limit: requested content for a file larger than %d (%d) bytes",
					s.maximumFileWithContentSizeBytes,
					int(header.Size),
				),
			}
		}

		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, tr, header.Size); err != nil {
			return nil, err
		}

		// Since we quoted all literal path specs on entry, we need to remove it from
		// the returned filepaths.
		contentsByPath[strings.TrimPrefix(header.Name, ":(literal)")] = buf.String()
	}

	return contentsByPath, nil
}

type registrationAPI struct {
	recognizers []*luatypes.Recognizer
}

func (api *registrationAPI) Register(recognizer *luatypes.Recognizer) {
	api.recognizers = append(api.recognizers, recognizer)
}

// invokeRecognizerChains invokes each of the given recognizer's callback function and combines
// their complete output.
func (s *Service) invokeRecognizerChains(
	ctx context.Context,
	invocationContext invocationContext,
	recognizers []*luatypes.Recognizer,
	paths []string,
	contentsByPath map[string]string,
) (jobOrHints []indexJobOrHint, _ error) {
	registrationAPI := &registrationAPI{}

	// Invoke the recognizers and gather the resulting jobs or hints
	for _, recognizer := range recognizers {
		additionalJobOrHints, err := s.invokeRecognizerChainUntilResults(
			ctx,
			invocationContext,
			recognizer,
			registrationAPI,
			paths,
			contentsByPath,
		)
		if err != nil {
			return nil, err
		}

		jobOrHints = append(jobOrHints, additionalJobOrHints...)
	}

	if len(registrationAPI.recognizers) != 0 {
		// Recursively call any recognizers that were registered from the previous invocation
		// of recognizers. This allows users to have control over conditional execution so that
		// gitserver data requests re minimal when requested with the expected query patterns.

		additionalJobOrHints, err := s.invokeRecognizers(ctx, invocationContext, registrationAPI.recognizers)
		if err != nil {
			return nil, err
		}
		jobOrHints = append(jobOrHints, additionalJobOrHints...)
	}

	return jobOrHints, nil
}

// invokeRecognizerChainUntilResults invokes the callback function from each recognizer reachable
// from the given root recognizer. Once a non-nil error or non-empty set of results are returned
// from a recognizer, the chain invocation halts and the recognizer's index job or hint values
// are returned.
func (s *Service) invokeRecognizerChainUntilResults(
	ctx context.Context,
	invocationContext invocationContext,
	recognizer *luatypes.Recognizer,
	registrationAPI *registrationAPI,
	paths []string,
	contentsByPath map[string]string,
) ([]indexJobOrHint, error) {
	for _, recognizer := range invocationContext.linearize(recognizer) {
		if jobOrHints, err := s.invokeLinearizedRecognizer(
			ctx,
			invocationContext,
			recognizer,
			registrationAPI,
			paths,
			contentsByPath,
		); err != nil || len(jobOrHints) > 0 {
			return jobOrHints, err
		}
	}

	return nil, nil
}

// invokeLinearizedRecognizer invokes a single recognizer callback.
func (s *Service) invokeLinearizedRecognizer(
	ctx context.Context,
	invocationContext invocationContext,
	recognizer *luatypes.Recognizer,
	registrationAPI *registrationAPI,
	paths []string,
	contentsByPath map[string]string,
) (_ []indexJobOrHint, err error) {
	ctx, _, endObservation := s.operations.invokeLinearizedRecognizer.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	callPaths, callContentsByPath, err := s.filterPathsForRecognizer(recognizer, paths, contentsByPath)
	if err != nil {
		return nil, err
	}
	if len(callPaths) == 0 && len(callContentsByPath) == 0 {
		return nil, nil
	}

	opts := luasandbox.RunOptions{}
	args := []any{registrationAPI, callPaths, callContentsByPath}
	value, err := invocationContext.sandbox.Call(ctx, opts, invocationContext.callback(recognizer), args...)
	if err != nil {
		return nil, err
	}

	jobOrHints, err := invocationContext.scanLuaValue(value)
	if err != nil {
		return nil, err
	}

	return jobOrHints, nil
}

// filterPathsForRecognizer creates a copy of the the given path slice and file content map
// that only contain elements matching the patterns attached to the given recognizer.
func (s *Service) filterPathsForRecognizer(
	recognizer *luatypes.Recognizer,
	paths []string,
	contentsByPath map[string]string,
) ([]string, map[string]string, error) {
	// Filter out paths which are not interesting to this recognizer
	filteredPaths, err := filterPathsByPatterns(paths, recognizer.Patterns(false))
	if err != nil {
		return nil, nil, err
	}

	// Filter out paths which are not interesting to this recognizer
	filteredPathsWithContent, err := filterPathsByPatterns(paths, recognizer.Patterns(true))
	if err != nil {
		return nil, nil, err
	}

	// Copy over content for remaining paths in map
	filteredContentsByPath := make(map[string]string, len(filteredPathsWithContent))
	for _, key := range filteredPathsWithContent {
		filteredContentsByPath[key] = contentsByPath[key]
	}

	return filteredPaths, filteredContentsByPath, nil
}
