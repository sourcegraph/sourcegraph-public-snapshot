package db

import (
	"log"
	"os/exec"
)

// InitTest creates a new test database (named with the given suffix) and configures
// this package to use it. It is called by integration tests (in a package init func)
// that need to use a real database.
func InitTest(nameSuffix string) {
	dbname := "sourcegraph-test-" + nameSuffix
	_ = exec.Command("dropdb", dbname).Run()
	if err := exec.Command("createdb", dbname).Run(); err != nil {
		log.Fatal(err)
	}
	ConnectToDB("dbname=" + dbname)
}
