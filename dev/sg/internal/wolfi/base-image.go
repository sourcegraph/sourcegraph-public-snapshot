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
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func (c PackageRepoConfig) SetupBaseImageBuild(name string) (manifestBaseName string, buildDir string, err error) {
	// Get root of repo
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", "", errors.Wrap(err, "unable to get repository root")
	}
	buildDir = filepath.Join(repoRoot, "wolfi-images")

	// Strip .yaml suffix if it exists
	manifestBaseName = strings.Replace(name, ".yaml", "", 1)
	manifestFileName := manifestBaseName + ".yaml"

	// Check manfest exists
	manifestPath := filepath.Join(repoRoot, "wolfi-images", manifestFileName)

	if _, err = os.Stat(manifestPath); os.IsNotExist(err) {
		return "", "", errors.Wrap(err, "manifest file does not exist")
	}

	return
}

// resolveImagePath takes an image name and returns the build path where the image's Bazel config can be found
func resolveImagePath(name string) (string, error) {
	// Special cases which are embedded in wolfi-images/
	if name == "sourcegraph" || name == "sourcegraph-dev" {
		return "wolfi-images", nil
	}

	// Search for requested image in standard locations
	imageDirs := []string{"cmd", "docker-images"}
	for _, dir := range imageDirs {
		imagePath := filepath.Join(dir, name)
		if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
			return imagePath, nil
		}
	}

	return "", errors.New(fmt.Sprintf("no such image (searched %+v)", imageDirs))
}

func (c PackageRepoConfig) DoBaseImageBuild(name string, buildDir string) error {
	std.Out.WriteLine(output.Linef("📦", output.StylePending, "Building base image %s...", name))
	std.Out.WriteLine(output.Linef("🤖", output.StylePending, "Apko build output:\n"))

	buildPath, err := resolveImagePath(name)
	if err != nil {
		return errors.Wrap(err, "failed to resolve image's Bazel build path")
	}
	bazelBuildPath := fmt.Sprintf("//%s:wolfi_base_tarball", buildPath)

	cmd := exec.Command(
		"bazel", "run", bazelBuildPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to build base image")
	}

	return nil
}

func (c PackageRepoConfig) DoBaseImageBuildLegacy(name string, buildDir string) error {
	std.Out.WriteLine(output.Linef("📦", output.StylePending, "Building base image %s...", name))
	std.Out.WriteLine(output.Linef("🤖", output.StylePending, "Apko build output:\n"))

	imageName := fmt.Sprintf("sourcegraph-wolfi/%s-base:latest", name)
	imageFileName := fmt.Sprintf("sourcegraph-wolfi-%s-base.tar", name)

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/work", buildDir),
		"-v", fmt.Sprintf("%s:/packages", c.PackageDir),
		"-v", fmt.Sprintf("%s:/keys", c.KeyDir),
		"-v", fmt.Sprintf("%s:/images", c.ImageDir),
		"-e", fmt.Sprintf("SOURCE_DATE_EPOCH=%d", time.Now().Unix()),
		"-w", "/work",
		"cgr.dev/chainguard/apko", "build",
		"--arch", "x86_64",
		"--repository-append", "@local /packages",
		"--keyring-append", fmt.Sprintf("/keys/%s.pub", c.KeyFilename),
		fmt.Sprintf("/work/%s.yaml", name),
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
	std.Out.WriteSuccessf("Successfully built base image %s\n", name)

	return nil
}

func DockerImageName(name string) string {
	return fmt.Sprintf("%s-base:latest", name)
}

func legacyDockerImageName(name string) string {
	return fmt.Sprintf("sourcegraph-wolfi/%s-base:latest-amd64", name)
}

func imageFileName(name string) string {
	return fmt.Sprintf("sourcegraph-wolfi-%s-base.tar", name)
}

func (c PackageRepoConfig) LoadBaseImage(name string) error {
	baseImagePath := filepath.Join(c.ImageDir, imageFileName(name))
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
	std.Out.WriteLine(output.Linef("🛠️ ", output.StyleBold, "Run base image locally using:\n\n\tdocker run -it --entrypoint /bin/sh %s\n", legacyDockerImageName(name)))

	return nil
}

func (c PackageRepoConfig) CleanupBaseImageBuild(name string) error {
	imageDir := c.ImageDir
	if !strings.HasSuffix(imageDir, "/wolfi-images/local-images") {
		return errors.New(fmt.Sprintf("directory '%s' does not look like the image output directory - not cleaning up", imageDir))
	}

	if err := os.RemoveAll(imageDir); err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to remove image output dir '%s'", imageDir))
	}

	return nil
}
