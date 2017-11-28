package localstore

import (
	"testing"
)

func TestUsers_MatchUsernameRegex(t *testing.T) {
	tests := []struct {
		username string
		isValid  bool
	}{
		{"nick", true},
		{"n1ck", true},
		{"Nick", true},
		{"N-S", true},
		{"nick-s", true},
		{"renfred-xh", true},
		{"renfred-x-h", true},
		{"deadmau5", true},
		{"deadmau-5", true},
		{"3blindmice", true},
		{"777", true},
		{"7-7", true},
		{"long-butnotquitelongenoughtoreachlimit", true},

		{"nick-", false},
		{"nick.com", false},
		{"nick_s", false},
		{"_", false},
		{"_nick", false},
		{"nick_", false},
		{"ke$ha", false},
		{"ni%k", false},
		{"#nick", false},
		{"@nick", false},
		{"", false},
		{"nick s", false},
		{" ", false},
		{"-", false},
		{"--", false},
		{"-s", false},
		{"レンフレッド", false},
		{"veryveryveryveryveryveryveryveryverylong", false},
	}

	for _, test := range tests {
		matched := MatchUsernameString.MatchString(test.username)
		if matched != test.isValid {
			t.Errorf("expected '%v' for username '%s'", test.isValid, test.username)
		}
	}
}

func TestUsers_CheckAndDecrementInviteQuota(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, "authid", "a@a.com", "u", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check default invite quota.
	var inviteQuota int
	row := globalDB.QueryRowContext(ctx, "SELECT invite_quota FROM users WHERE id=$1", user.ID)
	if err := row.Scan(&inviteQuota); err != nil {
		t.Fatal(err)
	}
	// Check that it's within some reasonable bounds. The upper bound number here can increased
	// if we increase the default.
	if lo, hi := 0, 15; inviteQuota <= lo || inviteQuota > hi {
		t.Fatalf("got default user invite quota %d, want in [%d,%d)", inviteQuota, lo, hi)
	}

	// Decrementing should succeed while we have remaining quota. Keep going until we exhaust it.
	// Since the quota is fairly low, this isn't too slow.
	for inviteQuota > 0 {
		if err := Users.CheckAndDecrementInviteQuota(ctx, user.ID); err != nil {
			t.Fatal("initial CheckAndDecrementInviteQuota failed:", err)
		}
		inviteQuota--
	}

	// Now our quota is exhausted, and CheckAndDecrementInviteQuota should fail.
	if err := Users.CheckAndDecrementInviteQuota(ctx, user.ID); err != ErrInviteQuotaExceeded {
		t.Fatalf("over-limit CheckAndDecrementInviteQuota #1: got error %v, want %q", err, ErrInviteQuotaExceeded)
	}

	// Check again that we're still over quota, just in case.
	if err := Users.CheckAndDecrementInviteQuota(ctx, user.ID); err != ErrInviteQuotaExceeded {
		t.Fatalf("over-limit CheckAndDecrementInviteQuota #2: got error %v, want %q", err, ErrInviteQuotaExceeded)
	}
}
