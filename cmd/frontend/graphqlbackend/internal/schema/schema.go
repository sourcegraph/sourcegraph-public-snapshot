package schema

import (
	"io/ioutil"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlfile"
)

// ReadFromDisk reads all GraphQL schemas matched by the
// given glob concatenated into a single byte slice with
// internal comments stripped out.
func ReadFromDisk(glob string) ([]byte, error) {
	fs, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	var schema []byte
	for _, f := range fs {
		bs, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		schema = append(schema, '\n')
		schema = append(schema, bs...)
	}

	schema, err = graphqlfile.StripInternalComments(schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}
