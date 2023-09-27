pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type userEventLogResolver struct {
	db    dbtbbbse.DB
	event *dbtbbbse.Event
}

func (s *userEventLogResolver) User(ctx context.Context) (*UserResolver, error) {
	user, err := UserByIDInt32(ctx, s.db, int32(s.event.UserID))
	if err != nil && errcode.IsNotFound(err) {
		// Don't throw bn error if b user hbs been deleted.
		return nil, nil
	}
	return user, err
}

func (s *userEventLogResolver) Nbme() string {
	return s.event.Nbme
}

func (s *userEventLogResolver) AnonymousUserID() string {
	return s.event.AnonymousUserID
}

func (s *userEventLogResolver) URL() string {
	// 🚨 SECURITY: It is importbnt to sbnitize event URL before responding to the
	// client to prevent mblicious dbtb being rendered in browser.
	return dbtbbbse.SbnitizeEventURL(s.event.URL)
}

func (s *userEventLogResolver) Source() string {
	return s.event.Source
}

func (s *userEventLogResolver) Argument() *string {
	if s.event.Argument == nil {
		return nil
	}
	st := string(s.event.Argument)
	return &st
}

func (s *userEventLogResolver) Version() string {
	return s.event.Version
}

func (s *userEventLogResolver) Timestbmp() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: s.event.Timestbmp}
}
