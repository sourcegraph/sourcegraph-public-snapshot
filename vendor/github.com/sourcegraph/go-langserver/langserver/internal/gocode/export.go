package gocode

import (
	"go/build"
)

var bctx go_build_context

func InitDaemon(bc *build.Context) {
	bctx = pack_build_context(bc)
	g_config.ProposeBuiltins = true
	g_config.Autobuild = true
	g_daemon = new(daemon)
	g_daemon.drop_cache()
}

func SetBuildContext(bc *build.Context) {
	bctx = pack_build_context(bc)
}

func AutoComplete(file []byte, filename string, offset int) ([]candidate, int) {
	return server_auto_complete(file, filename, offset, bctx)
}

// dumb vars for unused parts of the package
var (
	g_sock              *string
	g_addr              *string
	fals                = false
	g_debug             = &fals
	get_socket_filename func() string
	config_dir          func() string
	config_file         func() string
)

// dumb types for unused parts of the package
type (
	RPC struct{}
)
