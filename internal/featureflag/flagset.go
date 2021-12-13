package featureflag

type FlagSet map[string]bool

func (f FlagSet) GetBool(flag string) (bool, bool) {
	if f == nil {
		return false, false
	}
	v, ok := f[flag]
	return v, ok
}

func (f FlagSet) GetBoolOr(flag string, defaultVal bool) bool {
	if f == nil {
		return defaultVal
	}
	if v, ok := f[flag]; ok {
		return v
	}
	return defaultVal
}
