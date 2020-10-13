package api

import (
	"regexp"

	"github.com/Masterminds/semver"
)

func CheckSourcegraphVersion(version, constraint, minDate string) (bool, error) {
	if version == "dev" || version == "0.0.0+dev" {
		return true, nil
	}

	buildDate := regexp.MustCompile(`^\d+_(\d{4}-\d{2}-\d{2})_[a-z0-9]{7}$`)
	matches := buildDate.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1] >= minDate, nil
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}
	return c.Check(v), nil
}
