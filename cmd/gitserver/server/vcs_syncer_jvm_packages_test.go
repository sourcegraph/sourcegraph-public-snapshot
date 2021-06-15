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

	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

func check(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}

const (
	exampleJar                = "sources.jar"
	exampleJar2               = "sources2.jar"
	exampleFilePath           = "Example.java"
	exampleFileContents       = "package example;\npublic class Example {}\n"
	exampleFileContents2      = "package example;\npublic class Example { public static final int x = 42; }\n"
	examplePackageVersion     = "1.0.0"
	examplePackageVersion2    = "2.0.0"
	examplePackageDependency  = "org.example:example:1.0.0"
	examplePackageDependency2 = "org.example:example:2.0.0"
)

func createPlaceholderSourcesJar(t *testing.T, dir, contents, jarName string) {
	t.Helper()
	sourcesPath, err := os.Create(path.Join(dir, jarName))
	check(t, err)
	zipWriter := zip.NewWriter(sourcesPath)
	exampleWriter, err := zipWriter.Create(exampleFilePath)
	check(t, err)
	_, err = exampleWriter.Write([]byte(contents))
	check(t, err)
	check(t, zipWriter.Close())
	check(t, sourcesPath.Close())
}

func assertCommandOutput(t *testing.T, cmd *exec.Cmd, workingDir, expectedOut string) {
	t.Helper()
	cmd.Dir = workingDir
	showOut, err := cmd.Output()
	check(t, errors.Wrapf(err, "cmd=%q", cmd))
	if string(showOut) != expectedOut {
		t.Fatalf("got %q, want %q", showOut, expectedOut)
	}
}

func coursierScript(t *testing.T, dir string) string {
	coursierPath, err := os.OpenFile(path.Join(dir, "coursier"), os.O_CREATE|os.O_RDWR, 07777)
	check(t, err)
	defer coursierPath.Close()
	script := fmt.Sprintf(`#!/usr/bin/env bash
ARG="$3"
if [[ "$ARG" =~ "%s" ]]; then
  echo "%s"
elif [[ "$ARG" =~ "%s" ]]; then
  echo "%s"
else
  echo "Invalid argument $1"
  exit 1
fi
`,
		examplePackageVersion, path.Join(dir, exampleJar),
		examplePackageVersion2, path.Join(dir, exampleJar2))
	_, err = coursierPath.WriteString(script)
	check(t, err)
	return coursierPath.Name()
}

func (s JvmPackagesArtifactSyncer) runCloneCommand(t *testing.T, bareGitDirectory string, artifacts []string) {
	url := vcs.URL{
		URL: url.URL{Path: "maven/org.example/example"},
	}
	s.Config.Maven.Artifacts = artifacts
	cmd, err := s.CloneCommand(context.Background(), &url, bareGitDirectory)
	check(t, err)
	check(t, cmd.Run())
}

func TestJvmCloneCommand(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	check(t, err)
	defer os.RemoveAll(dir)

	createPlaceholderSourcesJar(t, dir, exampleFileContents, exampleJar)
	createPlaceholderSourcesJar(t, dir, exampleFileContents2, exampleJar2)

	coursier.CoursierBinary = coursierScript(t, dir)

	s := JvmPackagesArtifactSyncer{Config: &schema.JvmPackagesConnection{
		Maven: &schema.Maven{Artifacts: []string{}},
	}}
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
