package images

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var verbose bool

func Parse(path string, v bool) error {

	verbose = v

	rw := &kio.LocalPackageReadWriter{
		KeepReaderAnnotations: false,
		PreserveSeqIndent:     true,
		PackagePath:           path,
		PackageFileName:       "",
		IncludeSubpackages:    true,
		ErrorIfNonResources:   false,
		OmitReaderAnnotations: false,
		SetAnnotations:        nil,
		NoDeleteFiles:         true, //modify in place
		WrapBareSeqNode:       false,
	}

	err := kio.Pipeline{
		Inputs:                []kio.Reader{rw},
		Filters:               []kio.Filter{imageFilter{}},
		Outputs:               []kio.Writer{rw},
		ContinueOnEmptyResult: true,
	}.Execute()

	return err
}

type imageFilter struct {
	options map[string]string
}

var _ kio.Filter = &imageFilter{}

// Filter implements kio.Filter (notably different from yaml.Filter)
// Analogous to http://www.linfo.org/filters.html
func (imageFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, r := range in {
		if err := findImage(r); err != nil {
			if errors.Is(err, ErrNoImages) || errors.Is(err, ErrNoUpdateNeeded) {
				if verbose {
					fmt.Printf("Encountered expected err: %v\n", err)
				}
				continue
			}
			return nil, err
		}
	}
	return in, nil
}

var conventionalInitContainerPaths = [][]string{
	{"spec", "initContainers"},
	{"spec", "template", "spec", "initContainers"},
}

func findImage(r *yaml.RNode) error {
	containers, err := r.Pipe(yaml.LookupFirstMatch(yaml.ConventionalContainerPaths))
	if err != nil {
		return fmt.Errorf("%v: %s", err, r.GetName())
	}
	initContainers, err := r.Pipe(yaml.LookupFirstMatch(conventionalInitContainerPaths))
	if err != nil {
		return err
	}
	if containers == nil && initContainers == nil {

		if verbose {
			fmt.Printf("no images founds in %s:%s \n", r.GetKind(), r.GetName())
		}
		return ErrNoImages
	}

	var lookupImage = func(node *yaml.RNode) error {
		image := node.Field("image")
		if image == nil {
			return fmt.Errorf("couldn't find image for container %s", node.GetName())
		}
		s, err := image.Value.String()
		if err != nil {
			return err
		}
		updatedImage, err := updateImage(s)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("found image %s for container %s in file %s+%s\n Replaced with %s", s, node.GetName(), r.GetKind(), r.GetName(), updatedImage)
		}

		err = node.PipeE(yaml.Lookup("image"), yaml.Set(yaml.NewStringRNode(updatedImage)))
		return err
	}

	if err := containers.VisitElements(lookupImage); err != nil {
		return err
	}
	if initContainers != nil {
		return initContainers.VisitElements(lookupImage)
	}
	return nil

}

type ImageReference struct {
	Registry string        // index.docker.io
	Name     string        // sourcegraph/frontend
	Digest   digest.Digest // sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc
	Tag      string        // insiders
}

func (image ImageReference) String() string {
	return fmt.Sprintf("%s/%s:%s@%s", image.Registry, image.Name, image.Tag, image.Digest)
}

