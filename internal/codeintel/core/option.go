package core

type Option[A any] struct {
	value *A
}

func Some[A any](value A) Option[A] {
	return Option[A]{
		value: &value,
	}
}

func None[A any]() Option[A] {
	return Option[A]{
		value: nil,
	}
}

func (o Option[A]) Compare(other Option[A], cmp func(A, A) int) int {
	if o.IsNone() {
		if other.IsNone() {
			return 0
		}
		return -1
	}
	if other.IsNone() {
		return 1
	}
	v1 := o.Unwrap()
	v2 := other.Unwrap()
	return cmp(v1, v2)
}

func (o Option[A]) Get() (A, bool) {
	if o.value == nil {
		return *new(A), false
	}
	return *o.value, true
}

func (o Option[A]) Unwrap() A {
	if o.IsSome() {
		return *o.value
	}
	panic("called Option.Unwrap on None")
}

func (o Option[A]) UnwrapOr(defaultValue A) A {
	if o.IsSome() {
		return o.Unwrap()
	}
	return defaultValue
}

func (o Option[A]) UnwrapOrElse(defaultFunc func() A) A {
	if o.IsSome() {
		return o.Unwrap()
	}
	return defaultFunc()
}

func (o Option[A]) IsSome() bool {
	return o.value != nil
}

func (o Option[A]) IsNone() bool {
	return !o.IsSome()
}
