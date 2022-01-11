package images

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var seenImageRepos = map[string]imageRepository{}

func Parse(path string) error {

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

type imageFilter struct{}

var _ kio.Filter = &imageFilter{}

// Filter implements kio.Filter (notably different from yaml.Filter)
// Analogous to http://www.linfo.org/filters.html
func (imageFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, r := range in {
		if err := findImage(r); err != nil {
			if errors.As(err, &ErrNoImage{}) || errors.Is(err, ErrNoUpdateNeeded) {
				stdout.Out.Verbosef("Encountered expected err: %v\n", err)
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

		return ErrNoImage{
			Kind: r.GetKind(),
			Name: r.GetName(),
		}
	}

	var lookupImage = func(node *yaml.RNode) error {
		image := node.Field("image")
		if image == nil {
			return fmt.Errorf("couldn't find image for container %s within %w", node.GetName(), ErrNoImage{r.GetKind(), r.GetName()})
		}
		s, err := image.Value.String()
		if err != nil {
			return err
		}
		updatedImage, err := updateImage(s)
		if err != nil {
			return err
		}

		stdout.Out.Verbosef("found image %s for container %s in file %s+%s\n Replaced with %s", s, node.GetName(), r.GetKind(), r.GetName(), updatedImage)

		return node.PipeE(yaml.Lookup("image"), yaml.Set(yaml.NewStringRNode(updatedImage)))
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

type imageRepository struct {
	name             string
	isDockerRegistry bool
	authToken        string
	imageRef         *ImageReference
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
	if prevRepo, ok := seenImageRepos[imgRef.Name]; ok {
		if imgRef.Tag == prevRepo.imageRef.Tag {
			// no update needed
			return imgRef.String(), ErrNoUpdateNeeded
		}
		if prevRepo.checkLegacy(rawImage) {
			prevRepo.imageRef.Registry = legacyDockerhub
			return prevRepo.imageRef.String(), nil
		}
		return prevRepo.imageRef.String(), nil
	}

	repo, err := createAndFillImageRepository(imgRef)
	if err != nil {
		if errors.Is(err, ErrNoUpdateNeeded) {
			return imgRef.String(), ErrNoUpdateNeeded
		}
		return "", err
	}

	seenImageRepos[imgRef.Name] = *repo

	if repo.checkLegacy(rawImage) {
		repo.imageRef.Registry = legacyDockerhub
		return repo.imageRef.String(), nil
	}
	return repo.imageRef.String(), nil
}

const (
	legacyDockerhub = "index.docker.io"
	dockerhub       = "docker.io"
)

var ErrNoUpdateNeeded = errors.New("no update needed")

type ErrNoImage struct {
	Kind string
	Name string
}

func (m ErrNoImage) Error() string {
	return fmt.Sprintf("no images found for resource: %s of kind: %s", m.Name, m.Kind)
}

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
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return "", errors.New(resp.Status + ": " + string(data))
	}

	result := struct {
		AccessToken string `json:"access_token"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func createAndFillImageRepository(ref *ImageReference) (repo *imageRepository, err error) {

	repo = &imageRepository{name: ref.Name, imageRef: ref}
	repo.authToken, err = repo.fetchAuthToken(ref.Registry)
	if err != nil {
		return nil, nil
	}
	tags, err := repo.fetchAllTags()
	if err != nil {
		return nil, err
	}

	repo.imageRef = &ImageReference{
		Registry: ref.Registry,
		Name:     ref.Name,
		Digest:   "",
		Tag:      ref.Tag,
	}

	latestTag := findLatestTag(tags)
	if latestTag == ref.Tag || latestTag == "" {
		return repo, ErrNoUpdateNeeded
	}
	repo.imageRef.Tag = latestTag

	dig, err := repo.fetchDigest(latestTag)
	if err != nil {
		return nil, err
	}
	repo.imageRef.Digest = dig

	return repo, nil
}

type SgImageTag struct {
	buildNum  int
	date      string
	shortSHA1 string
}

// ParseTag creates SgImageTag structs for strings that follow MainBranchTagPublishFormat
func ParseTag(t string) (*SgImageTag, error) {
	s := SgImageTag{}
	t = strings.TrimSpace(t)
	var err error
	n := strings.Split(t, "_")
	if len(n) != 3 {
		return nil, fmt.Errorf("unable to convert tag: %s", t)
	}
	s.buildNum, err = strconv.Atoi(n[0])
	if err != nil {
		return nil, fmt.Errorf("unable to convert tag: %v", err)
	}

	s.date = n[1]
	s.shortSHA1 = n[2]
	return &s, nil
}

// Assume we use 'sourcegraph' tag format of :[build_number]_[date]_[short SHA1]
func findLatestTag(tags []string) string {
	maxBuildID := 0
	targetTag := ""

	for _, tag := range tags {
		stag, err := ParseTag(tag)
		if err != nil {
			stdout.Out.Verbosef("%v\n", err)
			continue
		}
		if stag.buildNum > maxBuildID {
			maxBuildID = stag.buildNum
			targetTag = tag
		}
	}
	return targetTag
}

// CheckLegacy prevents changing the registry if they are equivalent, internally legacyDockerhub is resolved to dockerhub
// Most helpful during printing
func (i *imageRepository) checkLegacy(rawImage string) bool {
	if i.imageRef.Registry == dockerhub && strings.Contains(rawImage, legacyDockerhub) {
		return true
	}
	return false
}

// snippets below from https://github.com/sourcegraph/update-docker-tags/blob/46711ff8882cfe09eaaef0f8b9f2d8c2ee7660ff/update-docker-tags.go#L258-L303

// Effectively the same as:
//
// 	$ curl -H "Authorization: Bearer $token" https://index.docker.io/v2/sourcegraph/server/tags/list
//
func (i *imageRepository) fetchDigest(tag string) (digest.Digest, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://index.docker.io/v2/%s/manifests/%s", i.name, tag), nil)
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
	if resp.StatusCode != 200 {
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
	if i.authToken == "" {
		return nil, fmt.Errorf("missing auth token")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf(dockerImageTagsURL, i.name), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+i.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.New(resp.Status + ": " + string(data))
	}
	defer resp.Body.Close()
	result := struct {
		Tags []string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Tags, nil
}
