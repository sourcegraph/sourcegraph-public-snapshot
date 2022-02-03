package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmtest"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	exampleTSFilepath           = "Example.ts"
	exampleJSFilepath           = "Example.js"
	exampleTSFileContents       = "export X; interface X { x: number }"
	exampleJSFileContents       = "var x = 1; var y = 'hello'; x = y;"
	exampleNPMVersion           = "1.0.0"
	exampleNPMVersion2          = "2.0.0-abc"
	exampleNPMVersionedPackage  = "example@1.0.0"
	exampleNPMVersionedPackage2 = "example@2.0.0-abc"
	exampleTgz                  = "example-1.0.0.tgz"
	exampleTgz2                 = "example-2.0.0-abc.tgz"
	exampleNPMPackageURL        = "npm/example"
)

func TestNoMaliciousFilesNPM(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	tgzPath := path.Join(dir, "malicious.tgz")
	extractPath := path.Join(dir, "extracted")
	assert.Nil(t, os.Mkdir(extractPath, os.ModePerm))

	createMaliciousTgz(t, tgzPath)

	s := NewNPMPackagesSyncer(
		schema.NPMPackagesConnection{Dependencies: []string{}},
		NewMockDBStore(),
		nil,
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel now  to prevent any network IO
	tgzFile, err := os.Open(tgzPath)
	require.Nil(t, err)
	defer func() { require.Nil(t, tgzFile.Close()) }()
	err = s.commitTgz(ctx, reposource.NPMDependency{}, extractPath, tgzFile)
	require.NotNil(t, err, "malicious tarball should not be committed successfully")

	dirEntries, err := os.ReadDir(extractPath)
	baseline := []string{"src"}
	assert.Nil(t, err)
	paths := []string{}
	for _, dirEntry := range dirEntries {
		paths = append(paths, dirEntry.Name())
	}
	if !reflect.DeepEqual(baseline, paths) {
		t.Errorf("expected paths: %v\n   found paths:%v", baseline, paths)
	}
}

func createMaliciousTgz(t *testing.T, tgzPath string) {
	fileInfos := []fileInfo{
		{harmlessPath, []byte{}},
	}
	for _, filepath := range maliciousPaths {
		fileInfos = append(fileInfos, fileInfo{filepath, []byte{}})
	}
	createTgz(t, tgzPath, fileInfos)
}

func TestNPMCloneCommand(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	defer func() { assert.Nil(t, os.RemoveAll(dir)) }()

	tgzPath := path.Join(dir, exampleTgz)
	createTgz(t, tgzPath, []fileInfo{{exampleJSFilepath, []byte(exampleJSFileContents)}})
	defer func() { assert.Nil(t, os.Remove(tgzPath)) }()
	tgzPath2 := path.Join(dir, exampleTgz2)
	createTgz(t, tgzPath2, []fileInfo{{exampleTSFilepath, []byte(exampleTSFileContents)}})
	defer func() { assert.Nil(t, os.Remove(tgzPath2)) }()

	client := npmtest.MockClient{
		TarballMap: map[string]string{exampleNPMVersionedPackage: tgzPath, exampleNPMVersionedPackage2: tgzPath2},
	}
	s := NewNPMPackagesSyncer(
		schema.NPMPackagesConnection{Dependencies: []string{}},
		NewMockDBStore(),
		&client,
	)
	bareGitDirectory := path.Join(dir, "git")
	s.runCloneCommand(t, bareGitDirectory, []string{exampleNPMVersionedPackage})
	checkSingleTag := func() {
		assertCommandOutput(t,
			exec.Command("git", "tag", "--list"),
			bareGitDirectory,
			fmt.Sprintf("v%s\n", exampleNPMVersion))
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNPMVersion, exampleJSFilepath)),
			bareGitDirectory,
			exampleJSFileContents,
		)
	}
	checkSingleTag()

	s.runCloneCommand(t, bareGitDirectory, []string{exampleNPMVersionedPackage, exampleNPMVersionedPackage2})
	checkTagAdded := func() {
		assertCommandOutput(t,
			exec.Command("git", "tag", "--list"),
			bareGitDirectory,
			fmt.Sprintf("v%s\nv%s\n", exampleNPMVersion, exampleNPMVersion2), // verify that a new tag was added
		)
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNPMVersion, exampleJSFilepath)),
			bareGitDirectory,
			exampleJSFileContents,
		)
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNPMVersion2, exampleTSFilepath)),
			bareGitDirectory,
			exampleTSFileContents,
		)
	}
	checkTagAdded()

	s.runCloneCommand(t, bareGitDirectory, []string{exampleNPMVersionedPackage})
	checkTagRemoved := func() {
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNPMVersion, exampleJSFilepath)),
			bareGitDirectory,
			exampleJSFileContents,
		)
		assertCommandOutput(t,
			exec.Command("git", "tag", "--list"),
			bareGitDirectory,
			fmt.Sprintf("v%s\n", exampleNPMVersion), // verify that second tag has been removed.
		)
	}
	checkTagRemoved()

	// Now run the same tests with the database output instead.
	mockStore := NewStrictMockDBStore()
	s.dbStore = mockStore

	mockStore.GetNPMDependencyReposFunc.PushReturn([]dbstore.NPMDependencyRepo{
		{"example", exampleNPMVersion, 0},
	}, nil)
	s.runCloneCommand(t, bareGitDirectory, []string{})
	checkSingleTag()

	mockStore.GetNPMDependencyReposFunc.PushReturn([]dbstore.NPMDependencyRepo{
		{"example", exampleNPMVersion, 0},
		{"example", exampleNPMVersion2, 1},
	}, nil)
	s.runCloneCommand(t, bareGitDirectory, []string{})
	checkTagAdded()

	mockStore.GetNPMDependencyReposFunc.PushReturn([]dbstore.NPMDependencyRepo{
		{"example", "1.0.0", 0},
	}, nil)
	s.runCloneCommand(t, bareGitDirectory, []string{})
	checkTagRemoved()
}

