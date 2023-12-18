package main

import (
	"fmt"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/internal/database/operator"
)

func main() {
	err := operator.Validate(semver.MustParse("0.0.0+dev"))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("OK")
	}
}
