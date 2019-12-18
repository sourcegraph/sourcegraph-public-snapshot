package confdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema/critical"
)

// RunMigrations runs configuration DB table migrations.
func RunMigrations(ctx context.Context) error {
	// Migrate critical configuration into the site configuration (merge the two).
	rawCritical, err := CriticalGetLatest(ctx)
	if err != nil {
		return err
	}
	rawSite, err := SiteGetLatest(ctx)
	if err != nil {
		return err
	}
	var critical critical.CriticalConfiguration
	if err := jsonc.Unmarshal(rawCritical.Contents, &critical); err != nil {
		return err
	}
	if critical.Migrated {
		return nil
	}
	if os.Getenv("SITE_CONFIG_FILE") != "" || os.Getenv("CRITICAL_CONFIG_FILE") != "" {
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("- IMPORTANT: Migrating critical configuration into site configuration.         -")
		fmt.Println("-                                                                              -")
		fmt.Println("- Please copy the updated contents of your site configuration found in the     -")
		fmt.Println("- Site Admin area into your SITE_CONFIG_FILE, otherwise your Sourcegraph       -")
		fmt.Println("- instance may be misconfigured when you next upgrade!                         -")
		fmt.Println("-                                                                              -")
		fmt.Println("--------------------------------------------------------------------------------")
	}
	for _, edit := range []struct {
		fieldName string
		value     interface{}
	}{
		{"auth.enableUsernameChanges", critical.AuthEnableUsernameChanges},
		{"auth.providers", critical.AuthProviders},
		{"auth.sessionExpiry", critical.AuthSessionExpiry},
		{"auth.userOrgMap", critical.AuthUserOrgMap},
		{"externalURL", critical.ExternalURL},
		{"htmlBodyBottom", critical.HtmlBodyBottom},
		{"htmlBodyTop", critical.HtmlBodyTop},
		{"htmlHeadBottom", critical.HtmlHeadBottom},
		{"htmlHeadTop", critical.HtmlHeadTop},
		{"licenseKey", critical.LicenseKey},
		{"lightstepAccessToken", critical.LightstepAccessToken},
		{"lightstepProject", critical.LightstepProject},
		{"log", critical.Log},
		{"update.channel", critical.UpdateChannel},
		{"useJaeger", critical.UseJaeger},
	} {
		// All of these fields are omitempty, so if they are zero values do not write them.
		if reflect.ValueOf(edit.value).IsZero() {
			continue
		}
		rawSite.Contents, err = jsonc.Edit(rawSite.Contents, edit.value, edit.fieldName)
		if err != nil {
			return err
		}
	}
	_, err = CriticalCreateIfUpToDate(ctx, &rawCritical.ID, `{"migrated": true}`)
	if err != nil {
		if err == ErrNewerEdit {
			// Since all frontends are racing to this point, we rely on the fact that one
			// of us will be the first to make this edit and that frontend owns performing
			// the migration.
			//
			// In theory there is a small chance we could have a DB connection failure
			// after doing this, or that in rare cases our process would die for some
			// unrelated reason -- but in practice this should be very rare and a site
			// admin would just need to copy/paste their critical configuration into their
			// site configuration via the escape hatch file.
			log15.Warn("migrating configuration: another frontend has already performed the migration, skipping")
			return nil
		}
		log15.Warn("migrating configuration: failed to update critical configuration", "error", err)
		return err
	}
	_, err = SiteCreateIfUpToDate(ctx, &rawSite.ID, rawSite.Contents)
	if err != nil {
		log15.Warn("migrating configuration: failed to update site configuration", "error", err)
		return err
	}
	return err
}

// Config contains the contents of a critical/site config along with associated metadata.
type Config struct {
	ID        int32     // the unique ID of this config
	Type      string    // either "critical" or "site"
	Contents  string    // the raw JSON content (with comments and trailing commas allowed)
	CreatedAt time.Time // the date when this config was created
	UpdatedAt time.Time // the date when this config was updated
}

// SiteConfig contains the contents of a site config along with associated metadata.
type SiteConfig Config

// CriticalConfig contains the contents of a critical config along with associated metadata.
type CriticalConfig Config

// ErrNewerEdit is returned by SiteCreateIfUpToDate and
// CriticalCreateifUpToDate when a newer edit has already been applied and the
// edit has been rejected.
var ErrNewerEdit = errors.New("someone else has already applied a newer edit")

