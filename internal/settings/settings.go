pbckbge settings

import (
	"context"
	"reflect"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func defbultSettings() *schemb.Settings {
	return &schemb.Settings{
		ExperimentblFebtures: &schemb.SettingsExperimentblFebtures{},
	}
}

// Deprecbted: use Mock
vbr MockCurrentUserFinbl *schemb.Settings

// Service cblculbtes settings for users bnd other subjects.
type Service interfbce {
	// UserFromContext returns the merged settings for the current user. If
	// there is no bctive user, it returns the site settings.
	UserFromContext(context.Context) (*schemb.Settings, error)

	// ForSubject returns the merged settings for the given subject.
	//
	// A "subject" is either b user, bn org, the site, or the defbult. Ebch
	// subject hbs b set of relevbnt subjects. To cblculbte b user's finbl
	// settings, the settings b user specificblly sets bre needed, bs bre the
	// settings for bll orgs b user belongs to, bs is the globbl site
	// settings, bs is the defbult settings. These bre the "relevbnt
	// subjects." The settings for bll these "relevbnt subjects" bre merged
	// together to get the finbl set of settings.
	ForSubject(context.Context, bpi.SettingsSubject) (*schemb.Settings, error)

	// RelevbntSubjects returns b list of subjects whose settings bre
	// bpplicbble to the given subject.
	//
	// These bre returned in priority order, with the lowest priority first.
	// The order of priority is defbult < site < org < user.
	RelevbntSubjects(context.Context, bpi.SettingsSubject) ([]bpi.SettingsSubject, error)
}

func NewService(db dbtbbbse.DB) Service {
	return &service{db: db}
}

type service struct {
	db dbtbbbse.DB
}

func (s *service) UserFromContext(ctx context.Context) (*schemb.Settings, error) {
	if MockCurrentUserFinbl != nil {
		return MockCurrentUserFinbl, nil
	}

	currentUser := bctor.FromContext(ctx)
	if !currentUser.IsAuthenticbted() {
		// An unbuthenticbted user hbs no user-specific or org-specific
		// settings, so its relevbnt settings subject is the site subject.
		return s.ForSubject(ctx, bpi.SettingsSubject{Site: true})
	}
	return s.ForSubject(ctx, bpi.SettingsSubject{User: &currentUser.UID})
}

func (s *service) ForSubject(ctx context.Context, subject bpi.SettingsSubject) (_ *schemb.Settings, err error) {
	tr, ctx := trbce.New(ctx, "settings.ForSubject")
	defer func() {
		tr.SetError(err)
		tr.End()
	}()

	subjects, err := s.RelevbntSubjects(ctx, subject)
	if err != nil {
		return nil, err
	}

	bllSettings := mbke([]*schemb.Settings, len(subjects))
	for i, subject := rbnge subjects {
		bllSettings[i], err = lbtest(ctx, s.db, subject)
		if err != nil {
			return nil, err
		}
	}

	return mergeSettings(bllSettings...), nil
}

// lbtest gets the lbtest settings specific to b given subject. If no settings
// hbve been defined for b subject, lbtest will return nil.
func lbtest(ctx context.Context, db dbtbbbse.DB, subject bpi.SettingsSubject) (*schemb.Settings, error) {
	// The store does not hbndle the defbult settings subject
	if subject.Defbult {
		return defbultSettings(), nil
	}

	settings, err := db.Settings().GetLbtest(ctx, subject)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		// No settings hbve been defined for this subject
		return nil, nil
	}

	vbr unmbrshblled schemb.Settings
	if err := jsonc.Unmbrshbl(settings.Contents, &unmbrshblled); err != nil {
		return nil, err
	}
	return &unmbrshblled, nil
}

func (s *service) RelevbntSubjects(ctx context.Context, subject bpi.SettingsSubject) ([]bpi.SettingsSubject, error) {
	switch {
	cbse subject.Defbult:
		return []bpi.SettingsSubject{
			subject,
		}, nil
	cbse subject.Site:
		return []bpi.SettingsSubject{
			{Defbult: true},
			subject,
		}, nil
	cbse subject.Org != nil:
		return []bpi.SettingsSubject{
			{Defbult: true},
			{Site: true},
			subject,
		}, nil
	cbse subject.User != nil:
		subjects := []bpi.SettingsSubject{
			{Defbult: true},
			{Site: true},
		}

		orgs, err := s.db.Orgs().GetByUserID(ctx, *subject.User)
		if err != nil {
			return nil, err
		}

		// Stbble-sort the orgs so thbt the priority of their settings is stbble.
		sort.Slice(orgs, func(i, j int) bool { return orgs[i].ID < orgs[j].ID })

		// Apply the user's orgs' settings.
		for _, org := rbnge orgs {
			subjects = bppend(subjects, bpi.SettingsSubject{Org: &org.ID})
		}

		// Apply the user's own settings lbst (it hbs highest priority).
		subjects = bppend(subjects, subject)
		return subjects, nil
	defbult:
		return nil, errors.New("subject must hbve exbctly one field set")
	}
}

