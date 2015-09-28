package sourcegraph

import "sourcegraph.com/sourcegraph/srclib/unit"

func NewUnitSpecFromUnit(u *unit.RepoSourceUnit) UnitSpec {
	return UnitSpec{
		RepoRevSpec: RepoRevSpec{
			RepoSpec: RepoSpec{URI: u.Repo},
			Rev:      u.CommitID,
			CommitID: u.CommitID,
		},
		UnitType: u.UnitType,
		Unit:     u.Unit,
	}
}

func UnmarshalUnitSpec(vars map[string]string) (UnitSpec, error) {
	repoRevSpec, err := UnmarshalRepoRevSpec(vars)
	if err != nil {
		return UnitSpec{}, err
	}
	return UnitSpec{
		RepoRevSpec: repoRevSpec,
		UnitType:    vars["UnitType"],
		Unit:        vars["Unit"],
	}, nil
}

func (s UnitSpec) RouteVars() map[string]string {
	v := s.RepoRevSpec.RouteVars()
	v["UnitType"] = s.UnitType
	v["Unit"] = s.Unit
	return v
}
