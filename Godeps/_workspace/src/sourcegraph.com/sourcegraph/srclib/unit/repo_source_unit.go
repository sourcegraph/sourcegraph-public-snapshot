package unit

import "encoding/json"

// NewRepoSourceUnit creates an equivalent RepoSourceUnit from a
// SourceUnit.
//
// It does not set the returned source unit's Private field (because
// it can't tell if it is private from the underlying source unit
// alone).
//
// It also doesn't set CommitID (for the same reason).
func NewRepoSourceUnit(u *SourceUnit) (*RepoSourceUnit, error) {
	unitJSON, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return &RepoSourceUnit{
		Repo:     u.Repo,
		UnitType: u.Type,
		Unit:     u.Name,
		Data:     unitJSON,
	}, nil
}

// SourceUnit decodes u's Data JSON field to the SourceUnit it
// represents.
func (u *RepoSourceUnit) SourceUnit() (*SourceUnit, error) {
	var u2 *SourceUnit
	if err := json.Unmarshal(u.Data, &u2); err != nil {
		return nil, err
	}
	return u2, nil
}
