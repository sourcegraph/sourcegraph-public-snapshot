package billing

import (
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestGetOrAssignUserCustomerID(t *testing.T) {
	ctx := dbtesting.TestContext(t)

	c := 0
	mockCreateCustomerID = func(userID int32) (string, error) {
		c++
		return fmt.Sprintf("cust%d", c), nil
	}
	defer func() { mockCreateCustomerID = nil }()

	u, err := db.Users.Create(ctx, db.NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("assigns and retrieves", func(t *testing.T) {
		custID1, err := GetOrAssignUserCustomerID(ctx, u.ID)
		if err != nil {
			t.Fatal(err)
		}
		custID2, err := GetOrAssignUserCustomerID(ctx, u.ID)
		if err != nil {
			t.Fatal(err)
		}
		if custID2 != custID1 {
			t.Errorf("got custID %q, want %q", custID2, custID2)
		}
	})

	t.Run("fails on nonexistent users", func(t *testing.T) {
		if _, err := GetOrAssignUserCustomerID(ctx, 123 /* no such user */); err == nil {
			t.Fatal("err == nil")
		}
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_640(size int) error {
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
