package period

// Set enables use of Period by the flag API.
func (period *Period) Set(p string) error {
	p2, err := Parse(p)
	if err != nil {
		return err
	}
	*period = p2
	return nil
}

// Type is for compatibility with the spf13/pflag library.
func (period Period) Type() string {
	return "period"
}
