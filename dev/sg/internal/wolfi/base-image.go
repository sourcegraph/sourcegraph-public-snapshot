package wolfi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// define a constant wolfi-images
const baseImageDir = "wolfi-images"

var localRepoRegex = lazyregexp.New(`(?m)^\s+-\s+.*@local`)

type BaseImageConfig struct {
	PackageRepoConfig PackageRepoConfig
	// ImageConfigDir is the directory containing all image configs
	ImageConfigDir string
	// ImageName is the name of the image e.g. gitserver
	ImageName string
	// ImageConfigPath is the path to the image config e.g. wolfi-images/gitserver.yaml
	ImageConfigPath string
	// LockfilePath is the path to the image lockfile e.g. wolfi-images/gitserver.lock.json
	LockfilePath string
	// BazelBuildPath is the Bazel build path for the image e.g. //cmd/gitserver:base_tarball
	BazelBuildPath string
	// KeyringAppend is the path to additional keys to include in the keyring
	KeyringAppend string
	// RepositoryAppend is the path to additional repositories to include in the keyring
	RepositoryAppend string
}

type BaseImageOpts struct {
	KeyringAppend    string
	RepositoryAppend string
}

func SetupBaseImageBuild(name string, pc PackageRepoConfig, opts BaseImageOpts) (bc BaseImageConfig, err error) {
	bc.PackageRepoConfig = pc

	// Get root of repo
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return bc, errors.Wrap(err, "unable to get repository root")
	}
	bc.ImageConfigDir = filepath.Join(repoRoot, baseImageDir)

	// Strip .yaml suffix if it exists
	bc.ImageName = strings.Replace(name, ".yaml", "", 1)
	bc.LockfilePath = filepath.Join(bc.ImageConfigDir, bc.ImageName+".lock.json")

	// Check manfest exists
	bc.ImageConfigPath = filepath.Join(bc.ImageConfigDir, bc.ImageName+".yaml")
	if _, err = os.Stat(bc.ImageConfigPath); os.IsNotExist(err) {
		return bc, errors.Wrap(err, "manifest file does not exist")
	}

	// Ignore error if no Bazel build path can be found - some images are not built in this repo
	imagePath, err := resolveImagePath(bc.ImageName)
	if err == nil {
		bc.BazelBuildPath = fmt.Sprintf("//%s:base_tarball", imagePath)
	}

	bc.KeyringAppend = opts.KeyringAppend
	bc.RepositoryAppend = opts.RepositoryAppend

	return bc, nil
}

// resolveImagePath takes an image name and returns the build path where the image's Bazel config can be found
func resolveImagePath(name string) (string, error) {
	// Handle special case mappings
	specialCase := map[string]string{
		"sourcegraph-base":  "wolfi-images/sourcegraph-base",
		"sourcegraph-dev":   "wolfi-images/sourcegraph-dev",
		"postgres-exporter": "docker-images/postgres_exporter",
		"redis-exporter":    "docker-images/redis_exporter",
		"redis":             "docker-images/redis-cache", // Or redis-store
		"blobstore":         "docker-images/blobstore",   // cmd/blobstore is unused
	}
	if val, exists := specialCase[name]; exists {
		std.Out.WriteLine(output.Linef(output.EmojiInfo, output.StylePending, "Mapping Bazel build path for image '%s' to '%s", name, val))
		return val, nil
	}

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", errors.Wrap(err, "unable to get repository root")
	}

	// Search for requested image in standard locations
	imageDirs := []string{"cmd", "docker-images"}
	for _, dir := range imageDirs {
		imagePath := filepath.Join(dir, name)
		fullImagePath := filepath.Join(repoRoot, imagePath)
		if _, err := os.Stat(fullImagePath); !os.IsNotExist(err) {
			return imagePath, nil
		}
	}

	return "", errors.New(fmt.Sprintf("no such image '%s' (searched %+v)", name, imageDirs))
}

