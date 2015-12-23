package dep

import (
	"encoding/json"
)

// START ResolvedTarget OMIT
// ResolvedTarget represents a resolved dependency target.
type ResolvedTarget struct {
	// ToRepoCloneURL is the clone URL of the repository that is depended on.
	//
	// When graphers emit ResolvedDependencies, they should fill in this field,
	// not ToRepo, so that the dependent repository can be added if it doesn't
	// exist. The ToRepo URI alone does not specify enough information to add
	// the repository (because it doesn't specify the VCS type, scheme, etc.).
	ToRepoCloneURL string

	// ToUnit is the name of the source unit that is depended on.
	ToUnit string

	// ToUnitType is the type of the source unit that is depended on.
	ToUnitType string

	// ToVersion is the version of the dependent repository (if known),
	// according to whatever version string specifier is used by FromRepo's
	// dependency management system.
	ToVersionString string

	// ToRevSpec specifies the desired VCS revision of the dependent repository
	// (if known).
	ToRevSpec string
}

// END ResolvedTarget OMIT

// START Resolution OMIT
// Resolution is the result of dependency resolution: either a successfully
// resolved target or an error.
type Resolution struct {
	// Raw is the original raw dep that this was resolution was attempted on.
	Raw interface{}

	// Target is the resolved dependency, if resolution succeeds.
	Target *ResolvedTarget `json:",omitempty"`

	// Error is the resolution error, if any.
	Error string `json:",omitempty"`
}

// END Resolution OMIT

func (r *Resolution) KeyId() string {
	return r.Target.ToRepoCloneURL + r.Target.ToUnit + r.Target.ToUnitType + r.Target.ToVersionString + r.Target.ToRevSpec
}

func (r *Resolution) RawKeyId() (string, error) {
	b, err := json.Marshal(r.Raw)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
