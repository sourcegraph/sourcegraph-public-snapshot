package server

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	exampleJar               = "sources.jar"
	exampleByteCodeJar       = "bytes.jar"
	exampleJar2              = "sources2.jar"
	exampleByteCodeJar2      = "bytes2.jar"
	exampleFilePath          = "Example.java"
	exampleClassfilePath     = "Example.class"
	exampleFileContents      = "package example;\npublic class Example {}\n"
	exampleFileContents2     = "package example;\npublic class Example { public static final int x = 42; }\n"
	exampleVersion           = "1.0.0"
	exampleVersion2          = "2.0.0"
	exampleVersionedPackage  = "org.example:example:1.0.0"
	exampleVersionedPackage2 = "org.example:example:2.0.0"
	examplePackageUrl        = "maven/org.example/example"

	// These magic numbers come from the table here https://en.wikipedia.org/wiki/Java_class_file#General_layout
	java5MajorVersion  = 49
	java11MajorVersion = 53
)

func createPlaceholderJar(t *testing.T, dir string, contents []byte, jarName, contentPath string) {
	t.Helper()
	jarPath, err := os.Create(path.Join(dir, jarName))
	require.Nil(t, err)
	zipWriter := zip.NewWriter(jarPath)
	exampleWriter, err := zipWriter.Create(contentPath)
	require.Nil(t, err)
	_, err = exampleWriter.Write(contents)
	require.Nil(t, err)
	require.Nil(t, zipWriter.Close())
	require.Nil(t, jarPath.Close())
}

func createPlaceholderSourcesJar(t *testing.T, dir, contents, jarName string) {
	t.Helper()
	createPlaceholderJar(t, dir, []byte(contents), jarName, exampleFilePath)
}

func createPlaceholderByteCodeJar(t *testing.T, contents []byte, dir, jarName string) {
	t.Helper()
	createPlaceholderJar(t, dir, contents, jarName, exampleClassfilePath)
}

func assertCommandOutput(t *testing.T, cmd *exec.Cmd, workingDir, expectedOut string) {
	t.Helper()
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error running command %q: %v with output: %q", cmd.String(), err, out)
	}
	if string(out) != expectedOut {
		t.Fatalf("unexpected command output: want=%q, got=%q", expectedOut, out)
	}
}

