package inference

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/lua"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	sandboxService SandboxService
	gitService     GitService
	operations     *operations
}

type invocationContext struct {
	sandbox    *luasandbox.Sandbox
	gitService GitService
	repo       api.RepoName
	commit     string
}

func newService(
	sandboxService SandboxService,
	gitService GitService,
	observationContext *observation.Context,
) *Service {
	return &Service{
		sandboxService: sandboxService,
		gitService:     gitService,
		operations:     newOperations(observationContext),
	}
}

// InferIndexJobs invokes the given script in a fresh Lua sandbox. The return value of this script
// is assumed to be a table of recognizer instances. Keys conflicting with the default recognizers
// will overwrite them (to disable or change default behavior).
func (s *Service) InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) (_ []config.IndexJob, err error) {
	ctx, _, endObservation := s.operations.inferIndexJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	sandbox, err := s.createSandbox(ctx)
	if err != nil {
		return nil, err
	}
	defer sandbox.Close()

	recognizers, err := s.setupRecognizers(ctx, sandbox, overrideScript)
	if err != nil || len(recognizers) == 0 {
		return nil, err
	}

	invocationContext := &invocationContext{
		sandbox:    sandbox,
		gitService: s.gitService,
		repo:       repo,
		commit:     commit,
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

	recognizers, err := luatypes.RecognizersFromUserDataMap(rawRecognizers)
	if err != nil {
		return nil, err
	}

	if overrideScript != "" {
		// TODO - run this script and merge recognizer results
		// See https://github.com/sourcegraph/sourcegraph/issues/33046
		return nil, errors.Newf("unimplemented")
	}

	return recognizers, nil
}

// invokeRecognizers invokes each generator function associated with one of the given recognizers
// and returns the resulting index job values. This function is called iteratively with recognizers
// registered by a previous invocation of a recognizer. Calls to gitserver are made in as few
// batches across all recognizer invocations as possible.
func (s *Service) invokeRecognizers(
	ctx context.Context,
	invocationContext *invocationContext,
	recognizers []*luatypes.Recognizer,
) (_ []config.IndexJob, err error) {
	ctx, _, endObservation := s.operations.invokeRecognizers.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	patternsForPaths, patternsForContent := partitionPatterns(recognizers)

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

	jobs, err := s.invokeRecognizerChains(ctx, invocationContext, recognizers, paths, contentsByPath)
	if err != nil {
		return nil, err
	}

	return jobs, err
}

// resolvePaths requests all paths matching the given combined regular expression from gitserver. This
// list will likely be a superset of any one recognizer, so we'll need to filter the data before each
// individual recognizer invocation so that we only pass in what matches the set of patterns specific to
// that recognizer instance.
func (s *Service) resolvePaths(
	ctx context.Context,
	invocationContext *invocationContext,
	patternsForPaths []*luatypes.PathPattern,
) (_ []string, err error) {
	ctx, _, endObservation := s.operations.resolvePaths.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	pathPattern, err := flattenPatterns(patternsForPaths, false)
	if err != nil {
		return nil, err
	}

	paths, err := invocationContext.gitService.ListFiles(ctx, invocationContext.repo, invocationContext.commit, pathPattern)
	if err != nil {
		return nil, err
	}

	return paths, err
}

// resolveFileContents requests the content of the paths that match teh given combined regular expression.
// The contents are fetched via a single git archive call.
func (s *Service) resolveFileContents(
	ctx context.Context,
	invocationContext *invocationContext,
	paths []string,
	patternsForContent []*luatypes.PathPattern,
) (_ map[string]string, err error) {
	ctx, _, endObservation := s.operations.resolveFileContents.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	relevantPaths, err := filterPathsByPatterns(paths, patternsForContent)
	if err != nil {
		return nil, err
	}
	pathspecs := make([]gitserver.Pathspec, 0, len(relevantPaths))
	for _, p := range relevantPaths {
		pathspecs = append(pathspecs, gitserver.PathspecLiteral(p))
	}

	opts := gitserver.ArchiveOptions{
		Treeish:   invocationContext.commit,
		Format:    "zip",
		Pathspecs: pathspecs,
	}
	rc, err := invocationContext.gitService.Archive(ctx, invocationContext.repo, opts)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, nil
	}

	contentsByPath := make(map[string]string, len(zr.File))
	for _, f := range zr.File {
		contents, err := readZipFile(f)
		if err != nil {
			return nil, err
		}

		// Since we quoted all literal path specs on entry, we need to remove it from
		// the returned filepaths.
		contentsByPath[strings.TrimPrefix(f.Name, ":(literal)")] = contents
	}

	return contentsByPath, nil
}

type registrationAPI struct {
	recognizers []*luatypes.Recognizer
}

func (api *registrationAPI) Register(recognizer *luatypes.Recognizer) {
	api.recognizers = append(api.recognizers, recognizer)
}

// invokeRecognizerChains invokes each of the recognizers and combines their complete output.
func (s *Service) invokeRecognizerChains(
	ctx context.Context,
	invocationContext *invocationContext,
	recognizers []*luatypes.Recognizer,
	paths []string,
	contentsByPath map[string]string,
) (jobs []config.IndexJob, _ error) {
	registrationAPI := &registrationAPI{}

	// Invoke the recognizers and gather the resulting jobs
	for _, recognizer := range recognizers {
		additionalJobs, err := s.invokeRecognizerChainUntilResults(
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

		jobs = append(jobs, additionalJobs...)
	}

	if len(registrationAPI.recognizers) != 0 {
		// Recursively call any recognizers that were registered from the previous invocation
		// of recognizers. This allows users to have control over conditional execution so that
		// gitserver data requests re minimal when requested with the expected query patterns.

		additionalJobs, err := s.invokeRecognizers(ctx, invocationContext, registrationAPI.recognizers)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, additionalJobs...)
	}

	return jobs, nil
}

// invokeRecognizerChainUntilResults invokes each of the linearized recognizers from the given
// recognizer root in sequence until a non-nil error or non-empty set of results are returned.
// Any remaining recognizers in the chain are not invoked.
func (s *Service) invokeRecognizerChainUntilResults(
	ctx context.Context,
	invocationContext *invocationContext,
	recognizer *luatypes.Recognizer,
	registrationAPI *registrationAPI,
	paths []string,
	contentsByPath map[string]string,
) ([]config.IndexJob, error) {
	for _, recognizer := range luatypes.LinearizeRecognizer(recognizer) {
		if values, err := s.invokeLinearizedRecognizer(
			ctx,
			invocationContext,
			recognizer,
			registrationAPI,
			paths,
			contentsByPath,
		); err != nil || len(values) > 0 {
			return values, err
		}
	}

	return nil, nil
}

// invokeLinearizedRecognizer invokes a single recognizer callback and converts the return
// value to a slice of index jobs.
func (s *Service) invokeLinearizedRecognizer(
	ctx context.Context,
	invocationContext *invocationContext,
	recognizer *luatypes.Recognizer,
	registrationAPI *registrationAPI,
	paths []string,
	contentsByPath map[string]string,
) (_ []config.IndexJob, err error) {
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
	args := []interface{}{registrationAPI, callPaths, callContentsByPath}
	value, err := invocationContext.sandbox.Call(ctx, opts, recognizer.Generator(), args...)
	if err != nil {
		return nil, err
	}

	jobs, err := luatypes.IndexJobsFromTable(value)
	if err != nil {
		return nil, err
	}

	return jobs, nil
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
