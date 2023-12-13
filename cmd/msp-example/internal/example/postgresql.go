package example

import (
	"context"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// initPostgreSQL connects to a database 'primary' based on a DSN provided by contract.
// It then sets up a few example databases using Gorm, in a manner similar to
// https://github.com/sourcegraph/accounts.sourcegraph.com
func initPostgreSQL(ctx context.Context, contract runtime.Contract) error {
	sqlDB, err := contract.PostgreSQL.OpenDatabase(ctx, "primary")
	if err != nil {
		return errors.Wrap(err, "GetPostgreSQLDB")
	}
	db, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{
			SkipDefaultTransaction: true,
			NowFunc: func() time.Time {
				return time.Now().UTC().Truncate(time.Microsecond)
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "gorm.Open")
	}

	for _, table := range []any{
		&User{},
		&Email{},
	} {
		if err = db.AutoMigrate(table); err != nil {
			return errors.Wrapf(err, "auto migrating table for %T", table)
		}
	}

	return nil
}

type User struct {
	ID        int64          `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ExternalID string `gorm:"size:36;not null;uniqueIndex,where:deleted_at IS NULL"`
	Name       string `gorm:"size:256;not null"`
	AvatarURL  string `gorm:"size:256;not null"`
}

type Email struct {
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID     int64  `gorm:"not null;uniqueIndex:,where:deleted_at IS NULL AND verified_at IS NOT NULL"`
	Email      string `gorm:"size:256;not null;uniqueIndex:,where:deleted_at IS NULL AND verified_at IS NOT NULL"`
	VerifiedAt *time.Time

	// ⚠️ DO NOT USE: This field is only used for creating foreign key constraint.
	User *User `gorm:"foreignKey:UserID"`
}
