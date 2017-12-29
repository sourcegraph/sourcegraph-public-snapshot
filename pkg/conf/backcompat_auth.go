package conf

// AuthProvider returns the auth.provider value, or applies the "builtin" default otherwise.
func AuthProvider() string {
	if cfg.AuthProvider != "" {
		return cfg.AuthProvider
	}
	return "builtin"
}
