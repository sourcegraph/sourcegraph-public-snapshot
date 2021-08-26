package server

import (
	"archive/zip"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	exampleJar                = "sources.jar"
	exampleByteCodeJar        = "bytes.jar"
	exampleJar2               = "sources2.jar"
	exampleByteCodeJar2       = "bytes2.jar"
	exampleFilePath           = "Example.java"
	exampleClassfilePath      = "Example.class"
	exampleFileContents       = "package example;\npublic class Example {}\n"
	exampleFileContents2      = "package example;\npublic class Example { public static final int x = 42; }\n"
	examplePackageVersion     = "1.0.0"
	examplePackageVersion2    = "2.0.0"
	examplePackageDependency  = "org.example:example:1.0.0"
	examplePackageDependency2 = "org.example:example:2.0.0"
	examplePackageUrl         = "maven/org.example/example"

	// These magic numbers come from the table here https://en.wikipedia.org/wiki/Java_class_file#General_layout
	java5MajorVersion  = 49
	java11MajorVersion = 53
)

func createPlaceholderSourcesJar(t *testing.T, dir, contents, jarName string) {
	t.Helper()
	sourcesPath, err := os.Create(path.Join(dir, jarName))
	assert.Nil(t, err)
	zipWriter := zip.NewWriter(sourcesPath)
	exampleWriter, err := zipWriter.Create(exampleFilePath)
	assert.Nil(t, err)
	_, err = exampleWriter.Write([]byte(contents))
	assert.Nil(t, err)
	assert.Nil(t, zipWriter.Close())
	assert.Nil(t, sourcesPath.Close())
}

func createPlaceholderByteCodeJar(t *testing.T, contents []byte, dir, jarName string) {
	t.Helper()
	byteCodePath, err := os.Create(path.Join(dir, jarName))
	assert.Nil(t, err)
	zipWriter := zip.NewWriter(byteCodePath)
	exampleWriter, err := zipWriter.Create(exampleClassfilePath)
	assert.Nil(t, err)
	_, err = exampleWriter.Write(contents)
	assert.Nil(t, err)
	assert.Nil(t, zipWriter.Close())
	assert.Nil(t, byteCodePath.Close())
}

func assertCommandOutput(t *testing.T, cmd *exec.Cmd, workingDir, expectedOut string) {
	t.Helper()
	cmd.Dir = workingDir
	showOut, err := cmd.Output()
	assert.Nil(t, errors.Wrapf(err, "cmd=%q", cmd))
	if string(showOut) != expectedOut {
		t.Fatalf("got %q, want %q", showOut, expectedOut)
	}
}

func coursierScript(t *testing.T, dir string) string {
	coursierPath, err := os.OpenFile(path.Join(dir, "coursier"), os.O_CREATE|os.O_RDWR, 07777)
	assert.Nil(t, err)
	defer coursierPath.Close()
	script := fmt.Sprintf(`#!/usr/bin/env bash
ARG="$5"
CLASSIFIER="$7"
if [[ "$ARG" =~ "%s" ]]; then
  if [[ "$CLASSIFIER" =~ "sources" ]]; then
    echo "%s"
  else
    echo "%s"
  fi
elif [[ "$ARG" =~ "%s" ]]; then
  if [[ "$CLASSIFIER" =~ "sources" ]]; then
    echo "%s"
  else
    echo "%s"
  fi
else
  echo "Invalid argument $1"
  exit 1
fi
`,
		examplePackageVersion, path.Join(dir, exampleJar), path.Join(dir, exampleByteCodeJar),
		examplePackageVersion2, path.Join(dir, exampleJar2), path.Join(dir, exampleByteCodeJar2),
	)
	_, err = coursierPath.WriteString(script)
	assert.Nil(t, err)
	return coursierPath.Name()
}

func (s JVMPackagesSyncer) runCloneCommand(t *testing.T, bareGitDirectory string, dependencies []string) {
	url := vcs.URL{
		URL: url.URL{Path: examplePackageUrl},
	}
	s.Config.Maven.Dependencies = dependencies
	cmd, err := s.CloneCommand(context.Background(), &url, bareGitDirectory)
	assert.Nil(t, err)
	assert.Nil(t, cmd.Run())
}

func TestJVMCloneCommand(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	createPlaceholderSourcesJar(t, dir, exampleFileContents, exampleJar)
	createPlaceholderByteCodeJar(t,
		[]byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, java5MajorVersion, 0xab}, dir, exampleByteCodeJar)
	createPlaceholderSourcesJar(t, dir, exampleFileContents2, exampleJar2)
	createPlaceholderByteCodeJar(t,
		[]byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, java11MajorVersion, 0xab}, dir, exampleByteCodeJar2)

	coursier.CoursierBinary = coursierScript(t, dir)

	s := JVMPackagesSyncer{
		Config:  &schema.JVMPackagesConnection{Maven: &schema.Maven{Dependencies: []string{}}},
		DBStore: &simpleJVMPackageDBStoreMock{},
	}
	bareGitDirectory := path.Join(dir, "git")

	s.runCloneCommand(t, bareGitDirectory, []string{examplePackageDependency})
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list"),
		bareGitDirectory,
		"v1.0.0\n",
	)
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", examplePackageVersion, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents,
	)

	s.runCloneCommand(t, bareGitDirectory, []string{examplePackageDependency, examplePackageDependency2})
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list"),
		bareGitDirectory,
		"v1.0.0\nv2.0.0\n", // verify that the v2.0.0 tag got added
	)

	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", examplePackageVersion, "lsif-java.json")),
		bareGitDirectory,
		// Assert that Java 8 is used for a library compiled with Java 5.
		fmt.Sprintf(`{"kind":"maven","jvm":"%s","dependencies":["%s"]}`, "8", examplePackageDependency),
	)
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", examplePackageVersion2, "lsif-java.json")),
		bareGitDirectory,
		// Assert that Java 11 is used for a library compiled with Java 11.
		fmt.Sprintf(`{"kind":"maven","jvm":"%s","dependencies":["%s"]}`, "11", examplePackageDependency2),
	)

	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", examplePackageVersion, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents,
	)

	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", examplePackageVersion2, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents2,
	)

	s.runCloneCommand(t, bareGitDirectory, []string{examplePackageDependency})
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", examplePackageVersion, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents,
	)
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list"),
		bareGitDirectory,
		"v1.0.0\n", // verify that the v2.0.0 tag has been removed.
	)
}

type simpleJVMPackageDBStoreMock struct{}

func (m *simpleJVMPackageDBStoreMock) GetJVMDependencyRepos(ctx context.Context, filter dbstore.GetJVMDependencyReposOpts) ([]dbstore.JVMDependencyRepo, error) {
	return []dbstore.JVMDependencyRepo{}, nil
}
