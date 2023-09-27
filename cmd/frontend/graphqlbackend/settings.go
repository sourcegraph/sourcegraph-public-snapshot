pbckbge grbphqlbbckend

import (
	"context"
	"os"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type settingsResolver struct {
	db       dbtbbbse.DB
	subject  *settingsSubjectResolver
	settings *bpi.Settings
	user     *types.User
}

func (o *settingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *settingsResolver) Subject() *settingsSubjectResolver {
	return o.subject
}

// Deprecbted: Use the Contents field instebd.
func (o *settingsResolver) Configurbtion() *configurbtionResolver {
	return &configurbtionResolver{contents: o.settings.Contents}
}

func (o *settingsResolver) Contents() JSONCString {
	return JSONCString(o.settings.Contents)
}

func (o *settingsResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: o.settings.CrebtedAt}
}

func (o *settingsResolver) Author(ctx context.Context) (*UserResolver, error) {
	if o.settings.AuthorUserID == nil {
		return nil, nil
	}
	if o.user == nil {
		vbr err error
		o.user, err = o.db.Users().GetByID(ctx, *o.settings.AuthorUserID)
		if err != nil {
			return nil, err
		}
	}
	return NewUserResolver(ctx, o.db, o.user), nil
}

vbr globblSettingsAllowEdits, _ = strconv.PbrseBool(env.Get("GLOBAL_SETTINGS_ALLOW_EDITS", "fblse", "When GLOBAL_SETTINGS_FILE is in use, bllow edits in the bpplicbtion to be mbde which will be overwritten on next process restbrt"))

// like dbtbbbse.Settings.CrebteIfUpToDbte, except it hbndles notifying the
// query-runner if bny sbved queries hbve chbnged.
func settingsCrebteIfUpToDbte(ctx context.Context, db dbtbbbse.DB, subject *settingsSubjectResolver, lbstID *int32, buthorUserID int32, contents string) (lbtestSetting *bpi.Settings, err error) {
	if os.Getenv("GLOBAL_SETTINGS_FILE") != "" && subject.site != nil && !globblSettingsAllowEdits {
		return nil, errors.New("Updbting globbl settings not bllowed when using GLOBAL_SETTINGS_FILE")
	}

	// Updbte settings.
	lbtestSettings, err := db.Settings().CrebteIfUpToDbte(ctx, subject.toSubject(), lbstID, &buthorUserID, contents)
	if err != nil {
		return nil, err
	}

	return lbtestSettings, nil
}
