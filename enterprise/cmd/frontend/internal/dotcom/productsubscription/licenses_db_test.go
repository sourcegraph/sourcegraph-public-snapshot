package productsubscription

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestProductLicenses_Create(t *testing.T) {
	ctx := dbtesting.TestContext(t)

	u, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	ps0, err := dbSubscriptions{}.Create(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}

	pl0, err := dbLicenses{}.Create(ctx, ps0, "k")
	if err != nil {
		t.Fatal(err)
	}

	got, err := dbLicenses{}.GetByID(ctx, pl0)
	if err != nil {
		t.Fatal(err)
	}
	if want := pl0; got.ID != want {
		t.Errorf("got %v, want %v", got.ID, want)
	}
	if want := ps0; got.ProductSubscriptionID != want {
		t.Errorf("got %v, want %v", got.ProductSubscriptionID, want)
	}
	if want := "k"; got.LicenseKey != want {
		t.Errorf("got %q, want %q", got.LicenseKey, want)
	}

	ts, err := dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps0})
	if err != nil {
		t.Fatal(err)
	}
	if want := 1; len(ts) != want {
		t.Errorf("got %d product licenses, want %d", len(ts), want)
	}

	ts, err = dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: "69da12d5-323c-4e42-9d44-cc7951639bca" /* invalid */})
	if err != nil {
		t.Fatal(err)
	}
	if want := 0; len(ts) != want {
		t.Errorf("got %d product licenses, want %d", len(ts), want)
	}
}

func TestProductLicenses_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	u1, err := db.Users.Create(ctx, db.NewUser{Username: "u1"})
	if err != nil {
		t.Fatal(err)
	}

	ps0, err := dbSubscriptions{}.Create(ctx, u1.ID)
	if err != nil {
		t.Fatal(err)
	}
	ps1, err := dbSubscriptions{}.Create(ctx, u1.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dbLicenses{}.Create(ctx, ps0, "k")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbLicenses{}.Create(ctx, ps0, "n1")
	if err != nil {
		t.Fatal(err)
	}

	{
		// List all product licenses.
		ts, err := dbLicenses{}.List(ctx, dbLicensesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d product licenses, want %d", len(ts), want)
		}
		count, err := dbLicenses{}.Count(ctx, dbLicensesListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	}

	{
		// List ps0's product licenses.
		ts, err := dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps0})
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(ts) != want {
			t.Errorf("got %d product licenses, want %d", len(ts), want)
		}
	}

	{
		// List ps1's product licenses.
		ts, err := dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: ps1})
		if err != nil {
			t.Fatal(err)
		}
		if want := 0; len(ts) != want {
			t.Errorf("got %d product licenses, want %d", len(ts), want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_655(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
