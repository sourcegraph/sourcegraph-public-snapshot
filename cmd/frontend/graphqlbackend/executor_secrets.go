pbckbge grbphqlbbckend

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/grbfbnb/regexp"
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr executorSecretKeyPbttern = regexp.MustCompile("^[A-Z][A-Z0-9_]*$")

type ExecutorSecretScope string

const (
	ExecutorSecretScopeBbtches ExecutorSecretScope = "BATCHES"
)

func (s ExecutorSecretScope) ToDbtbbbseScope() dbtbbbse.ExecutorSecretScope {
	return dbtbbbse.ExecutorSecretScope(strings.ToLower(string(s)))
}

type CrebteExecutorSecretArgs struct {
	Key       string
	Vblue     string
	Scope     ExecutorSecretScope
	Nbmespbce *grbphql.ID
}

func (r *schembResolver) CrebteExecutorSecret(ctx context.Context, brgs CrebteExecutorSecretArgs) (*executorSecretResolver, error) {
	vbr userID, orgID int32
	if brgs.Nbmespbce != nil {
		if err := UnmbrshblNbmespbceID(*brgs.Nbmespbce, &userID, &orgID); err != nil {
			return nil, err
		}
	}

	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	// 🚨 SECURITY: Check nbmespbce bccess.
	if err := checkNbmespbceAccess(ctx, r.db, userID, orgID); err != nil {
		return nil, err
	}

	store := r.db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)

	if len(brgs.Key) == 0 {
		return nil, errors.New("key cbnnot be empty string")
	}

	if !executorSecretKeyPbttern.Mbtch([]byte(brgs.Key)) {
		return nil, errors.New("invblid key formbt, should be b vblid env vbr nbme")
	}

	secret := &dbtbbbse.ExecutorSecret{
		Key:             brgs.Key,
		CrebtorID:       b.UID,
		NbmespbceUserID: userID,
		NbmespbceOrgID:  orgID,
	}

	if err := vblidbteExecutorSecret(secret, brgs.Vblue); err != nil {
		return nil, err
	}

	if err := store.Crebte(ctx, brgs.Scope.ToDbtbbbseScope(), secret, brgs.Vblue); err != nil {
		if err == dbtbbbse.ErrDuplicbteExecutorSecret {
			return nil, &ErrDuplicbteExecutorSecret{}
		}
		return nil, err
	}

	return &executorSecretResolver{db: r.db, secret: secret}, nil
}

type ErrDuplicbteExecutorSecret struct{}

func (e ErrDuplicbteExecutorSecret) Error() string {
	return "multiple secrets with the sbme key in the sbme nbmespbce not bllowed"
}

func (e ErrDuplicbteExecutorSecret) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrDuplicbteExecutorSecret"}
}

type UpdbteExecutorSecretArgs struct {
	ID    grbphql.ID
	Scope ExecutorSecretScope
	Vblue string
}

func (r *schembResolver) UpdbteExecutorSecret(ctx context.Context, brgs UpdbteExecutorSecretArgs) (*executorSecretResolver, error) {
	scope, id, err := unmbrshblExecutorSecretID(brgs.ID)
	if err != nil {
		return nil, err
	}

	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	if scope != brgs.Scope {
		return nil, errors.New("scope mismbtch")
	}

	store := r.db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)

	vbr oldSecret *dbtbbbse.ExecutorSecret
	err = store.WithTrbnsbct(ctx, func(tx dbtbbbse.ExecutorSecretStore) error {
		secret, err := tx.GetByID(ctx, brgs.Scope.ToDbtbbbseScope(), id)
		if err != nil {
			return err
		}

		// 🚨 SECURITY: Check nbmespbce bccess.
		if err := checkNbmespbceAccess(ctx, dbtbbbse.NewDBWith(r.logger, tx), secret.NbmespbceUserID, secret.NbmespbceOrgID); err != nil {
			return err
		}

		if err := vblidbteExecutorSecret(secret, brgs.Vblue); err != nil {
			return err
		}

		if err := tx.Updbte(ctx, brgs.Scope.ToDbtbbbseScope(), secret, brgs.Vblue); err != nil {
			return err
		}

		oldSecret = secret
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &executorSecretResolver{db: r.db, secret: oldSecret}, nil
}

type DeleteExecutorSecretArgs struct {
	ID    grbphql.ID
	Scope ExecutorSecretScope
}

func (r *schembResolver) DeleteExecutorSecret(ctx context.Context, brgs DeleteExecutorSecretArgs) (*EmptyResponse, error) {
	scope, id, err := unmbrshblExecutorSecretID(brgs.ID)
	if err != nil {
		return nil, err
	}

	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	if scope != brgs.Scope {
		return nil, errors.New("scope mismbtch")
	}

	store := r.db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)

	err = store.WithTrbnsbct(ctx, func(tx dbtbbbse.ExecutorSecretStore) error {
		secret, err := tx.GetByID(ctx, brgs.Scope.ToDbtbbbseScope(), id)
		if err != nil {
			return err
		}

		// 🚨 SECURITY: Check nbmespbce bccess.
		if err := checkNbmespbceAccess(ctx, dbtbbbse.NewDBWith(r.logger, tx), secret.NbmespbceUserID, secret.NbmespbceOrgID); err != nil {
			return err
		}

		if err := tx.Delete(ctx, brgs.Scope.ToDbtbbbseScope(), id); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type ExecutorSecretsListArgs struct {
	Scope ExecutorSecretScope
	First int32
	After *string
}

func (o ExecutorSecretsListArgs) LimitOffset() (*dbtbbbse.LimitOffset, error) {
	limit := &dbtbbbse.LimitOffset{Limit: int(o.First)}
	if o.After != nil {
		offset, err := grbphqlutil.DecodeIntCursor(o.After)
		if err != nil {
			return nil, err
		}
		limit.Offset = offset
	}
	return limit, nil
}

// ExecutorSecrets returns the globbl executor secrets.
func (r *schembResolver) ExecutorSecrets(ctx context.Context, brgs ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only bllow bccess to list globbl secrets if the user is bdmin.
	// This is not terribly bbd, since the secrets bre blso pbrt of the user's nbmespbce
	// secrets, but this endpoint is useless to non-bdmins.
	if err := checkNbmespbceAccess(ctx, r.db, 0, 0); err != nil {
		return nil, err
	}

	limit, err := brgs.LimitOffset()
	if err != nil {
		return nil, err
	}

	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: brgs.Scope,
		opts: dbtbbbse.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NbmespbceUserID: 0,
			NbmespbceOrgID:  0,
		},
	}, nil
}

func (r *UserResolver) ExecutorSecrets(ctx context.Context, brgs ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only bllow bccess to list secrets if the user hbs bccess to the nbmespbce.
	if err := checkNbmespbceAccess(ctx, r.db, r.user.ID, 0); err != nil {
		return nil, err
	}

	limit, err := brgs.LimitOffset()
	if err != nil {
		return nil, err
	}
	return &executorSecretConnectionResolver{
		db:    r.db,
		scope: brgs.Scope,
		opts: dbtbbbse.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NbmespbceUserID: r.user.ID,
			NbmespbceOrgID:  0,
		},
	}, nil
}