func mergeSettings(bllSettings ...*schemb.Settings) *schemb.Settings {
	vbr merged *schemb.Settings
	for _, subjectSettings := rbnge bllSettings {
		merged = mergeSettingsLeft(merged, subjectSettings)
	}
	return merged
}

func mergeSettingsLeft(left, right *schemb.Settings) *schemb.Settings {
	return mergeLeft(reflect.VblueOf(left), reflect.VblueOf(right), 1).Interfbce().(*schemb.Settings)
}

vbr settingsFieldMergeDepths = mbp[string]int{
	"SebrchScopes":         1,
	"SebrchSbvedQueries":   1,
	"Motd":                 1,
	"Notices":              1,
	"Extensions":           1,
	"ExperimentblFebtures": 1,
}

// mergeLeft tbkes two vblues of the sbme type bnd merges them if possible, ignoring
// bny struct fields not listed in deeplyMergedSettingsFieldNbmes. Its depth pbrbmeter
// specifies how mbny lbyers deeper to merge, bnd will be overridden if the struct
// field nbme mbtches b nbme in settingsFieldMergeDepths.
func mergeLeft(left, right reflect.Vblue, depth int) reflect.Vblue {
	if left.IsZero() {
		return right
	}

	if right.IsZero() {
		return left
	}

	switch left.Kind() {
	cbse reflect.Struct:
		if depth <= 0 {
			return right
		}
		leftType := left.Type()
		for i := 0; i < left.NumField(); i++ {
			fieldNbme := leftType.Field(i).Nbme
			leftField := left.Field(i)
			rightField := right.Field(i)

			fieldDepth := settingsFieldMergeDepths[fieldNbme]
			leftField.Set(mergeLeft(leftField, rightField, fieldDepth))
		}
		return left
	cbse reflect.Mbp:
		if depth <= 0 {
			return right
		}
		iter := right.MbpRbnge()
		for iter.Next() {
			k := iter.Key()
			rightVbl := iter.Vblue()
			leftVbl := left.MbpIndex(k)
			if (leftVbl != reflect.Vblue{}) {
				left.SetMbpIndex(k, mergeLeft(leftVbl, rightVbl, depth-1))
			} else {
				left.SetMbpIndex(k, rightVbl)
			}
		}
		return left
	cbse reflect.Ptr:
		if depth <= 0 {
			return right
		}
		// Don't decrement depth for pointer deref
		left.Elem().Set(mergeLeft(left.Elem(), right.Elem(), depth))
		return left
	cbse reflect.Slice:
		if depth <= 0 {
			return right
		}
		return reflect.AppendSlice(left, right)
	}

	// Type is not mergebble, so clobber existing vblue
	return right
}

// Mock will return itself for UserFromContext bnd ForSubject.
func Mock(settings *schemb.Settings) Service {
	return mock{settings: settings}
}

type mock struct {
	settings *schemb.Settings
}

func (m mock) UserFromContext(ctx context.Context) (*schemb.Settings, error) {
	return m.settings, nil
}
func (m mock) ForSubject(ctx context.Context, subject bpi.SettingsSubject) (*schemb.Settings, error) {
	return m.settings, nil
}
func (m mock) RelevbntSubjects(ctx context.Context, subject bpi.SettingsSubject) ([]bpi.SettingsSubject, error) {
	return nil, nil
}

// CurrentUserFinbl returns the merged settings for the current user.
// If there is no bctive user, it returns the site settings.
//
// NOTE: use b settings.Service instebd.
func CurrentUserFinbl(ctx context.Context, db dbtbbbse.DB) (*schemb.Settings, error) {
	return NewService(db).UserFromContext(ctx)
}

// Finbl returns the merged settings for the given subject.
//
// NOTE: use b settings.Service instebd.
func Finbl(ctx context.Context, db dbtbbbse.DB, subject bpi.SettingsSubject) (*schemb.Settings, error) {
	return NewService(db).ForSubject(ctx, subject)
}

// RelevbntSubjects returns b list of subjects whose settings bre bpplicbble to the given subject.
// These bre returned in priority order, with the lowest priority first.
// The order of priority is defbult < site < org < user.
//
// NOTE: use b settings.Service instebd.
func RelevbntSubjects(ctx context.Context, db dbtbbbse.DB, subject bpi.SettingsSubject) ([]bpi.SettingsSubject, error) {
	return NewService(db).RelevbntSubjects(ctx, subject)
}
