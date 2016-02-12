package pgsql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/kr/fs"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	_, err := cli.CLI.AddCommand("internal-store-migrate",
		"migrate from FS to PostgreSQL store",
		"Migrate from FS to PostgreSQL store.",
		&storeMigrateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type sqlOp struct {
	sql  string
	args []interface{}

	insert interface{}
}

type storeMigrateCmd struct {
	SGPATH string `long:"sgpath" description:"SGPATH directory on filesystem to migrate from" env:"SGPATH" default:"~/.sourcegraph"`
	Exec   bool   `short:"x" description:"actually insert data into PostgreSQL (don't just print what it'd do)"`
}

func (c *storeMigrateCmd) Execute(args []string) error {
	var ops []sqlOp

	repoOps, err := c.repos()
	if err != nil {
		return err
	}
	ops = append(ops, repoOps...)

	userOps, err := c.users()
	if err != nil {
		return err
	}
	ops = append(ops, userOps...)

	appDataOps, err := c.appData()
	if err != nil {
		return err
	}
	ops = append(ops, appDataOps...)

	isDuplicateKeyError := func(err error) bool {
		return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
	}

	dbh, err := globalDB()
	if err != nil {
		return err
	}
	for _, op := range ops {
		if op.sql != "" {
			fmt.Printf("%s [%v]\n", op.sql, op.args)
			if c.Exec {
				if _, err := dbh.Exec(op.sql, op.args...); err != nil {
					if isDuplicateKeyError(err) {
						log15.Warn("Duplicate key value; continuing after SQL statement failed", "err", err)
						continue
					}
					return err
				}
			}
		} else if op.insert != nil {
			fmt.Printf("insert %T %v\n", op.insert, op.insert)
			if c.Exec {
				if err := dbh.Insert(op.insert); err != nil {
					if isDuplicateKeyError(err) {
						log15.Warn("Duplicate key value; continuing without inserting new data", "err", err)
						continue
					}
					return err
				}
			}
		}
	}
	return nil
}

func (c *storeMigrateCmd) repos() ([]sqlOp, error) {
	var ops []sqlOp

	addRepo := func(uri, dir string) error {
		repo := &sourcegraph.Repo{
			URI:           uri,
			VCS:           "git",
			Name:          path.Base(uri),
			DefaultBranch: "master",
		}

		var dbRepo dbRepo
		dbRepo.fromRepo(repo)
		ops = append(ops, sqlOp{insert: &dbRepo})

		return nil
	}

	root := filepath.Join(c.SGPATH, "repos")
	w := fs.Walk(root)
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}

		fi := w.Stat()
		isGitRepo := fi.Name() == ".git" || fi.Name() == "HEAD"
		if isGitRepo {
			uri, err := filepath.Rel(root, filepath.Dir(w.Path()))
			if err != nil {
				return nil, err
			}
			if err := addRepo(uri, filepath.Dir(w.Path())); err != nil {
				return nil, err
			}
			w.SkipDir()
		}
	}
	return ops, nil
}

func (c *storeMigrateCmd) users() ([]sqlOp, error) {
	var ops []sqlOp

	var users []struct {
		sourcegraph.User
		EmailAddrs []*sourcegraph.EmailAddr
	}
	rawUsers, err := ioutil.ReadFile(filepath.Join(c.SGPATH, "db/users.json"))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawUsers, &users); err != nil {
		return nil, err
	}

	var passwordHashes map[string]string
	rawPasswordHashes, err := ioutil.ReadFile(filepath.Join(c.SGPATH, "db/passwords.json"))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawPasswordHashes, &passwordHashes); err != nil {
		return nil, err
	}

	for _, user := range users {
		var dbUser dbUser
		dbUser.fromUser(&user.User)
		ops = append(ops, sqlOp{insert: &dbUser})

		ops = append(ops, sqlOp{insert: &dbPassword{UID: user.UID, HashedPassword: []byte(passwordHashes[fmt.Sprint(user.UID)])}})

		for _, email := range user.EmailAddrs {
			ops = append(ops, sqlOp{insert: &userEmailAddrRow{UID: int(user.UID), EmailAddr: *email}})
		}
	}
	return ops, nil
}

func (c *storeMigrateCmd) appData() ([]sqlOp, error) {
	var ops []sqlOp

	root := filepath.Join(c.SGPATH, "appdata")
	w := fs.Walk(root)
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}

		fi := w.Stat()
		if fi.Mode().IsRegular() {
			contents, err := ioutil.ReadFile(w.Path())
			if err != nil {
				return nil, err
			}

			path, err := filepath.Rel(root, w.Path())
			if err != nil {
				return nil, err
			}

			bucketParts := strings.Split(filepath.Dir(path), string(os.PathSeparator))
			var bucket string
			if len(bucketParts) == 3 {
				// Global data
				bucket = strings.Join(bucketParts, "-")
			} else {
				// Repo data
				bucket = strings.Join([]string{
					bucketParts[0],
					strings.Join(bucketParts[1:len(bucketParts)-2], "/"), // retain repo URI slash-joined
					bucketParts[len(bucketParts)-2],
					bucketParts[len(bucketParts)-1],
				}, "-")
			}
			key := filepath.Base(path)
			ops = append(ops, sqlOp{
				sql: `WITH upsert AS (UPDATE appdata SET objects = objects || $1 WHERE name = $2 RETURNING *)
INSERT INTO appdata (name, objects) SELECT $2, $1 WHERE NOT EXISTS (SELECT * FROM upsert)`,
				args: []interface{}{
					hQuote(url.QueryEscape(key)) + "=>" + hQuote(base64.StdEncoding.EncodeToString(contents)),
					bucket,
				},
			})
		}
	}

	return ops, nil
}
