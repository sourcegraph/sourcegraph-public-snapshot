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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dependenciesStore "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmtest"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	exampleTSFilepath           = "Example.ts"
	exampleJSFilepath           = "Example.js"
	exampleTSFileContents       = "export X; interface X { x: number }"
	exampleJSFileContents       = "var x = 1; var y = 'hello'; x = y;"
	exampleNpmVersion           = "1.0.0"
	exampleNpmVersion2          = "2.0.0-abc"
	exampleNpmVersionedPackage  = "example@1.0.0"
	exampleNpmVersionedPackage2 = "example@2.0.0-abc"
	exampleTgz                  = "example-1.0.0.tgz"
	exampleTgz2                 = "example-2.0.0-abc.tgz"
	exampleNpmPackageURL        = "npm/example"
)

func TestNoMaliciousFilesNpm(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	extractPath := path.Join(dir, "extracted")
	assert.Nil(t, os.Mkdir(extractPath, os.ModePerm))

	tgz := bytes.NewReader(createMaliciousTgz(t))

	s := NewNpmPackagesSyncer(
		schema.NpmPackagesConnection{Dependencies: []string{}},
		NewMockDependenciesStore(),
		nil,
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel now  to prevent any network IO

	dep := &reposource.NpmDependency{NpmPackage: &reposource.NpmPackage{}}
	err = s.commitTgz(ctx, dep, extractPath, tgz)
	require.NotNil(t, err, "malicious tarball should not be committed successfully")

	dirEntries, err := os.ReadDir(extractPath)
	baseline := []string{"harmless.java"}
	assert.Nil(t, err)
	paths := []string{}
	for _, dirEntry := range dirEntries {
		paths = append(paths, dirEntry.Name())
	}
	if !reflect.DeepEqual(baseline, paths) {
		t.Errorf("expected paths: %v\n   found paths:%v", baseline, paths)
	}
}

func createMaliciousTgz(t *testing.T) []byte {
	fileInfos := []fileInfo{
		{harmlessPath, []byte("harmless")},
	}
	for _, filepath := range maliciousPaths {
		fileInfos = append(fileInfos, fileInfo{filepath, []byte("malicious")})
	}
	return createTgz(t, fileInfos)
}

func TestNpmCloneCommand(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})

	tgz1 := createTgz(t, []fileInfo{{exampleJSFilepath, []byte(exampleJSFileContents)}})
	tgz2 := createTgz(t, []fileInfo{{exampleTSFilepath, []byte(exampleTSFileContents)}})

	client := npmtest.MockClient{
		Packages: map[string]*npm.PackageInfo{
			"example": {
				Versions: map[string]*npm.DependencyInfo{
					exampleNpmVersion: {
						Dist: npm.DependencyInfoDist{TarballURL: exampleNpmVersion},
					},
					exampleNpmVersion2: {
						Dist: npm.DependencyInfoDist{TarballURL: exampleNpmVersion2},
					},
				},
			},
		},
		Tarballs: map[string][]byte{
			exampleNpmVersion:  tgz1,
			exampleNpmVersion2: tgz2,
		},
	}
	s := NewNpmPackagesSyncer(
		schema.NpmPackagesConnection{Dependencies: []string{}},
		NewMockDependenciesStore(),
		&client,
	)
	bareGitDirectory := path.Join(dir, "git")
	s.runCloneCommand(t, bareGitDirectory, []string{exampleNpmVersionedPackage})
	checkSingleTag := func() {
		assertCommandOutput(t,
			exec.Command("git", "tag", "--list"),
			bareGitDirectory,
			fmt.Sprintf("v%s\n", exampleNpmVersion))
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNpmVersion, exampleJSFilepath)),
			bareGitDirectory,
			exampleJSFileContents,
		)
	}
	checkSingleTag()

	s.runCloneCommand(t, bareGitDirectory, []string{exampleNpmVersionedPackage, exampleNpmVersionedPackage2})
	checkTagAdded := func() {
		assertCommandOutput(t,
			exec.Command("git", "tag", "--list"),
			bareGitDirectory,
			fmt.Sprintf("v%s\nv%s\n", exampleNpmVersion, exampleNpmVersion2), // verify that a new tag was added
		)
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNpmVersion, exampleJSFilepath)),
			bareGitDirectory,
			exampleJSFileContents,
		)
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNpmVersion2, exampleTSFilepath)),
			bareGitDirectory,
			exampleTSFileContents,
		)
	}
	checkTagAdded()

	s.runCloneCommand(t, bareGitDirectory, []string{exampleNpmVersionedPackage})
	checkTagRemoved := func() {
		assertCommandOutput(t,
			exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNpmVersion, exampleJSFilepath)),
			bareGitDirectory,
			exampleJSFileContents,
		)
		assertCommandOutput(t,
			exec.Command("git", "tag", "--list"),
			bareGitDirectory,
			fmt.Sprintf("v%s\n", exampleNpmVersion), // verify that second tag has been removed.
		)
	}
	checkTagRemoved()

	// Now run the same tests with the database output instead.
	mockStore := NewStrictMockDependenciesStore()
	s.depsStore = mockStore

	mockStore.ListDependencyReposFunc.PushReturn([]dependenciesStore.DependencyRepo{
		{ID: 0, Name: "example", Version: exampleNpmVersion},
	}, nil)
	s.runCloneCommand(t, bareGitDirectory, []string{})
	checkSingleTag()

	mockStore.ListDependencyReposFunc.PushReturn([]dependenciesStore.DependencyRepo{
		{ID: 0, Name: "example", Version: exampleNpmVersion},
		{ID: 1, Name: "example", Version: exampleNpmVersion2},
	}, nil)
	s.runCloneCommand(t, bareGitDirectory, []string{})
	checkTagAdded()

	mockStore.ListDependencyReposFunc.PushReturn([]dependenciesStore.DependencyRepo{
		{ID: 0, Name: "example", Version: "1.0.0"},
	}, nil)
	s.runCloneCommand(t, bareGitDirectory, []string{})
	checkTagRemoved()
}