func updateImage(rawImage string) (string, error) {
	ref, err := reference.ParseNormalizedNamed(strings.TrimSpace(rawImage))
	if err != nil {
		return "", err
	}

	imgRef := &ImageReference{}

	// TODO Handle images without registry specified
	imgRef.Registry = reference.Domain(ref)
	if nameTagged, ok := ref.(reference.NamedTagged); ok {
		imgRef.Tag = nameTagged.Tag()
		imgRef.Name = reference.Path(nameTagged)
		if canonical, ok := ref.(reference.Canonical); ok {
			newNamed, err := reference.WithName(canonical.Name())
			if err != nil {
				return "", err
			}
			newCanonical, err := reference.WithDigest(newNamed, canonical.Digest())
			if err != nil {
				return "", err
			}
			imgRef.Digest = newCanonical.Digest()
		}
	}

	imgRepo := &imageRepository{
		name: imgRef.Name,
	}
	imgRepo.authToken, err = imgRepo.fetchAuthToken(imgRef.Registry)
	if err != nil {
		return "", err
	}

	tags, err := imgRepo.fetchAllTags()
	if err != nil {
		return "", err
	}

	latestTag := findLatestTag(tags)
	if latestTag == imgRef.Tag {
		// do nothing
		return imgRef.String(), ErrNoUpdateNeeded
	}

	// also get digest for latestTag
	newDigest, err := imgRepo.fetchDigest(latestTag)
	if err != nil {
		return "", err
	}
	newImgRef := imgRef
	newImgRef.Tag = latestTag
	newImgRef.Digest = newDigest

	// prevent changing the registry if they are equivalent
	if newImgRef.Registry == dockerhub && strings.Contains(rawImage, legacyDockerhub) {
		newImgRef.Registry = legacyDockerhub
	}

	return newImgRef.String(), nil
}

type imageRepository struct {
	name             string
	isDockerRegistry bool
	authToken        string
}

const (
	legacyDockerhub = "index.docker.io"
	dockerhub       = "docker.io"
)

var ErrNoUpdateNeeded = errors.New("no update needed")
var ErrNoImages = errors.New("no images found in resource")
var ErrUnsupportedRegistry = errors.New("unsupported registry")

func (i *imageRepository) fetchAuthToken(registryName string) (string, error) {
	if registryName != legacyDockerhub && registryName != dockerhub {
		i.isDockerRegistry = false
		return "", ErrUnsupportedRegistry
	} else {
		i.isDockerRegistry = true
	}

	resp, err := http.Get(fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", i.name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	result := struct {
		AccessToken string `json:"access_token"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

// Assume we use 'sourcegraph' tag format of :[build_number]_[date]_[short SHA1]
func findLatestTag(tags []string) string {
	maxBuildID := 0
	targetTag := ""

	for _, tag := range tags {
		s := strings.Split(tag, "_")
		if len(s) != 3 {
			continue
		}
		b, err := strconv.Atoi(s[0])
		if err != nil {
			if verbose {
				fmt.Printf("encountered err converting tag: %v\n", err)
			}
			continue
		}
		if b > maxBuildID {
			maxBuildID = b
			targetTag = tag
		}
	}
	return targetTag
}

// snippets below from https://github.com/sourcegraph/update-docker-tags/blob/46711ff8882cfe09eaaef0f8b9f2d8c2ee7660ff/update-docker-tags.go#L258-L303

// Effectively the same as:
//
// 	$ curl -H "Authorization: Bearer $token" https://index.docker.io/v2/sourcegraph/server/tags/list
//
func (i *imageRepository) fetchDigest(tag string) (digest.Digest, error) {
	req, err := http.NewRequest("GET", "https://index.docker.io/v2/"+i.name+"/manifests/"+tag, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", i.authToken))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return "", errors.New(resp.Status + ": " + string(data))
	}

	d := resp.Header.Get("Docker-Content-Digest")
	g, err := digest.Parse(d)
	if err != nil {
		return "", err
	}
	return g, nil

}

const dockerImageTagsURL = "https://index.docker.io/v2/%s/tags/list"

// Effectively the same as:
//
// 	$ export token=$(curl -s "https://auth.docker.io/token?service=registry.docker.io&scope=repository:sourcegraph/server:pull" | jq -r .token)
//
func (i *imageRepository) fetchAllTags() ([]string, error) {
	if !i.isDockerRegistry {
		return nil, ErrUnsupportedRegistry
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(dockerImageTagsURL, i.name), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+i.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode > 200 {
		b, _ := io.ReadAll(resp.Body)
		println(b)
		return nil, err
	}
	result := struct {
		Tags []string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Tags, nil
}
