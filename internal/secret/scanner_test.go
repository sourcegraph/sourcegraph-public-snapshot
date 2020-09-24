package secret

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	dbtesting.DBNameSuffix = "secret"
	os.Exit(m.Run())
}

func TestScanner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dbtesting.SetupGlobalTestDB(t)
	defaultEncryptor = newAESGCMEncodedEncryptor(mustGenerateRandomAESKey(), nil)

	t.Run("base", func(t *testing.T) {
		message := "Able was I ere I saw Elba"
		encryptedMessage := StringValue(message)

		_, err := dbconn.Global.Exec(`CREATE TABLE IF NOT EXISTS secret_scanner_test(name text, message text)`)
		if err != nil {
			t.Fatal(err)
		}

		_, err = dbconn.Global.Exec(`INSERT INTO secret_scanner_test(name,message) VALUES ($1,$2)`, t.Name(), encryptedMessage)
		if err != nil {
			t.Fatal(err)
		}

		rows, err := dbconn.Global.Query(`SELECT name,message FROM secret_scanner_test`)
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

		if gotName != t.Name() {
			t.Fatalf("expected %q, got %q for name", t.Name(), gotName)
		}
		if gotEncryptedMessage != encryptedMessage {
			t.Fatalf("expected %q, got %q", encryptedMessage, gotEncryptedMessage)
		}
	})

	t.Run("null example", func(t *testing.T) {

		_, err := dbconn.Global.Exec(`CREATE TABLE IF NOT EXISTS secret_null_test(name text, message text)`)
		if err != nil {
			t.Fatal(err)
		}

		nullMessage := NullStringValue{}
		_, err = dbconn.Global.Exec(`INSERT INTO secret_null_test(name, message) VALUES ($1,$2)`, t.Name(), nullMessage)
		if err != nil {
			t.Fatal(err)
		}

		rows, err := dbconn.Global.Query(`SELECT name,message FROM secret_null_test`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var gotName string
		var gotEncryptedMessage NullStringValue
		for rows.Next() {
			if err := rows.Scan(&gotName, &gotEncryptedMessage); err != nil {
				t.Fatal(err)
			}
		}

		if gotName != t.Name() {
			t.Fatalf("expected %q, got %q for name", t.Name(), gotName)
		}
		if gotEncryptedMessage.S != nil {
			t.Fatal("expected nil, got non-nil result")
		}
	})
}
