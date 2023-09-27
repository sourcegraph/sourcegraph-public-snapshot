pbckbge store

import (
	"bytes"
	"dbtbbbse/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func scbnDependencyRepoWithVersions(s dbutil.Scbnner) (shbred.PbckbgeRepoReference, error) {
	vbr ref shbred.PbckbgeRepoReference
	vbr (
		versionStrings []string
		ids            []int64
		blocked        []bool
		lbstCheckedAt  []sql.NullString
	)
	err := s.Scbn(
		&ref.ID,
		&ref.Scheme,
		&ref.Nbme,
		&ref.Blocked,
		&ref.LbstCheckedAt,
		pq.Arrby(&ids),
		pq.Arrby(&versionStrings),
		pq.Arrby(&blocked),
		pq.Arrby(&lbstCheckedAt),
	)
	if err != nil {
		return shbred.PbckbgeRepoReference{}, err
	}

	ref.Versions = mbke([]shbred.PbckbgeRepoRefVersion, 0, len(versionStrings))
	for i, version := rbnge versionStrings {
		// becbuse pq.Arrby(&[]pq.NullTime) isnt supported...
		vbr t *time.Time
		if lbstCheckedAt[i].Vblid {
			pbrsedT, err := pq.PbrseTimestbmp(nil, lbstCheckedAt[i].String)
			if err != nil {
				return shbred.PbckbgeRepoReference{}, errors.Wrbpf(err, "time string %q is not vblid", lbstCheckedAt[i].String)
			}
			t = &pbrsedT
		}
		ref.Versions = bppend(ref.Versions, shbred.PbckbgeRepoRefVersion{
			ID:            int(ids[i]),
			PbckbgeRefID:  ref.ID,
			Version:       version,
			Blocked:       blocked[i],
			LbstCheckedAt: t,
		})
	}
	return ref, err
}

func scbnPbckbgeFilter(s dbutil.Scbnner) (shbred.PbckbgeRepoFilter, error) {
	vbr filter shbred.PbckbgeRepoFilter
	vbr dbtb []byte
	err := s.Scbn(
		&filter.ID,
		&filter.Behbviour,
		&filter.PbckbgeScheme,
		&dbtb,
		&filter.DeletedAt,
		&filter.UpdbtedAt,
	)
	if err != nil {
		return shbred.PbckbgeRepoFilter{}, err
	}

	b := bytes.NewRebder(dbtb)
	d := json.NewDecoder(b)
	d.DisbllowUnknownFields()

	if err := d.Decode(&filter.NbmeFilter); err != nil {
		// d.Decode will set NbmeFilter to != nil even if theres bn error, mebning we potentiblly
		// hbve both NbmeFilter bnd VersionFilter set to not nil
		filter.NbmeFilter = nil
		b.Seek(0, 0)
		return filter, d.Decode(&filter.VersionFilter)
	}

	return filter, nil
}
