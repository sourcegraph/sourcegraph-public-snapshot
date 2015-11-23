package pgsql

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// TODO(slimsag): in the case of errors we must return zero-value non-nil
// structs:
//
//  2015/11/21 10:31:18 grpc: Server failed to encode response proto: Marshal called with nil
//
// Identify why this is and fix it.

func init() {
	// TODO(slimsag): the below doesn't play nicely with `src drop` etc. Use more
	// standard creation scheme like other alter type code that works with hstore
	// type?
	Schema.CreateSQL = append(Schema.CreateSQL,
		"CREATE TABLE appdata (name text, objects hstore)",
	)
}

// Storage is a DB-backed implementation of the Storage store.
type Storage struct{}

var _ store.Storage = (*Storage)(nil)

// Get implements the store.Storage interface.
func (s *Storage) Get(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageValue, error) {
	// Validate the key. We don't care what it is, as long as it's something.
	if opt.Key == "" {
		return &sourcegraph.StorageValue{}, errors.New("key must be specified")
	}

	// Compose the bucket key.
	bucket, err := bucketKey(opt.Bucket)
	if err != nil {
		return &sourcegraph.StorageValue{}, err
	}

	var value []string
	err = dbh(ctx).Select(&value, "SELECT objects -> $1 FROM appdata WHERE name = $2", url.QueryEscape(opt.Key), bucket)
	if err != nil {
		return &sourcegraph.StorageValue{}, err
	}
	if len(value) != 1 {
		return &sourcegraph.StorageValue{}, errors.New("no such object")
	}
	v, err := base64.StdEncoding.DecodeString(value[0])
	return &sourcegraph.StorageValue{Value: []byte(v)}, nil
}

// Put implements the store.Storage interface.
func (s *Storage) Put(ctx context.Context, opt *sourcegraph.StoragePutOp) (*pbtypes.Void, error) {
	// Validate the key. We don't care what it is, as long as it's something.
	if opt.Key.Key == "" {
		return &pbtypes.Void{}, errors.New("key must be specified")
	}

	// Compose the bucket key.
	bucket, err := bucketKey(opt.Key.Bucket)
	if err != nil {
		return &pbtypes.Void{}, err
	}

	// Put a K/V pair into the bucket creating the bucket if needed.
	_, err = dbh(ctx).Exec(
		`WITH upsert AS (UPDATE appdata SET objects = objects || $1 WHERE name = $2 RETURNING *)
	  INSERT INTO appdata (name, objects) SELECT $2, $1 WHERE NOT EXISTS (SELECT * FROM upsert)`,
		hQuote(url.QueryEscape(opt.Key.Key))+"=>"+hQuote(base64.StdEncoding.EncodeToString(opt.Value)), bucket)
	return &pbtypes.Void{}, err
}

// Delete implements the store.Storage interface.
func (s *Storage) Delete(ctx context.Context, opt *sourcegraph.StorageKey) (*pbtypes.Void, error) {
	// Compose the bucket key.
	bucket, err := bucketKey(opt.Bucket)
	if err != nil {
		return &pbtypes.Void{}, err
	}

	if opt.Bucket.Name != "" {
		// Delete the entire bucket.
		_, err := dbh(ctx).Exec("DELETE FROM appdata WHERE name = $1", bucket)
		return &pbtypes.Void{}, err
	}

	// Delete just a single key.
	_, err = dbh(ctx).Exec("UPDATE appdata SET objects = delete(objects, $1) WHERE name = $2", url.QueryEscape(opt.Key), bucket)
	return &pbtypes.Void{}, err
}

// Exists implements the store.Storage interface.
func (s *Storage) Exists(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageExists, error) {
	// Validate the key. We don't care what it is, as long as it's something.
	if opt.Key == "" {
		return &sourcegraph.StorageExists{}, errors.New("key must be specified")
	}

	// Compose the bucket key.
	bucket, err := bucketKey(opt.Bucket)
	if err != nil {
		return &sourcegraph.StorageExists{}, err
	}

	var exists bool
	err = dbh(ctx).Select(&exists, "SELECT exist(objects, $1) FROM appdata WHERE name = $2", url.QueryEscape(opt.Key), bucket)
	if err != nil {
		return &sourcegraph.StorageExists{}, err
	}
	return &sourcegraph.StorageExists{Exists: exists}, nil
}

// List implements the store.Storage interface.
func (s *Storage) List(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageList, error) {
	// Compose the bucket key.
	bucket, err := bucketKey(opt.Bucket)
	if err != nil {
		return &sourcegraph.StorageList{}, err
	}

	var rawKeys []string
	err = dbh(ctx).Select(&rawKeys, "SELECT skeys(objects) FROM appdata WHERE name = $1", bucket)
	if err != nil {
		return &sourcegraph.StorageList{}, err
	}

	// Decode keys.
	keys := make([]string, len(rawKeys))
	for i, raw := range rawKeys {
		keys[i], err = url.QueryUnescape(raw)
		if err != nil {
			return &sourcegraph.StorageList{}, err
		}
	}
	return &sourcegraph.StorageList{Keys: keys}, nil
}

// hQuote takes an input string and makes it a valid hstore quoted string.
func hQuote(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	return `"` + strings.Replace(s, "\"", "\\\"", -1) + `"`
}

// bucketKey returns the key for a bucket. The composed key will be in the
// format of:
//
//  <RepoURI|global>-<AppName>-<BucketName>
//
// For example:
//
//  repo-github.com/foo/bar-issues-comments
//  global-issues-comments
//
// It returns an error only if the app name or bucket name are invalid.
func bucketKey(bucket *sourcegraph.StorageBucket) (string, error) {
	// Be very strict about what names may look like. The goal here is to keep
	// them human-readable and also make errors obvious.
	//
	// TODO(slimsag): duplicated in ../fs/storage.go
	validateName := func(field, v string) error {
		if !isAlphaNumeric(v) {
			return fmt.Errorf("%s must only be alphanumeric with underscores and dashes", field)
		}
		if strings.TrimSpace(v) != v {
			return fmt.Errorf("%s may not start or end with a space", field)
		}
		if v == "" {
			return fmt.Errorf("%s must be specified", field)
		}
		return nil
	}
	if err := validateName("app name", bucket.AppName); err != nil {
		return "", err
	}
	if err := validateName("bucket name", bucket.Name); err != nil {
		return "", err
	}

	// Determine the location, global or local to a repo.
	location := "global"
	if bucket.Repo != "" {
		location = "repo-" + bucket.Repo
	}

	return location + "-" + bucket.AppName + "-" + bucket.Name, nil
}

// isAlphaNumeric reports whether the string is alphabetic, digit, underscore,
// or dash.
//
// TODO(slimsag): duplicated in ../fs/storage.go
func isAlphaNumeric(s string) bool {
	for _, r := range s {
		if r != '_' && r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
