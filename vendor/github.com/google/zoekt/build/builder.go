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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"

	"github.com/google/zoekt"
	"github.com/google/zoekt/ctags"
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

	// RepositoryDescription holds names and URLs for the repository.
	RepositoryDescription zoekt.Repository

	// SubRepositories is a path => sub repository map.
	SubRepositories map[string]*zoekt.Repository

	// Path to exuberant ctags binary to run
	CTags string

	// If set, ctags must succeed.
	CTagsMustSucceed bool

	// Write memory profiles to this file.
	MemProfile string
}

// Builder manages (parallel) creation of uniformly sized shards.
type Builder struct {
	opts     Options
	throttle chan int

	nextShardNum int
	todo         []*zoekt.Document
	size         int

	parser ctags.Parser

	building sync.WaitGroup

	errMu      sync.Mutex
	buildError error

	// temp name => final name for finished shards. We only rename
	// them once all shards succeed to avoid Frankstein corpuses.
	finishedShards map[string]string
}

type finishedShard struct {
	temp, final string
}

// SetDefaults sets reasonable default options.
func (o *Options) SetDefaults() {
	if o.CTags == "" {
		ctags, err := exec.LookPath("universal-ctags")
		if err == nil {
			o.CTags = ctags
		}
	}

	if o.CTags == "" {
		ctags, err := exec.LookPath("ctags-exuberant")
		if err == nil {
			o.CTags = ctags
		}
	}
	if o.Parallelism == 0 {
		o.Parallelism = 1
	}
	if o.SizeMax == 0 {
		o.SizeMax = 128 << 10
	}
	if o.ShardMax == 0 {
		o.ShardMax = 128 << 20
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
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ShardName returns the name the given index shard.
func (o *Options) shardName(n int) (string, error) {
	abs := url.QueryEscape(o.RepositoryDescription.Name)
	if len(abs) > 200 {
		abs = abs[:200] + hashString(abs)[:8]
	}
	return filepath.Join(o.IndexDir,
		fmt.Sprintf("%s_v%d.%05d.zoekt", abs, zoekt.IndexFormatVersion, n)), nil
}

// IndexVersions returns the versions as present in the index, for
// implementing incremental indexing.
func (o *Options) IndexVersions() []zoekt.RepositoryBranch {
	fn, err := o.shardName(0)
	if err != nil {
		return nil
	}
	f, err := os.Open(fn)
	if err != nil {
		return nil
	}

	iFile, err := zoekt.NewIndexFile(f)
	if err != nil {
		return nil
	}
	defer iFile.Close()

	repo, index, err := zoekt.ReadMetadata(iFile)
	if err != nil {
		return nil
	}

	if index.IndexFeatureVersion != zoekt.FeatureVersion {
		return nil
	}

	return repo.Branches
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

	if b.opts.CTags == "" && b.opts.CTagsMustSucceed {
		return nil, fmt.Errorf("ctags binary not found, but CTagsMustSucceed set.")
	}

	if strings.Contains(opts.CTags, "universal-ctags") {
		parser, err := ctags.NewParser(opts.CTags)
		if err != nil && opts.CTagsMustSucceed {
			return nil, fmt.Errorf("ctags.NewParser: %v", err)
		}

		b.parser = parser
	}
	if _, err := b.newShardBuilder(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Builder) AddFile(name string, content []byte) error {
	return b.Add(zoekt.Document{Name: name, Content: content})
}

func (b *Builder) Add(doc zoekt.Document) error {
	if len(doc.Content) > b.opts.SizeMax {
		return nil
	}

	if !zoekt.IsText(doc.Content) {
		return nil
	}

	b.todo = append(b.todo, &doc)
	b.size += len(doc.Name) + len(doc.Content)
	if b.size > b.opts.ShardMax {
		return b.flush()
	}

	return nil
}

func (b *Builder) Finish() error {
	b.flush()
	b.building.Wait()

	if b.buildError != nil {
		for tmp := range b.finishedShards {
			log.Printf("Builder.Finish %s", tmp)
			os.Remove(tmp)
		}
		return b.buildError
	}

	for tmp, final := range b.finishedShards {
		if err := os.Rename(tmp, final); err != nil {
			b.buildError = err
		}
	}

	if b.nextShardNum > 0 {
		b.deleteRemainingShards()
	}
	return b.buildError
}

func (b *Builder) deleteRemainingShards() {
	for {
		shard := b.nextShardNum
		b.nextShardNum++
		name, err := b.opts.shardName(shard)
		if err != nil {
			break
		}

		log.Printf("Builder.deleteRemainingShards %s", name)
		if err := os.Remove(name); os.IsNotExist(err) {
			break
		}
	}
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

	if b.opts.Parallelism > 1 {
		b.building.Add(1)
		go func() {
			b.throttle <- 1
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

var testRe = regexp.MustCompile("test")

type rankedDoc struct {
	*zoekt.Document
	rank []float64
}

func rank(d *zoekt.Document, origIdx int) []float64 {
	test := 0.0
	if testRe.MatchString(d.Name) {
		test = 1.0
	}

	// Smaller is earlier (=better).
	return []float64{
		// Prefer docs that are not tests
		test,

		// With many symbols
		1.0 - squashRange(len(d.Symbols)),

		// With short content
		squashRange(len(d.Content)),

		// With short names
		squashRange(len(d.Name)),

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
	if b.opts.CTags != "" {
		err := ctagsAddSymbols(todo, b.parser, b.opts.CTags)
		if b.opts.CTagsMustSucceed && err != nil {
			return nil, err
		}
		if err != nil {
			log.Printf("ignoring %s error: %v", b.opts.CTags, err)
		}
	}

	name, err := b.opts.shardName(nextShardNum)
	if err != nil {
		return nil, err
	}

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
	desc.SubRepoMap = b.opts.SubRepositories

	shardBuilder, err := zoekt.NewIndexBuilder(&desc)
	if err != nil {
		return nil, err
	}
	return shardBuilder, nil
}

func (b *Builder) writeShard(fn string, ib *zoekt.IndexBuilder) (*finishedShard, error) {
	dir := filepath.Dir(fn)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(dir, filepath.Base(fn))
	if err != nil {
		return nil, err
	}
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0666 &^ umask); err != nil {
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

	log.Printf("finished %s: %d index bytes (overhead %3.1f)", fn, fi.Size(),
		float64(fi.Size())/float64(ib.ContentSize()+1))

	return &finishedShard{f.Name(), fn}, nil
}

var umask os.FileMode
