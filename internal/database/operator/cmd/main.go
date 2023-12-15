package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/internal/database/operator"
)

func main() {
	dsn, err := url.Parse("postgres://sourcegraph:sourcegraph@localhost:5432/sourcegraph?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	err = operator.Validate(semver.MustParse("0.0.0+dev"), dsn)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("OK")
	}
}
