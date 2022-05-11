package output

type NoopWriter struct{}

func (NoopWriter) Write(s string)                      {}
func (NoopWriter) Writef(format string, args ...any)   {}
func (NoopWriter) WriteLine(line FancyLine)            {}
func (NoopWriter) Verbose(s string)                    {}
func (NoopWriter) Verbosef(format string, args ...any) {}
func (NoopWriter) VerboseLine(line FancyLine)          {}
