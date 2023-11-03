package wolfi

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func trivyScanImage(image *ImageInfo, trivyPath string) error {

	std.Out.Writef("starting scan for image %s:%s", image.Image, image.Digest)

	output, err := exec.Command(trivyPath, "image", fmt.Sprintf("%s@%s", image.Image, image.Digest)).Output()
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s.txt", image.Name), output, 0o644)
	if err != nil {
		return err
	}

	std.Out.Writef("written scan to %s.txt", image.Name)

	return nil

}

func ScanImages() error {

	trivyPath, err := exec.LookPath("trivy")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	root, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	bzl_deps_file := "dev/oci_deps.bzl"
	bzl_deps := filepath.Join(root, bzl_deps_file)

	file, err := os.Open(bzl_deps)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var currentImage *ImageInfo
	for _, line := range lines {
		switch {
		case NamePattern.MatchString(line):
			match := NamePattern.FindStringSubmatch(line)
			if len(match) > 1 {
				imageName := strings.Trim(match[1], `"`)
				currentImage = &ImageInfo{Name: imageName}
			}
		case DigestPattern.MatchString(line):
			match := DigestPattern.FindStringSubmatch(line)
			if len(match) > 1 && currentImage != nil {
				currentImage.Digest = strings.Trim(match[1], `"`)
			}
		case ImagePattern.MatchString(line):
			match := ImagePattern.FindStringSubmatch(line)
			if len(match) > 1 && currentImage != nil {
				currentImage.Image = strings.Trim(match[1], `"`)
			}

			if currentImage != nil {
				wg.Add(1)
				go trivyScanImage(currentImage, trivyPath)
				currentImage = nil
			}
		}
	}

	wg.Wait()

	return nil
}
