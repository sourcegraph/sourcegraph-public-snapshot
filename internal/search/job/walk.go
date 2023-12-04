package job

type MapFunc func(Job) Job

// Map applies fn to every job in tree recursively, returning a new job.
// The provided function should return a copied job with the mutations
// applied rather than mutating the job in-place.
func Map(j Job, fn MapFunc) Job {
	j = j.MapChildren(fn)
	return fn(j)
}

// MapType works the same way as Map, except the provided function
// is only called for jobs of the type the function accepts. This
// is useful for when you want to modify only jobs of a certain type.
func MapType[T Job](j Job, fn func(T) Job) Job {
	mapFn := func(current Job) Job {
		if t, ok := current.(T); ok {
			return fn(t)
		}
		return current
	}
	return Map(j, mapFn)
}

// Visit iterates through each job and partial job in preorder.
func Visit(j Describer, fn func(Describer)) {
	fn(j)
	for _, child := range j.Children() {
		Visit(child, fn)
	}
}

// VisitType works the same way as Visit, except the callback is only
// called on describers with the type that the caller accepts. This is
// useful when you want to visit all jobs of a certain type.
func VisitType[T Describer](j Describer, fn func(T)) {
	Visit(j, func(current Describer) {
		if t, ok := current.(T); ok {
			fn(t)
		}
	})
}

// HasDescendent returns whether the job has any descendents with type T.
func HasDescendent[T Describer](j Job) (res bool) {
	VisitType(j, func(T) {
		res = true
	})
	return res
}
