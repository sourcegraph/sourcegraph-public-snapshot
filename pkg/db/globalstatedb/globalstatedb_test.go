package globalstatedb

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestGet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)
	config, err := Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if config.SiteID == "" {
		t.Fatal("expected site_id to be set")
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	pw, err := generateRandomPassword()
	if err != nil {
		t.Fatal(err)
	}
	if len(pw) != 128 {
		t.Fatal("expected len == 128")
	}

	pw2, err := generateRandomPassword()
	if err != nil {
		t.Fatal(err)
	}
	if pw == pw2 {
		t.Fatal("generated passwords must be random")
	}
}

func TestManagementConsoleState(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	// Site must be initialized to get mgmt console state.
	mgmt, err := getManagementConsoleState(ctx)
	if err != ErrCannotUseManagementConsole {
		t.Fatal("expected error")
	}
	if mgmt != nil {
		t.Fatal("expected nil")
	}

	// Initialize the site.
	_, err = EnsureInitialized(ctx, dbconn.Global)
	if err != nil {
		t.Fatal(err)
	}

	// Now we can get mgmt console state.
	mgmt, err = getManagementConsoleState(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// We expect auto generated plaintext + bcrypt passwords.
	if len(mgmt.PasswordPlaintext) != 128 {
		t.Fatal("expected 128 character password")
	}
	if mgmt.PasswordBcrypt == "" {
		t.Fatal("expected bcrypt password hash")
	}
	passwordPlaintext := mgmt.PasswordPlaintext

	// Clear the plaintext password so it is no longer stored insecurely in the DB.
	if err := ClearManagementConsolePlaintextPassword(ctx); err != nil {
		t.Fatal(err)
	}

	// Should now be impossible to fetch plaintext password.
	plaintext, err := GetManagementConsolePlaintextPassword(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "" {
		t.Fatal("expected plaintext password to be cleared")
	}

	// Now we should be able to authenticate.
	err = AuthenticateManagementConsole(ctx, passwordPlaintext)
	if err != nil {
		t.Fatal("failed to authenticate", err)
	}
}
