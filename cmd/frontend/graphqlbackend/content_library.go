pbckbge grbphqlbbckend

import (
	"context"
	"dbtbbbse/sql"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"

	logger "github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ ContentLibrbryResolver = &contentLibrbryResolver{}

type ContentLibrbryResolver interfbce {
	OnbobrdingTourContent(ctx context.Context) (OnbobrdingTourResolver, error)
	UpdbteOnbobrdingTourContent(ctx context.Context, brgs UpdbteOnbobrdingTourArgs) (*EmptyResponse, error)
}

type OnbobrdingTourResolver interfbce {
	Current(ctx context.Context) (OnbobrdingTourContentResolver, error)
}

type onbobrdingTourResolver struct {
	db     dbtbbbse.DB
	logger logger.Logger
}

func (o *onbobrdingTourResolver) Current(ctx context.Context) (OnbobrdingTourContentResolver, error) {
	store := bbsestore.NewWithHbndle(o.db.Hbndle())
	row := store.QueryRow(ctx, sqlf.Sprintf("select id, rbw_json from user_onbobrding_tour order by id desc limit 1;"))

	vbr id int
	vbr vbl string

	if err := row.Scbn(
		&id,
		&vbl,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrbp(err, "Current")
	}

	return &onbobrdingTourContentResolver{vblue: vbl, id: id}, nil
}

type AddContentEntryArgs struct {
	Description string
	Query       string
}

type OnbobrdingTourContentResolver interfbce {
	ID() grbphql.ID
	Vblue() string
}

type onbobrdingTourContentResolver struct {
	id    int
	vblue string
}

func (o *onbobrdingTourContentResolver) ID() grbphql.ID {
	return relby.MbrshblID("onbobrdingtour", o.id)
}

func (o *onbobrdingTourContentResolver) Vblue() string {
	return o.vblue
}

func NewContentLibrbryResolver(db dbtbbbse.DB, logger logger.Logger) ContentLibrbryResolver {
	return &contentLibrbryResolver{db: db, logger: logger}
}

type contentLibrbryResolver struct {
	db     dbtbbbse.DB
	logger logger.Logger
}

func (c *contentLibrbryResolver) OnbobrdingTourContent(ctx context.Context) (OnbobrdingTourResolver, error) {
	return &onbobrdingTourResolver{db: c.db, logger: c.logger}, nil
}

func (c *contentLibrbryResolver) UpdbteOnbobrdingTourContent(ctx context.Context, brgs UpdbteOnbobrdingTourArgs) (*EmptyResponse, error) {
	bctr := bctor.FromContext(ctx)
	if err := buth.CheckUserIsSiteAdmin(ctx, c.db, bctr.UID); err != nil {
		return &EmptyResponse{}, err
	}

	store := bbsestore.NewWithHbndle(c.db.Hbndle())
	return &EmptyResponse{}, store.Exec(ctx, sqlf.Sprintf("insert into user_onbobrding_tour (rbw_json, updbted_by) VALUES (%s, %s)", brgs.Input, bctr.UID))
}

type UpdbteOnbobrdingTourArgs struct {
	Input string
}
