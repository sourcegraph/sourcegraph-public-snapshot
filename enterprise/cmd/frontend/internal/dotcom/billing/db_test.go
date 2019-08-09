package billing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func init() {
	dbtesting.DBNameSuffix = "billing"
}

func TestDBUsersBillingCustomerID(t *testing.T) {
	ctx := dbtesting.TestContext(t)

	t.Run("existing user", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
		if err != nil {
			t.Fatal(err)
		}

		if custID, err := (dbBilling{}).getUserBillingCustomerID(ctx, nil, u.ID); err != nil {
			t.Fatal(err)
		} else if custID != nil {
			t.Errorf("got %q, want nil", *custID)
		}

		t.Run("set to non-nil", func(t *testing.T) {
			if err := (dbBilling{}).setUserBillingCustomerID(ctx, nil, u.ID, strptr("x")); err != nil {
				t.Fatal(err)
			}
			if custID, err := (dbBilling{}).getUserBillingCustomerID(ctx, nil, u.ID); err != nil {
				t.Fatal(err)
			} else if want := "x"; custID == nil || *custID != want {
				t.Errorf("got %v, want %q", custID, want)
			}
		})

		t.Run("set to nil", func(t *testing.T) {
			if err := (dbBilling{}).setUserBillingCustomerID(ctx, nil, u.ID, nil); err != nil {
				t.Fatal(err)
			}
			if custID, err := (dbBilling{}).getUserBillingCustomerID(ctx, nil, u.ID); err != nil {
				t.Fatal(err)
			} else if custID != nil {
				t.Errorf("got %q, want nil", *custID)
			}
		})
	})

	t.Run("nonexistent user", func(t *testing.T) {
		if _, err := (dbBilling{}).getUserBillingCustomerID(ctx, nil, 123 /* doesn't exist */); !errcode.IsNotFound(err) {
			t.Errorf("got %v, want errcode.IsNotFound(err) == true", err)
		}
		if err := (dbBilling{}).setUserBillingCustomerID(ctx, nil, 123 /* doesn't exist */, strptr("x")); !errcode.IsNotFound(err) {
			t.Errorf("got %v, want errcode.IsNotFound(err) == true", err)
		}
	})
}

func strptr(s string) *string {
	return &s
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_642(size int) error {
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