func (bc BaseImageConfig) DoBaseImageBuild() error {
	// If we're already running in Bazel we can't run Bazel again inside its own builddir,
	// so ensure we're running in the base repo
	buildDir, err := os.Getwd()
	if err != nil {
		return err
	}
	bwd := os.Getenv("BUILD_WORKING_DIRECTORY")
	if bwd != "" {
		buildDir = bwd
	}

	if bc.BazelBuildPath == "" {
		return errors.Newf("no Bazel build path found for image '%s'", bc.ImageName)
	}

	bazelArgs := append(getBazelArgs(), "run")
	commandArgs := append(bazelArgs, bc.BazelBuildPath)

	std.Out.WriteLine(output.Linef("📦", output.StylePending, "Building base image %s using `bazel %s`", bc.ImageName, strings.Join(commandArgs, " ")))
	std.Out.WriteLine(output.Linef("🤖", output.StylePending, "rules_apko build output:\n"))

	cmd := exec.Command(
		"bazel", commandArgs...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = buildDir
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to build base image")
	}

	return nil
}

func (bc BaseImageConfig) DoBaseImageBuildLegacy() error {
	std.Out.WriteLine(output.Linef("📦", output.StylePending, "Building base image %s...", bc.ImageName))
	std.Out.WriteLine(output.Linef("🤖", output.StylePending, "Apko build output:\n"))

	imageName := legacyDockerImageName(bc.ImageName)
	imageFileName := imageFileName(bc.ImageName)

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/work", bc.ImageConfigDir+"/"),
		"-v", fmt.Sprintf("%s:/packages", bc.PackageRepoConfig.PackageDir),
		"-v", fmt.Sprintf("%s:/keys", bc.PackageRepoConfig.KeyDir),
		"-v", fmt.Sprintf("%s:/images", bc.PackageRepoConfig.ImageDir),
		"-e", fmt.Sprintf("SOURCE_DATE_EPOCH=%d", time.Now().Unix()),
		"-w", "/work",
		"cgr.dev/chainguard/apko", "build",
		"--arch", "x86_64",
		"--repository-append", "@local /packages",
		"--keyring-append", fmt.Sprintf("/keys/%s.pub", bc.PackageRepoConfig.KeyFilename),
		fmt.Sprintf("/work/%s.yaml", bc.ImageName),
		imageName,
		filepath.Join("/images", imageFileName),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to build base image")
	}

	std.Out.Write("")
	std.Out.WriteSuccessf("Successfully built base image %s\n", bc.ImageName)

	return nil
}

func DockerImageName(name string) string {
	return fmt.Sprintf("%s-base:latest", name)
}

func legacyDockerImageName(name string) string {
	return fmt.Sprintf("sourcegraph-wolfi/%s-base:latest", name)
}

func imageFileName(name string) string {
	return fmt.Sprintf("sourcegraph-wolfi-%s-base.tar", name)
}

func (bc BaseImageConfig) LoadBaseImage() error {
	baseImagePath := filepath.Join(bc.PackageRepoConfig.ImageDir, imageFileName(bc.ImageName))
	std.Out.WriteLine(output.Linef("🐳", output.StylePending, "Loading base image into Docker... (%s)", baseImagePath))

	f, err := os.Open(baseImagePath)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"docker", "load",
	)
	cmd.Stdin = f
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to load base image in Docker")
	}

	std.Out.Write("")
	std.Out.WriteLine(output.Linef("🛠️ ", output.StyleBold, "Run base image locally using:\n\n\tdocker run -it --entrypoint /bin/sh %s-amd64\n", legacyDockerImageName(bc.ImageName)))

	return nil
}

func (bc BaseImageConfig) CleanupBaseImageBuild() error {
	imageDir := bc.PackageRepoConfig.ImageDir
	if !strings.HasSuffix(imageDir, "/wolfi-images/local-images") {
		return errors.New(fmt.Sprintf("directory '%s' does not look like the image output directory - not cleaning up", imageDir))
	}

	if err := os.RemoveAll(imageDir); err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to remove image output dir '%s'", imageDir))
	}

	return nil
}

// ContainsLocalPackages checks if a BaseImageConfig contains a reference to a
// @local package repository.
func (bc BaseImageConfig) ContainsLocalPackages() (bool, error) {
	imageConfigData, err := os.ReadFile(bc.ImageConfigPath)
	if err != nil {
		return false, err
	}

	imageConfig := string(imageConfigData)
	localRepoMatch := localRepoRegex.MatchString(imageConfig)

	return localRepoMatch, nil
}
