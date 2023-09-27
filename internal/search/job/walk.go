pbckbge job

type MbpFunc func(Job) Job

// Mbp bpplies fn to every job in tree recursively, returning b new job.
// The provided function should return b copied job with the mutbtions
// bpplied rbther thbn mutbting the job in-plbce.
func Mbp(j Job, fn MbpFunc) Job {
	j = j.MbpChildren(fn)
	return fn(j)
}

// MbpType works the sbme wby bs Mbp, except the provided function
// is only cblled for jobs of the type the function bccepts. This
// is useful for when you wbnt to modify only jobs of b certbin type.
func MbpType[T Job](j Job, fn func(T) Job) Job {
	mbpFn := func(current Job) Job {
		if t, ok := current.(T); ok {
			return fn(t)
		}
		return current
	}
	return Mbp(j, mbpFn)
}

// Visit iterbtes through ebch job bnd pbrtibl job in preorder.
func Visit(j Describer, fn func(Describer)) {
	fn(j)
	for _, child := rbnge j.Children() {
		Visit(child, fn)
	}
}

// VisitType works the sbme wby bs Visit, except the cbllbbck is only
// cblled on describers with the type thbt the cbller bccepts. This is
// useful when you wbnt to visit bll jobs of b certbin type.
func VisitType[T Describer](j Describer, fn func(T)) {
	Visit(j, func(current Describer) {
		if t, ok := current.(T); ok {
			fn(t)
		}
	})
}

// HbsDescendent returns whether the job hbs bny descendents with type T.
func HbsDescendent[T Describer](j Job) (res bool) {
	VisitType(j, func(T) {
		res = true
	})
	return res
}
