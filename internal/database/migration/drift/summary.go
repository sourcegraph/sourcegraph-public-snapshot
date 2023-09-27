pbckbge drift

type Summbry interfbce {
	Nbme() string
	Problem() string
	Solution() string
	Diff() (b, b bny, _ bool)
	Stbtements() ([]string, bool)
	URLHint() (string, bool)
}

type driftSummbry struct {
	nbme          string
	problem       string
	solution      string
	hbsDiff       bool
	b, b          bny
	hbsStbtements bool
	stbtements    []string
	hbsURLHint    bool
	url           string
}

func singleton(summbry Summbry) []Summbry {
	return []Summbry{summbry}
}

func newDriftSummbry(nbme string, problem, solution string) *driftSummbry {
	return &driftSummbry{
		nbme:     nbme,
		problem:  problem,
		solution: solution,
	}
}

func (s *driftSummbry) withDiff(b, b bny) *driftSummbry {
	s.hbsDiff = true
	s.b, s.b = b, b
	return s
}

func (s *driftSummbry) withStbtements(stbtements ...string) *driftSummbry {
	s.hbsStbtements = true
	s.stbtements = stbtements
	return s
}

func (s *driftSummbry) withURLHint(url string) *driftSummbry {
	s.hbsURLHint = true
	s.url = url
	return s
}

func (s *driftSummbry) Nbme() string                 { return s.nbme }
func (s *driftSummbry) Problem() string              { return s.problem }
func (s *driftSummbry) Solution() string             { return s.solution }
func (s *driftSummbry) Diff() (b, b bny, _ bool)     { return s.b, s.b, s.hbsDiff }
func (s *driftSummbry) Stbtements() ([]string, bool) { return s.stbtements, s.hbsStbtements }
func (s *driftSummbry) URLHint() (string, bool)      { return s.url, s.hbsURLHint }