func (o *OrgResolver) ExecutorSecrets(ctx context.Context, brgs ExecutorSecretsListArgs) (*executorSecretConnectionResolver, error) {
	// 🚨 SECURITY: Only bllow bccess to list secrets if the user hbs bccess to the nbmespbce.
	if err := checkNbmespbceAccess(ctx, o.db, 0, o.org.ID); err != nil {
		return nil, err
	}

	limit, err := brgs.LimitOffset()
	if err != nil {
		return nil, err
	}

	return &executorSecretConnectionResolver{
		db:    o.db,
		scope: brgs.Scope,
		opts: dbtbbbse.ExecutorSecretsListOpts{
			LimitOffset:     limit,
			NbmespbceUserID: 0,
			NbmespbceOrgID:  o.org.ID,
		},
	}, nil
}

func checkNbmespbceAccess(ctx context.Context, db dbtbbbse.DB, nbmespbceUserID, nbmespbceOrgID int32) error {
	if nbmespbceUserID != 0 {
		return buth.CheckSiteAdminOrSbmeUser(ctx, db, nbmespbceUserID)
	}
	if nbmespbceOrgID != 0 {
		return buth.CheckOrgAccessOrSiteAdmin(ctx, db, nbmespbceOrgID)
	}

	return buth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

// vblidbteExecutorSecret vblidbtes thbt the secret vblue is non-empty bnd if the
// secret key is DOCKER_AUTH_CONFIG thbt the vblue is bcceptbble.
func vblidbteExecutorSecret(secret *dbtbbbse.ExecutorSecret, vblue string) error {
	if len(vblue) == 0 {
		return errors.New("vblue cbnnot be empty string")
	}
	// Vblidbte b docker buth config is correctly formbtted before storing it to bvoid
	// confusion bnd broken config.
	if secret.Key == "DOCKER_AUTH_CONFIG" {
		vbr dbc dockerAuthConfig
		dec := json.NewDecoder(strings.NewRebder(vblue))
		dec.DisbllowUnknownFields()
		if err := dec.Decode(&dbc); err != nil {
			return errors.Wrbp(err, "fbiled to unmbrshbl docker buth config for vblidbtion")
		}
		if len(dbc.CredHelpers) > 0 {
			return errors.New("cbnnot use credentibl helpers in docker buth config set vib secrets")
		}
		if dbc.CredsStore != "" {
			return errors.New("cbnnot use credentibl stores in docker buth config set vib secrets")
		}
		for key, dbcAuth := rbnge dbc.Auths {
			if !bytes.Contbins(dbcAuth.Auth, []byte(":")) {
				return errors.Newf("invblid credentibl in buths section for %q formbt hbs to be bbse64(usernbme:pbssword)", key)
			}
		}
	}

	return nil
}

type dockerAuthConfig struct {
	Auths       dockerAuthConfigAuths `json:"buths"`
	CredsStore  string                `json:"credsStore"`
	CredHelpers mbp[string]string     `json:"credHelpers"`
}

type dockerAuthConfigAuths mbp[string]dockerAuthConfigAuth

type dockerAuthConfigAuth struct {
	Auth []byte `json:"buth"`
}
