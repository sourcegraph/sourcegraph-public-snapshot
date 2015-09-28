package pgsql

import (
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil2"
)

func init() {
	c, err := cli.CLI.AddCommand("pgsql", "manage the PostgreSQL database", "", &baseCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("create", "create the databases", "", &createCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("drop", "drop the databases (DELETES ALL DATA)", "", &dropCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("reset", "drop and re-create the databases (DELETES ALL DATA)", "", &resetCmd{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("truncate", "truncates (removes all rows from) all tables in the databases (DELETES ALL DATA)", "", &truncateCmd{})
	if err != nil {
		log.Fatal(err)
	}
}

type baseCmd struct{}

func (c *baseCmd) Execute(args []string) error {
	return nil
}

type createCmd struct {
	CreateDatabases bool `short:"c" long:"createdb" description:"create PostgreSQL databases as needed" default:"yes"`
}

func (c *createCmd) Execute(args []string) error {
	// TODO(sqs): respect the c.CreateDatabases value (and change default to no once it respects it)
	db, err := OpenDB(dbutil2.CreateDBIfNotExists)
	if err != nil {
		return err
	}
	return db.CreateSchema()
}

type dropCmd struct{}

func (c *dropCmd) Execute(args []string) error {
	db, err := OpenDB(dbutil2.CreateDBIfNotExists)
	if err != nil {
		return err
	}
	return db.DropSchema()
}

type resetCmd struct{}

func (c *resetCmd) Execute(args []string) error {
	if err := (&dropCmd{}).Execute(nil); err != nil {
		return err
	}
	if err := (&createCmd{}).Execute(nil); err != nil {
		return err
	}
	return nil
}

type truncateCmd struct{}

func (c *truncateCmd) Execute(args []string) error {
	db, err := OpenDB(dbutil2.CreateDBIfNotExists)
	if err != nil {
		return err
	}
	return db.TruncateAllTables()
}
