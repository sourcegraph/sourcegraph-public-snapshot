package drift

type Summary interface {
	Name() string
	Problem() string
	Solution() string
	Diff() (a, b any, _ bool)
	Statements() ([]string, bool)
	URLHint() (string, bool)
}

type driftSummary struct {
	name          string
	problem       string
	solution      string
	hasDiff       bool
	a, b          any
	hasStatements bool
	statements    []string
	hasURLHint    bool
	url           string
}

func singleton(summary Summary) []Summary {
	return []Summary{summary}
}

func newDriftSummary(name string, problem, solution string) *driftSummary {
	return &driftSummary{
		name:     name,
		problem:  problem,
		solution: solution,
	}
}

func (s *driftSummary) withDiff(a, b any) *driftSummary {
	s.hasDiff = true
	s.a, s.b = a, b
	return s
}

func (s *driftSummary) withStatements(statements ...string) *driftSummary {
	s.hasStatements = true
	s.statements = statements
	return s
}

func (s *driftSummary) withURLHint(url string) *driftSummary {
	s.hasURLHint = true
	s.url = url
	return s
}

func (s *driftSummary) Name() string                 { return s.name }
func (s *driftSummary) Problem() string              { return s.problem }
func (s *driftSummary) Solution() string             { return s.solution }
func (s *driftSummary) Diff() (a, b any, _ bool)     { return s.a, s.b, s.hasDiff }
func (s *driftSummary) Statements() ([]string, bool) { return s.statements, s.hasStatements }
func (s *driftSummary) URLHint() (string, bool)      { return s.url, s.hasURLHint }
