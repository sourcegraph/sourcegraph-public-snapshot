package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

var (
	email    = flag.String("e", "test@sourcegraph.com", "email")
	username = flag.String("u", "seed-user", "username")
	password = flag.String("p", "123456", "password")
)

func main() {
	flag.Parse()

	if *username == "" || *password == "" || *email == "" {
		flag.Usage()
		os.Exit(1)
	}

	dsn := dbutil.PostgresDSN("", os.Getenv)
	conn, err := dbutil.NewDB(dsn, "seed_database")
	if err != nil {
		log.Fatalf("failed to initialize db store: %v", err)
	}

	dbconn.Global = conn

	u, err := db.Users.Create(context.Background(), db.NewUser{
		Email:                 *email,
		Username:              *username,
		DisplayName:           *username,
		Password:              *password,
		EmailVerificationCode: "verification-code",
	})
	if err != nil {
		if db.IsEmailExists(err) {
			fmt.Printf("User with that email already exists\n")
			return
		}
		if db.IsUsernameExists(err) {
			fmt.Printf("User with that username already exists\n")
			return
		}
		log.Fatal(err)
	}

	fmt.Printf("User created. username=%s, password=%s\n", u.Username, *password)
}
