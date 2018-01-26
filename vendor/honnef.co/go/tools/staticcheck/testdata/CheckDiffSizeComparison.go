package pkg

func fn1() {
	var s1 = "foobar"
	_ = "a"[:] == s1           // MATCH /comparing strings of different sizes/
	_ = s1 == "a"[:]           // MATCH /comparing strings of different sizes/
	_ = "a"[:] == s1[:2]       // MATCH /comparing strings of different sizes/
	_ = "ab"[:] == s1[1:2]     // MATCH /comparing strings of different sizes/
	_ = "ab"[:] == s1[0+1:2]   // MATCH /comparing strings of different sizes/
	_ = "a"[:] == "abc"        // MATCH /comparing strings of different sizes/
	_ = "a"[:] == "a"+"bc"     // MATCH /comparing strings of different sizes/
	_ = "foobar"[:] == s1+"bc" // MATCH /comparing strings of different sizes/
	_ = "a"[:] == "abc"[:]     // MATCH /comparing strings of different sizes/
	_ = "a"[:] == "abc"[:2]    // MATCH /comparing strings of different sizes/

	_ = "a" == s1 // ignores
	_ = s1 == "a" // ignored
	_ = "abcdef"[:] == s1
	_ = "ab"[:] == s1[:2]
	_ = "a"[:] == s1[1:2]
	_ = "a"[:] == s1[0+1:2]
	_ = "abc"[:] == "abc"
	_ = "abc"[:] == "a"+"bc"
	_ = s1[:] == "foo"+"bar"
	_ = "abc"[:] == "abc"[:] // MATCH /identical expressions on the left and right side/
	_ = "ab"[:] == "abc"[:2]
}

func fn2() {
	s1 := "123"
	if true {
		s1 = "1234"
	}

	_ = s1 == "12345"[:] // MATCH /comparing strings of different sizes/
	_ = s1 == "1234"[:]
	_ = s1 == "123"[:]
	_ = s1 == "12"[:] // MATCH /comparing strings of different sizes/
}
