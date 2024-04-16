package mock

type ProgressCall struct {
	Done  int
	Total int
}

type Progress struct {
	Calls []ProgressCall
}

func (p *Progress) Callback() func(int, int) {
	return func(done, total int) {
		p.Calls = append(p.Calls, ProgressCall{done, total})
	}
}
