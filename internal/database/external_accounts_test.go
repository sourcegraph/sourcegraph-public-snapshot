package database

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalAccounts_LookupUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	userID, err := ExternalAccounts(db).CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	lookedUpUserID, err := ExternalAccounts(db).LookupUserAndSave(ctx, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}
	if lookedUpUserID != userID {
		t.Errorf("got %d, want %d", lookedUpUserID, userID)
	}
}

func TestExternalAccounts_AssociateUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	user, err := Users(db).Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: &authData,
		Data:     &data,
	}
	if err := ExternalAccounts(db).AssociateUserAndSave(ctx, user.ID, spec, accountData); err != nil {
		t.Fatal(err)
	}

	accounts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0

	want := extsvc.Account{
		UserID:      user.ID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_CreateUserAndSave(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: &authData,
		Data:     &data,
	}
	userID, err := ExternalAccounts(db).CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}

	user, err := Users(db).GetByID(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u"; user.Username != want {
		t.Errorf("got %q, want %q", user.Username, want)
	}

	accounts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0

	want := extsvc.Account{
		UserID:      userID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_CreateUserAndSave_NilData(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	userID, err := ExternalAccounts(db).CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	user, err := Users(db).GetByID(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u"; user.Username != want {
		t.Errorf("got %q, want %q", user.Username, want)
	}

	accounts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
	}
	account := *accounts[0]
	simplifyExternalAccount(&account)
	account.ID = 0

	want := extsvc.Account{
		UserID:      userID,
		AccountSpec: spec,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestExternalAccounts_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	specs := []extsvc.AccountSpec{
		{
			ServiceType: "xa",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "11",
		},
		{
			ServiceType: "xa",
			ServiceID:   "xb",
			ClientID:    "xc",
			AccountID:   "12",
		},
		{
			ServiceType: "ya",
			ServiceID:   "yb",
			ClientID:    "yc",
			AccountID:   "3",
		},
	}
	userIDs := make([]int32, 0, len(specs))

	for i, spec := range specs {
		id, err := ExternalAccounts(db).CreateUserAndSave(ctx, NewUser{Username: fmt.Sprintf("u%d", i)}, spec, extsvc.AccountData{})
		if err != nil {
			t.Fatal(err)
		}
		userIDs = append(userIDs, id)
	}

	specByID := make(map[int32]extsvc.AccountSpec)
	for i, id := range userIDs {
		specByID[id] = specs[i]
	}

	tc := []struct {
		name        string
		args        ExternalAccountsListOptions
		expectedIDs []int32
	}{
		{
			name:        "ListAll",
			expectedIDs: userIDs,
		},
		{
			name:        "ListByAccountID",
			expectedIDs: []int32{userIDs[2]},
			args: ExternalAccountsListOptions{
				AccountID: 3,
			},
		},
		{
			name:        "ListByAccountNotFound",
			expectedIDs: []int32{},
			args: ExternalAccountsListOptions{
				AccountID: 33333,
			},
		},
		{
			name:        "ListByService",
			expectedIDs: []int32{userIDs[0], userIDs[1]},
			args: ExternalAccountsListOptions{
				ServiceType: "xa",
				ServiceID:   "xb",
				ClientID:    "xc",
			},
		},
		{
			name:        "ListByServiceNotFound",
			expectedIDs: []int32{},
			args: ExternalAccountsListOptions{
				ServiceType: "notfound",
				ServiceID:   "notfound",
				ClientID:    "notfound",
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			accounts, err := ExternalAccounts(db).List(ctx, c.args)
			if err != nil {
				t.Fatal(err)
			}
			if got, expected := len(accounts), len(c.expectedIDs); got != expected {
				t.Fatalf("len(accounts) got %d, want %d", got, expected)
			}
			for i, id := range c.expectedIDs {
				account := accounts[i]
				simplifyExternalAccount(account)
				want := extsvc.Account{
					UserID:      id,
					ID:          id,
					AccountSpec: specByID[id],
				}
				if diff := cmp.Diff(want, *account); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExternalAccounts_Encryption(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	store := ExternalAccounts(db).WithEncryptionKey(testKey{})

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}

	authData := json.RawMessage(`"authData"`)
	data := json.RawMessage(`"data"`)
	accountData := extsvc.AccountData{
		AuthData: &authData,
		Data:     &data,
	}

	// store with encrypted authdata
	userID, err := store.CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}

	listFirstAccount := func(s *UserExternalAccountsStore) extsvc.Account {
		t.Helper()

		accounts, err := s.List(ctx, ExternalAccountsListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(accounts) != 1 {
			t.Fatalf("got len(accounts) == %d, want 1", len(accounts))
		}
		account := *accounts[0]
		simplifyExternalAccount(&account)
		account.ID = 0
		return account
	}

	// create a store with a NoopKey to read the raw encrypted value
	noopStore := store.WithEncryptionKey(&encryption.NoopKey{})

	account := listFirstAccount(noopStore)

	// if the testKey worked, the data should just be a base64 encoded version
	if string(*account.AuthData) != base64.StdEncoding.EncodeToString([]byte(*accountData.AuthData)) {
		t.Fatalf("expected base64 encoded auth data, got %s", string(*account.AuthData))
	}
	if string(*account.Data) != base64.StdEncoding.EncodeToString([]byte(*accountData.Data)) {
		t.Fatalf("expected base64 encoded data, got %s", string(*account.Data))
	}

	// List should return decrypted data
	account = listFirstAccount(store)
	want := extsvc.Account{
		UserID:      userID,
		AccountSpec: spec,
		AccountData: accountData,
	}
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// LookupUserAndSave should encrypt the accountData correctly
	userID, err = store.LookupUserAndSave(ctx, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}
	account = listFirstAccount(store)
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}

	// AssociateUserAndSave should encrypt the accountData correctly
	err = store.AssociateUserAndSave(ctx, userID, spec, accountData)
	if err != nil {
		t.Fatal(err)
	}
	account = listFirstAccount(store)
	if diff := cmp.Diff(want, account); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func simplifyExternalAccount(account *extsvc.Account) {
	account.CreatedAt = time.Time{}
	account.UpdatedAt = time.Time{}
}

func TestExternalAccounts_expiredAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	ctx := context.Background()

	spec := extsvc.AccountSpec{
		ServiceType: "xa",
		ServiceID:   "xb",
		ClientID:    "xc",
		AccountID:   "xd",
	}
	userID, err := ExternalAccounts(db).CreateUserAndSave(ctx, NewUser{Username: "u"}, spec, extsvc.AccountData{})
	if err != nil {
		t.Fatal(err)
	}

	accts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{UserID: userID})
	if err != nil {
		t.Fatal(err)
	} else if len(accts) != 1 {
		t.Fatalf("Want 1 external accounts but got %d", len(accts))
	}
	acct := accts[0]

	t.Run("Exclude expired", func(t *testing.T) {
		err := ExternalAccounts(db).TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		accts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{
			UserID:         userID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) > 0 {
			t.Fatalf("Want no external accounts but got %d", len(accts))
		}
	})

	t.Run("LookupUserAndSave should set expired_at to NULL", func(t *testing.T) {
		err := ExternalAccounts(db).TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ExternalAccounts(db).LookupUserAndSave(ctx, spec, extsvc.AccountData{})
		if err != nil {
			t.Fatal(err)
		}

		accts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{
			UserID:         userID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) != 1 {
			t.Fatalf("Want 1 external accounts but got %d", len(accts))
		}
	})

	t.Run("AssociateUserAndSave should set expired_at to NULL", func(t *testing.T) {
		err := ExternalAccounts(db).TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		err = ExternalAccounts(db).AssociateUserAndSave(ctx, userID, spec, extsvc.AccountData{})
		if err != nil {
			t.Fatal(err)
		}

		accts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{
			UserID:         userID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(accts) != 1 {
			t.Fatalf("Want 1 external accounts but got %d", len(accts))
		}
	})

	t.Run("TouchLastValid should set expired_at to NULL", func(t *testing.T) {
		err := ExternalAccounts(db).TouchExpired(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		err = ExternalAccounts(db).TouchLastValid(ctx, acct.ID)
		if err != nil {
			t.Fatal(err)
		}

		accts, err := ExternalAccounts(db).List(ctx, ExternalAccountsListOptions{
			UserID:         userID,
			ExcludeExpired: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(accts) != 1 {
			t.Fatalf("Want 1 external accounts but got %d", len(accts))
		}
	})
}

func TestExternalAccountsMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	// ensure no keyring is configured
	keyring.SetDefault(keyring.Ring{})

	setupKey := func() func() {
		// configure key
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				EncryptionKeys: &schema.EncryptionKeys{
					UserExternalAccountKey: &schema.EncryptionKey{
						Base64: &schema.Base64EncryptionKey{},
					},
				},
			},
		})
		if err := keyring.Init(ctx); err != nil {
			t.Fatal(err)
		}

		return func() {
			conf.Mock(&conf.Unified{})
			keyring.SetDefault(keyring.Ring{})
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

		key := &encryption.Base64Key{}

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

			if data == string(*accounts[i].Data) {
				t.Fatalf("stored data is the same as before migration")
			}
			secret, err = key.Decrypt(ctx, []byte(data))
			if err != nil {
				t.Fatal(err)
			}
			if secret.Secret() != string(*accounts[i].Data) {
				t.Fatalf("decrypted data is different from the original one")
			}

			if id, _ := key.ID(ctx); keyID != id {
				t.Fatalf("wrong encryption_key_id, want %s, got %s", id, keyID)
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

			if data != string(*accounts[i].Data) {
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
		keyring.SetDefault(keyring.Ring{UserExternalAccountKey: &invalidKey{}})
		defer keyring.SetDefault(keyring.Ring{})

		// migrate the accounts, should fail
		err := migrator.Up(ctx)
		if err == nil {
			t.Fatal("migrating the service with an invalid key should fail")
		}
		if err.Error() != "invalid encryption round-trip" {
			t.Fatal(err)
		}
	})
}
