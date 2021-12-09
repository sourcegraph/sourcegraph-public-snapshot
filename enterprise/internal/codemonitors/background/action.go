package background

// actionArgs is the shared set of arguments needed to execute any
// action for code monitors.
type actionArgs struct {
	MonitorDescription string
	MonitorURL         string
	Query              string
	QueryURL           string
	NumResults         int
}
