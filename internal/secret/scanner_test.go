package secret

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestScannerInsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dbtesting.DBNameSuffix = "secret"
	dbtesting.SetupGlobalTestDB(t)
	defaultEncryptor = newAESGCMEncodedEncryptor(mustGenerateRandomAESKey(), nil)

	message := "Able was I ere I saw Elba"
	encryptedMessage := StringValue(message)
	name := "base"

	_, err := dbconn.Global.Exec(`CREATE TABLE IF NOT EXISTS secret_test (name text, message text)`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dbconn.Global.Exec(`INSERT INTO secret_test(name,message) VALUES ($1,$2)`, name, encryptedMessage)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := dbconn.Global.Query(`SELECT name,message FROM secret_test`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var gotName string
	var gotEncryptedMessage StringValue
	for rows.Next() {
		if err := rows.Scan(&gotName, &gotEncryptedMessage); err != nil {
			t.Fatal(err)
		}
	}

	if gotName != name {
		t.Fatalf("expected %q, got %q for name", name, gotName)
	}
	if gotEncryptedMessage != encryptedMessage {
		t.Fatalf("expected %q, got %q", encryptedMessage, gotEncryptedMessage)
	}
	return
}
