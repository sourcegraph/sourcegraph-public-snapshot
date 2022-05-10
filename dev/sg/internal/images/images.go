package images

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/distribution/distribution/v3/reference"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/opencontainers/go-digest"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var seenImageRepos = map[string]imageRepository{}

type DeploymentType string

const (
	DeploymentTypeK8S  DeploymentType = "k8s"
	DeploymentTypeHelm DeploymentType = "helm"
)

func Parse(path string, creds credentials.Credentials, deploy DeploymentType, pinTag string) error {
	if deploy == DeploymentTypeK8S {
		return ParseK8S(path, creds, pinTag)
	} else if deploy == DeploymentTypeHelm {
		return ParseHelm(path, creds, pinTag)
	}
	return errors.Newf("deployment kind %s is not supported", deploy)
}

func ParseK8S(path string, creds credentials.Credentials, pinTag string) error {
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
		Filters:               []kio.Filter{imageFilter{credentials: &creds, pinTag: pinTag}},
		Outputs:               []kio.Writer{rw},
		ContinueOnEmptyResult: true,
	}.Execute()

	return err
}

func isImgMap(m map[string]any) bool {
	if m["defaultTag"] != nil && m["name"] != nil {
		return true
	}
	return false
}

func extraImages(m any, acc *[]string) {
	for m != nil {
		switch m := m.(type) {
		case map[string]any:
			for k, v := range m {
				if k == "image" && reflect.TypeOf(v).Kind() == reflect.Map && isImgMap(v.(map[string]any)) {
					imgMap := v.(map[string]any)
					*acc = append(*acc, fmt.Sprintf("index.docker.io/sourcegraph/%s:%s", imgMap["name"], imgMap["defaultTag"]))
				}
				extraImages(v, acc)
			}
		case []any:
			for _, v := range m {
				extraImages(v, acc)
			}
		}
		m = nil
	}
}

func ParseHelm(path string, creds credentials.Credentials, pinTag string) error {
	valuesFilePath := filepath.Join(path, "values.yaml")
	valuesFile, err := os.ReadFile(valuesFilePath)
	if err != nil {
		return errors.Wrapf(err, "couldn't read %s", valuesFilePath)
	}

	var rawValues []byte
	rawValues, err = k8syaml.YAMLToJSON(valuesFile)
	if err != nil {
		return errors.Wrapf(err, "couldn't unmarshal %s", valuesFilePath)
	}

	var values map[string]any
	err = json.Unmarshal(rawValues, &values)
	if err != nil {
		return errors.Wrapf(err, "couldn't unmarshal %s", valuesFilePath)
	}

	var images []string
	extraImages(values, &images)

	valuesFileString := string(valuesFile)
	for _, img := range images {
		var updatedImg string
		updatedImg, err = updateImage(img, creds, pinTag)
		if err != nil {
			return errors.Wrapf(err, "couldn't update image %s", img)
		}

		var oldImgRef, newImgRef *ImageReference
		oldImgRef, err = parseImgString(img)
		if err != nil {
			return err
		}
		newImgRef, err = parseImgString(updatedImg)
		if err != nil {
			return err
		}

		oldImgDefaultTag := fmt.Sprintf("%s@%s", oldImgRef.Tag, oldImgRef.Digest)
		newImgDefaultTag := fmt.Sprintf("%s@%s", newImgRef.Tag, newImgRef.Digest)
		valuesFileString = strings.ReplaceAll(valuesFileString, oldImgDefaultTag, newImgDefaultTag)
	}

	if err := os.WriteFile(valuesFilePath, []byte(valuesFileString), 0644); err != nil {
		return errors.Newf("WriteFile: %w", err)
	}

	return nil
}

type imageFilter struct {
	credentials *credentials.Credentials
	pinTag      string
}

var _ kio.Filter = &imageFilter{}

