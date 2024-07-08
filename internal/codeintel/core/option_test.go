package core

import (
	"cmp"
	"testing"

	"golang.org/x/exp/rand"
	"pgregory.net/rapid"

	"github.com/sourcegraph/sourcegraph/internal/pbt"
)

func OptionChanceGenerator[A any](someChance float64, gen *rapid.Generator[A]) *rapid.Generator[Option[A]] {
	return rapid.Custom(func(t *rapid.T) Option[A] {
		seed := rapid.Uint64().Draw(t, "seed")
		rng := rand.New(rand.NewSource(seed))
		pbt.Bool(rng, someChance)
		if rapid.Bool().Draw(t, "IsSome") {
			return Some[A](gen.Draw(t, "value"))
		}
		return None[A]()
	})
}

func OptionGenerator[A any](gen *rapid.Generator[A]) *rapid.Generator[Option[A]] {
	// 90% chance of being Some, 10% chance of being None
	return OptionChanceGenerator(0.9, gen)
}

func TestOption_IsSomeIsNone(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		i := rapid.Int().Draw(t, "value")
		some := Some[int](i)
		if !some.IsSome() || some.IsNone() {
			t.Errorf("Expected Some, got None")
		}
		none := None[int]()
		if none.IsSome() || !none.IsNone() {
			t.Errorf("Expected None, got Some")
		}
	})
}

func TestOption_Get(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		i := rapid.Int().Draw(t, "value")
		some := Some[int](i)
		if v, isSome := some.Get(); isSome {
			if v != i {
				t.Errorf("Expected value %d, got %d", i, v)
			}
		} else {
			t.Errorf("expected Some")
		}
	})
}

func TestOption_UnwrapOr(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		none := None[int]()
		i := rapid.Int().Draw(t, "value")
		v := none.UnwrapOr(i)
		if v != i {
			t.Errorf("Expected %v, got %v", i, v)
		}

		j := rapid.Int().Draw(t, "value")
		some := Some[int](j)
		v2 := some.UnwrapOr(i)
		if v2 != j {
			t.Errorf("Expected %v, got %v", j, v2)
		}
	})
}

func TestOption_UnwrapOrElse(t *testing.T) {
	opt := None[int]()
	val := opt.UnwrapOrElse(func() int {
		return 21
	})
	if val != 21 {
		t.Errorf("Expected %v, got %v", 21, val)
	}

	opt2 := Some[int](42)
	val2 := opt2.UnwrapOrElse(func() int {
		panic("I'm lazy and shouldn't be called")
	})
	if val2 != 42 {
		t.Errorf("Expected %v, got %v", 42, val2)
	}
}

func TestOption_Compare(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		a := rapid.Int().Draw(t, "value")
		b := rapid.Int().Draw(t, "value")
		none := None[int]()
		someA := Some[int](a)
		someB := Some[int](b)
		if none.Compare(none, cmp.Compare[int]) != 0 {
			t.Errorf("expected None.Compare(None) to be 0")
		}
		if none.Compare(someA, cmp.Compare[int]) != -1 {
			t.Errorf("expected None.Compare(Some) to be -1")
		}
		if someA.Compare(none, cmp.Compare[int]) != 1 {
			t.Errorf("expected Some.Compare(None) to be 1")
		}
		if someA.Compare(someB, cmp.Compare[int]) != cmp.Compare[int](a, b) {
			t.Errorf("expected Some(%v).Compare(Some(%v)) to be equal to cmp.Compare(%v, %v)", a, b, a, b)
		}
	})
}
