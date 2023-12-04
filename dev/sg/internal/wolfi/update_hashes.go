package wolfi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ImagePattern = regexp.MustCompile(`image\s=\s"(.*?)"`)
var DigestPattern = regexp.MustCompile(`digest\s=\s"(.*?)"`)
var NamePattern = regexp.MustCompile(`name\s=\s"(.*?)"`)

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

type ImageInfo struct {
	Name   string
	Digest string
	Image  string
}

func getAnonDockerAuthToken(repo string) (string, error) {
	// get a token so we can fetch manifests
	if !strings.Contains(repo, "/") {
		repo = fmt.Sprintf("%s/%s", repo, repo)
	}

	client := http.Client{}

	url := "https://auth.docker.io/token"
	scope := fmt.Sprintf("repository:%s:pull", repo)

	req, _ := http.NewRequest("GET", url, nil)

	q := req.URL.Query()
	q.Add("service", "registry.docker.io")
	q.Add("scope", scope)

	req.URL.RawQuery = q.Encode()

	resp, _ := client.Do(req)

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("unexpected status code while fetching token %d\n", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var tr TokenResponse

	err := json.Unmarshal([]byte(body), &tr)

	if err != nil {
		return "", err
	}

	return tr.Token, nil
}

func getImageManifest(image string, tag string) (string, error) {
	token, err := getAnonDockerAuthToken(image)

	if err != nil {
		return "", err
	}

	reg := "https://registry.hub.docker.com/v2/%s/manifests/%s"
	url := fmt.Sprintf(reg, image, tag)

	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, _ := client.Do(req)

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("unexpected status code while fetching manifest %d\n", resp.StatusCode)
	}
	defer resp.Body.Close()

	digest := resp.Header.Get("Docker-Content-Digest")

	return digest, nil
}

func UpdateHashes(_ *cli.Context, updateImageName string) error {
	if updateImageName != "" {
		updateImageName = strings.ReplaceAll(updateImageName, "-", "_")
		updateImageName = fmt.Sprintf("wolfi_%s_base", updateImageName)
	}

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

	if updateImageName == "" {
		std.Out.Write("Checking for hash updates to all images...")
	} else {
		std.Out.Writef("Checking for hash updates to '%s'...", updateImageName)
	}

	var updated, updateImageNameMatch bool

	var currentImage *ImageInfo
	for i, line := range lines {
		switch {
		case NamePattern.MatchString(line):
			match := NamePattern.FindStringSubmatch(line)
			if len(match) > 1 {
				imageName := strings.Trim(match[1], `"`)

				// Only update an image if updateImageName matches the name, or if it's empty (in which case update all images)
				if updateImageName == imageName || updateImageName == "" {
					updateImageNameMatch = true
					currentImage = &ImageInfo{Name: imageName}
				}
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

				if strings.HasPrefix(currentImage.Image, "index.docker.io") {
					// fetch new digest for latest tag
					newDigest, err := getImageManifest(strings.Replace(currentImage.Image, "index.docker.io/", "", 1), "latest")

					if err != nil {
						std.Out.WriteWarningf("%v", err)
					} else if currentImage.Digest != newDigest {
						updated = true
						// replace old digest with new digest in the previous line
						lines[i-1] = DigestPattern.ReplaceAllString(lines[i-1], fmt.Sprintf(`digest = "%s"`, newDigest))
						std.Out.WriteSuccessf("Found new digest for %s", currentImage.Image)
					}
				}

				currentImage = nil
			}
		}
	}

	// write lines back to file
	if updated {
		std.Out.Write("Updating file ...")
		file, err = os.Create(bzl_deps)
		if err != nil {
			return err
		}
		writer := bufio.NewWriter(file)
		for _, line := range lines {
			fmt.Fprintln(writer, line)
		}

		err = writer.Flush()
		if err != nil {
			return err
		}
		std.Out.WriteSuccessf("Succesfully updated digests in %s", bzl_deps_file)

	} else {
		// No digests were updated - determine why and print status message
		if updateImageName == "" {
			std.Out.WriteSuccessf("No digests needed to be updated in %s", bzl_deps_file)
		} else {
			if updateImageNameMatch {
				std.Out.WriteSuccessf("No digests needed to be updated in %s", bzl_deps_file)
			} else {
				std.Out.WriteFailuref("Did not find any images matching '%s' in %s", updateImageName, bzl_deps_file)
			}
		}
	}

	return nil
}
