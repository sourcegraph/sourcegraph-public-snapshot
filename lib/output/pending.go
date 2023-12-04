package output

type Pending interface {
	// Anything sent to the Writer methods will be displayed as a log message
	// above the pending line.
	Context

	// Update and Updatef change the message shown after the spinner.
	Update(s string)
	Updatef(format string, args ...any)

	// Complete stops the spinner and replaces the pending line with the given
	// message.
	Complete(message FancyLine)

	// Destroy stops the spinner and removes the pending line.
	Destroy()
}

func newPending(message FancyLine, o *Output) Pending {
	if !o.caps.Isatty {
		return newPendingSimple(message, o)
	}

	return newPendingTTY(message, o)
}
