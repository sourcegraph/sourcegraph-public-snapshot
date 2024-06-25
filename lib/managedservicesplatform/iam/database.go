package iam

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	openfga_assets "github.com/openfga/openfga/assets"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/redislock"
)

const databaseName = "msp_iam"

type metadata struct {
	AuthorizationModelID  string    `gorm:"column:authorization_model_id"`
	AuthorizationModelDSL string    `gorm:"column:authorization_model_dsl"`
	UpdatedAt             time.Time `gorm:"not null"`
}

// migrateAndReconcile migrates the "msp-iam" database schema (when needed) and
// reconciles the framework metadata.
func migrateAndReconcile(ctx context.Context, logger log.Logger, sqlDB *sql.DB, redisClient *redis.Client) (_ *metadata, err error) {
	ctx, span := iamTracer.Start(ctx, "iam.migrateAndReconcile",
		trace.WithAttributes(
			attribute.String("database", databaseName),
		))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	conn, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{
			SkipDefaultTransaction: true,
			NowFunc: func() time.Time {
				return time.Now().UTC().Truncate(time.Microsecond)
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open connection")
	}

	goose.SetBaseFS(openfga_assets.EmbedMigrations)
	goose.SetLogger(&gooseLoggerShim{Logger: logger})
	currentVersion, err := goose.GetDBVersionContext(ctx, sqlDB)
	if err != nil {
		return nil, errors.Wrap(err, "get DB version")
	}
	span.SetAttributes(attribute.Int64("currentVersion", currentVersion))

	// We want to make sure only one instance of the server is doing auto-migration
	// at a time.
	err = redislock.OnlyOne(
		logger,
		redisClient,
		fmt.Sprintf("%s:auto-migrate", databaseName),
		15*time.Second,
		func() error {
			ctx := context.WithoutCancel(ctx) // do not interrupt once we start
			span.AddEvent("lock.acquired")

			// Create a session that ignore debug logging.
			sess := conn.Session(&gorm.Session{
				Context: ctx,
				Logger:  gormlogger.Default.LogMode(gormlogger.Warn),
			})
			// Auto-migrate database table definitions.
			for _, table := range []any{&metadata{}} {
				span.AddEvent(fmt.Sprintf("automigrate.%s", fmt.Sprintf("%T", table)))
				err := sess.AutoMigrate(table)
				if err != nil {
					return errors.Wrapf(err, "auto migrating table for %s", errors.Safe(fmt.Sprintf("%T", table)))
				}
			}

			// Migrate OpenFGA's database schema.
			span.AddEvent("automigrate.openfga")
			err = goose.UpContext(
				ctx,
				sqlDB,
				openfga_assets.PostgresMigrationDir,
			)
			if err != nil {
				return errors.Wrap(err, "run OpenFGA migrations")
			}
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "auto-migrate")
	}

	var md metadata
	if err = conn.FirstOrCreate(&md).Error; err != nil {
		return nil, errors.Wrap(err, "init metadata")
	}
	return &md, nil
}

type gooseLoggerShim struct {
	log.Logger
}

func (l *gooseLoggerShim) Fatalf(format string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(format, v...),
		log.Error(errors.New("fatal Goose error")), // Sentinel error to trigger Sentry alerts.
	)
}

func (l *gooseLoggerShim) Printf(format string, v ...interface{}) {
	l.Debug(fmt.Sprintf(format, v...))
}
