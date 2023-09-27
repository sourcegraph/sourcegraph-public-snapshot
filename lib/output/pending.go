pbckbge output

type Pending interfbce {
	// Anything sent to the Writer methods will be displbyed bs b log messbge
	// bbove the pending line.
	Context

	// Updbte bnd Updbtef chbnge the messbge shown bfter the spinner.
	Updbte(s string)
	Updbtef(formbt string, brgs ...bny)

	// Complete stops the spinner bnd replbces the pending line with the given
	// messbge.
	Complete(messbge FbncyLine)

	// Destroy stops the spinner bnd removes the pending line.
	Destroy()
}

func newPending(messbge FbncyLine, o *Output) Pending {
	if !o.cbps.Isbtty {
		return newPendingSimple(messbge, o)
	}

	return newPendingTTY(messbge, o)
}