func coursierScript(t *testing.T, dir string) string {
	coursierPath, err := os.OpenFile(path.Join(dir, "coursier"), os.O_CREATE|os.O_RDWR, 0o7777)
	require.Nil(t, err)
	defer coursierPath.Close()

	script := fmt.Sprintf(`#!/usr/bin/env bash
ARG="$5"
CLASSIFIER="$7"

if [[ $1 =~ "complete-dep" ]]; then
	echo "%s"
	echo "%s"
elif [[ "$ARG" =~ "%s" ]]; then
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
		exampleVersion, exampleVersion2,
		exampleVersion, path.Join(dir, exampleJar), path.Join(dir, exampleByteCodeJar),
		exampleVersion2, path.Join(dir, exampleJar2), path.Join(dir, exampleByteCodeJar2),
	)
	_, err = coursierPath.WriteString(script)
	require.Nil(t, err)
	return coursierPath.Name()
}

var maliciousPaths = []string{
	// Absolute paths
	"/sh", "/usr/bin/sh",
	// Paths into .git which may trigger when git runs a hook
	".git/blah", ".git/hooks/pre-commit",
	// Paths into a nested .git which may trigger when git runs a hook
	"src/.git/blah", "src/.git/hooks/pre-commit",
	// Relative paths which stray outside
	"../foo/../bar", "../../../usr/bin/sh",
}

const harmlessPath = "src/harmless.java"

func TestNoMaliciousFiles(t *testing.T) {
	dir := t.TempDir()

	extractPath := path.Join(dir, "extracted")
	require.Nil(t, os.Mkdir(extractPath, os.ModePerm))

	s := jvmPackagesSyncer{
		config: &schema.JVMPackagesConnection{Maven: &schema.Maven{Dependencies: []string{}}},
		fetch: func(ctx context.Context, config *schema.JVMPackagesConnection, dependency *reposource.MavenVersionedPackage) (sourceCodeJarPath string, err error) {
			jarPath := path.Join(dir, "sampletext.zip")
			createMaliciousJar(t, jarPath)
			return jarPath, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel now  to prevent any network IO
	dep := &reposource.MavenVersionedPackage{MavenModule: &reposource.MavenModule{}}
	err := s.Download(ctx, extractPath, dep)
	require.NotNil(t, err)

	dirEntries, err := os.ReadDir(extractPath)
	require.Nil(t, err)

	_, err = filepath.EvalSymlinks(filepath.Join(extractPath, "symlink"))
	require.Error(t, err)

	baseline := map[string]int{"lsif-java.json": 0, strings.Split(harmlessPath, string(os.PathSeparator))[0]: 0}
	paths := map[string]int{}
	for _, dirEntry := range dirEntries {
		paths[dirEntry.Name()] = 0
	}
	if !reflect.DeepEqual(baseline, paths) {
		t.Errorf("expected paths: %v\n   found paths:%v", baseline, paths)
	}
}

func createMaliciousJar(t *testing.T, name string) {
	f, err := os.Create(name)
	require.Nil(t, err)
	defer f.Close()
	writer := zip.NewWriter(f)
	defer writer.Close()

	for _, filepath := range maliciousPaths {
		_, err = writer.Create(filepath)
		require.Nil(t, err)
	}

	os.Symlink("/etc/passwd", "symlink")
	defer os.Remove("symlink")

	fi, _ := os.Lstat("symlink")
	header, _ := zip.FileInfoHeader(fi)
	_, err = writer.CreateRaw(header)

	require.Nil(t, err)

	_, err = writer.Create(harmlessPath)
	require.Nil(t, err)
}

func TestJVMCloneCommand(t *testing.T) {
	logger := logtest.Scoped(t)
	dir := t.TempDir()

	createPlaceholderSourcesJar(t, dir, exampleFileContents, exampleJar)
	createPlaceholderByteCodeJar(t,
		[]byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, java5MajorVersion, 0xab}, dir, exampleByteCodeJar)
	createPlaceholderSourcesJar(t, dir, exampleFileContents2, exampleJar2)
	createPlaceholderByteCodeJar(t,
		[]byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, java11MajorVersion, 0xab}, dir, exampleByteCodeJar2)

	coursier.CoursierBinary = coursierScript(t, dir)

	depsSvc := dependencies.TestService(database.NewDB(logger, dbtest.NewDB(logger, t)), nil)
	s := NewJVMPackagesSyncer(&schema.JVMPackagesConnection{Maven: &schema.Maven{Dependencies: []string{}}}, depsSvc).(*vcsPackagesSyncer)
	bareGitDirectory := path.Join(dir, "git")

	s.runCloneCommand(t, examplePackageUrl, bareGitDirectory, []string{exampleVersionedPackage})
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list", "--format=%(*objectname):%(refname:short)"),
		bareGitDirectory,
		"9d94a147bd5968393e189540e50dd4d315704cc3:v1.0.0\n"+gitdomain.EmptyGitObject+":v2.0.0\n",
	)
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleVersion, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents,
	)

	s.runCloneCommand(t, examplePackageUrl, bareGitDirectory, []string{exampleVersionedPackage, exampleVersionedPackage2})
	assertCommandOutput(t,
		exec.Command("git", "tag", "--list", "--format=%(*objectname):%(refname:short)"),
		bareGitDirectory,
		"9d94a147bd5968393e189540e50dd4d315704cc3:v1.0.0\n7b2e2f2b8910b0ad20a8b0f745ad50eb57ad6d7d:v2.0.0\n",
	)

	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleVersion, "lsif-java.json")),
		bareGitDirectory,
		// Assert that Java 8 is used for a library compiled with Java 5.
		fmt.Sprintf(`{"kind":"maven","jvm":"%s","dependencies":["%s"]}`, "8", exampleVersionedPackage),
	)
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleVersion2, "lsif-java.json")),
		bareGitDirectory,
		// Assert that Java 11 is used for a library compiled with Java 11.
		fmt.Sprintf(`{"kind":"maven","jvm":"%s","dependencies":["%s"]}`, "11", exampleVersionedPackage2),
	)

	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleVersion, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents,
	)

	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleVersion2, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents2,
	)

	s.runCloneCommand(t, examplePackageUrl, bareGitDirectory, []string{exampleVersionedPackage})
	assertCommandOutput(t,
		exec.Command("git", "show", fmt.Sprintf("v%s:%s", exampleVersion, exampleFilePath)),
		bareGitDirectory,
		exampleFileContents,
	)
}
