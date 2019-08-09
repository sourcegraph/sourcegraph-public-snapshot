package db

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestUsers_SetTag(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	// Create user.
	u, err := Users.Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	checkTags := func(t *testing.T, userID int32, wantTags []string) {
		t.Helper()
		u, err := Users.GetByID(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(u.Tags)
		sort.Strings(wantTags)
		if !reflect.DeepEqual(u.Tags, wantTags) {
			t.Errorf("got tags %v, want %v", u.Tags, wantTags)
		}
	}
	checkUsersWithTag := func(t *testing.T, tag string, wantUsers []int32) {
		t.Helper()
		users, err := Users.List(ctx, &UsersListOptions{Tag: tag})
		if err != nil {
			t.Fatal(err)
		}
		userIDs := make([]int32, len(users))
		for i, u := range users {
			userIDs[i] = u.ID
		}
		if !reflect.DeepEqual(userIDs, wantUsers) {
			t.Errorf("got user IDs %v, want %v", userIDs, wantUsers)
		}
	}

	t.Run("fails on nonexistent user", func(t *testing.T) {
		if err := Users.SetTag(ctx, 1234 /* doesn't exist */, "t", true); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
		if err := Users.SetTag(ctx, 1234 /* doesn't exist */, "t", false); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

	t.Run("tags begins empty", func(t *testing.T) {
		checkTags(t, u.ID, []string{})
		checkUsersWithTag(t, "t1", []int32{})
	})

	t.Run("adds and removes tag", func(t *testing.T) {
		if err := Users.SetTag(ctx, u.ID, "t1", true); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{"t1"})
		checkUsersWithTag(t, "t1", []int32{u.ID})

		t.Run("deduplicates", func(t *testing.T) {
			if err := Users.SetTag(ctx, u.ID, "t1", true); err != nil {
				t.Fatal(err)
			}
			checkTags(t, u.ID, []string{"t1"})
		})

		if err := Users.SetTag(ctx, u.ID, "t2", true); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{"t1", "t2"})
		checkUsersWithTag(t, "t1", []int32{u.ID})
		checkUsersWithTag(t, "t2", []int32{u.ID})

		if err := Users.SetTag(ctx, u.ID, "t1", false); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{"t2"})
		checkUsersWithTag(t, "t1", []int32{})
		checkUsersWithTag(t, "t2", []int32{u.ID})

		t.Run("removing nonexistent tag is noop", func(t *testing.T) {
			if err := Users.SetTag(ctx, u.ID, "t1", false); err != nil {
				t.Fatal(err)
			}
			checkTags(t, u.ID, []string{"t2"})
		})

		if err := Users.SetTag(ctx, u.ID, "t2", false); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{})
		checkUsersWithTag(t, "t2", []int32{})
	})
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_96(size int) error {
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
