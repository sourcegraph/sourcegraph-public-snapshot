package dbconn

import (
	"os"
	"testing"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	rdsauthmocks "github.com/sourcegraph/sourcegraph/internal/database/dbconn/rds"
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

func TestBuildConfig_RDSIAMAuth(t *testing.T) {
	t.Run("yes rds iam auth", func(t *testing.T) {
		_ = os.Setenv(pgAWSUseEC2RoleCredentialsEnvKey, "true")
		defer func() {
			_ = os.Unsetenv(pgAWSUseEC2RoleCredentialsEnvKey)
		}()

		logger := logtest.Scoped(t)
		mocks := rdsauthmocks.NewMockAuthProvider()
		mocks.AuthTokenFunc.SetDefaultReturn("rds-auth-token", nil)

		cfg, err := buildConfig(logger, "postgres://sourcegraph@postgresmydb.123456789012.us-east-1.rds.amazonaws.com/sourcegraph?sslmode=verify-full", "", mocks)
		require.NoError(t, err)
		require.Equal(t, cfg.Password, "rds-auth-token")
	})

	t.Run("no rds iam auth", func(t *testing.T) {
		_ = os.Unsetenv(pgAWSUseEC2RoleCredentialsEnvKey)

		logger := logtest.Scoped(t)
		mocks := rdsauthmocks.NewMockAuthProvider()
		mocks.AuthTokenFunc.SetDefaultReturn("rds-auth-token", nil)

		_, err := buildConfig(logger, "postgres://sourcegraph@postgresmydb.123456789012.us-east-1.rds.amazonaws.com/sourcegraph?sslmode=verify-full", "", mocks)
		require.NoError(t, err)
		mockassert.NotCalled(t, mocks.AuthTokenFunc)
	})

	t.Run("broken rds iam auth", func(t *testing.T) {
		_ = os.Setenv(pgAWSUseEC2RoleCredentialsEnvKey, "true")
		defer func() {
			_ = os.Unsetenv(pgAWSUseEC2RoleCredentialsEnvKey)
		}()

		logger := logtest.Scoped(t)
		mocks := rdsauthmocks.NewMockAuthProvider()
		mocks.AuthTokenFunc.SetDefaultHook(func(string, uint16, string) (string, error) {
			return "", errors.New("broken")
		})

		_, err := buildConfig(logger, "postgres://sourcegraph@postgresmydb.123456789012.us-east-1.rds.amazonaws.com/sourcegraph?sslmode=verify-full", "", mocks)
		require.Error(t, err)
		autogold.Expect("Error retrieving auth token for RDS IAM auth: broken").Equal(t, err.Error())
	})
}