func (s NpmPackagesSyncer) runCloneCommand(t *testing.T, bareGitDirectory string, dependencies []string) {
	t.Helper()
	packageURL := vcs.URL{URL: url.URL{Path: exampleNpmPackageURL}}
	s.connection.Dependencies = dependencies
	cmd, err := s.CloneCommand(context.Background(), &packageURL, bareGitDirectory)
	require.NoError(t, err)
	require.NoError(t, cmd.Run())
}

func createTgz(t *testing.T, fileInfos []fileInfo) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, fileinfo := range fileInfos {
		require.NoError(t, addFileToTarball(t, tarWriter, fileinfo))
	}

	require.NoError(t, tarWriter.Close())
	require.NoError(t, gzipWriter.Close())

	return buf.Bytes()
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
		{[]string{"f1", "d1/f2", "d1/f3"}, []string{"d1", "f1"}},
	}

	for i, testData := range table {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, os.RemoveAll(dir)) })

			var fileInfos []fileInfo
			for _, path := range testData.paths {
				fileInfos = append(fileInfos, fileInfo{path: path, contents: []byte("x")})
			}

			tgz := bytes.NewReader(createTgz(t, fileInfos))

			require.NoError(t, decompressTgz(tgz, dir))

			have, err := fs.Glob(os.DirFS(dir), "*")
			require.NoError(t, err)

			require.Equal(t, testData.expect, have)
		})
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

	reader := bytes.NewReader(buffer.Bytes())

	outDir, err := os.MkdirTemp("", "decompress-oobfix-")
	require.Nil(t, err)
	defer os.RemoveAll(outDir)

	require.NotPanics(t, func() {
		decompressTgz(reader, outDir)
	})
}
