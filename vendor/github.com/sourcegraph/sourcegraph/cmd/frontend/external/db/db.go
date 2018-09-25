package db

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"

type (
	NewUser             = db.NewUser
	ExternalAccountSpec = db.ExternalAccountSpec
	ExternalAccountData = db.ExternalAccountData
)

var Pkgs = db.Pkgs
var Users = db.Users
