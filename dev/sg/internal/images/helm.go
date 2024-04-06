package images

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	k8syaml "sigs.k8s.io/yaml"
)

func UpdateHelmManifest(ctx context.Context, registry Registry, path string, op UpdateOperation) error {
	valuesFilePath := filepath.Join(path, "values.yaml")
	valuesFile, err := os.ReadFile(valuesFilePath)
	if err != nil {
		return errors.Wrapf(err, "couldn't read %s", valuesFilePath)
	}
	valuesFileString := string(valuesFile)

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

	// If we switch registries, we need to update it in
	// sourcegraph.image.repository.
	existingReg, err := readRegistry(values)
	if err != nil {
		// github.com/sourcegraph/deploy-sourcegraph-dogfood-k8s doesn't have the
		// sourcegraph.image.repository object being defined, so we can't swap registries
		// if we want to use an internal release.
		//
		// This is fine as long as we go for using the public registry, so we can safely
		// ignore the error. But we don't we try to build a private release, in which
		// case we *need* to define explictly the registry.
		//
		// This is a case of deviations between the two helm repos, and best be addressed
		// by the release team.
		if !registry.Public() {
			return err
		} else {
			std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("skipping updating registry as it's public which we assume is the default, %q", err.Error())))
		}
	}

	if existingReg != "" {
		valuesFileString = strings.ReplaceAll(
			valuesFileString,
			existingReg,
			filepath.Join(registry.Host(), registry.Org()),
		)
	}

	// Collect all images.
	var imgs []string
	extraImages(values, &imgs)

	for _, img := range imgs {
		r, err := ParseRepository(img)
		if err != nil {
			if errors.Is(err, ErrNoUpdateNeeded) {
				std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("skipping %q", img)))
				continue
			} else {
				return err
			}
		}

		newRef, err := op(registry, r)
		if err != nil {
			if errors.Is(err, ErrNoUpdateNeeded) {
				std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("skipping %q", r.Ref())))
				continue
			} else {
				return errors.Wrapf(err, "couldn't update image %s", img)
			}
		}

		oldRaw := fmt.Sprintf("%s@%s", r.Tag(), r.digest)
		newRaw := fmt.Sprintf("%s@%s", newRef.Tag(), newRef.digest)

		valuesFileString = strings.ReplaceAll(valuesFileString, oldRaw, newRaw)
	}

	if err := os.WriteFile(valuesFilePath, []byte(valuesFileString), 0644); err != nil {
		return errors.Newf("WriteFile: %w", err)
	}

	return nil
}

func readRegistry(m map[string]any) (string, error) {
	if top, ok := m["sourcegraph"].(map[string]any); ok {
		if image, ok := top["image"].(map[string]any); ok {
			if repo, ok := image["repository"].(string); ok {
				return repo, nil
			}
		}
	}

	return "", errors.New("cannot find sourcegraph.image.registry in values.yml")
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
					//TODO
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
