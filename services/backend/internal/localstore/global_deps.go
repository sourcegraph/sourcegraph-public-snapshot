package localstore

import (
	"fmt"
	"strings"

	"context"

	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	GraphSchema.Map.AddTableWithName(resolution{}, "global_deps").SetKeys(false, "Repo", "CommitID", "UnitType", "Unit")
	GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
		`ALTER TABLE global_deps ALTER COLUMN repo TYPE citext`,
		`CREATE INDEX resolver_idx ON global_deps USING btree (raw_unit_type, raw_unit)`,
	)
}

type resolution struct {
	Repo        string `db:"repo"`
	CommitID    string `db:"commit"`
	UnitType    string `db:"unit_type"`
	Unit        string `db:"unit"`
	RawUnitType string `db:"raw_unit_type"`
	RawUnit     string `db:"raw_unit"`
	RawVersion  string `db:"raw_version"`
}

func (r *resolution) toResolution() *unit.Resolution {
	return &unit.Resolution{
		Resolved: unit.Key{
			Repo:     r.Repo,
			CommitID: r.CommitID,
			Type:     r.UnitType,
			Name:     r.Unit,
		},
		Raw: unit.Key{
			Repo:    unit.UnitRepoUnresolved,
			Version: r.RawVersion,
			Type:    r.RawUnitType,
			Name:    r.RawUnit,
		},
	}
}

func fromResolution(r *unit.Resolution) *resolution {
	return &resolution{
		Repo:        r.Resolved.Repo,
		CommitID:    r.Resolved.CommitID,
		UnitType:    r.Resolved.Type,
		Unit:        r.Resolved.Name,
		RawUnitType: r.Raw.Type,
		RawUnit:     r.Raw.Name,
		RawVersion:  r.Raw.Version,
	}
}

type globalDeps struct{}

func (g *globalDeps) Upsert(ctx context.Context, resolutions []*unit.Resolution) error {
	if TestMockGlobalDeps != nil {
		return TestMockGlobalDeps.Upsert(ctx, resolutions)
	}

	for _, res_ := range resolutions {
		res := fromResolution(res_)
		var args []interface{}
		arg := func(v interface{}) string {
			args = append(args, v)
			return gorp.PostgresDialect{}.BindVar(len(args) - 1)
		}

		upsertSQL := `
WITH upsert AS (
UPDATE global_deps SET
	raw_unit_type=` + arg(res.RawUnitType) + `,
	raw_unit=` + arg(res.RawUnit) + `,
	raw_version=` + arg(res.RawVersion) + `
WHERE
	repo=` + arg(res.Repo) + ` AND
	commit=` + arg(res.CommitID) + ` AND
	unit_type=` + arg(res.UnitType) + ` AND
	unit=` + arg(res.Unit) + `
RETURNING *
)
INSERT into global_deps (repo, commit, unit_type, unit, raw_unit_type, raw_unit, raw_version)
SELECT ` +
			arg(res.Repo) + `,` +
			arg(res.CommitID) + `,` +
			arg(res.UnitType) + `,` +
			arg(res.Unit) + `,` +
			arg(res.RawUnitType) + `,` +
			arg(res.RawUnit) + `,` +
			arg(res.RawVersion) + `
WHERE NOT EXISTS (SELECT * FROM upsert);`

		if _, err := graphDBH(ctx).Exec(upsertSQL, args...); err != nil {
			return err
		}
	}
	return nil
}

func (g *globalDeps) Resolve(ctx context.Context, raw *unit.Key) ([]unit.Key, error) {
	if TestMockGlobalDeps != nil {
		return TestMockGlobalDeps.Resolve(ctx, raw)
	}

	if raw.IsResolved() {
		return nil, fmt.Errorf("raw unit %+v was already resolved", raw)
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	whereClauses := []string{
		`raw_unit_type=` + arg(raw.Type),
		`raw_unit=` + arg(raw.Name),
	}
	if raw.CommitID != "" {
		whereClauses = append(whereClauses, `raw_version=`+arg(raw.CommitID))
	}
	whereSQL := strings.Join(whereClauses, " AND ")

	var res_ []*resolution
	_, err := graphDBH(ctx).Select(&res_, fmt.Sprintf(`SELECT * FROM global_deps WHERE %s`, whereSQL), args...)
	if err != nil {
		return nil, err
	}

	resolved := make([]unit.Key, len(res_))
	for i, _ := range res_ {
		resolved[i] = res_[i].toResolution().Resolved
	}
	return resolved, nil
}

var TestMockGlobalDeps *MockGlobalDeps

type MockGlobalDeps struct {
	Upsert  func(ctx context.Context, resolutions []*unit.Resolution) error
	Resolve func(ctx context.Context, raw *unit.Key) ([]unit.Key, error)
}
