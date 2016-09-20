package sourcegraph

func (u *User) Spec() UserSpec {
	return UserSpec{UID: u.UID}
}

// InBeta tells whether or not the given user is in the given beta.
func (u *User) InBeta(beta string) bool {
	for _, b := range u.Betas {
		if b == beta {
			return true
		}
	}
	return false
}

// InAnyBeta tells whether or not the given user is in any beta program.
func (u *User) InAnyBeta() bool {
	return len(u.Betas) > 0
}

// BetaPending tells if the given user is registered for beta access but is not
// yet participating in any beta programs.
func (u *User) BetaPending() bool {
	return u.BetaRegistered && len(u.Betas) == 0
}
