// Package s3cache provides an implementation of httpcache.Cache that stores and
// retrieves data using Amazon S3.
package s3cache // import "sourcegraph.com/sourcegraph/s3cache"

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sqs/s3"
	"github.com/sqs/s3/s3util"
)

// Cache objects store and retrieve data using Amazon S3.
type Cache struct {
	// Config is the Amazon S3 configuration.
	Config s3util.Config

	// BucketURL is the URL to the bucket on Amazon S3, which includes the
	// bucket name and the AWS region. Example:
	// "https://s3-us-west-2.amazonaws.com/mybucket".
	BucketURL string

	// Gzip indicates whether cache entries should be gzipped in Set and
	// gunzipped in Get. If true, cache entry keys will have the suffix ".gz"
	// appended.
	Gzip bool
}

var noLogErrors, _ = strconv.ParseBool(os.Getenv("NO_LOG_S3CACHE_ERRORS"))

func (c *Cache) Get(key string) (resp []byte, ok bool) {
	rdr, err := s3util.Open(c.url(key), &c.Config)
	if err != nil {
		return []byte{}, false
	}
	defer rdr.Close()
	if c.Gzip {
		rdr, err = gzip.NewReader(rdr)
		if err != nil {
			return nil, false
		}
		defer rdr.Close()
	}
	resp, err = ioutil.ReadAll(rdr)
	if err != nil {
		if !noLogErrors {
			log.Printf("s3cache.Get failed: %s", err)
		}
	}
	return resp, err == nil
}

func (c *Cache) Set(key string, resp []byte) {
	w, err := s3util.Create(c.url(key), nil, &c.Config)
	if err != nil {
		if !noLogErrors {
			log.Printf("s3util.Create failed: %s", err)
		}
		return
	}
	defer w.Close()
	if c.Gzip {
		w = gzip.NewWriter(w)
		defer w.Close()
	}
	_, err = w.Write(resp)
	if err != nil {
		if !noLogErrors {
			log.Printf("s3cache.Set failed: %s", err)
		}
	}
}

func (c *Cache) Delete(key string) {
	rdr, err := s3util.Delete(c.url(key), &c.Config)
	if err != nil {
		if !noLogErrors {
			log.Printf("s3cache.Delete failed: %s", err)
		}
		return
	}
	defer rdr.Close()
}

func (c *Cache) url(key string) string {
	key = cacheKeyToObjectKey(key)
	if c.Gzip {
		key += ".gz"
	}
	if strings.HasSuffix(c.BucketURL, "/") {
		return c.BucketURL + key
	}
	return c.BucketURL + "/" + key
}

func cacheKeyToObjectKey(key string) string {
	h := md5.New()
	io.WriteString(h, key)
	return hex.EncodeToString(h.Sum(nil))
}

// New returns a new Cache with underlying storage in Amazon S3. The bucketURL
// is the full URL to the bucket on Amazon S3, including the bucket name and AWS
// region (e.g., "https://s3-us-west-2.amazonaws.com/mybucket").
//
// The environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_KEY are used as the AWS
// credentials. To use different credentials, modify the returned Cache object
// or construct a Cache object manually.
func New(bucketURL string) *Cache {
	return &Cache{
		Config: s3util.Config{
			Keys: &s3.Keys{
				AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
				SecretKey: os.Getenv("AWS_SECRET_KEY"),
			},
			Service: s3.DefaultService,
		},
		BucketURL: bucketURL,
	}
}
