package output

type pendingSimple struct {
	*Output
}

func (p *pendingSimple) Update(s string) {
	p.Write(s + "...")
}

func (p *pendingSimple) Updatef(format string, args ...any) {
	p.Writef(format+"...", args...)
}

func (p *pendingSimple) Complete(message FancyLine) {
	p.WriteLine(message)
}

func (p *pendingSimple) Close()   {}
func (p *pendingSimple) Destroy() {}

func newPendingSimple(message FancyLine, o *Output) *pendingSimple {
	message.format += "..."
	o.WriteLine(message)
	return &pendingSimple{o}
}
