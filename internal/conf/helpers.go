package conf

func HasExternalAuthProvider(c Unified) bool {
	for _, p := range c.AuthProviders {
		if p.Builtin == nil { // not builtin implies SSO
			return true
		}
	}
	return false
}
