package migrators

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
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
			ExternalServiceKey: et.TestKey{},
		})

		return func() {
			keyring.MockDefault(keyring.Ring{})
		}
	}

	t.Run("Up/Down/Progress", func(t *testing.T) {
		db := database.NewDB(dbtest.NewDB(t))

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
		svcs := typestest.GenerateExternalServices(10, typestest.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := db.ExternalServices().Create(ctx, confGet, svc); err != nil {
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
		db := database.NewDB(dbtest.NewDB(t))

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := typestest.GenerateExternalServices(10, typestest.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := db.ExternalServices().Create(ctx, confGet, svc); err != nil {
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
		rows, err := db.Handle().DBUtilDB().QueryContext(ctx, "SELECT config, encryption_key_id FROM external_services ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		key := et.TestKey{}

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
		db := database.NewDB(dbtest.NewDB(t))

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10
		migrator.AllowDecrypt = true

		// Create 10 external services
		svcs := typestest.GenerateExternalServices(10, typestest.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := db.ExternalServices().Create(ctx, confGet, svc); err != nil {
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
		rows, err := db.Handle().DBUtilDB().QueryContext(ctx, "SELECT config, encryption_key_id FROM external_services ORDER BY id")
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
		db := database.NewDB(dbtest.NewDB(t))

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := typestest.GenerateExternalServices(10, typestest.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := db.ExternalServices().Create(ctx, confGet, svc); err != nil {
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
		db := database.NewDB(dbtest.NewDB(t))

		migrator := NewExternalServiceConfigMigratorWithDB(db)
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := typestest.GenerateExternalServices(10, typestest.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := db.ExternalServices().Create(ctx, confGet, svc); err != nil {
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
		rows, err := db.Handle().DBUtilDB().QueryContext(ctx, "SELECT config, encryption_key_id FROM external_services ORDER BY id")
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