// Filter implements kio.Filter (notably different from yaml.Filter)
// Analogous to http://www.linfo.org/filters.html
func (filter imageFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, r := range in {
		if err := findImage(r, *filter.credentials, filter.pinTag); err != nil {
			if errors.As(err, &ErrNoImage{}) || errors.Is(err, ErrNoUpdateNeeded) {
				std.Out.Verbosef("Encountered expected err: %v\n", err)
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

func findImage(r *yaml.RNode, credential credentials.Credentials, pinTag string) error {
	containers, err := r.Pipe(yaml.LookupFirstMatch(yaml.ConventionalContainerPaths))
	if err != nil {
		return errors.Newf("%v: %s", err, r.GetName())
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
			return errors.Newf("couldn't find image for container %s within %w", node.GetName(), ErrNoImage{r.GetKind(), r.GetName()})
		}
		s, err := image.Value.String()
		if err != nil {
			return err
		}
		updatedImage, err := updateImage(s, credential, pinTag)
		if err != nil {
			return err
		}

		std.Out.Verbosef("found image %s for container %s in file %s+%s\n Replaced with %s", s, node.GetName(), r.GetKind(), r.GetName(), updatedImage)

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
	Registry    string // index.docker.io
	Credentials *credentials.Credentials
	Name        string        // sourcegraph/frontend
	Digest      digest.Digest // sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc
	Tag         string        // insiders
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

func parseImgString(rawImg string) (*ImageReference, error) {
	ref, err := reference.ParseNormalizedNamed(strings.TrimSpace(rawImg))
	if err != nil {
		return nil, err
	}

	imgRef := &ImageReference{
		Registry: reference.Domain(ref),
	}

	if nameTagged, ok := ref.(reference.NamedTagged); ok {
		imgRef.Tag = nameTagged.Tag()
		imgRef.Name = reference.Path(nameTagged)
		if canonical, ok := ref.(reference.Canonical); ok {
			newNamed, err := reference.WithName(canonical.Name())
			if err != nil {
				return nil, err
			}
			newCanonical, err := reference.WithDigest(newNamed, canonical.Digest())
			if err != nil {
				return nil, err
			}
			imgRef.Digest = newCanonical.Digest()
		}
	}

	return imgRef, nil
}

func updateImage(rawImage string, credential credentials.Credentials, pinTag string) (string, error) {
	imgRef, err := parseImgString(rawImage)
	if err != nil {
		return "", err
	}
	imgRef.Credentials = &credential

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

	repo, err := createAndFillImageRepository(imgRef, pinTag)
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

	req, err := http.NewRequest("GET", fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", i.name), nil)
	if err != nil {
		return "", err
	}

	if i.imageRef.Credentials.Username != "" && i.imageRef.Credentials.Secret != "" {
		req.SetBasicAuth(i.imageRef.Credentials.Username, i.imageRef.Credentials.Secret)
	}

	resp, err := http.DefaultClient.Do(req)
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

func createAndFillImageRepository(ref *ImageReference, pinTag string) (repo *imageRepository, err error) {
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

	var targetTag string
	isDevTag := pinTag == ""
	if isDevTag {
		targetTag, err = findLatestTag(tags)
		if err != nil {
			std.Out.Verbose("findLatestTag: " + err.Error())
		}
	} else {
		targetTag = pinTag
	}

	_, semverErr := semver.NewVersion(targetTag)
	isReleaseTag := semverErr == nil
	isAlreadyLatest := targetTag == ref.Tag
	// for release build, we use semver tags and they are immutable, so no update is needed if the current tag is the same as target tag
	// for dev builds, if the current tag is the same as the latest tag, also no update is needed
	// for mutable tags (neither release nor dev tag, e.g. `insiders`), we always need to fetch the latest digest.
	if (isReleaseTag || isDevTag) && isAlreadyLatest {
		return repo, ErrNoUpdateNeeded
	}
	repo.imageRef.Tag = targetTag

	dig, err := repo.fetchDigest(targetTag)
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
		return nil, errors.Newf("unable to convert tag: %s", t)
	}
	s.buildNum, err = strconv.Atoi(n[0])
	if err != nil {
		return nil, errors.Newf("unable to convert tag: %v", err)
	}

	s.date = n[1]
	s.shortSHA1 = n[2]
	return &s, nil
}

// Assume we use 'sourcegraph' tag format of :[build_number]_[date]_[short SHA1]
func findLatestTag(tags []string) (string, error) {
	maxBuildID := 0
	targetTag := ""

	var errs error
	for _, tag := range tags {
		stag, err := ParseTag(tag)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if stag.buildNum > maxBuildID {
			maxBuildID = stag.buildNum
			targetTag = tag
		}
	}
	return targetTag, errs
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
		return "", errors.Newf("GET https://index.docker.io/v2/%s/manifests/%s %s: %s", i.name, tag, resp.Status, string(data))
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
		return nil, errors.Newf("missing auth token")
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
