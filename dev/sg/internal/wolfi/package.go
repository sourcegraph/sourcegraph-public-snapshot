package wolfi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PackageRepoConfig represents config for a local package repo
type PackageRepoConfig struct {
	PackageDir  string
	Arch        string
	KeyDir      string
	KeyFilename string
	KeyFilepath string
}

// InitLocalPackageRepo initializes a local package repository
func InitLocalPackageRepo() (PackageRepoConfig, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return PackageRepoConfig{}, err
	}

	c := PackageRepoConfig{
		PackageDir:  filepath.Join(repoRoot, "wolfi-packages/local-repo/packages"),
		Arch:        "x86_64",
		KeyDir:      filepath.Join(repoRoot, "wolfi-packages/local-repo/keys"),
		KeyFilename: "sourcegraph-dev-local.rsa",
	}
	c.KeyFilepath = filepath.Join(c.KeyDir, c.KeyFilename)

	// Make directories
	if err := os.MkdirAll(filepath.Join(c.PackageDir, c.Arch), os.ModePerm); err != nil {
		return c, err
	}
	if err := os.MkdirAll(c.KeyDir, os.ModePerm); err != nil {
		return c, err
	}

	// Generate keys for local repository
	f, err := os.Open(c.KeyFilepath)
	if err == nil {
		f.Close()
	} else if os.IsNotExist(err) {
		fmt.Printf("Error is %s\n", err)
		if err := c.GenerateKeypair(); err != nil {
			return c, err
		}
	} else {
		return c, err
	}

	return c, nil
}

// GenerateKeypair generates a new RSA keypair for signing packages
func (c PackageRepoConfig) GenerateKeypair() error {
	// Run docker command
	fmt.Printf(" 🗝️  Initializing keypair for local repo... under %s\n", c.KeyDir)

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/keys", c.KeyDir),
		"cgr.dev/chainguard/melange", "keygen",
		fmt.Sprintf("/keys/%s", c.KeyFilename),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to generate keypair")
	}

	fmt.Printf(" 🔐 Keypair initialized\n")

	return nil
}

// SetupPackageBuild sets up the build directory for a package
func SetupPackageBuild(name string) (manifestBaseName string, buildDir string, err error) {
	// Search for `.yaml` suffix in `name` and remove it if present

	// Get root of repo
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", "", errors.Wrap(err, "unable to get repository root")
	}

	// Strip .yaml suffix if it exists
	manifestBaseName = strings.Replace(name, ".yaml", "", 1)
	manifestFileName := manifestBaseName + ".yaml"

	// Check manfest exists
	manifestPath := filepath.Join(repoRoot, "wolfi-packages", manifestFileName)
	manifestDir := filepath.Join(repoRoot, "wolfi-packages", manifestBaseName)

	if _, err := os.Open(manifestPath); err != nil {
		return "", "", errors.Wrap(err, "manifest file does not exist")
	}

	// Create a temp dir
	buildDir, err = os.MkdirTemp("/tmp", "sg-wolfi-package-tmp")
	if err != nil {
		return "", "", errors.Wrap(err, "unable to create temporary build directory")
	}

	// Copy files
	cmd := exec.Command("cp", "-r", manifestPath, buildDir)
	err = cmd.Run()
	if err != nil {
		return "", "", errors.Wrap(err, "error copying build config to temp dir")
	}
	if _, err := os.Stat(manifestDir); !os.IsNotExist(err) {
		cmd := exec.Command("cp", "-r", manifestDir, buildDir)
		err = cmd.Run()
		if err != nil {
			return "", "", errors.Wrap(err, "error copying build config dir to temp dir")
		}
	}

	return
}

// DoPackageBuild builds a package using the provided build config
func (c PackageRepoConfig) DoPackageBuild(name string, buildDir string) error {
	fmt.Printf("📦 Building package '%s'...\n\n", name)
	fmt.Printf("Melange build output:\n\n")

	cmd := exec.Command(
		"docker", "run", "--rm", "--privileged",
		"-v", fmt.Sprintf("%s:/work", buildDir),
		"-v", fmt.Sprintf("%s:/work/packages", c.PackageDir),
		"-v", fmt.Sprintf("%s:/keys", c.KeyDir),
		"cgr.dev/chainguard/melange", "build",
		"--arch", "x86_64",
		"--signing-key", filepath.Join("/keys", c.KeyFilename),
		fmt.Sprintf("/work/%s.yaml", name),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to build package")
	}

	fmt.Printf("\n")

	if err := deleteBuildDir(buildDir); err != nil {
		return err
	}

	fmt.Printf(" ✅  Successfully built package '%s'\n", name)
	fmt.Printf(" 🛠️  Use this package in local image builds by adding the package '%s@local' to the base image config\n", name)

	return nil
}

// deleteTempDir deletes a build directory after checking that it's a temporary directory
func deleteBuildDir(path string) error {
	if !strings.HasPrefix(path, "/tmp/") {
		return errors.New(fmt.Sprintf("directory '%s' is not a temporary directory - not cleaning up", path))
	}

	if err := os.RemoveAll(path); err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to remove build dir '%s'", path))
	}

	return nil
}
