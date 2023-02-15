package dbconn

import (
	"testing"
	"time"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/hexops/autogold/v2"
	"github.com/jackc/pgx/v4"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	tests := []struct {
		name                    string
		dataSource              string
		expectedApplicationName string
		fails                   bool
	}{
		{
			name:                    "empty dataSource",
			dataSource:              "",
			expectedApplicationName: defaultApplicationName,
			fails:                   false,
		}, {
			name:                    "connection string",
			dataSource:              "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph",
			expectedApplicationName: defaultApplicationName,
			fails:                   false,
		}, {
			name:                    "connection string with application name",
			dataSource:              "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph application_name=foo",
			expectedApplicationName: "foo",
			fails:                   false,
		}, {
			name:                    "postgres URL",
			dataSource:              "postgres://sourcegraph@localhost/sourcegraph?sslmode=verify-full",
			expectedApplicationName: defaultApplicationName,
			fails:                   false,
		}, {
			name:                    "postgres URL with fallback",
			dataSource:              "postgres://sourcegraph@localhost/sourcegraph?sslmode=verify-full&application_name=foo",
			expectedApplicationName: "foo",
			fails:                   false,
		}, {
			name:       "invalid URL",
			dataSource: "invalid string",
			fails:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := buildConfig(logger, tt.dataSource, "", nil)
			if tt.fails {
				if err == nil {
					t.Fatal("error expected")
				}

				return
			}

			fb, ok := cfg.RuntimeParams["application_name"]
			if !ok || fb != tt.expectedApplicationName {
				t.Fatalf("wrong application_name: got %q want %q", fb, tt.expectedApplicationName)
			}
		})
	}
}

func TestBuildConfig_AuthProvider(t *testing.T) {
	t.Run("should be able to apply token during start up", func(t *testing.T) {
		logger := logtest.Scoped(t)
		mocks := NewMockAuthProvider()

		// prevent provider from ever being refreshed
		mocks.IsRefreshFunc.SetDefaultReturn(false)
		mocks.ApplyFunc.SetDefaultHook(func(logger log.Logger, cfg *pgx.ConnConfig) error {
			cfg.Password = "rds-auth-token"
			return nil
		})

		cfg, err := buildConfig(logger, "postgres://sourcegraph@postgresmydb.123456789012.us-east-1.rds.amazonaws.com/sourcegraph?sslmode=verify-full", "", mocks)
		require.NoError(t, err)
		require.Equal(t, cfg.Password, "rds-auth-token")
		mockassert.CalledOnce(t, mocks.ApplyFunc)
	})

	t.Run("broken auth provider", func(t *testing.T) {
		logger := logtest.Scoped(t)
		mocks := NewMockAuthProvider()

		// prevent provider from ever being refreshed
		mocks.IsRefreshFunc.SetDefaultReturn(false)
		mocks.ApplyFunc.SetDefaultHook(func(logger log.Logger, cfg *pgx.ConnConfig) error {
			return errors.New("broken")
		})

		_, err := buildConfig(logger, "postgres://sourcegraph@postgresmydb.123456789012.us-east-1.rds.amazonaws.com/sourcegraph?sslmode=verify-full", "", mocks)
		require.Error(t, err)
		autogold.Expect("Error applying auth provider: broken").Equal(t, err.Error())
		mockassert.CalledOnce(t, mocks.ApplyFunc)
	})

	// this test mimics the behaviour of buildConfig internal
	// to simulate token refresh
	// unfortunately we can't test whether pgx actually uses the new token
	// since it does not expose internal state, but according
	t.Run("auth provider can refresh", func(t *testing.T) {
		logger := logtest.Scoped(t)
		mocks := NewMockAuthProvider()

		// always expire to force refresh
		mocks.IsRefreshFunc.SetDefaultReturn(true)
		// PushHook is provides the canned reply during the first Apply call
		mocks.ApplyFunc.PushHook(func(logger log.Logger, cfg *pgx.ConnConfig) error {
			cfg.Password = "old-token"
			return nil
		})
		// SetDefaultHook provides the canned reply during all subsequent Apply calls
		mocks.ApplyFunc.SetDefaultHook(func(logger log.Logger, cfg *pgx.ConnConfig) error {
			cfg.Password = "new-token"
			return nil
		})

		cfg, err := pgx.ParseConfig("postgres://sourcegraph@postgresmydb.123456789012.us-east-1.rds.amazonaws.com/sourcegraph?sslmode=verify-full")
		require.NoError(t, err)

		err = mocks.Apply(logger, cfg)
		require.NoError(t, err)
		doneC := make(chan struct{})
		go func() {
			for range time.Tick(200 * time.Millisecond) {
				if !mocks.IsRefresh(logger, cfg) {
					continue
				}
				err := mocks.Apply(logger, cfg)
				require.NoError(t, err)

				// signal refresh is done so we can assert it
				doneC <- struct{}{}
			}
		}()

		<-doneC
		require.Equal(t, cfg.Password, "new-token")
		mockassert.Called(t, mocks.IsRefreshFunc)
		mockassert.Called(t, mocks.ApplyFunc)
	})
}
