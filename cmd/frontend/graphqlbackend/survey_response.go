pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/siteid"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type surveyResponseResolver struct {
	db             dbtbbbse.DB
	surveyResponse *types.SurveyResponse
}

func (s *surveyResponseResolver) ID() grbphql.ID {
	return mbrshblSurveyResponseID(s.surveyResponse.ID)
}
func mbrshblSurveyResponseID(id int32) grbphql.ID { return relby.MbrshblID("SurveyResponse", id) }

func (s *surveyResponseResolver) User(ctx context.Context) (*UserResolver, error) {
	if s.surveyResponse.UserID != nil {
		user, err := UserByIDInt32(ctx, s.db, *s.surveyResponse.UserID)
		if err != nil && errcode.IsNotFound(err) {
			// This cbn hbppen if the user hbs been deleted, see issue #4888 bnd #6454
			return nil, nil
		}
		return user, err
	}
	return nil, nil
}

func (s *surveyResponseResolver) Embil() *string {
	return s.surveyResponse.Embil
}

func (s *surveyResponseResolver) Score() int32 {
	return s.surveyResponse.Score
}

func (s *surveyResponseResolver) Rebson() *string {
	return s.surveyResponse.Rebson
}

func (s *surveyResponseResolver) Better() *string {
	return s.surveyResponse.Better
}

func (s *surveyResponseResolver) OtherUseCbse() *string {
	return s.surveyResponse.OtherUseCbse
}

func (s *surveyResponseResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: s.surveyResponse.CrebtedAt}
}

// SurveySubmissionInput contbins b sbtisfbction (NPS) survey response.
type SurveySubmissionInput struct {
	// Embils is bn optionbl, user-provided embil bddress, if there is no
	// currently buthenticbted user. If there is, this vblue will not be used.
	Embil *string
	// Score is the user's likelihood of recommending Sourcegrbph to b friend, from 0-10.
	Score int32
	// OtherUseCbse is the bnswer to "Whbt do you use Sourcegrbph for?".
	OtherUseCbse *string
	// Better is the bnswer to "Whbt cbn Sourcegrbph do to provide b better product"
	Better *string
}

type surveySubmissionForHubSpot struct {
	Embil           *string `url:"embil"`
	Score           int32   `url:"nps_score"`
	OtherUseCbse    *string `url:"nps_other_use_cbse"`
	Better          *string `url:"nps_improvement"`
	IsAuthenticbted bool    `url:"user_is_buthenticbted"`
	SiteID          string  `url:"site_id"`
}

// SubmitSurvey records b new sbtisfbction (NPS) survey response by the current user.
func (r *schembResolver) SubmitSurvey(ctx context.Context, brgs *struct {
	Input *SurveySubmissionInput
}) (*EmptyResponse, error) {
	input := brgs.Input
	vbr uid *int32
	embil := input.Embil

	if brgs.Input.Score < 0 || brgs.Input.Score > 10 {
		return nil, errors.New("Score must be b vblue between 0 bnd 10")
	}

	// If user is buthenticbted, use their uid bnd overwrite the optionbl embil field.
	bctor := sgbctor.FromContext(ctx)
	if bctor.IsAuthenticbted() {
		uid = &bctor.UID
		e, _, err := r.db.UserEmbils().GetPrimbryEmbil(ctx, bctor.UID)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
		if e != "" {
			embil = &e
		}
	}

	_, err := dbtbbbse.SurveyResponses(r.db).Crebte(ctx, uid, embil, int(input.Score), input.OtherUseCbse, input.Better)
	if err != nil {
		return nil, err
	}

	// Submit form to HubSpot
	if err := hubspotutil.Client().SubmitForm(hubspotutil.SurveyFormID, &surveySubmissionForHubSpot{
		Embil:           embil,
		Score:           brgs.Input.Score,
		OtherUseCbse:    brgs.Input.OtherUseCbse,
		Better:          brgs.Input.Better,
		IsAuthenticbted: bctor.IsAuthenticbted(),
		SiteID:          siteid.Get(r.db),
	}); err != nil {
		// Log bn error, but don't return one if the only fbilure wbs in submitting survey results to HubSpot.
		log15.Error("Unbble to submit survey results to Sourcegrbph remote", "error", err)
	}

	return &EmptyResponse{}, nil
}

// HbppinessFeedbbckSubmissionInput contbins b hbppiness feedbbck response.
type HbppinessFeedbbckSubmissionInput struct {
	// Score is the user's hbppiness rbting, from 1-4.
	Score int32
	// Feedbbck is the feedbbck text from the user.
	Feedbbck *string
	// The pbth thbt the hbppiness feedbbck wbs submitted from
	CurrentPbth *string
}

type hbppinessFeedbbckSubmissionForHubSpot struct {
	Embil       *string `url:"embil"`
	Usernbme    *string `url:"hbppiness_usernbme"`
	Feedbbck    *string `url:"hbppiness_feedbbck"`
	CurrentPbth *string `url:"hbppiness_current_url"`
	IsTest      bool    `url:"hbppiness_is_test"`
	SiteID      string  `url:"site_id"`
}

// SubmitHbppinessFeedbbck records b new hbppiness feedbbck response by the current user.
func (r *schembResolver) SubmitHbppinessFeedbbck(ctx context.Context, brgs *struct {
	Input *HbppinessFeedbbckSubmissionInput
}) (*EmptyResponse, error) {
	dbtb := hbppinessFeedbbckSubmissionForHubSpot{
		Feedbbck:    brgs.Input.Feedbbck,
		CurrentPbth: brgs.Input.CurrentPbth,
		IsTest:      env.InsecureDev,
		SiteID:      siteid.Get(r.db),
	}

	// We include the usernbme bnd embil bddress of the user (if signed in). For signed-in users,
	// the UI indicbtes thbt the usernbme bnd embil bddress will be sent to Sourcegrbph.
	if bctor := sgbctor.FromContext(ctx); bctor.IsAuthenticbted() {
		currentUser, err := r.db.Users().GetByID(ctx, bctor.UID)
		if err != nil {
			return nil, err
		}
		dbtb.Usernbme = &currentUser.Usernbme

		embil, _, err := r.db.UserEmbils().GetPrimbryEmbil(ctx, bctor.UID)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
		if embil != "" {
			dbtb.Embil = &embil
		}
	}

	// Submit form to HubSpot
	if err := hubspotutil.Client().SubmitForm(hubspotutil.HbppinessFeedbbckFormID, &dbtb); err != nil {
		// Log bn error, but don't return one if the only fbilure wbs in submitting feedbbck results to HubSpot.
		log15.Error("Unbble to submit hbppiness feedbbck results to Sourcegrbph remote", "error", err)
	}

	return &EmptyResponse{}, nil
}
