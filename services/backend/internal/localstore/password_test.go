package localstore

import (
	"sync/atomic"
	"testing"
	"time"
)

var testUID int32

// nextUID returns a unique test user UID for this process. This is needed
// since we do sets and compares on passwords for users, and if tests are
// running in parallel the results returned will be racey.
func nextUID() int32 {
	return atomic.AddInt32(&testUID, 1)
}

// TestPasswords_CheckUIDPassword_valid tests the behavior of
// Passwords.CheckUIDPassword when called with valid credentials.
func TestPasswords_CheckUIDPassword_valid(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}
}

// TestPasswords_CheckUIDPassword_invalid tests the behavior of
// Passwords.CheckUIDPassword when called with invalid credentials.
func TestPasswords_CheckUIDPassword_invalid(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}
	oldDBPass, err := unmarshalPassword(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, uid, "WRONG"); err == nil {
		t.Fatal("err == nil")
	}

	newDBPass, err := unmarshalPassword(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}

	if !newDBPass.LastFail.After(oldDBPass.LastFail) {
		t.Errorf("password's last_fail field not updated - old: %s, new: %s", oldDBPass.LastFail.String(), newDBPass.LastFail.String())
	}

	if oldDBPass.ConsecutiveFails+1 != newDBPass.ConsecutiveFails {
		t.Errorf("password's consecutive_fails field not incremented - old: %d, new: %d", oldDBPass.ConsecutiveFails, newDBPass.ConsecutiveFails)
	}
}

// TestPasswords_CheckUIDPassword_empty tests the behavior of
// Passwords.CheckUIDPassword when called with empty credentials.
func TestPasswords_CheckUIDPassword_empty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	if err := s.CheckUIDPassword(ctx, uid, ""); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_CheckUIDPassword_noneSet tests the behavior of
// Passwords.CheckUIDPassword when there is no password set.
func TestPasswords_CheckUIDPassword_noneSet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_CheckUIDPassword_noneSetForUser tests the behavior of
// Passwords.CheckUIDPassword when there is no password set for the
// given user (but other users have passwords).
func TestPasswords_CheckUIDPassword_noneSetForUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	uid = nextUID()
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_SetPassword_ok tests changing the password.
func TestPasswords_SetPassword_ok(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	// Password is p.
	if err := s.CheckUIDPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, uid, "p2"); err == nil {
		t.Fatal("err == nil")
	}

	// Change to p2.
	oldPass, err := unmarshalPassword(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.SetPassword(ctx, uid, "p2"); err != nil {
		t.Fatal(err)
	}
	newPass, err := unmarshalPassword(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if newPass.ConsecutiveFails != 0 || !newPass.LastFail.After(oldPass.LastFail) {
		t.Fatalf("the password should be completely reset after the password is reset - old: %+v, new: %+v", oldPass, newPass)
	}
	if err := s.CheckUIDPassword(ctx, uid, "p2"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_SetPassword_empty tests changing the password to an
// empty password.
func TestPasswords_SetPassword_empty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, ""); err != nil {
		t.Fatal(err)
	}

	// No password should be accepted.
	if err := s.CheckUIDPassword(ctx, uid, ""); err == nil {
		t.Fatal("err == nil")
	}
}

// TestPasswords_SetPassword_setToEmpty tests changing the password FROM a
// valid password to an empty password.
func TestPasswords_SetPassword_setToEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	s := newPassword()
	uid := nextUID()
	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	// Set to empty
	if err := s.SetPassword(ctx, uid, ""); err != nil {
		t.Fatal(err)
	}

	// No password should work: the old password should no longer work, and "" should also not work.
	if err := s.CheckUIDPassword(ctx, uid, "p"); err == nil {
		t.Fatal("err == nil")
	}
	if err := s.CheckUIDPassword(ctx, uid, ""); err == nil {
		t.Fatal("err == nil")
	}
}

// testClock gives us the ability to inject whatever we want
// for the next time Now() is called.
type testClock struct {
	nextTime time.Time
}

func (c *testClock) Now() time.Time {
	return c.nextTime
}

// TestPasswords_CheckUIDPassword_WaitPeriod tests to see if the waiting periods in between password
// retries is properly enforced.
func TestPasswords_CheckUIDPassword_WaitPeriod(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx, _, done := testContext()
	defer done()

	tc := &testClock{nextTime: time.Now()}
	s := &password{clock: tc}
	uid := nextUID()

	// checkAndRestore is a lambda helper function with a closure containing this test's context
	// and uid. It checks the given password 'pw' against uid's stored password
	// with CheckUIDPassword(), and restores uid's consecutive_fails and last_fail values
	// before returning the result of that 'pw' check.
	var checkAndRestore = func(pw string, consecFails int, time time.Time) error {
		checkErr := s.CheckUIDPassword(ctx, uid, pw)
		pass, err := unmarshalPassword(ctx, uid)
		if err != nil {
			t.Fatalf(err.Error())
		}
		pass.ConsecutiveFails = int32(consecFails)
		pass.LastFail = time
		_, err = appDBH(ctx).Update(&pass)
		if err != nil {
			t.Fatalf("error: %v  when restoring consecutive_fails and fail_time for uid: %d: ", err, uid)
		}
		return checkErr
	}

	if err := s.SetPassword(ctx, uid, "p"); err != nil {
		t.Fatal(err)
	}

	for i := 1; calcWaitPeriod(i) < maxWaitPeriod; i++ {
		before := tc.nextTime.Add(-calcWaitPeriod(i - 1))
		now := tc.nextTime
		next := tc.nextTime.Add(calcWaitPeriod(i + 1))

		// Before - Current system time is set to a value well before the user should be allowed a login attempt.
		tc.nextTime = before
		if err := checkAndRestore("WRONG", i, now); err == nil {
			t.Fatal("should always reject an incorrect password, no matter what time it is (before)")
		}
		if err := checkAndRestore("p", i, now); err == nil {
			t.Fatal("should reject even the correct password when not enough time has passed in between attempts (before)")
		}
		// On the nose - Current system time is set to a value that is right on the threshold, immediately after which the user should be
		// allowed a login attempt.
		tc.nextTime = now
		if err := checkAndRestore("WRONG", i, now); err == nil {
			t.Fatal("should always reject an incorrect password, no matter what time it is (on the nose)")
		}
		if err := checkAndRestore("p", i, now); err == nil {
			t.Fatal("should reject even the correct password when not enough time has passed in between attempts (on the nose)")
		}
		// After - Current system time is set to a value well after the user should be allowed a login attempt.
		tc.nextTime = next
		if err := checkAndRestore("WRONG", i, now); err == nil {
			t.Fatal("should always reject an incorrect password, no matter what time it is (after)")
		}
		if err := checkAndRestore("p", i+1, next); err != nil {
			t.Fatalf("got err: %s, should accept the correct password when enough time has passed (after)", err)
		}
	}
}
