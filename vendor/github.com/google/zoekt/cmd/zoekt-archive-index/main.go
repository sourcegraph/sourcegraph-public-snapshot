// Command zoekt-archive-index indexes an archive.
//
// Example via github.com:
//
//   zoekt-archive-index -index $PWD/index -incremental -commit b57cb1605fd11ba2ecfa7f68992b4b9cc791934d -name github.com/gorilla/mux https://codeload.github.com/gorilla/mux/legacy.tar.gz/b57cb1605fd11ba2ecfa7f68992b4b9cc791934d
package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/google/zoekt"
	"github.com/google/zoekt/build"
)

func main() {
	var sizeMax = flag.Int("file_limit", 128*1024, "maximum file size")
	var shardLimit = flag.Int("shard_limit", 100<<20, "maximum corpus size for a shard")
	var parallelism = flag.Int("parallelism", 4, "maximum number of parallel indexing processes.")
	name := flag.String("name", "", "The repository name for the archive")
	branch := flag.String("branch", "HEAD", "The branch name for the archive")
	commit := flag.String("commit", "", "The commit sha for the archive. If incremental this will avoid updating shards already at commit")

	indexDir := flag.String("index", build.DefaultDir, "index directory for *.zoekt files.")
	incremental := flag.Bool("incremental", true, "only index changed repositories")
	ctags := flag.Bool("require_ctags", false, "If set, ctags calls must succeed.")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	opts := build.Options{
		Parallelism:      *parallelism,
		SizeMax:          *sizeMax,
		ShardMax:         *shardLimit,
		IndexDir:         *indexDir,
		CTagsMustSucceed: *ctags,
	}
	opts.SetDefaults()

	// For now just make all these args required. In the future we can read
	// extended attributes.
	if *name == "" {
		log.Fatal("-name required")
	}
	if *branch == "" {
		log.Fatal("-branch required")
	}
	if len(*commit) != 40 {
		log.Fatal("-commit required to be absolute commit sha")
	}
	if len(flag.Args()) != 1 {
		log.Fatal("expected argument for archive location")
	}
	archive := flag.Args()[0]

	opts.RepositoryDescription.Name = *name
	opts.RepositoryDescription.Branches = []zoekt.RepositoryBranch{{Name: *branch, Version: *commit}}
	brs := []string{*branch}

	if *incremental {
		versions := opts.IndexVersions()
		if reflect.DeepEqual(versions, opts.RepositoryDescription.Branches) {
			return
		}
	}

	var r io.Reader
	if strings.HasPrefix(archive, "https://") || strings.HasPrefix(archive, "http://") {
		resp, err := http.Get(archive)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		r = resp.Body
		if resp.Header.Get("Content-Type") == "application/x-gzip" {
			r, err = gzip.NewReader(r)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else if archive == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(archive)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = f
	}

	builder, err := build.NewBuilder(opts)
	if err != nil {
		log.Fatal(err)
	}

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// We only care about files
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		// We do not index large files
		if hdr.Size > int64(opts.SizeMax) {
			continue
		}

		contents, err := ioutil.ReadAll(tr)
		if err != nil {
			log.Fatal(err)
		}

		err = builder.Add(zoekt.Document{
			Name:     hdr.Name,
			Content:  contents,
			Branches: brs,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	err = builder.Finish()
	if err != nil {
		log.Fatal(err)
	}
}
