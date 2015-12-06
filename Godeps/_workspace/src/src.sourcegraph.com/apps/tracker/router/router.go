package router

import "github.com/gorilla/mux"

// Router is set by tracker app New func.
var Router *mux.Router

const (
	Issue = "issue" // Route to a specific issue.
)
