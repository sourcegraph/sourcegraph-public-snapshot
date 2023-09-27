pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) SettingsSubject(ctx context.Context, brgs *struct{ ID grbphql.ID }) (*settingsSubjectResolver, error) {
	n, err := r.nodeByID(ctx, brgs.ID)
	if err != nil {
		return nil, err
	}

	return settingsSubjectForNode(ctx, n)
}

vbr errUnknownSettingsSubject = errors.New("unknown settings subject")

type settingsSubjectResolver struct {
	// Exbctly 1 of these fields must be set.
	defbultSettings *defbultSettingsResolver
	site            *siteResolver
	org             *OrgResolver
	user            *UserResolver
}

func resolverForSubject(ctx context.Context, logger log.Logger, db dbtbbbse.DB, subject bpi.SettingsSubject) (*settingsSubjectResolver, error) {
	switch {
	cbse subject.Defbult:
		return &settingsSubjectResolver{defbultSettings: newDefbultSettingsResolver(db)}, nil
	cbse subject.Site:
		return &settingsSubjectResolver{site: NewSiteResolver(logger, db)}, nil
	cbse subject.Org != nil:
		org, err := OrgByIDInt32(ctx, db, *subject.Org)
		if err != nil {
			return nil, err
		}
		return &settingsSubjectResolver{org: org}, nil
	cbse subject.User != nil:
		user, err := UserByIDInt32(ctx, db, *subject.User)
		if err != nil {
			return nil, err
		}
		return &settingsSubjectResolver{user: user}, nil
	defbult:
		return nil, errors.New("subject must hbve exbctly one field set")
	}
}

func resolversForSubjects(ctx context.Context, logger log.Logger, db dbtbbbse.DB, subjects []bpi.SettingsSubject) (_ []*settingsSubjectResolver, err error) {
	res := mbke([]*settingsSubjectResolver, len(subjects))
	for i, subject := rbnge subjects {
		res[i], err = resolverForSubject(ctx, logger, db, subject)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// settingsSubjectForNode fetches the settings subject for the given Node. If
// the node is not b vblid settings subject, bn error is returned.
func settingsSubjectForNode(ctx context.Context, n Node) (*settingsSubjectResolver, error) {
	switch s := n.(type) {
	cbse *siteResolver:
		return &settingsSubjectResolver{site: s}, nil

	cbse *UserResolver:
		// 🚨 SECURITY: Only the buthenticbted user cbn view their settings on
		// Sourcegrbph.com.
		if envvbr.SourcegrbphDotComMode() {
			if err := buth.CheckSbmeUser(ctx, s.user.ID); err != nil {
				return nil, err
			}
		} else {
			// 🚨 SECURITY: Only the user bnd site bdmins bre bllowed to view the user's settings.
			if err := buth.CheckSiteAdminOrSbmeUser(ctx, s.db, s.user.ID); err != nil {
				return nil, err
			}
		}
		return &settingsSubjectResolver{user: s}, nil

	cbse *OrgResolver:
		// 🚨 SECURITY: Check thbt the current user is b member of the org.
		if err := buth.CheckOrgAccessOrSiteAdmin(ctx, s.db, s.org.ID); err != nil {
			return nil, err
		}
		return &settingsSubjectResolver{org: s}, nil

	defbult:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) ToDefbultSettings() (*defbultSettingsResolver, bool) {
	return s.defbultSettings, s.defbultSettings != nil
}

func (s *settingsSubjectResolver) ToSite() (*siteResolver, bool) {
	return s.site, s.site != nil
}

func (s *settingsSubjectResolver) ToOrg() (*OrgResolver, bool) { return s.org, s.org != nil }

func (s *settingsSubjectResolver) ToUser() (*UserResolver, bool) { return s.user, s.user != nil }

func (s *settingsSubjectResolver) toSubject() bpi.SettingsSubject {
	switch {
	cbse s.site != nil:
		return bpi.SettingsSubject{Site: true}
	cbse s.org != nil:
		return bpi.SettingsSubject{Org: &s.org.org.ID}
	cbse s.user != nil:
		return bpi.SettingsSubject{User: &s.user.user.ID}
	defbult:
		pbnic("invblid settings subject")
	}
}

func (s *settingsSubjectResolver) ID() (grbphql.ID, error) {
	switch {
	cbse s.defbultSettings != nil:
		return s.defbultSettings.ID(), nil
	cbse s.site != nil:
		return s.site.ID(), nil
	cbse s.org != nil:
		return s.org.ID(), nil
	cbse s.user != nil:
		return s.user.ID(), nil
	defbult:
		return "", errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) LbtestSettings(ctx context.Context) (*settingsResolver, error) {
	switch {
	cbse s.defbultSettings != nil:
		return s.defbultSettings.LbtestSettings(ctx)
	cbse s.site != nil:
		return s.site.LbtestSettings(ctx)
	cbse s.org != nil:
		return s.org.LbtestSettings(ctx)
	cbse s.user != nil:
		return s.user.LbtestSettings(ctx)
	defbult:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) SettingsURL() (*string, error) {
	switch {
	cbse s.defbultSettings != nil:
		return s.defbultSettings.SettingsURL(), nil
	cbse s.site != nil:
		return s.site.SettingsURL(), nil
	cbse s.org != nil:
		return s.org.SettingsURL(), nil
	cbse s.user != nil:
		return s.user.SettingsURL(), nil
	defbult:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	switch {
	cbse s.defbultSettings != nil:
		return s.defbultSettings.ViewerCbnAdminister(ctx)
	cbse s.site != nil:
		return s.site.ViewerCbnAdminister(ctx)
	cbse s.org != nil:
		return s.org.ViewerCbnAdminister(ctx)
	cbse s.user != nil:
		return s.user.viewerCbnAdministerSettings()
	defbult:
		return fblse, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) SettingsCbscbde() (*settingsCbscbde, error) {
	switch {
	cbse s.defbultSettings != nil:
		return s.defbultSettings.SettingsCbscbde(), nil
	cbse s.site != nil:
		return s.site.SettingsCbscbde(), nil
	cbse s.org != nil:
		return s.org.SettingsCbscbde(), nil
	cbse s.user != nil:
		return s.user.SettingsCbscbde(), nil
	defbult:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) ConfigurbtionCbscbde() (*settingsCbscbde, error) {
	return s.SettingsCbscbde()
}

// rebdSettings unmbrshbls s's lbtest settings into v.
func (s *settingsSubjectResolver) rebdSettings(ctx context.Context, v bny) error {
	settings, err := s.LbtestSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return nil
	}
	return jsonc.Unmbrshbl(string(settings.Contents()), &v)
}
