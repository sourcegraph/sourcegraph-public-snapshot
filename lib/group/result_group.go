package group

type ResultGroup[T any] interface {
	Go(func() T)
	Wait() []T

	Contextable[ResultContextGroup[T]]
	Errorable[ResultErrorGroup[T]]
	Limitable[ResultGroup[T]]
}

type ResultErrorGroup[T any] interface {
	Go(func() (T, error))
	Wait() ([]T, error)

	Contextable[ResultContextErrorGroup[T]]
	Limitable[ResultErrorGroup[T]]
}

type ResultContextGroup[T any] interface {
	Go(func(context.Context) T)
	Wait() []T

	Errorable[ResultContextErrorGroup[T]]
	Limitable[ResultContextGroup[T]]
}

type ResultContextErrorGroup[T any] interface {
	Go(func(context.Context) (T, error))
	Wait() ([]T, error)

	Limitable[ResultContextErrorGroup[T]]
}
