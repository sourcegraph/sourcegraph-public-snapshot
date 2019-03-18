package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"

	_ "github.com/lib/pq"
)

// This script generates markdown formatted output containing descriptions of
// the current dabase schema, obtained from postgres. The correct PGHOST,
// PGPORT, PGUSER etc. env variables must be set to run this script.
//
// First CLI argument is an optional filename to write the output to.
func main() {
	const dbname = "schemadoc-gen-temp"
	_ = exec.Command("dropdb", dbname).Run()
	if out, err := exec.Command("createdb", dbname).CombinedOutput(); err != nil {
		log.Fatalf("createdb: %s, %v", out, err)
	}
	defer exec.Command("dropdb", dbname).Run()

	if err := dbconn.ConnectToDB("dbname=" + dbname); err != nil {
		log.Fatal(err)
	}

	db, err := dbconn.Open("dbname=" + dbname)
	if err != nil {
		log.Fatal("db.Open", err)
	}

	// Query names of all public tables.
	rows, err := db.Query(`
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public' AND table_type='BASE TABLE';
	`)
	if err != nil {
		log.Fatal("db.Query", err)
	}
	tables := []string{}
	defer rows.Close()
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal("rows.Scan", err)
		}
		tables = append(tables, name)
	}
	if err = rows.Err(); err != nil {
		log.Fatal("rows.Err", err)
	}

	docs := []string{}
	for _, table := range tables {
		// Get postgres "describe table" output.
		cmd := exec.Command("psql", "-X", "--quiet", "--dbname", dbname, "-c", fmt.Sprintf("\\d %s", table))
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal("cmd.CombinedOutput", out, err)
		}

		lines := strings.Split(string(out), "\n")
		doc := "# " + strings.TrimSpace(lines[0]) + "\n"
		doc += "```\n" + strings.Join(lines[1:], "\n") + "```\n"
		docs = append(docs, doc)
	}
	sort.Strings(docs)

	out := strings.Join(docs, "\n")
	if len(os.Args) > 1 {
		ioutil.WriteFile(os.Args[1], []byte(out), 0644)
	} else {
		fmt.Print(out)
	}
}
