package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// This script generates markdown formatted output containing descriptions of
// the current dabase schema, obtained from postgres. The correct PGHOST,
// PGPORT, PGUSER etc. env variables must be set to run this script.
//
// First CLI argument is an optional filename to write the output to.
func main() {
	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Query names of all public tables.
	rows, err := db.Query(`
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public' AND table_type='BASE TABLE';
	`)
	if err != nil {
		log.Fatal(err)
	}
	tables := []string{}
	defer rows.Close()
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	docs := []string{}
	for _, table := range tables {
		// Get postgres "describe table" output.
		cmd := exec.Command("psql", "-c", fmt.Sprintf("\\d %s", table))
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
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
