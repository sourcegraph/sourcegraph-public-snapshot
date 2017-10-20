package localstore

import "testing"

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
