package example

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/service"
)

type Config struct {
	StatelessMode bool
	Variable      string
}

func (c *Config) Load(env *service.Env) {
	c.StatelessMode = env.GetBool("STATELESSMODE", "false", "if true, disable dependencies")
	c.Variable = env.Get("VARIABLE", "13", "variable value")
}

type Service struct{}

var _ service.Service[Config] = Service{}

func (s Service) Name() string    { return "example" }
func (s Service) Version() string { return "dev" }
func (s Service) Initialize(
	ctx context.Context,
	logger log.Logger,
	contract service.Contract,
	config Config,
) (background.CombinedRoutine, error) {
	logger.Info("starting service")

	if !config.StatelessMode {
		if err := initDB(ctx, contract); err != nil {
			return nil, errors.Wrap(err, "initDB")
		}
		logger.Info("database configured")
	}

	return background.CombinedRoutine{
		&httpRoutine{
			log: logger,
			Server: &http.Server{
				Addr: fmt.Sprintf(":%d", contract.Port),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(fmt.Sprintf("Variable: %s", config.Variable)))
				}),
			},
		},
	}, nil
}

type httpRoutine struct {
	log log.Logger
	*http.Server
}

func (s *httpRoutine) Start() {
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("error stopping server", log.Error(err))
	}
}

func (s *httpRoutine) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		s.log.Error("error shutting down server", log.Error(err))
	} else {
		s.log.Info("server stopped")
	}
}

// initDB connects to a database 'primary' based on a DSN provided by contract.
// It then sets up a few example databases using Gorm, in a manner similar to
// https://github.com/sourcegraph/accounts.sourcegraph.com
func initDB(ctx context.Context, contract service.Contract) error {
	sqlDB, err := contract.GetPostgreSQLDB(ctx, "primary")
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
