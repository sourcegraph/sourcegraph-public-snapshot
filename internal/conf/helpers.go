package conf

func HasExternalAuthProvider(c Unified) bool {
	for _, p := range c.Critical.AuthProviders {
		if p.Builtin == nil { // not builtin implies SSO
			return true
		}
	}
	return false
}
