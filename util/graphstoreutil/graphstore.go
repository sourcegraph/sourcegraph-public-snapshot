package graphstoreutil

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/rwvfs/cloudstoragevfs"
	"sourcegraph.com/sourcegraph/s3vfs"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil"
	"sourcegraph.com/sourcegraph/srclib/store"
)

var keepAliveTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,

	// Allow more keep-alive connections per host to avoid
	// ephemeral port exhaustion due to getting stuck in
	// TIME_WAIT. Some systems have a very limited ephemeral port
	// supply (~1024). 30 connections is perfectly reasonable,
	// since this client will only ever hit one host.
	MaxIdleConnsPerHost: 30,
}

// New creates a new multi-repo store using the given HTTP transport
// (for stores accessed via HTTP).
func New(graphstore string, transport http.RoundTripper) store.MultiRepoStoreImporterIndexer {
	var rp store.RepoPaths
	if strings.HasPrefix(graphstore, "s3://") {
		// Use hashed paths for better S3 perf.
		bucket, err := parseS3Bucket(graphstore)
		if err != nil {
			log.Fatal("graphstoreutil.New:", err)
		}
		rp = &s3RepoPaths{
			bucket:    bucket,
			config:    &s3vfs.DefaultS3Config,
			client:    http.DefaultClient,
			delim:     "/",
			RepoPaths: &EvenlyDistributedRepoPaths{},
		}
	}
	if strings.HasPrefix(graphstore, "gcloudstorage://") {
		rp = &EvenlyDistributedRepoPaths{}
	}
	s := store.NewFSMultiRepoStore(VFS(graphstore, transport), &store.FSMultiRepoStoreConf{
		RepoPaths: rp,
	})
	return withMetrics(graphstore, s)
}

// VFS returns the virtual filesystem where the graph data is stored.
func VFS(graphstore string, transport http.RoundTripper) rwvfs.WalkableFileSystem {
	if transport == nil {
		transport = keepAliveTransport
	}

	var fs rwvfs.FileSystem
	if strings.HasPrefix(graphstore, "http://") || strings.HasPrefix(graphstore, "https://") {
		url, err := url.Parse(graphstore)
		if err != nil {
			log.Fatalf("Error parsing graphstore URL %q: %s", graphstore, err)
		}
		fs = rwvfs.HTTP(url, &http.Client{Transport: transport})
	} else if strings.HasPrefix(graphstore, "gcloudstorage://") {
		var err error
		fs, err = cloudstoragevfs.NewDefault(strings.TrimPrefix(graphstore, "gcloudstorage://"))
		if err != nil {
			log.Fatal("graphstoreutil.VFS:", err)
		}
		fs = withVFSCache(fs)
	} else if s3URL, err := parseS3Bucket(graphstore); err != nil {
		log.Fatal("Parsing graphstore:", err)
	} else if s3URL != nil {
		s3Conf := s3vfs.DefaultS3Config
		s3Conf.Client = &http.Client{Transport: transport}
		if logVFS {
			s3Conf.Client.Transport = &httputil.LoggedTransport{Writer: os.Stderr, Transport: s3Conf.Client.Transport}
		}
		fs = s3vfs.S3(s3URL, &s3Conf)
	} else {
		fs = rwvfs.OS(graphstore)
	}

	if fs, ok := fs.(interface {
		CreateParentDirs(bool)
	}); ok {
		fs.CreateParentDirs(true)
	}

	if logVFS {
		greenbg := func(s string) string { return "\x1b[42;37;1m" + s + "\x1b[39;49m" }
		fs = rwvfs.Logged(log.New(os.Stderr, greenbg("GFS: "), log.Lmicroseconds), fs)
	}

	return rwvfs.Walkable(fs)
}

var logVFS, _ = strconv.ParseBool(os.Getenv("VFSLOG"))

// parseS3Bucket parses s as an Amazon S3 bucket URL, either of the
// form "s3://mybucket" or
// "https://mybucket.s3-us-west-2.amazonaws.com". If s is not an S3
// bucket URL, the returned URL and error are both nil
func parseS3Bucket(s string) (httpURL *url.URL, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		const suffix = ".amazonaws.com"
		if !strings.HasSuffix(u.Host, suffix) {
			return nil, fmt.Errorf("S3 bucket URL hostname must have %q suffix (was %q)", suffix, u.Host)
		}
		return u, nil
	case "s3":
		region := os.Getenv("AWS_REGION")
		if region == "" {
			return nil, fmt.Errorf("Missing S3 bucket region; set your AWS_REGION environment variable")
		}
		// Need to use CNAME, not bucket path, because s3util.NewFile assumes that.
		return &url.URL{
			Scheme: "https",
			Host:   u.Host + ".s3-" + region + ".amazonaws.com",
			Path:   u.Path,
		}, nil
	}
	return nil, nil
}