// SiteCreateIfUpToDate saves the given site config "contents" to the database iff the
// supplied "lastID" is equal to the one that was most recently saved to the database.
//
// The site config that was most recently saved to the database is returned.
// An error is returned if "contents" is invalid JSON.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func SiteCreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (latest *SiteConfig, err error) {
	tx, done, err := newTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer done()

	newLastID, err := addDefault(ctx, tx, typeSite, confdefaults.Default.Site)
	if err != nil {
		return nil, err
	}
	if newLastID != nil {
		lastID = newLastID
	}

	criticalSite, err := createIfUpToDate(ctx, tx, typeSite, lastID, contents)
	return (*SiteConfig)(criticalSite), err
}

// CriticalCreateIfUpToDate saves the given critical config "contents" to the
// database iff the supplied "lastID" is equal to the one that was most
// recently saved to the database (i.e. SiteGetlatest's ID field).
//
// The critical config that was most recently saved to the database is returned.
// An error is returned if "contents" is invalid JSON.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func CriticalCreateIfUpToDate(ctx context.Context, lastID *int32, contents string) (latest *CriticalConfig, err error) {
	tx, done, err := newTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer done()

	newLastID, err := addDefault(ctx, tx, typeCritical, confdefaults.Default.Critical)
	if err != nil {
		return nil, err
	}
	if newLastID != nil {
		lastID = newLastID
	}

	criticalSite, err := createIfUpToDate(ctx, tx, typeCritical, lastID, contents)
	return (*CriticalConfig)(criticalSite), err
}

// SiteGetLatest returns the site config that was most recently saved to the database.
// This returns nil, nil if there is not yet a site config in the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func SiteGetLatest(ctx context.Context) (latest *SiteConfig, err error) {
	tx, done, err := newTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer done()

	_, err = addDefault(ctx, tx, typeSite, confdefaults.Default.Site)
	if err != nil {
		return nil, err
	}

	site, err := getLatest(ctx, tx, typeSite)
	return (*SiteConfig)(site), err
}

// CriticalGetLatest returns critical site config that was most recently saved to the database.
// This returns nil, nil if there is not yet a critical config in the database.
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func CriticalGetLatest(ctx context.Context) (latest *CriticalConfig, err error) {
	tx, done, err := newTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer done()

	_, err = addDefault(ctx, tx, typeCritical, confdefaults.Default.Critical)
	if err != nil {
		return nil, err
	}

	critical, err := getLatest(ctx, tx, typeCritical)
	return (*CriticalConfig)(critical), err
}

func newTransaction(ctx context.Context) (tx queryable, done func(), err error) {
	rtx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	return rtx, func() {
		if err != nil {
			rollErr := rtx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = rtx.Commit()
	}, nil
}

func addDefault(ctx context.Context, tx queryable, configType configType, contents string) (newLastID *int32, err error) {
	latest, err := getLatest(ctx, tx, configType)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		// We have an existing config!
		return nil, nil
	}

	// Create the default.
	latest, err = createIfUpToDate(ctx, tx, configType, nil, contents)
	if err != nil {
		return nil, err
	}
	return &latest.ID, nil
}

func createIfUpToDate(ctx context.Context, tx queryable, configType configType, lastID *int32, contents string) (latest *Config, err error) {
	// Validate JSON syntax before saving.
	if _, errs := jsonx.Parse(contents, jsonx.ParseOptions{Comments: true, TrailingCommas: true}); len(errs) > 0 {
		return nil, fmt.Errorf("invalid settings JSON: %v", errs)
	}

	new := Config{
		Contents: contents,
	}

	latest, err = getLatest(ctx, tx, configType)
	if err != nil {
		return nil, err
	}
	if latest != nil && lastID != nil && latest.ID != *lastID {
		return nil, ErrNewerEdit
	}

	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO critical_and_site_config(type, contents) VALUES($1, $2) RETURNING id, created_at, updated_at",
		configType, new.Contents,
	).Scan(&new.ID, &new.CreatedAt, &new.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &new, nil
}

func getLatest(ctx context.Context, tx queryable, configType configType) (*Config, error) {
	q := sqlf.Sprintf("SELECT s.id, s.type, s.contents, s.created_at, s.updated_at FROM critical_and_site_config s WHERE type=%s ORDER BY id DESC LIMIT 1", configType)
	rows, err := tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	versions, err := parseQueryRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(versions) != 1 {
		// No config has been written yet.
		return nil, nil
	}
	return versions[0], nil
}

func parseQueryRows(ctx context.Context, rows *sql.Rows) ([]*Config, error) {
	versions := []*Config{}
	defer rows.Close()
	for rows.Next() {
		f := Config{}
		err := rows.Scan(&f.ID, &f.Type, &f.Contents, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		versions = append(versions, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

// queryable allows us to reuse the same logic for certain operations both
// inside and outside an explicit transaction.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type configType string

const (
	typeCritical configType = "critical"
	typeSite     configType = "site"
)
