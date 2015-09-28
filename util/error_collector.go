package util

type ErrorCollector []error

func (ec ErrorCollector) Error() string {
	if len(ec) == 0 {
		return ""
	} else if len(ec) == 1 {
		return ec[0].Error()
	}
	msg := "Collected the following errors:\n"
	tab := "  "
	for _, e := range ec {
		msg += tab + e.Error() + "\n"
	}
	return msg
}
