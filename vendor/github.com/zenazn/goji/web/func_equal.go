package web

import (
	"reflect"
)

func isFunc(fn interface{}) bool {
	return reflect.ValueOf(fn).Kind() == reflect.Func
}

/*
This is more than a little sketchtacular. Go's rules for function pointer
equality are pretty restrictive: nil function pointers always compare equal, and
all other pointer types never do. However, this is pretty limiting: it means
that we can't let people reference the middleware they've given us since we have
no idea which function they're referring to.

To get better data out of Go, we sketch on the representation of interfaces. We
happen to know that interfaces are pairs of pointers: one to the real data, one
to data about the type. Therefore, two interfaces, including two function
interface{}'s, point to exactly the same objects iff their interface
representations are identical. And it turns out this is sufficient for our
purposes.

If you're curious, you can read more about the representation of functions here:
http://golang.org/s/go11func
We're in effect comparing the pointers of the indirect layer.
*/
func funcEqual(a, b interface{}) bool {
	if !isFunc(a) || !isFunc(b) {
		panic("funcEqual: type error!")
	}

	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()

	return av.InterfaceData() == bv.InterfaceData()
}
