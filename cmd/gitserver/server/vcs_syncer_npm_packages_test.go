package server

import (
	"archive/tar"
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

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npmpackages/npm"
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

	s := NPMPackagesSyncer{
		Config:  &schema.NPMPackagesConnection{Dependencies: []string{}},
		DBStore: NewMockDBStore(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel now  to prevent any network IO
	err = s.commitTgz(ctx, reposource.NPMDependency{}, extractPath, tgzPath, s.Config)
	assert.NotNil(t, err)

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

	npm.NPMBinary = npmScript(t, dir)
	s := NPMPackagesSyncer{
		Config:  &schema.NPMPackagesConnection{Dependencies: []string{}},
		DBStore: NewMockDBStore(),
	}
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
	s.DBStore = mockStore

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

func npmScript(t *testing.T, dir string) string {
	t.Helper()
	npmPath, err := os.OpenFile(path.Join(dir, "npm"), os.O_CREATE|os.O_RDWR, 07777)
	assert.Nil(t, err)
	defer func() { assert.Nil(t, npmPath.Close()) }()
	script := fmt.Sprintf(`#!/usr/bin/env bash
CLASSIFIER="$1"
ARG="$2"
if [[ "$CLASSIFIER" =~ "view" ]]; then
  if [[ "$ARG" =~ "%[1]s" || "$ARG" =~ "%[2]s" ]]; then
    echo "$ARG"
  else
    # Mimicking NPM's buggy behavior:
    # https://github.com/npm/cli/issues/3184#issuecomment-963387099
    exit 0
  fi
elif [[ "$CLASSIFIER" =~ "pack" ]]; then
  # Copy tarball as it will be deleted after npm pack's output is parsed.
  if [[ "$ARG" =~ "%[1]s" ]]; then
    mkdir -p %[7]s
    cp %[3]s %[5]s
    echo "[{\"filename\": \"%[5]s\"}]"
  elif [[ "$ARG" =~ "%[2]s" ]]; then
    mkdir -p %[7]s
    cp %[4]s %[6]s
    echo "[{\"filename\": \"%[6]s\"}]"
  fi
else
  echo "invalid arguments for fake npm script: $@"
  exit 1
fi
`,
		exampleNPMVersionedPackage, exampleNPMVersionedPackage2,
		path.Join(dir, exampleTgz), path.Join(dir, exampleTgz2),
		path.Join(dir, "tarballs", exampleTgz), path.Join(dir, "tarballs", exampleTgz2),
		path.Join(dir, "tarballs"),
	)
	_, err = npmPath.WriteString(script)
	assert.Nil(t, err)
	return npmPath.Name()
}

func (s NPMPackagesSyncer) runCloneCommand(t *testing.T, bareGitDirectory string, dependencies []string) {
	t.Helper()
	packageURL := vcs.URL{URL: url.URL{Path: exampleNPMPackageURL}}
	s.Config.Dependencies = dependencies
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
		assert.Nil(t, decompressTgz(tgzPath, dir))
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
