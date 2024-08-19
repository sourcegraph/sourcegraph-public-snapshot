// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// package build implements a more convenient interface for building
// zoekt indices.
package build

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/grafana/regexp"
	"github.com/rs/xid"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/ctags"
)

var DefaultDir = filepath.Join(os.Getenv("HOME"), ".zoekt")

// Branch describes a single branch version.
type Branch struct {
	Name    string
	Version string
}

// Options sets options for the index building.
type Options struct {
	// IndexDir is a directory that holds *.zoekt index files.
	IndexDir string

	// SizeMax is the maximum file size
	SizeMax int

	// Parallelism is the maximum number of shards to index in parallel
	Parallelism int

	// ShardMax sets the maximum corpus size for a single shard
	ShardMax int

	// TrigramMax sets the maximum number of distinct trigrams per document.
	TrigramMax int

	// RepositoryDescription holds names and URLs for the repository.
	RepositoryDescription zoekt.Repository

	// SubRepositories is a path => sub repository map.
	SubRepositories map[string]*zoekt.Repository

	// DisableCTags disables the generation of ctags metadata.
	DisableCTags bool

	// CtagsPath is the path to the ctags binary to run, or empty
	// if a valid binary couldn't be found.
	CTagsPath string

	// Same as CTagsPath but for scip-ctags
	ScipCTagsPath string

	// If set, ctags must succeed.
	CTagsMustSucceed bool

	// Write memory profiles to this file.
	MemProfile string

	// LargeFiles is a slice of glob patterns, including ** for any number
	// of directories, where matching file paths should be indexed
	// regardless of their size. The full pattern syntax is here:
	// https://github.com/bmatcuk/doublestar/tree/v1#patterns.
	LargeFiles []string

	// IsDelta is true if this run contains only the changed documents since the
	// last run.
	IsDelta bool

	// DocumentRanksPath is the path to the file with document ranks. If empty,
	// ranks will be computed on-the-fly.
	DocumentRanksPath string

	// DocumentRanksVersion is a string which when changed will cause us to
	// reindex a shard. This field is used so that when the contents of
	// DocumentRanksPath changes, we can reindex.
	DocumentRanksVersion string

	// changedOrRemovedFiles is a list of file paths that have been changed or removed
	// since the last indexing job for this repository. These files will be tombstoned
	// in the older shards for this repository.
	changedOrRemovedFiles []string

	LanguageMap ctags.LanguageMap

	// ShardMerging is true if builder should respect compound shards. This is a
	// Sourcegraph specific option.
	ShardMerging bool
}

// HashOptions contains only the options in Options that upon modification leads to IndexState of IndexStateMismatch during the next index building.
type HashOptions struct {
	sizeMax          int
	disableCTags     bool
	ctagsPath        string
	cTagsMustSucceed bool
	largeFiles       []string

	// documentRankVersion is an experimental field which will change when the
	// DocumentRanksPath content changes. If empty we ignore it.
	documentRankVersion string
}

func (o *Options) HashOptions() HashOptions {
	return HashOptions{
		sizeMax:             o.SizeMax,
		disableCTags:        o.DisableCTags,
		ctagsPath:           o.CTagsPath,
		cTagsMustSucceed:    o.CTagsMustSucceed,
		largeFiles:          o.LargeFiles,
		documentRankVersion: o.DocumentRanksVersion,
	}
}

