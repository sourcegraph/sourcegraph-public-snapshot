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
		Config: &schema.NPMPackagesConnection{Dependencies: []string{}},
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

	createPlaceholderSourcesTgz(t, dir, exampleJSFileContents, exampleJSFilepath, exampleTgz)
	createPlaceholderSourcesTgz(t, dir, exampleTSFileContents, exampleTSFilepath, exampleTgz2)

	npm.NPMBinary = npmScript(t, dir)
	s := NPMPackagesSyncer{
		Config: &schema.NPMPackagesConnection{Dependencies: []string{}},
	}
	bareGitDirectory := path.Join(dir, "git")
	s.runCloneCommand(t, bareGitDirectory, []string{exampleNPMVersionedPackage})
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list"),
		bareGitDirectory,
		"v1.0.0\n")
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNPMVersion, exampleJSFilepath)),
		bareGitDirectory,
		exampleJSFileContents,
	)

	s.runCloneCommand(t, bareGitDirectory, []string{exampleNPMVersionedPackage, exampleNPMVersionedPackage2})

	assertCommandOutput(t,
		exec.Command("git", "tag", "--list"),
		bareGitDirectory,
		"v1.0.0\nv2.0.0-abc\n", // verify that the v2.0.0 tag got added
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

	s.runCloneCommand(t, bareGitDirectory, []string{exampleNPMVersionedPackage})
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleNPMVersion, exampleJSFilepath)),
		bareGitDirectory,
		exampleJSFileContents,
	)
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list"),
		bareGitDirectory,
		"v1.0.0\n", // verify that the v2.0.0 tag has been removed.
	)
}

func createPlaceholderSourcesTgz(t *testing.T, dir, contents, filename, tgzFilename string) {
	createTgz(t, path.Join(dir, tgzFilename),
		[]fileInfo{
			{filename, []byte(contents)},
		})
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
  if [[ "$ARG" =~ "%[1]s" ]]; then
    echo "[{\"filename\": \"%[3]s\"}]"
  elif [[ "$ARG" =~ "%[2]s" ]]; then
    echo "[{\"filename\": \"%[4]s\"}]"
  fi
else
  echo "invalid arguments for fake npm script: $@"
  exit 1
fi
`,
		exampleNPMVersionedPackage, exampleNPMVersionedPackage2,
		path.Join(dir, exampleTgz), path.Join(dir, exampleTgz2),
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
