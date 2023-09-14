package wolfi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

func UpdateHashes(ctx *cli.Context) error {

	root, err := root.RepositoryRoot()

	if err != nil {
		return err
	}

	bzl_deps := filepath.Join(root, "dev/oci_deps.bzl")

	file, err := os.Open(bzl_deps)
	if err != nil {
		return err
	}
	defer file.Close()

	imagePattern := regexp.MustCompile(`image\s=\s"(.*?)"`)
	digestPattern := regexp.MustCompile(`digest\s=\s"(.*?)"`)
	namePattern := regexp.MustCompile(`name\s=\s"(.*?)"`)

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	std.Out.Write("Checking for image updates...")

	var updated bool

	var currentImage *ImageInfo
	for i, line := range lines {
		switch {
		case namePattern.MatchString(line):
			match := namePattern.FindStringSubmatch(line)
			if len(match) > 1 {
				currentImage = &ImageInfo{Name: strings.Trim(match[1], `"`)}
			}
		case digestPattern.MatchString(line):
			match := digestPattern.FindStringSubmatch(line)
			if len(match) > 1 && currentImage != nil {
				currentImage.Digest = strings.Trim(match[1], `"`)
			}
		case imagePattern.MatchString(line):
			match := imagePattern.FindStringSubmatch(line)
			if len(match) > 1 && currentImage != nil {
				currentImage.Image = strings.Trim(match[1], `"`)

				if strings.HasPrefix(currentImage.Image, "index.docker.io") {
					// fetch new digest for latest tag
					newDigest, err := getImageManifest(strings.Replace(currentImage.Image, "index.docker.io/", "", 1), "latest")

					if err != nil {
						std.Out.WriteWarningf("%v", err)
					}

					if currentImage.Digest != newDigest {
						updated = true
						// replace old digest with new digest in the previous line
						lines[i-1] = digestPattern.ReplaceAllString(lines[i-1], fmt.Sprintf(`digest = "%s"`, newDigest))
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
		writer.Flush()
		std.Out.WriteSuccessf("Succesfully updated digests in %s", "oci_deps.bzl")
	}

	return nil
}
