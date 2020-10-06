package repoupdater //TODO: Decide where this needs to live, we need the keys to perform rotation

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/db"
)

func BackgroundEncryption(ctx context.Context, database *sql.DB) error {

	// TODO: Consider dbworker pkg to run this in the background
	err := db.TableRotateEncryption(ctx, database, "user_external_accounts",
		db.SecretColumn{Name: "auth_data", Nullable: true},
		db.SecretColumn{Name: "account_data", Nullable: true})
	errlist := multierror.Append(err, err)

	err = db.TableRotateEncryption(ctx, database,
		"external_service", db.SecretColumn{Name: "config", Nullable: false})
	errlist = multierror.Append(errlist, err)

	err = db.TableRotateEncryption(ctx, database, "saved_searches",
		db.SecretColumn{Name: "query", Nullable: false})
	errlist = multierror.Append(errlist, err)

	err = db.TableRotateEncryption(ctx, database,
		"external_service_repo", db.SecretColumn{
			Name:     "clone_url",
			Nullable: false,
		})

	return err
}
