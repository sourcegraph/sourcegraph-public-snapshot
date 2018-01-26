package pkg

import (
	"log"
	"regexp"
)

const c1 = `[`
const c2 = `abc`

var re1 = regexp.MustCompile(`ab\yef`) // MATCH /error parsing regexp/
var re2 = regexp.MustCompile(c1)       // MATCH /error parsing regexp/
var re3 = regexp.MustCompile(c2)

func fn() {
	_, err := regexp.Compile(`foo(`) // MATCH /error parsing regexp/
	if err != nil {
		panic(err)
	}
	if re2.MatchString("foo(") {
		log.Println("of course 'foo(' matches 'foo('")
	}
}
