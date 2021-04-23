package output

type NoopWriter struct{}

func (NoopWriter) Write(s string)                              {}
func (NoopWriter) Writef(format string, args ...interface{})   {}
func (NoopWriter) WriteLine(line FancyLine)                    {}
func (NoopWriter) Verbose(s string)                            {}
func (NoopWriter) Verbosef(format string, args ...interface{}) {}
func (NoopWriter) VerboseLine(line FancyLine)                  {}