func (s NPMPackagesSyncer) runCloneCommand(t *testing.T, bareGitDirectory string, dependencies []string) {
	t.Helper()
	packageURL := vcs.URL{URL: url.URL{Path: exampleNPMPackageURL}}
	s.connection.Dependencies = dependencies
	cmd, err := s.CloneCommand(context.Background(), &packageURL, bareGitDirectory)
	assert.Nil(t, err)
	assert.Nil(t, cmd.Run())
}

func createTgz(t *testing.T, tgzPath string, fileInfos []fileInfo) {
	t.Helper()
	ioWriter, err := os.Create(tgzPath)
	assert.Nil(t, err)
	gzipWriter := gzip.NewWriter(ioWriter)
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		assert.Nil(t, tarWriter.Close())
		assert.Nil(t, gzipWriter.Close())
		assert.Nil(t, ioWriter.Close())
	}()
	for _, fileinfo := range fileInfos {
		assert.Nil(t, addFileToTarball(t, tarWriter, fileinfo))
	}
}

func addFileToTarball(t *testing.T, tarWriter *tar.Writer, info fileInfo) error {
	t.Helper()
	header, err := tar.FileInfoHeader(&info, "")
	if err != nil {
		return err
	}
	header.Name = info.path
	if err = tarWriter.WriteHeader(header); err != nil {
		return errors.Wrapf(err, "failed to write header for %s", info.path)
	}
	_, err = tarWriter.Write(info.contents)
	return err
}

type fileInfo struct {
	path     string
	contents []byte
}

var _ fs.FileInfo = &fileInfo{}

func (info *fileInfo) Name() string       { return path.Base(info.path) }
func (info *fileInfo) Size() int64        { return int64(len(info.contents)) }
func (info *fileInfo) Mode() fs.FileMode  { return 0600 }
func (info *fileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (info *fileInfo) IsDir() bool        { return false }
func (info *fileInfo) Sys() interface{}   { return nil }

func TestDecompressTgz(t *testing.T) {
	table := []struct {
		paths  []string
		expect []string
	}{
		// Check that stripping the outermost shared directory works if all
		// paths have a common outermost directory.
		{[]string{"d/f1", "d/f2"}, []string{"f1", "f2"}},
		{[]string{"d1/d2/f1", "d1/d2/f2"}, []string{"d2"}},
		{[]string{"d1/f1", "d2/f2", "d3/f3"}, []string{"d1", "d2", "d3"}},
	}

	for _, testData := range table {
		dir, err := os.MkdirTemp("", "")
		assert.Nil(t, err)
		defer func() { assert.Nil(t, os.RemoveAll(dir)) }()

		tarballName := "test.tgz"
		tgzPath := path.Join(dir, tarballName)
		// Creating a tarball with empty files fails???
		fileInfos := []fileInfo{}
		for _, path := range testData.paths {
			fileInfos = append(fileInfos, fileInfo{path: path, contents: []byte("x")})
		}
		createTgz(t, tgzPath, fileInfos)
		tgzFile, err := os.Open(tgzPath)
		require.Nil(t, err)
		defer tgzFile.Close()
		assert.Nil(t, decompressTgz(namedReadSeeker{tgzPath, tgzFile}, dir))
		dirEntries, err := os.ReadDir(dir)
		assert.Nil(t, err)
		dirEntryNames := []string{}
		for _, entry := range dirEntries {
			if entry.Name() == tarballName {
				continue
			}
			dirEntryNames = append(dirEntryNames, entry.Name())
		}
		assert.True(t, reflect.DeepEqual(dirEntryNames, testData.expect))
	}
}

// Regression test for: https://github.com/sourcegraph/sourcegraph/issues/30554
func TestDecompressTgzNoOOB(t *testing.T) {
	testCases := [][]tar.Header{
		{
			{Typeflag: tar.TypeDir, Name: "non-empty"},
			{Typeflag: tar.TypeReg, Name: "non-empty/f1"},
		},
		{
			{Typeflag: tar.TypeDir, Name: "empty"},
			{Typeflag: tar.TypeReg, Name: "non-empty/f1"},
		},
		{
			{Typeflag: tar.TypeDir, Name: "empty"},
			{Typeflag: tar.TypeDir, Name: "non-empty/"},
			{Typeflag: tar.TypeReg, Name: "non-empty/f1"},
		},
	}

	for _, testCase := range testCases {
		testDecompressTgzNoOOBImpl(t, testCase)
	}
}

func testDecompressTgzNoOOBImpl(t *testing.T, entries []tar.Header) {
	buffer := bytes.NewBuffer([]byte{})

	gzipWriter := gzip.NewWriter(buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		tarWriter.WriteHeader(&entry)
		if entry.Typeflag == tar.TypeReg {
			tarWriter.Write([]byte("filler"))
		}
	}
	tarWriter.Close()
	gzipWriter.Close()

	readSeeker := bytes.NewReader(buffer.Bytes())

	outDir, err := os.MkdirTemp("", "decompress-oobfix-")
	require.Nil(t, err)
	defer os.RemoveAll(outDir)

	require.NotPanics(t, func() {
		decompressTgz(namedReadSeeker{"buffer", readSeeker}, outDir)
	})
}
