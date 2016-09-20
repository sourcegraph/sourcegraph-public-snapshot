// package langp implements the Sourcegraph Language Processor REST API.
//
// Overview
//
// The application (i.e. the main codebase serving sourcegraph.com today) talks
// to a Language Processor server using a custom REST protocol. Note: VS Code
// LSP servers do not implement this.
//
// The REST protocol solely serves the needs of the App → Language Processor
// exchange and as such is inherently a (modified) version of what LSP provides.
//
// Every method is only HTTP POST, even for things that traditionally would be
// an HTTP GET. This enables us to speak purely in JSON objects and not need
// query parameter parsing/encoding (parameters are just JSON objects in the
// body of the POST request).
//
// To get an overview of the specification, see the Server interface in
// server.go. The actual JSON object structures are defined in proto.go.
package langp
