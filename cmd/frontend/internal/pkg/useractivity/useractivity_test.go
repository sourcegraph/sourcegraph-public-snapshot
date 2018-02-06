package useractivity

import (
	"reflect"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

func TestUserActivity_None(t *testing.T) {
	setupForTest(t)

	want := &types.UserActivity{
		UserID: 42,
	}
	got, err := GetByUserID(42)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("got %+v != %+v", got, want)
	}
}

func TestUserActivity_LogPageView(t *testing.T) {
	setupForTest(t)

	user := types.User{
		ID: 1,
	}
	err := LogPageView(true, user.ID, "test-cookie-id")
	if err != nil {
		t.Fatal(err)
	}

	a, err := GetByUserID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if wantViews := int32(1); a.PageViews != wantViews {
		t.Errorf("got %d, want %d", a.PageViews, wantViews)
	}
	diff := (*a.LastPageViewTime).Unix() - time.Now().Unix()
	if wantMaxDiff := 10; diff > int64(wantMaxDiff) || diff < -int64(wantMaxDiff) {
		t.Errorf("got %d seconds apart, wanted less than %d seconds apart", diff, wantMaxDiff)
	}
}

func TestUserActivity_LogSearchQuery(t *testing.T) {
	setupForTest(t)

	user := types.User{
		ID: 1,
	}
	err := LogSearchQuery(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	a, err := GetByUserID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := int32(1); a.SearchQueries != want {
		t.Errorf("got %d, want %d", a.SearchQueries, want)
	}
}

func TestUserActivity_GetUsersActiveToday(t *testing.T) {
	setupForTest(t)

	user1 := types.User{
		ID: 1,
	}
	user2 := types.User{
		ID: 2,
	}

	// Test single user
	err := LogPageView(true, user1.ID, "test-cookie-id-1")
	if err != nil {
		t.Fatal(err)
	}

	n, err := GetUsersActiveTodayCount()
	if err != nil {
		t.Fatal(err)
	}
	if want := 1; n != want {
		t.Errorf("got %d, want %d", n, want)
	}

	// Test multiple users, with repeats
	err = LogPageView(true, user2.ID, "test-cookie-id-2")
	if err != nil {
		t.Fatal(err)
	}
	err = LogPageView(true, user1.ID, "test-cookie-id-1")
	if err != nil {
		t.Fatal(err)
	}
	err = LogPageView(false, 0, "test-cookie-id-3")
	if err != nil {
		t.Fatal(err)
	}
	err = LogPageView(true, user2.ID, "test-cookie-id-2")
	if err != nil {
		t.Fatal(err)
	}

	n, err = GetUsersActiveTodayCount()
	if err != nil {
		t.Fatal(err)
	}
	if want := 3; n != want {
		t.Errorf("got %d, want %d", n, want)
	}
}

func setupForTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	keyPrefix = "__test__" + t.Name() + ":"
	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("EVAL", `local keys = redis.call('keys', ARGV[1])
if #keys > 0 then
	return redis.call('del', unpack(keys))
else
	return ''
end`, 0, keyPrefix+"*")
	if err != nil {
		t.Log("Could not clear test prefix:", err)
	}
}
