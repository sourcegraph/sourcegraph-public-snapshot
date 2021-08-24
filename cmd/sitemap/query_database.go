package main

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	bolt "go.etcd.io/bbolt"
)

var requestsBucket = []byte("requests")

type requestValue struct {
	Time     time.Time
	Response []byte
}

// queryDatabase is a bolt DB key-value store which contains all of the GraphQL queries and
// responses that we need to make in order to generate the sitemap. This is basically just a
// glorified HTTP query disk cache.
type queryDatabase struct {
	handle *bolt.DB
}

// request performs a request to fetch `key`. If it already exists in the cache, the cached value
// is returned. Otherwise, fetch is invoked and the result is stored and returned if not an error.
func (db *queryDatabase) request(key interface{}, fetch func() ([]byte, error)) ([]byte, error) {
	// Our key (i.e. the info needed to perform the request) will be the key in our bucket, as a
	// JSON string.
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return nil, errors.Wrap(err, "Marshal")
	}

	// Check if the bucket already has the request response or not.
	var value []byte
	err = db.handle.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(requestsBucket)
		if bucket != nil {
			value = bucket.Get(keyBytes)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "View")
	}
	if value != nil {
		var rv requestValue
		if err := json.Unmarshal(value, &rv); err != nil {
			return nil, errors.Wrap(err, "Unmarshal")
		}
		return value, nil
	}

	// Fetch and store the result.
	result, err := fetch()
	if err != nil {
		return nil, errors.Wrap(err, "fetch")
	}
	err = db.handle.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(requestsBucket)
		if err != nil {
			return errors.Wrap(err, "CreateBucketIfNotExists")
		}
		bucket.Put(keyBytes, result)
		return nil
	})
	return result, nil
}

func (db *queryDatabase) close() error {
	return db.handle.Close()
}

func openQueryDatabase(path string) (*queryDatabase, error) {
	db := &queryDatabase{}

	var err error
	db.handle, err = bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, errors.Wrap(err, "bolt.Open")
	}
	return db, nil
}
