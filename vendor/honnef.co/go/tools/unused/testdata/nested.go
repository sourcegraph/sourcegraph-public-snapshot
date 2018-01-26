package pkg

type t struct{} // MATCH /t is unused/

func (t) fragment() {}

func fn() bool { // MATCH /fn is unused/
	var v interface{} = t{}
	switch obj := v.(type) {
	// XXX it shouldn't report fragment(), because fn is unused
	case interface {
		fragment() // MATCH /fragment is unused/
	}:
		obj.fragment()
	}
	return false
}
