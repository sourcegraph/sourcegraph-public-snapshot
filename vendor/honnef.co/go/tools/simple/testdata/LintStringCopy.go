package pkg

import "fmt"

type S1 string

func (S1) String() string { return "hi" }

type S2 string

func (S2) Error() string { return "hi" }

func fn(s string) {
	_ = string([]byte(s)) // MATCH "should use s instead of string([]byte(s))"
	_ = "" + s            // MATCH /should use s instead of "" \+ s/
	_ = s + ""            // MATCH /should use s instead of s \+ ""/
	_ = fmt.Sprint(s)     // MATCH "should use s instead of fmt.Sprint(s)"

	_ = s
	_ = s + "foo"
	_ = s == ""
	_ = s != ""
	_ = "" +
		"really long lines follow" +
		"that need pretty formatting"
	var a1 S1
	var a2 S2
	_ = fmt.Sprint(a1)
	_ = fmt.Sprint(a2)

	_ = string([]rune(s))
	{
		string := func(v interface{}) string {
			return "foo"
		}
		_ = string([]byte(s))
	}
	{
		type byte rune
		_ = string([]byte(s))
	}
	{
		type T []byte
		var x T
		_ = string([]byte(x))
	}
}