func (o *Options) GetHash() string {
	h := o.HashOptions()
	hasher := sha1.New()

	hasher.Write([]byte(h.ctagsPath))
	hasher.Write([]byte(fmt.Sprintf("%t", h.cTagsMustSucceed)))
	hasher.Write([]byte(fmt.Sprintf("%d", h.sizeMax)))
	hasher.Write([]byte(fmt.Sprintf("%q", h.largeFiles)))
	hasher.Write([]byte(fmt.Sprintf("%t", h.disableCTags)))

	if h.documentRankVersion != "" {
		hasher.Write([]byte{0})
		io.WriteString(hasher, h.documentRankVersion)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

type largeFilesFlag struct{ *Options }

func (f largeFilesFlag) String() string {
	// From flag.Value documentation:
	//
	// The flag package may call the String method with a zero-valued receiver,
	// such as a nil pointer.
	if f.Options == nil {
		return ""
	}
	s := append([]string{""}, f.LargeFiles...)
	return strings.Join(s, "-large_file ")
}

func (f largeFilesFlag) Set(value string) error {
	f.LargeFiles = append(f.LargeFiles, value)
	return nil
}

// Flags adds flags for build options to fs. It is the "inverse" of Args.
func (o *Options) Flags(fs *flag.FlagSet) {
	x := *o
	x.SetDefaults()
	fs.IntVar(&o.SizeMax, "file_limit", x.SizeMax, "maximum file size")
	fs.IntVar(&o.TrigramMax, "max_trigram_count", x.TrigramMax, "maximum number of trigrams per document")
	fs.IntVar(&o.ShardMax, "shard_limit", x.ShardMax, "maximum corpus size for a shard")
	fs.IntVar(&o.Parallelism, "parallelism", x.Parallelism, "maximum number of parallel indexing processes.")
	fs.StringVar(&o.IndexDir, "index", x.IndexDir, "directory for search indices")
	fs.BoolVar(&o.CTagsMustSucceed, "require_ctags", x.CTagsMustSucceed, "If set, ctags calls must succeed.")
	fs.Var(largeFilesFlag{o}, "large_file", "A glob pattern where matching files are to be index regardless of their size. You can add multiple patterns by setting this more than once.")
	fs.StringVar(&o.MemProfile, "memprofile", "", "write memory profile(s) to `file.shardnum`. Note: sets parallelism to 1.")

	// Sourcegraph specific
	fs.BoolVar(&o.DisableCTags, "disable_ctags", x.DisableCTags, "If set, ctags will not be called.")
	fs.BoolVar(&o.ShardMerging, "shard_merging", x.ShardMerging, "If set, builder will respect compound shards.")
}

// Args generates command line arguments for o. It is the "inverse" of Flags.
func (o *Options) Args() []string {
	var args []string

	if o.SizeMax != 0 {
		args = append(args, "-file_limit", strconv.Itoa(o.SizeMax))
	}

	if o.TrigramMax != 0 {
		args = append(args, "-max_trigram_count", strconv.Itoa(o.TrigramMax))
	}

	if o.ShardMax != 0 {
		args = append(args, "-shard_limit", strconv.Itoa(o.ShardMax))
	}

	if o.Parallelism != 0 {
		args = append(args, "-parallelism", strconv.Itoa(o.Parallelism))
	}

	if o.IndexDir != "" {
		args = append(args, "-index", o.IndexDir)
	}

	if o.CTagsMustSucceed {
		args = append(args, "-require_ctags")
	}

	for _, a := range o.LargeFiles {
		args = append(args, "-large_file", a)
	}

	// Sourcegraph specific
	if o.DisableCTags {
		args = append(args, "-disable_ctags")
	}

	if o.ShardMerging {
		args = append(args, "-shard_merging")
	}

	return args
}

// Builder manages (parallel) creation of uniformly sized shards. The
// builder buffers up documents until it collects enough documents and
// then builds a shard and writes.
type Builder struct {
	opts     Options
	throttle chan int

	nextShardNum int
	todo         []*zoekt.Document
	docChecker   zoekt.DocChecker
	size         int

	parserBins ctags.ParserBinMap
	building   sync.WaitGroup

	errMu      sync.Mutex
	buildError error

	// temp name => final name for finished shards. We only rename
	// them once all shards succeed to avoid Frankstein corpuses.
	finishedShards map[string]string

	// indexTime is set by tests for doing reproducible builds.
	indexTime time.Time

	// a sortable 20 chars long id.
	id string

	finishCalled bool
}

type finishedShard struct {
	temp, final string
}

func checkCTags() string {
	if ctags := os.Getenv("CTAGS_COMMAND"); ctags != "" {
		return ctags
	}

	if ctags, err := exec.LookPath("universal-ctags"); err == nil {
		return ctags
	}

	return ""
}

func checkScipCTags() string {
	if ctags := os.Getenv("SCIP_CTAGS_COMMAND"); ctags != "" {
		return ctags
	}

	if ctags, err := exec.LookPath("scip-ctags"); err == nil {
		return ctags
	}

	return ""
}

// SetDefaults sets reasonable default options.
func (o *Options) SetDefaults() {
	if o.CTagsPath == "" && !o.DisableCTags {
		o.CTagsPath = checkCTags()
	}

	if o.ScipCTagsPath == "" && !o.DisableCTags {
		o.ScipCTagsPath = checkScipCTags()
	}

	if o.Parallelism == 0 {
		o.Parallelism = 4
	}
	if o.SizeMax == 0 {
		o.SizeMax = 2 << 20
	}
	if o.ShardMax == 0 {
		o.ShardMax = 100 << 20
	}
	if o.TrigramMax == 0 {
		o.TrigramMax = 20000
	}

	if o.RepositoryDescription.Name == "" && o.RepositoryDescription.URL != "" {
		parsed, _ := url.Parse(o.RepositoryDescription.URL)
		if parsed != nil {
			o.RepositoryDescription.Name = filepath.Join(parsed.Host, parsed.Path)
		}
	}
}

func hashString(s string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ShardName returns the name the given index shard.
func (o *Options) shardName(n int) string {
	return o.shardNameVersion(zoekt.IndexFormatVersion, n)
}

func (o *Options) shardNameVersion(version, n int) string {
	abs := url.QueryEscape(o.RepositoryDescription.Name)
	if len(abs) > 200 {
		abs = abs[:200] + hashString(abs)[:8]
	}
	return filepath.Join(o.IndexDir,
		fmt.Sprintf("%s_v%d.%05d.zoekt", abs, version, n))
}

type IndexState string

const (
	IndexStateMissing IndexState = "missing"
	IndexStateCorrupt IndexState = "corrupt"
	IndexStateVersion IndexState = "version-mismatch"
	IndexStateOption  IndexState = "option-mismatch"
	IndexStateMeta    IndexState = "meta-mismatch"
	IndexStateContent IndexState = "content-mismatch"
	IndexStateEqual   IndexState = "equal"
)

var readVersions = []struct {
	IndexFormatVersion int
	FeatureVersion     int
}{{
	IndexFormatVersion: zoekt.IndexFormatVersion,
	FeatureVersion:     zoekt.FeatureVersion,
}, {
	IndexFormatVersion: zoekt.NextIndexFormatVersion,
	FeatureVersion:     zoekt.FeatureVersion,
}}

// IncrementalSkipIndexing returns true if the index present on disk matches
// the build options.
func (o *Options) IncrementalSkipIndexing() bool {
	state, _ := o.IndexState()
	return state == IndexStateEqual
}

// IndexState checks how the index present on disk compares to the build
// options and returns the IndexState and the name of the first shard.
func (o *Options) IndexState() (IndexState, string) {
	// Open the latest version we support that is on disk.
	fn := o.findShard()
	if fn == "" {
		return IndexStateMissing, fn
	}

	repos, index, err := zoekt.ReadMetadataPathAlive(fn)
	if os.IsNotExist(err) {
		return IndexStateMissing, fn
	} else if err != nil {
		return IndexStateCorrupt, fn
	}

	for _, v := range readVersions {
		if v.IndexFormatVersion == index.IndexFormatVersion && v.FeatureVersion != index.IndexFeatureVersion {
			return IndexStateVersion, fn
		}
	}

	var repo *zoekt.Repository
	for _, cand := range repos {
		if cand.Name == o.RepositoryDescription.Name {
			repo = cand
			break
		}
	}

	if repo == nil {
		return IndexStateCorrupt, fn
	}

	if repo.IndexOptions != o.GetHash() {
		return IndexStateOption, fn
	}

	if !reflect.DeepEqual(repo.Branches, o.RepositoryDescription.Branches) {
		return IndexStateContent, fn
	}

	// We can mutate repo since it lives in the scope of this function call.
	if updated, err := repo.MergeMutable(&o.RepositoryDescription); err != nil {
		// non-nil err means we are trying to update an immutable field =>
		// reindex content.
		log.Printf("warn: immutable field changed, requires re-index: %s", err)
		return IndexStateContent, fn
	} else if updated {
		return IndexStateMeta, fn
	}

	return IndexStateEqual, fn
}

// FindRepositoryMetadata returns the index metadata for the repository
// specified in the options. 'ok' is false if the repository's metadata
// couldn't be found or if an error occurred.
func (o *Options) FindRepositoryMetadata() (repository *zoekt.Repository, metadata *zoekt.IndexMetadata, ok bool, err error) {
	shard := o.findShard()
	if shard == "" {
		return nil, nil, false, nil
	}

	repositories, metadata, err := zoekt.ReadMetadataPathAlive(shard)
	if err != nil {
		return nil, nil, false, fmt.Errorf("reading metadata for shard %q: %w", shard, err)
	}

	ID := o.RepositoryDescription.ID
	for _, r := range repositories {
		// compound shards contain multiple repositories, so we
		// have to pick only the one we're looking for
		if r.ID == ID {
			return r, metadata, true, nil
		}
	}

	// If we're here, then we're somehow in a state where we found a matching
	// shard that's missing the repository metadata we're looking for. This
	// should never happen.
	name := o.RepositoryDescription.Name
	return nil, nil, false, fmt.Errorf("matching shard %q doesn't contain metadata for repo id %d (%q)", shard, ID, name)
}

func (o *Options) findShard() string {
	for _, v := range readVersions {
		fn := o.shardNameVersion(v.IndexFormatVersion, 0)
		if _, err := os.Stat(fn); err == nil {
			return fn
		}
	}

	// Brute force finding the shard in compound shards. We should only hit this
	// code path for repositories that are not already existing or are in
	// compound shards.
	//
	// TODO add an oracle which can speed this up in the case of repositories
	// already in compound shards.
	compoundShards, err := filepath.Glob(path.Join(o.IndexDir, "compound-*.zoekt"))
	if err != nil {
		return ""
	}
	for _, fn := range compoundShards {
		repos, _, err := zoekt.ReadMetadataPathAlive(fn)
		if err != nil {
			continue
		}
		for _, repo := range repos {
			if repo.ID == o.RepositoryDescription.ID {
				return fn
			}
		}
	}

	return ""
}

func (o *Options) FindAllShards() []string {
	for _, v := range readVersions {
		fn := o.shardNameVersion(v.IndexFormatVersion, 0)
		if _, err := os.Stat(fn); err == nil {
			shards := []string{fn}
			for i := 1; ; i++ {
				fn := o.shardNameVersion(v.IndexFormatVersion, i)
				if _, err := os.Stat(fn); err != nil {
					return shards
				}
				shards = append(shards, fn)
			}
		}
	}

	// lazily fallback to findShard which will look for a compound shard.
	if fn := o.findShard(); fn != "" {
		return []string{fn}
	}

	return nil
}

// IgnoreSizeMax determines whether the max size should be ignored.
func (o *Options) IgnoreSizeMax(name string) bool {
	// A pattern match will override preceding pattern matches.
	for i := len(o.LargeFiles) - 1; i >= 0; i-- {
		pattern := strings.TrimSpace(o.LargeFiles[i])
		negated, validatedPattern := checkIsNegatePattern(pattern)

		if m, _ := doublestar.PathMatch(validatedPattern, name); m {
			if negated {
				return false
			} else {
				return true
			}
		}
	}

	return false
}

func checkIsNegatePattern(pattern string) (bool, string) {
	negate := "!"

	// if negated then strip prefix meta character which identifies negated filter pattern
	if strings.HasPrefix(pattern, negate) {
		return true, pattern[len(negate):]
	}

	return false, pattern
}

// NewBuilder creates a new Builder instance.
func NewBuilder(opts Options) (*Builder, error) {
	opts.SetDefaults()
	if opts.RepositoryDescription.Name == "" {
		return nil, fmt.Errorf("builder: must set Name")
	}

	b := &Builder{
		opts:           opts,
		throttle:       make(chan int, opts.Parallelism),
		finishedShards: map[string]string{},
	}

	parserBins, err := ctags.NewParserBinMap(
		b.opts.CTagsPath,
		b.opts.ScipCTagsPath,
		opts.LanguageMap,
		b.opts.CTagsMustSucceed,
	)
	if err != nil {
		return nil, err
	}

	b.parserBins = parserBins

	if opts.IsDelta {
		// Delta shards build on top of previously existing shards.
		// As a consequence, the shardNum for delta shards starts from
		// the number following the most recently generated shard - not 0.
		//
		// Using this numbering scheme allows all the shards to be
		// discovered as a set.
		shards := b.opts.FindAllShards()
		b.nextShardNum = len(shards) // shards are zero indexed, so len() provides the next number after the last one
	}

	if _, err := b.newShardBuilder(); err != nil {
		return nil, err
	}

	now := time.Now()
	b.indexTime = now
	b.id = xid.NewWithTime(now).String()

	return b, nil
}

// AddFile is a convenience wrapper for the Add method
func (b *Builder) AddFile(name string, content []byte) error {
	return b.Add(zoekt.Document{Name: name, Content: content})
}

func (b *Builder) Add(doc zoekt.Document) error {
	if b.finishCalled {
		return nil
	}

	allowLargeFile := b.opts.IgnoreSizeMax(doc.Name)
	if len(doc.Content) > b.opts.SizeMax && !allowLargeFile {
		// We could pass the document on to the shardbuilder, but if
		// we pass through a part of the source tree with binary/large
		// files, the corresponding shard would be mostly empty, so
		// insert a reason here too.
		doc.SkipReason = fmt.Sprintf("document size %d larger than limit %d", len(doc.Content), b.opts.SizeMax)
	} else if err := b.docChecker.Check(doc.Content, b.opts.TrigramMax, allowLargeFile); err != nil {
		doc.SkipReason = err.Error()
		doc.Language = "binary"
	}

	b.todo = append(b.todo, &doc)

	if doc.SkipReason == "" {
		b.size += len(doc.Name) + len(doc.Content)
	} else {
		b.size += len(doc.Name) + len(doc.SkipReason)
		// Drop the content if we are skipping the document. Skipped content is not counted towards the
		// shard size limit, so otherwise we might buffer too much data in memory before flushing.
		doc.Content = nil
	}

	if b.size > b.opts.ShardMax {
		return b.flush()
	}

	return nil
}

// MarkFileAsChangedOrRemoved indicates that the file specified by the given path
// has been changed or removed since the last indexing job for this repository.
//
// If this build is a delta build, these files will be tombstoned in the older shards for this repository.
func (b *Builder) MarkFileAsChangedOrRemoved(path string) {
	b.opts.changedOrRemovedFiles = append(b.opts.changedOrRemovedFiles, path)
}

// Finish creates a last shard from the buffered documents, and clears
// stale shards from previous runs. This should always be called, also
// in failure cases, to ensure cleanup.
//
// It is safe to call Finish() multiple times.
func (b *Builder) Finish() error {
	if b.finishCalled {
		return b.buildError
	}

	b.finishCalled = true

	b.flush()
	b.building.Wait()

	if b.buildError != nil {
		for tmp := range b.finishedShards {
			log.Printf("Builder.Finish %s", tmp)
			os.Remove(tmp)
		}
		b.finishedShards = map[string]string{}
		return b.buildError
	}

	// map of temporary -> final names for all updated shards + shard metadata files
	artifactPaths := make(map[string]string)
	for tmp, final := range b.finishedShards {
		artifactPaths[tmp] = final
	}

	oldShards := b.opts.FindAllShards()

	if b.opts.IsDelta {
		// Delta shard builds need to update FileTombstone and branch commit information for all
		// existing shards
		for _, shard := range oldShards {
			repositories, _, err := zoekt.ReadMetadataPathAlive(shard)
			if err != nil {
				return fmt.Errorf("reading metadata from shard %q: %w", shard, err)
			}

			if len(repositories) > 1 {
				return fmt.Errorf("delta shard builds don't support repositories contained in compound shards (shard %q)", shard)
			}

			if len(repositories) == 0 {
				return fmt.Errorf("failed to update repository metadata for shard %q - shard contains no repositories", shard)
			}

			repository := repositories[0]
			if repository.ID != b.opts.RepositoryDescription.ID {
				return fmt.Errorf("shard %q doesn't contain repository ID %d (%q)", shard, b.opts.RepositoryDescription.ID, b.opts.RepositoryDescription.Name)
			}

			if len(b.opts.changedOrRemovedFiles) > 0 && repository.FileTombstones == nil {
				repository.FileTombstones = make(map[string]struct{}, len(b.opts.changedOrRemovedFiles))
			}

			for _, f := range b.opts.changedOrRemovedFiles {
				repository.FileTombstones[f] = struct{}{}
			}

			if !BranchNamesEqual(repository.Branches, b.opts.RepositoryDescription.Branches) {
				return deltaBranchSetError{
					shardName: shard,
					old:       repository.Branches,
					new:       b.opts.RepositoryDescription.Branches,
				}
			}

			if b.opts.GetHash() != repository.IndexOptions {
				return &deltaIndexOptionsMismatchError{
					shardName:  shard,
					newOptions: b.opts.HashOptions(),
				}
			}

			repository.Branches = b.opts.RepositoryDescription.Branches

			repository.LatestCommitDate = b.opts.RepositoryDescription.LatestCommitDate

			tempPath, finalPath, err := zoekt.JsonMarshalRepoMetaTemp(shard, repository)
			if err != nil {
				return fmt.Errorf("writing repository metadta for shard %q: %w", shard, err)
			}

			artifactPaths[tempPath] = finalPath
		}
	}

	// We mark finished shards as empty when we successfully finish. Return now
	// to allow call sites to call Finish idempotently.
	if len(artifactPaths) == 0 {
		return b.buildError
	}

	// Collect a map of the old shards on disk. For each new shard we replace we
	// delete it from toDelete. Anything remaining in toDelete will be removed
	// after we have renamed everything into place.

	var toDelete map[string]struct{}
	if !b.opts.IsDelta {
		// Non-delta shard builds delete all existing shards before they write out
		// new ones.
		// By contrast, delta shard builds work by stacking changes on top of existing shards.
		// So, we skip populating the toDelete map if we're building delta shards.

		toDelete = make(map[string]struct{})
		for _, name := range oldShards {
			paths, err := zoekt.IndexFilePaths(name)
			if err != nil {
				b.buildError = fmt.Errorf("failed to find old paths for %s: %w", name, err)
			}
			for _, p := range paths {
				toDelete[p] = struct{}{}
			}
		}
	}

	for tmp, final := range artifactPaths {
		if err := os.Rename(tmp, final); err != nil {
			b.buildError = err
			continue
		}

		delete(toDelete, final)
	}

	b.finishedShards = map[string]string{}

	for p := range toDelete {
		// Don't delete compound shards, set tombstones instead.
		if b.opts.ShardMerging && strings.HasPrefix(filepath.Base(p), "compound-") {
			if !strings.HasSuffix(p, ".zoekt") {
				continue
			}
			err := zoekt.SetTombstone(p, b.opts.RepositoryDescription.ID)
			b.buildError = err
			continue
		}
		log.Printf("removing old shard file: %s", p)
		if err := os.Remove(p); err != nil {
			b.buildError = err
		}
	}

	return b.buildError
}

// BranchNamesEqual compares the given zoekt.RepositoryBranch slices, and returns true
// iff both slices specify the same set of branch names in the same order.
func BranchNamesEqual(a, b []zoekt.RepositoryBranch) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		x, y := a[i], b[i]
		if x.Name != y.Name {
			return false
		}
	}

	return true
}

func (b *Builder) flush() error {
	todo := b.todo
	b.todo = nil
	b.size = 0
	b.errMu.Lock()
	defer b.errMu.Unlock()
	if b.buildError != nil {
		return b.buildError
	}

	hasShard := b.nextShardNum > 0
	if len(todo) == 0 && hasShard {
		return nil
	}

	shard := b.nextShardNum
	b.nextShardNum++

	if b.opts.Parallelism > 1 && b.opts.MemProfile == "" {
		b.building.Add(1)
		b.throttle <- 1
		go func() {
			done, err := b.buildShard(todo, shard)
			<-b.throttle

			b.errMu.Lock()
			defer b.errMu.Unlock()
			if err != nil && b.buildError == nil {
				b.buildError = err
			}
			if err == nil {
				b.finishedShards[done.temp] = done.final
			}
			b.building.Done()
		}()
	} else {
		// No goroutines when we're not parallel. This
		// simplifies memory profiling.
		done, err := b.buildShard(todo, shard)
		b.buildError = err
		if err == nil {
			b.finishedShards[done.temp] = done.final
		}
		if b.opts.MemProfile != "" {
			// drop memory, and profile.
			todo = nil
			b.writeMemProfile(b.opts.MemProfile)
		}

		return b.buildError
	}

	return nil
}

var profileNumber int

func (b *Builder) writeMemProfile(name string) {
	nm := fmt.Sprintf("%s.%d", name, profileNumber)
	profileNumber++
	f, err := os.Create(nm)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	f.Close()
	log.Printf("wrote mem profile %q", nm)
}

// map [0,inf) to [0,1) monotonically
func squashRange(j int) float64 {
	x := float64(j)
	return x / (1 + x)
}

// IsLowPriority takes a file name and makes an educated guess about its priority
// in search results. A file is considered low priority if it looks like a test,
// vendored, or generated file.
//
// These 'priority' criteria affects how documents are ordered within a shard. It's
// also used to help guess a file's rank when we're missing ranking information.
func IsLowPriority(file string) bool {
	return testRe.MatchString(file) || isGenerated(file) || isVendored(file)
}

var testRe = regexp.MustCompile("[Tt]est")

func isGenerated(file string) bool {
	return strings.HasSuffix(file, "min.js") || strings.HasSuffix(file, "js.map")
}

func isVendored(file string) bool {
	return strings.Contains(file, "vendor/") || strings.Contains(file, "node_modules/")
}

type rankedDoc struct {
	*zoekt.Document
	rank []float64
}

// rank returns a vector of scores which is used at index-time to sort documents
// before writing them to disk. The order of documents in the shard is important
// at query time, because earlier documents receive a boost at query time and
// have a higher chance of being searched before limits kick in.
func rank(d *zoekt.Document, origIdx int) []float64 {
	skipped := 0.0
	if d.SkipReason != "" {
		skipped = 1.0
	}

	generated := 0.0
	if isGenerated(d.Name) {
		generated = 1.0
	}

	vendor := 0.0
	if isVendored(d.Name) {
		vendor = 1.0
	}

	test := 0.0
	if testRe.MatchString(d.Name) {
		test = 1.0
	}

	// Smaller is earlier (=better).
	return []float64{
		// Always place skipped docs last
		skipped,

		// Prefer docs that are not generated
		generated,

		// Prefer docs that are not vendored
		vendor,

		// Prefer docs that are not tests
		test,

		// With short names
		squashRange(len(d.Name)),

		// With many symbols
		1.0 - squashRange(len(d.Symbols)),

		// With short content
		squashRange(len(d.Content)),

		// That is present is as many branches as possible
		1.0 - squashRange(len(d.Branches)),

		// Preserve original ordering.
		squashRange(origIdx),
	}
}

func sortDocuments(todo []*zoekt.Document) {
	rs := make([]rankedDoc, 0, len(todo))
	for i, t := range todo {
		rd := rankedDoc{t, rank(t, i)}
		rs = append(rs, rd)
	}
	sort.Slice(rs, func(i, j int) bool {
		r1 := rs[i].rank
		r2 := rs[j].rank
		for i := range r1 {
			if r1[i] < r2[i] {
				return true
			}
			if r1[i] > r2[i] {
				return false
			}
		}

		return false
	})
	for i := range todo {
		todo[i] = rs[i].Document
	}
}

func (b *Builder) buildShard(todo []*zoekt.Document, nextShardNum int) (*finishedShard, error) {
	if !b.opts.DisableCTags && (b.opts.CTagsPath != "" || b.opts.ScipCTagsPath != "") {
		err := parseSymbols(todo, b.opts.LanguageMap, b.parserBins)
		if b.opts.CTagsMustSucceed && err != nil {
			return nil, err
		}
		if err != nil {
			log.Printf("ignoring universal:%s or scip:%s error: %v", b.opts.CTagsPath, b.opts.ScipCTagsPath, err)
		}
	}

	name := b.opts.shardName(nextShardNum)

	shardBuilder, err := b.newShardBuilder()
	if err != nil {
		return nil, err
	}

	sortDocuments(todo)

	for _, t := range todo {
		if err := shardBuilder.Add(*t); err != nil {
			return nil, err
		}
	}

	return b.writeShard(name, shardBuilder)
}

func (b *Builder) newShardBuilder() (*zoekt.IndexBuilder, error) {
	desc := b.opts.RepositoryDescription
	desc.HasSymbols = !b.opts.DisableCTags && b.opts.CTagsPath != ""
	desc.SubRepoMap = b.opts.SubRepositories
	desc.IndexOptions = b.opts.GetHash()

	shardBuilder, err := zoekt.NewIndexBuilder(&desc)
	if err != nil {
		return nil, err
	}
	shardBuilder.IndexTime = b.indexTime
	shardBuilder.ID = b.id
	return shardBuilder, nil
}

func (b *Builder) writeShard(fn string, ib *zoekt.IndexBuilder) (*finishedShard, error) {
	dir := filepath.Dir(fn)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	f, err := os.CreateTemp(dir, filepath.Base(fn)+".*.tmp")
	if err != nil {
		return nil, err
	}
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0o666 &^ umask); err != nil {
			return nil, err
		}
	}

	defer f.Close()
	if err := ib.Write(f); err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	log.Printf("finished shard %s: %d index bytes (overhead %3.1f), %d files processed \n",
		fn,
		fi.Size(),
		float64(fi.Size())/float64(ib.ContentSize()+1),
		ib.NumFiles())

	return &finishedShard{f.Name(), fn}, nil
}

type deltaBranchSetError struct {
	shardName string
	old, new  []zoekt.RepositoryBranch
}

func (e deltaBranchSetError) Error() string {
	return fmt.Sprintf("repository metadata in shard %q contains a different set of branch names than what was requested, which is unsupported in a delta shard build. old: %+v, new: %+v", e.shardName, e.old, e.new)
}

type deltaIndexOptionsMismatchError struct {
	shardName  string
	newOptions HashOptions
}

func (e *deltaIndexOptionsMismatchError) Error() string {
	return fmt.Sprintf("one or more index options for shard %q do not match Builder's index options. These index option updates are incompatible with delta build. New index options: %+v", e.shardName, e.newOptions)
}

// umask holds the Umask of the current process
var umask os.FileMode
