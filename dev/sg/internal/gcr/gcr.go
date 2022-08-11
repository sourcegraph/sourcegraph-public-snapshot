package gcr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

const (
	registry = "gcr.io"
)

type TagListResp struct {
	Child    []string
	Manifest map[string]Manifest
	Name     string
	Tags     []string
}
type Manifest struct {
	ImageSizeBytes string
	LayerID        string
	MediaType      string
	Tag            []string
	TimeCreatedMs  string
	TimeUploadedMs string
}

func fetchAuthToken(ctx context.Context) (string, error) {
	creds, err := transport.Creds(ctx, option.WithScopes(compute.CloudPlatformScope))
	if err != nil {
		return "", err
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", err
	}
	accessToken := token.AccessToken

	return accessToken, nil
}
func findUpdatedTag(tagsList []string) string {
	var tags []string
	//var updatedManifest manifest
	for _, tag := range tagsList {
		if strings.Contains(tag, "-") || strings.Contains(tag, "latest") {
			continue
		}
		tags = append(tags, tag)
	}
	updatedTag := tags[len(tags)-1]
	return updatedTag
}

func findUpdatedManifest(manifest map[string]Manifest, tag string) string {
	for sha, manifest := range manifest {
		if len(manifest.Tag) > 0 {
			if manifest.Tag[0] == tag {
				return sha
			}

		}
	}
	return ""
}

func fetchUpdatedImage(imageName string, currentTag string) (string, error) {
	ctx := context.Background()
	token, err := fetchAuthToken(ctx)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/v2/%s/tags/list", registry, imageName), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	//req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("the access token may have expired")
	}

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return "", errors.Newf("GET https://%s/v2/%s/tags/list", registry, imageName, string(data))
	}

	result := TagListResp{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	updatedTag := findUpdatedTag(result.Tags)
	updatedManifest := findUpdatedManifest(result.Manifest, updatedTag)
	if updatedTag == "" || updatedManifest == "" {
		return "", errors.New("updated image not found")
	}
	updatedImage := registry + "/" + imageName + ":" + updatedTag + "@" + updatedManifest
	return updatedImage, nil
}
