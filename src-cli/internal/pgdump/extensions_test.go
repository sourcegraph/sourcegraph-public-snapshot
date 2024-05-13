package pgdump

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartialCopyWithoutExtensions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test doesn't work on Windows of weirdness with t.TempDir() handling")
	}

	// Create test data - there is no stdlib in-memory io.ReadSeeker implementation
	src, err := os.Create(filepath.Join(t.TempDir(), t.Name()))
	require.NoError(t, err)
	_, err = src.WriteString(`-- Some comment

CREATE EXTENSION foobar

COMMENT ON EXTENSION barbaz

CREATE TYPE asdf

CREATE TABLE robert (
	...
)

CREATE TABLE bobhead (
	...
)`)
	require.NoError(t, err)
	_, err = src.Seek(0, io.SeekStart)
	require.NoError(t, err)

	// Set up target to assert against
	var dst bytes.Buffer

	// Perform partial copy
	_, err = PartialCopyWithoutExtensions(&dst, src, func(i int64) {})
	assert.NoError(t, err)

	// Copy rest of contents
	_, err = io.Copy(&dst, src)
	assert.NoError(t, err)

	expected := `-- Some comment

CREATE EXTENSION foobar

-- COMMENT ON EXTENSION barbaz

CREATE TYPE asdf

CREATE TABLE robert (
	...
)

CREATE TABLE bobhead (
	...
)`
	assert.Equal(t, expected, dst.String())
}
