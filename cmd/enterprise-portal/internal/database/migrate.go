package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
	"github.com/sourcegraph/sourcegraph/lib/redislock"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/tables"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/tables/custommigrator"
)

// maybeMigrate runs the auto-migration for the database when needed based on
// the given version.
func maybeMigrate(ctx context.Context, logger log.Logger, contract runtime.Contract, redisClient *redis.Client, currentVersion string) (err error) {
	dbName := databaseName(contract.MSP)
	ctx, span := databaseTracer.Start(
		ctx,
		"database.maybeMigrate",
		trace.WithAttributes(
			attribute.String("currentVersion", currentVersion),
			attribute.String("database", dbName),
		),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	logger = logger.
		WithTrace(log.TraceContext{
			TraceID: span.SpanContext().TraceID().String(),
			SpanID:  span.SpanContext().SpanID().String(),
		}).
		With(log.String("database", dbName))

	sqlDB, err := contract.PostgreSQL.OpenDatabase(ctx, dbName)
	if err != nil {
		return errors.Wrap(err, "open database")
	}
	defer func() {
		err := sqlDB.Close()
		if err != nil {
			logger.Error("failed to close database for migration", log.Error(err))
		}
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
		return errors.Wrap(err, "open connection")
	}

	if err = conn.Use(tracing.NewPlugin()); err != nil {
		return errors.Wrap(err, "initialize tracing plugin")
	}

	// We want to make sure only one instance of the server is doing auto-migration
	// at a time.
	return redislock.OnlyOne(
		logger,
		redisClient,
		fmt.Sprintf("%s:auto-migrate", dbName),
		15*time.Second,
		func() error {
			ctx := context.WithoutCancel(ctx) // do not interrupt once we start
			span.AddEvent("lock.acquired")

			versionKey := fmt.Sprintf("%s:db_version", dbName)
			liveVersion := redisClient.Get(ctx, versionKey).Val()
			if shouldSkipMigration(
				liveVersion,
				currentVersion,
			) {
				logger.Info("skipped auto-migration",
					log.String("currentVersion", currentVersion),
				)
				span.SetAttributes(attribute.Bool("skipped", true))
				return nil
			}
			logger.Info("executing auto-migration",
				log.String("liveVersion", liveVersion),
				log.String("currentVersion", currentVersion))
			span.SetAttributes(attribute.Bool("skipped", false))

			// Create a session that ignore debug logging.
			sess := conn.Session(&gorm.Session{
				Context: ctx,
				Logger:  gormlogger.Default.LogMode(gormlogger.Warn),
			})
			// Initialize extensions.
			if err := conn.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`).Error; err != nil {
				return errors.Wrap(err, "install pgcrypto extension")
			}
			// Auto-migrate database table definitions.
			for _, table := range tables.All() {
				span.AddEvent(fmt.Sprintf("automigrate.%s", table.TableName()))
				err := sess.AutoMigrate(table)
				if err != nil {
					return errors.Wrapf(err, "auto migrating table for %s", errors.Safe(fmt.Sprintf("%T", table)))
				}
				if m, ok := table.(custommigrator.CustomTableMigrator); ok {
					if err := m.RunCustomMigrations(sess.Migrator()); err != nil {
						return errors.Wrapf(err, "running custom migrations for %s", errors.Safe(fmt.Sprintf("%T", table)))
					}
				}
			}

			return redisClient.Set(ctx, versionKey, currentVersion, 0).Err()
		},
	)
}

// shouldSkipMigration returns true if the migration should be skipped.
func shouldSkipMigration(previousVersion, currentVersion string) bool {
	// Skip for PR-builds.
	if strings.HasPrefix(currentVersion, "_candidate") {
		return true
	}

	const releaseBuildVersionExample = "277307_2024-06-06_5.4-9185da3c3e42"
	// We always run the full auto-migration if the version is not release-build like.
	if len(currentVersion) < len(releaseBuildVersionExample) ||
		len(previousVersion) < len(releaseBuildVersionExample) {
		return false
	}

	// The release build version is sorted lexicographically, so we can compare it as a string.
	return previousVersion >= currentVersion
}
