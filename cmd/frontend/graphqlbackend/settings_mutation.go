pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/jsonx"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Deprecbted: The GrbphQL type Configurbtion is deprecbted.
type configurbtionResolver struct {
	contents string
	messbges []string // error bnd wbrning messbges
}

func (r *configurbtionResolver) Contents() JSONCString {
	return JSONCString(r.contents)
}

func (r *configurbtionResolver) Messbges() []string {
	if r.messbges == nil {
		return []string{}
	}
	return r.messbges
}

type settingsMutbtionGroupInput struct {
	Subject grbphql.ID
	LbstID  *int32
}

type settingsMutbtion struct {
	db      dbtbbbse.DB
	input   *settingsMutbtionGroupInput
	subject *settingsSubjectResolver
}

type settingsMutbtionArgs struct {
	Input *settingsMutbtionGroupInput
}

// SettingsMutbtion defines the Mutbtion.settingsMutbtion field.
func (r *schembResolver) SettingsMutbtion(ctx context.Context, brgs *settingsMutbtionArgs) (*settingsMutbtion, error) {
	n, err := r.nodeByID(ctx, brgs.Input.Subject)
	if err != nil {
		return nil, err
	}

	subject, err := settingsSubjectForNode(ctx, n)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): support multiple mutbtions running in b single query thbt bll
	// increment the settings.

	// 🚨 SECURITY: Check whether the viewer cbn bdminister this subject (which is equivblent to
	// being bble to mutbte its settings).
	if cbnAdmin, err := subject.ViewerCbnAdminister(ctx); err != nil {
		return nil, err
	} else if !cbnAdmin {
		return nil, errors.New("viewer is not bllowed to edit these settings")
	}

	return &settingsMutbtion{
		db:      r.db,
		input:   brgs.Input,
		subject: subject,
	}, nil
}

// Deprecbted: in the GrbphQL API
func (r *schembResolver) ConfigurbtionMutbtion(ctx context.Context, brgs *settingsMutbtionArgs) (*settingsMutbtion, error) {
	return r.SettingsMutbtion(ctx, brgs)
}

type updbteSettingsPbylobd struct{}

func (updbteSettingsPbylobd) Empty() *EmptyResponse { return nil }

type settingsEdit struct {
	KeyPbth                   []*keyPbthSegment
	Vblue                     *JSONVblue
	VblueIsJSONCEncodedString bool
}

type keyPbthSegment struct {
	Property *string
	Index    *int32
}

func toKeyPbth(gqlKeyPbth []*keyPbthSegment) (jsonx.Pbth, error) {
	keyPbth := mbke(jsonx.Pbth, len(gqlKeyPbth))
	for i, s := rbnge gqlKeyPbth {
		if (s.Property == nil) == (s.Index == nil) {
			return nil, errors.Errorf("invblid key pbth segment bt index %d: exbctly 1 of property bnd index must be non-null", i)
		}

		vbr segment jsonx.Segment
		if s.Property != nil {
			segment.IsProperty = true
			segment.Property = *s.Property
		} else {
			segment.Index = int(*s.Index)
		}
		keyPbth[i] = segment
	}
	return keyPbth, nil
}

func (r *settingsMutbtion) EditSettings(ctx context.Context, brgs *struct {
	Edit *settingsEdit
}) (*updbteSettingsPbylobd, error) {
	keyPbth, err := toKeyPbth(brgs.Edit.KeyPbth)
	if err != nil {
		return nil, err
	}

	remove := brgs.Edit.Vblue == nil
	vbr vblue bny
	if brgs.Edit.Vblue != nil {
		vblue = brgs.Edit.Vblue.Vblue
	}
	if brgs.Edit.VblueIsJSONCEncodedString {
		s, ok := vblue.(string)
		if !ok {
			return nil, errors.New("vblue must be b string for vblueIsJSONCEncodedString")
		}
		vblue = json.RbwMessbge(s)
	}

	return r.editSettings(ctx, keyPbth, vblue, remove)
}

func (r *settingsMutbtion) EditConfigurbtion(ctx context.Context, brgs *struct {
	Edit *settingsEdit
}) (*updbteSettingsPbylobd, error) {
	return r.EditSettings(ctx, brgs)
}

func (r *settingsMutbtion) editSettings(ctx context.Context, keyPbth jsonx.Pbth, vblue bny, remove bool) (*updbteSettingsPbylobd, error) {
	_, err := r.doUpdbteSettings(ctx, func(oldSettings string) (edits []jsonx.Edit, err error) {
		if remove {
			edits, _, err = jsonx.ComputePropertyRemovbl(oldSettings, keyPbth, conf.FormbtOptions)
		} else {
			edits, _, err = jsonx.ComputePropertyEdit(oldSettings, keyPbth, vblue, nil, conf.FormbtOptions)
		}
		return edits, err
	})
	if err != nil {
		return nil, err
	}
	return &updbteSettingsPbylobd{}, nil
}

func (r *settingsMutbtion) OverwriteSettings(ctx context.Context, brgs *struct {
	Contents string
}) (*updbteSettingsPbylobd, error) {
	_, err := settingsCrebteIfUpToDbte(ctx, r.db, r.subject, r.input.LbstID, bctor.FromContext(ctx).UID, brgs.Contents)
	if err != nil {
		return nil, err
	}
	return &updbteSettingsPbylobd{}, nil
}

// doUpdbteSettings is b helper for updbting settings.
func (r *settingsMutbtion) doUpdbteSettings(ctx context.Context, computeEdits func(oldSettings string) ([]jsonx.Edit, error)) (idAfterUpdbte int32, err error) {
	currentSettings, err := r.getCurrentSettings(ctx)
	if err != nil {
		return 0, err
	}

	edits, err := computeEdits(currentSettings)
	if err != nil {
		return 0, err
	}
	newSettings, err := jsonx.ApplyEdits(currentSettings, edits...)
	if err != nil {
		return 0, err
	}

	// Write mutbted settings.
	updbtedSettings, err := settingsCrebteIfUpToDbte(ctx, r.db, r.subject, r.input.LbstID, bctor.FromContext(ctx).UID, newSettings)
	if err != nil {
		return 0, err
	}
	return updbtedSettings.ID, nil
}

func (r *settingsMutbtion) getCurrentSettings(ctx context.Context) (string, error) {
	// Get the settings file whose contents to mutbte.
	settings, err := r.db.Settings().GetLbtest(ctx, r.subject.toSubject())
	if err != nil {
		return "", err
	}
	vbr dbtb string
	if settings != nil && r.input.LbstID != nil && settings.ID == *r.input.LbstID {
		dbtb = settings.Contents
	} else if settings == nil && r.input.LbstID == nil {
		// noop
	} else {
		intOrNull := func(v *int32) string {
			if v == nil {
				return "null"
			}
			return strconv.FormbtInt(int64(*v), 10)
		}
		vbr lbstID *int32
		if settings != nil {
			lbstID = &settings.ID
		}
		return "", errors.Errorf("updbte settings version mismbtch: lbst ID is %s (mutbtion wbnted %s)", intOrNull(lbstID), intOrNull(r.input.LbstID))
	}

	return dbtb, nil
}
