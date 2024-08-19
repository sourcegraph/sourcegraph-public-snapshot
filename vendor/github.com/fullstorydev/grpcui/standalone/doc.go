// Package standalone provides a stand-alone HTTP handler that provides a
// complete gRPC web UI that includes both the web page and relevant server-side
// handlers. The web page includes HTML, CSS, and JS sources, which can be
// customized only regarding the "target" in the page header (which is intended
// to show the server to which RPC requests are directed). The actual list of
// methods exposed via the form is also customizable.
//
// This is the same web content that the `grpcui` command-line program exposes.
// It starts up a web server that includes only this stand-alone handler. If you
// need more control over the page in which the UI form gets embedded or over
// the CSS/styling of the page, instead use the various functions in the
// "github.com/fullstorydev/grpcui" package to assemble your own web UI.
//
// Note that this package bundles the JS dependencies of the web form. It
// provides jQuery 3.3.1 and jQuery-UI 1.12.1.
//
// An example of using the package with a server supporting reflection:
//
//	cc, err := grpc.DialContext(ctx, "dns:///my-grpc-server:8080", grpc.WithBlock())
//	if err != nil {
//		return err
//	}
//	h, err := standalone.HandlerViaReflection(ctx, cc, "dns:///my-grpc-server:8080")
//	if err != nil {
//		return err
//	}
//	http.ListenAndServe(":8080", h)
package standalone
