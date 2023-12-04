package images

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DeploymentType string

const (
	DeploymentTypeK8S     DeploymentType = "k8s"
	DeploymentTypeHelm    DeploymentType = "helm"
	DeploymentTypeCompose DeploymentType = "compose"
)

var ErrNoUpdateNeeded = errors.New("no update needed")

type ErrNoImage struct {
	Kind string
	Name string
}

func (m ErrNoImage) Error() string {
	return fmt.Sprintf("no images found for resource: %s of kind: %s", m.Name, m.Kind)
}

// ParsedMainBranchImageTag is a structured representation of a parsed tag created by
// images.ParsedMainBranchImageTag.
type ParsedMainBranchImageTag struct {
	Build       int
	Date        string
	ShortCommit string
}

// ParseMainBranchImageTag creates MainTag structs for tags created by
// images.BranchImageTag with a branch of "main".
func ParseMainBranchImageTag(t string) (*ParsedMainBranchImageTag, error) {
	s := ParsedMainBranchImageTag{}
	t = strings.TrimSpace(t)
	var err error
	n := strings.Split(t, "_")
	if len(n) != 3 {
		return nil, errors.Newf("unable to convert tag: %q", t)
	}
	s.Build, err = strconv.Atoi(n[0])
	if err != nil {
		return nil, errors.Newf("unable to convert tag: %q", err)
	}

	s.Date = n[1]
	s.ShortCommit = n[2]
	return &s, nil
}

// Assume we use 'sourcegraph' tag format of ':[build_number]_[date]_[short SHA1]'
func FindLatestMainTag(tags []string) (string, error) {
	maxBuildID := 0
	targetTag := ""

	var errs error
	for _, tag := range tags {
		stag, err := ParseMainBranchImageTag(tag)
		if err != nil {
			errs = errors.Append(errs, err)
			continue
		}
		if stag.Build > maxBuildID {
			maxBuildID = stag.Build
			targetTag = tag
		}
	}
	return targetTag, errs
}
