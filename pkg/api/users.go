package sourcegraph

func (u *User) Spec() UserSpec {
	return UserSpec{UID: u.UID}
}
