package database

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExternalServiceConfigMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	// ensure no keyring is configured
	keyring.MockDefault(keyring.Ring{})

	setupKey := func() func() {
		keyring.MockDefault(keyring.Ring{
			ExternalServiceKey: testKey{},
		})

		return func() {
			keyring.MockDefault(keyring.Ring{})
		}
	}

	t.Run("Up/Down/Progress", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 2
		migrator.AllowDecrypt = true

		requireProgressEqual := func(want float64) {
			t.Helper()

			got, err := migrator.Progress(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%.3f", want) != fmt.Sprintf("%.3f", got) {
				t.Fatalf("invalid progress: want %f, got %f", want, got)
			}
		}

		// progress on empty table should be 1
		requireProgressEqual(1)

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// progress on non-migrated table should be 0
		requireProgressEqual(0)

		// Up with no configured key shouldn't do anything
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}
		requireProgressEqual(0)

		// configure key ring
		defer setupKey()()

		// Up should migrate two configs
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}
		// services: 10, migrated: 2, progress: 20%
		requireProgressEqual(0.2)

		// Let's migrate the other services
		for i := 2; i <= 5; i++ {
			if err := migrator.Up(ctx); err != nil {
				t.Fatal(err)
			}
			requireProgressEqual(float64(i) * 0.2)
		}
		requireProgressEqual(1)

		// Down should revert the migration for 2 services
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}
		// services: 10, migrated: 8, progress: 80%
		requireProgressEqual(0.8)

		// Let's revert the other services
		for i := 3; i >= 0; i-- {
			if err := migrator.Down(ctx); err != nil {
				t.Fatal(err)
			}
			requireProgressEqual(float64(i) * 0.2)
		}
		requireProgressEqual(0)
	})

	t.Run("Up/Encryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// setup key after storing the services
		defer setupKey()()

		// migrate the services
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually encrypted?
		rows, err := db.Query("SELECT config, encryption_key_id FROM external_services ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		key := testKey{}

		var i int
		for rows.Next() {
			var config, keyID string

			err = rows.Scan(&config, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if config == svcs[i].Config {
				t.Fatalf("stored config is the same as before migration")
			}

			secret, err := key.Decrypt(ctx, []byte(config))
			if err != nil {
				t.Fatal(err)
			}

			if secret.Secret() != svcs[i].Config {
				t.Fatalf("decrypted config is different from the original one")
			}

			if version, _ := key.Version(ctx); keyID != version.JSON() {
				t.Fatalf("wrong encryption_key_id, want %s, got %s", version.JSON(), keyID)
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})

	t.Run("Down/Decryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10
		migrator.AllowDecrypt = true

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// setup key after storing the services
		defer setupKey()()

		// migrate the services
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// revert the migration
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually reverted?
		rows, err := db.Query("SELECT config, encryption_key_id FROM external_services ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var i int
		for rows.Next() {
			var config, keyID string

			err = rows.Scan(&config, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if keyID != "" {
				t.Fatalf("encryption_key_id is still stored in the table")
			}

			if config != svcs[i].Config {
				t.Fatalf("stored config is still encrypted")
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})

	t.Run("Up/InvalidKey", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// setup invalid key after storing the services
		keyring.MockDefault(keyring.Ring{ExternalServiceKey: &invalidKey{}})
		defer keyring.MockDefault(keyring.Ring{})

		// migrate the services, should fail
		err := migrator.Up(ctx)
		if err == nil {
			t.Fatal("migration the service with an invalid key should fail")
		}
		if err.Error() != "invalid encryption round-trip" {
			t.Fatal(err)
		}
	})

	t.Run("Down/Disabled Decryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// setup key after storing the services
		defer setupKey()()

		// migrate the services
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// revert the migration
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually reverted?
		rows, err := db.Query("SELECT config, encryption_key_id FROM external_services ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var i int
		for rows.Next() {
			var config, keyID string

			err = rows.Scan(&config, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if keyID == "" {
				t.Fatalf("data was decrypted")
			}

			if config == svcs[i].Config {
				t.Fatalf("stored config was decrypted")
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})
}

// invalidKey is an encryption.Key that just base64 encodes the plaintext,
// but silently fails to decrypt the secret.
type invalidKey struct{}

func (k invalidKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(plaintext)), nil
}

func (k invalidKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	s := encryption.NewSecret(string(ciphertext))
	return &s, nil
}

func (k invalidKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "invalidkey"}, nil
}

func TestExternalAccountsMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	// ensure no keyring is configured
	keyring.MockDefault(keyring.Ring{})

	setupKey := func() func() {
		keyring.MockDefault(keyring.Ring{
			UserExternalAccountKey: testKey{},
		})

		return func() {
			keyring.MockDefault(keyring.Ring{})
		}
	}

	createAccounts := func(db dbutil.DB, n int) []*extsvc.Account {
		accounts := make([]*extsvc.Account, 0, n)

		for i := 0; i < n; i++ {
			spec := extsvc.AccountSpec{
				ServiceType: fmt.Sprintf("x-%d", i),
				ServiceID:   fmt.Sprintf("x-%d", i),
				ClientID:    fmt.Sprintf("x-%d", i),
				AccountID:   fmt.Sprintf("x-%d", i),
			}
			authData := json.RawMessage(fmt.Sprintf("auth-%d", i))
			data := json.RawMessage(fmt.Sprintf("data-%d", i))
			accData := extsvc.AccountData{
				AuthData: &authData,
				Data:     &data,
			}
			_, err := ExternalAccounts(db).CreateUserAndSave(ctx, NewUser{Username: fmt.Sprintf("u-%d", i)}, spec, accData)
			if err != nil {
				t.Fatal(err)
			}

			accounts = append(accounts, &extsvc.Account{
				AccountData: accData,
			})
		}

		return accounts
	}

	t.Run("Up/Down/Progress", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalAccountsMigratorWithDB(db)
		migrator.BatchSize = 2
		migrator.AllowDecrypt = true

		requireProgressEqual := func(want float64) {
			t.Helper()

			got, err := migrator.Progress(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%.3f", want) != fmt.Sprintf("%.3f", got) {
				t.Fatalf("invalid progress: want %f, got %f", want, got)
			}
		}

		// progress on empty table should be 1
		requireProgressEqual(1)

		// Create 10 user accounts
		createAccounts(db, 10)

		// progress on non-migrated table should be 0
		requireProgressEqual(0)

		// Up with no configured key shouldn't do anything
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}
		requireProgressEqual(0)

		// configure key ring
		defer setupKey()()

		// Up should migrate two configs
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}
		// accounts: 10, migrated: 2, progress: 20%
		requireProgressEqual(0.2)

		// Let's migrate the other accounts
		for i := 2; i <= 5; i++ {
			if err := migrator.Up(ctx); err != nil {
				t.Fatal(err)
			}
			requireProgressEqual(float64(i) * 0.2)
		}
		requireProgressEqual(1)

		// Down should revert the migration for 2 accounts
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}
		// accounts: 10, migrated: 8, progress: 80%
		requireProgressEqual(0.8)

		// Let's revert the other accounts
		for i := 3; i >= 0; i-- {
			if err := migrator.Down(ctx); err != nil {
				t.Fatal(err)
			}
			requireProgressEqual(float64(i) * 0.2)
		}
		requireProgressEqual(0)
	})

	t.Run("Up/Encryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalAccountsMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 accounts
		accounts := createAccounts(db, 10)

		// setup key after storing the accounts
		defer setupKey()()

		// migrate the accounts
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// was the data actually encrypted?
		rows, err := db.Query("SELECT auth_data, account_data, encryption_key_id FROM user_external_accounts ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		key := &testKey{}

		var i int
		for rows.Next() {
			var authData, data, keyID string

			err = rows.Scan(&authData, &data, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if authData == string(*accounts[i].AuthData) {
				t.Fatalf("stored data is the same as before migration")
			}
			secret, err := key.Decrypt(ctx, []byte(authData))
			if err != nil {
				t.Fatal(err)
			}
			if secret.Secret() != string(*accounts[i].AuthData) {
				t.Fatalf("decrypted data is different from the original one")
			}

			if version, _ := key.Version(ctx); keyID != version.JSON() {
				t.Fatalf("wrong encryption_key_id, want %s, got %s", version.JSON(), keyID)
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})

	t.Run("Down/Decryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalAccountsMigratorWithDB(db)
		migrator.BatchSize = 10
		migrator.AllowDecrypt = true

		// Create 10 accounts
		accounts := createAccounts(db, 10)

		// setup key after storing the accounts
		defer setupKey()()

		// migrate the accounts
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// revert the migration
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually reverted?
		rows, err := db.Query("SELECT auth_data, account_data, encryption_key_id FROM user_external_accounts ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var i int
		for rows.Next() {
			var authData, data, keyID string

			err = rows.Scan(&authData, &data, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if keyID != "" {
				t.Fatalf("encryption_key_id is still stored in the table")
			}

			if authData != string(*accounts[i].AuthData) {
				t.Fatalf("stored data is still encrypted")
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})

	t.Run("Up/InvalidKey", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalAccountsMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 accounts
		createAccounts(db, 10)

		// setup invalid key after storing the accounts
		keyring.MockDefault(keyring.Ring{UserExternalAccountKey: &invalidKey{}})
		defer keyring.MockDefault(keyring.Ring{})

		// migrate the accounts, should fail
		err := migrator.Up(ctx)
		if err == nil {
			t.Fatal("migrating the service with an invalid key should fail")
		}
		if err.Error() != "invalid encryption round-trip" {
			t.Fatal(err)
		}
	})

	t.Run("Down/Disabled Decryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalAccountsMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 accounts
		accounts := createAccounts(db, 10)

		// setup key after storing the accounts
		defer setupKey()()

		// migrate the accounts
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// revert the migration
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually reverted?
		rows, err := db.Query("SELECT auth_data, account_data, encryption_key_id FROM user_external_accounts ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var i int
		for rows.Next() {
			var authData, data, keyID string

			err = rows.Scan(&authData, &data, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if keyID == "" {
				t.Fatalf("data was decrypted")
			}

			if authData == string(*accounts[i].AuthData) {
				t.Fatalf("stored data was decrypted")
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})
}
