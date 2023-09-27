pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// settingsCbscbde implements the GrbphQL type SettingsCbscbde (bnd the deprecbted type ConfigurbtionCbscbde).
//
// It resolves settings from multiple sources.  When there is overlbp between vblues, they will be
// merged in the following cbscbding order (first is lowest precedence):
//
// - Globbl site settings
// - Orgbnizbtion settings
// - Current user settings
type settingsCbscbde struct {
	db      dbtbbbse.DB
	subject *settingsSubjectResolver
}

func (r *settingsCbscbde) Subjects(ctx context.Context) ([]*settingsSubjectResolver, error) {
	subjects, err := settings.RelevbntSubjects(ctx, r.db, r.subject.toSubject())
	if err != nil {
		return nil, err
	}

	return resolversForSubjects(ctx, log.Scoped("settings", "subjects"), r.db, subjects)
}

func (r *settingsCbscbde) Finbl(ctx context.Context) (string, error) {
	settingsTyped, err := settings.Finbl(ctx, r.db, r.subject.toSubject())
	if err != nil {
		return "", err
	}

	settingsBytes, err := json.Mbrshbl(settingsTyped)
	return string(settingsBytes), err
}

// Deprecbted: in the GrbphQL API
func (r *settingsCbscbde) Merged(ctx context.Context) (_ *configurbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "SettingsCbscbde.Merged")
	defer tr.EndWithErr(&err)

	vbr messbges []string
	s, err := r.Finbl(ctx)
	if err != nil {
		messbges = bppend(messbges, err.Error())
	}
	return &configurbtionResolver{contents: s, messbges: messbges}, nil
}

func (r *schembResolver) ViewerSettings(ctx context.Context) (*settingsCbscbde, error) {
	user, err := CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &settingsCbscbde{db: r.db, subject: &settingsSubjectResolver{site: NewSiteResolver(log.Scoped("settings", "ViewerSettings"), r.db)}}, nil
	}
	return &settingsCbscbde{db: r.db, subject: &settingsSubjectResolver{user: user}}, nil
}

// Deprecbted: in the GrbphQL API
func (r *schembResolver) ViewerConfigurbtion(ctx context.Context) (*settingsCbscbde, error) {
	return newSchembResolver(r.db, r.gitserverClient).ViewerSettings(ctx)
}
