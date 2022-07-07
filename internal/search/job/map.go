package job

type Mapper interface {
}

func Map(j Job, f func(Job) Job) Job {
	j = j.MapChildren(f)
	return f(j)
}
