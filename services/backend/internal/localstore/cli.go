package localstore

import (
	"errors"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
)

func init() {
	base := &baseCmd{}
	c, err := cli.CLI.AddCommand("pgsql", "manage the PostgreSQL database", "", base)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("create", "create the databases", "", &createCmd{baseCmd: base})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("drop", "drop the databases (DELETES ALL DATA)", "", &dropCmd{baseCmd: base})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("reset", "drop and re-create the databases (DELETES ALL DATA)", "", &resetCmd{baseCmd: base})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddCommand("truncate", "truncates (removes all rows from) all tables in the databases (DELETES ALL DATA)", "", &truncateCmd{baseCmd: base})
	if err != nil {
		log.Fatal(err)
	}
}

type baseCmd struct {
	// No default name specified for this arg to prevent unintended operations on any database by forcing
	// user to specify a db name.
	Db              string `long:"db"`
	CreateDatabases bool   `short:"c" long:"createdb" description:"create PostgreSQL databases as needed"`

	dataSource string
	schema     dbutil2.Schema
	mode       dbutil2.Mode
}

func (c *baseCmd) Execute(args []string) error {
	ds, schema, err := getDataSourceAndSchema(c.Db)
	if err != nil {
		return err
	}
	c.dataSource = ds
	c.schema = *schema
	if c.CreateDatabases {
		c.mode = dbutil2.CreateDBIfNotExists
	}
	return nil
}

type createCmd struct {
	*baseCmd
}

func (c *createCmd) Execute(args []string) error {
	if err := c.baseCmd.Execute(args); err != nil {
		return err
	}
	db, err := openDB(c.dataSource, c.schema, c.mode)
	if err != nil {
		return err
	}
	return db.CreateSchema()
}

type dropCmd struct {
	*baseCmd
}

func (c *dropCmd) Execute(args []string) error {
	if err := c.baseCmd.Execute(args); err != nil {
		return err
	}
	db, err := openDB(c.dataSource, c.schema, c.mode)
	if err != nil {
		return err
	}
	return db.DropSchema()
}

type resetCmd struct {
	*baseCmd
}

func (c *resetCmd) Execute(args []string) error {
	if err := (&dropCmd{c.baseCmd}).Execute(nil); err != nil {
		return err
	}
	if err := (&createCmd{c.baseCmd}).Execute(nil); err != nil {
		return err
	}
	return nil
}

type truncateCmd struct {
	*baseCmd
}

func (c *truncateCmd) Execute(args []string) error {
	if err := c.baseCmd.Execute(args); err != nil {
		return err
	}
	db, err := openDB(c.dataSource, c.schema, c.mode)
	if err != nil {
		return err
	}
	return db.TruncateTables()
}

func getDataSourceAndSchema(dbName string) (string, *dbutil2.Schema, error) {
	switch dbName {
	case "app":
		return getAppDBDataSource(), &AppSchema, nil
	case "graph":
		return getGraphDBDataSource(), &GraphSchema, nil
	default:
		return "", nil, errors.New(fmt.Sprintf("bad db name %q, please specify one of 'app' or 'graph', eg. src pgsql --db=app <cmd>", dbName))
	}
}
