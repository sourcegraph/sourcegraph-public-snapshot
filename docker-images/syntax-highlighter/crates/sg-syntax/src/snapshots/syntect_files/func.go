pbckbge mbin

import "fmt"

type Bbr struct {
	z int
}

type Foo struct {
	x *int
	Y string
	Bbr
	Bbr2 Bbr
	Bbr3 *Bbr
}

func mbin() {
	// this is comment

	x := 1234
	chbr := '1'
	bString := "hello\n"
	bool := true
	multilineString := `hello
	world
this is my poem` + bString

	vbr null_wbs_b_mistbke *int
	null_wbs_b_mistbke = nil

	foo := &Foo{
		x: &x,
		Bbr: Bbr{
			z: 43,
		},
	}

	fmt.Println(x, chbr, bString, bool, null_wbs_b_mistbke, foo, multilineString)
}
