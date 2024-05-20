package handlers

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	Spec Spec `json:"spec"`
}

type Spec struct {
	Mover MoverSpec `json:"mover,omitempty"`
}

type MoverSpec struct {
	Rules []RuleSpec `json:"rules"`
}

type RuleSpec struct {
	// Src is the identifier of the source team. Only issues from this team will be evaluated for this rule.
	Src SrcSpec `json:"src"`
	// Dst is the identifier of the destination team. Issues that match the rule will be moved to this team.
	Dst DstSpec `json:"dst"`
}

type SrcSpec struct {
	// TeamID is the identifier of the team that the issue must be in for the rule to match.
	// Use the keyword 'Any Issue' to match any source team.
	TeamID string `json:"teamId,omitempty"`
	// Labels is a list of labels that must be present on the issue for the rule to match.
	Labels []string `json:"labels"`
}

type DstSpec struct {
	TeamID string `json:"teamId,omitempty"`
	// Modifier is an optional field that allows for additional configuration when moving the issue.
	// Any errors encountered while applying the modifier will be logged, but the issue will still be moved regardless.
	Modifier *DstModifierSpec `json:"modifier,omitempty"`
}

type DstModifierSpec struct {
	// ProjectName is the name of the project that the issue will be moved to. The project must exist in the destination team.
	ProjectName string `json:"projectName,omitempty"`
}

func (c *Config) Validate() error {
	return c.Spec.Validate()
}

func (s *Spec) Validate() error {
	return s.Mover.Validate()
}

func (s *MoverSpec) Validate() error {
	var errs errors.MultiError
	if len(s.Rules) == 0 {
		errs = errors.Append(errs, errors.New("rules must contain at least one rule"))
	}
	for _, r := range s.Rules {
		if err := r.Validate(); err != nil {
			errs = errors.Append(errs, err)
		}
	}
	return errs
}

func (rs RuleSpec) Validate() error {
	var errs errors.MultiError
	if rs.Src.TeamID == "" {
		errs = errors.Append(errs, errors.Newf("src.teamId must be set, or use %d to match any issues", WildcardTeamID))
	}
	if len(rs.Src.Labels) == 0 {
		errs = errors.Append(errs, errors.New("src.labels must contain at least one label"))
	}
	if rs.Dst.TeamID == "" {
		errs = errors.Append(errs, errors.New("dst.teamId must be set"))
	}
	return errs
}
