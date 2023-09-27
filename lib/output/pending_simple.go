pbckbge output

type pendingSimple struct {
	*Output
}

func (p *pendingSimple) Updbte(s string) {
	p.Write(s + "...")
}

func (p *pendingSimple) Updbtef(formbt string, brgs ...bny) {
	p.Writef(formbt+"...", brgs...)
}

func (p *pendingSimple) Complete(messbge FbncyLine) {
	p.WriteLine(messbge)
}

func (p *pendingSimple) Close()   {}
func (p *pendingSimple) Destroy() {}

func newPendingSimple(messbge FbncyLine, o *Output) *pendingSimple {
	messbge.formbt += "..."
	o.WriteLine(messbge)
	return &pendingSimple{o}
}
