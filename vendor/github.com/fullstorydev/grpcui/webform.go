package grpcui

import (
	"bytes"
	"fmt"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/desc/protoprint"
	"html/template"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/jhump/protoreflect/desc"

	"github.com/fullstorydev/grpcui/internal/resources/webform"
)

var (
	webFormTemplate = template.Must(template.New("grpc web form").Parse(string(webform.Template())))

	protoPrinter = protoprint.Printer{
		Compact: true,
		Indent:  "   ",
	}
)

// WebFormContents returns an HTML form that can be embedded into a web UI to
// provide an interactive form for issuing RPCs.
//
// For a fully self-contained handler that provides both an HTML UI and the
// needed server handlers, see grpcui.UIHandler instead.
//
// The given invokeURI and metadataURI indicate the URI paths where server
// handlers are registered for invoking RPCs and querying RPC metadata,
// respectively. Handlers for these endpoints are provided via the
// RPCInvokeHandler and RPCMetadataHandler functions:
//
//   // This example uses "/rpcs" as the base URI.
//   pageHandler := func(w http.ResponseWriter, r *http.Request) {
//     webForm := grpcui.WebFormContents("/rpcs/invoke/", "/rpcs/metadata", descs)
//     webFormJs := grpcui.WebFormScript()
//     generateHTMLPage(w, r, webForm, webFormJs)
//   }
//
//   // Make sure the RPC handlers are registered at the same URI paths
//   // that were used in the call to WebFormContents:
//   rpcInvokeHandler := http.StripPrefix("/rpcs/invoke", grpcui.RPCInvokeHandler(conn, descs))
//   mux.Handle("/rpcs/invoke/", rpcInvokeHandler)
//   mux.Handle("/rpcs/metadata", grpcui.RPCMetadataHandler(descs))
//   mux.HandleFunc("/rpcs/index.html", pageHandler)
//
// The given descs is a slice of methods which are exposed through the web form.
// You can use AllMethodsForServices, AllMethodsForServer, and
// AllMethodsViaReflection helper functions to build this list.
//
// The returned HTML form requires that the contents of WebFormScript() have
// already been loaded as a script in the page.
func WebFormContents(invokeURI, metadataURI string, descs []*desc.MethodDescriptor) []byte {
	return WebFormContentsWithOptions(invokeURI, metadataURI, descs, WebFormOptions{})
}

// WebFormOptions contains optional arguments when creating a gRPCui web form.
type WebFormOptions struct {
	// The set of metadata to show in the web form by default. Each value in
	// the slice should be in the form "name: value"
	DefaultMetadata []string
	// If non-nil and true, the web form JS code will log debug information
	// to the JS console. If nil, whether debug is enabled or not depends on
	// an environment variable: GRPC_WEBFORM_DEBUG (if it's not blank, then
	// debug is enabled).
	Debug *bool
}

// WebFormContentsWithOptions is the same as WebFormContents except that it
// accepts an additional argument, options. This can be used to toggle the JS
// code into debug logging and can also be used to define the set of metadata to
// show in the web form by default (empty if unspecified).
func WebFormContentsWithOptions(invokeURI, metadataURI string, descs []*desc.MethodDescriptor, opts WebFormOptions) []byte {
	type metadataEntry struct {
		Name, Value string
	}
	params := struct {
		InvokeURI       string
		MetadataURI     string
		Services        []string
		SvcDescs        map[string]string
		Methods         map[string][]string
		MtdDescs        map[string]string
		DefaultMetadata []metadataEntry
		Debug           bool
	}{
		InvokeURI:   invokeURI,
		MetadataURI: metadataURI,
		SvcDescs:    map[string]string{},
		Methods:     map[string][]string{},
		MtdDescs:    map[string]string{},
		Debug:       os.Getenv("GRPC_WEBFORM_DEBUG") != "",
	}

	if opts.Debug != nil {
		params.Debug = *opts.Debug
	}
	for _, md := range opts.DefaultMetadata {
		parts := strings.SplitN(md, ":", 2)
		key := strings.TrimSpace(parts[0])
		var val string
		if len(parts) > 1 {
			val = strings.TrimLeftFunc(parts[1], unicode.IsSpace)
		}
		params.DefaultMetadata = append(params.DefaultMetadata, metadataEntry{Name: key, Value: val})
	}

	// build list of distinct service and method names and sort them
	uniqueServices := map[string]*desc.ServiceDescriptor{}
	for _, md := range descs {
		sd := md.GetService()
		svcName := sd.GetFullyQualifiedName()
		uniqueServices[svcName] = sd
		params.Methods[svcName] = append(params.Methods[svcName], md.GetName())

		desc, err := protoPrinter.PrintProtoToString(md)
		if err != nil {
			// generate simple description with no comments or options
			var reqStr, respStr string
			if md.IsClientStreaming() {
				reqStr = "stream "
			}
			if md.IsServerStreaming() {
				respStr = "stream "
			}
			desc = fmt.Sprintf("   rpc %s (%s%s) returns (%s%s);", md.GetName(), reqStr, md.GetInputType().GetFullyQualifiedName(), respStr, md.GetOutputType().GetFullyQualifiedName())
		} else {
			// indent and remove trailing newline
			desc = strings.TrimSuffix(desc, "\n")
			parts := strings.Split(desc, "\n")
			for i := range parts {
				parts[i] = "   " + parts[i]
			}
			desc = strings.Join(parts, "\n")
		}
		params.MtdDescs[md.GetFullyQualifiedName()] = desc
	}
	for svcName, sd := range uniqueServices {
		params.Services = append(params.Services, svcName)
		// for the service description, omit the methods (we just want the
		//  service's comments and options)
		sb, err := builder.FromService(sd)
		var desc string
		if err == nil {
			for _, md := range sd.GetMethods() {
				sb.RemoveMethod(md.GetName())
			}
			if s, e := sb.Build(); e == nil {
				desc, err = protoPrinter.PrintProtoToString(s)
			}
		}
		if err != nil {
			desc = fmt.Sprintf("service %s {", sd.GetName())
		} else {
			// strip last line, trailing close brace
			tr := "\n}"
			if strings.HasSuffix(desc, "\n") {
				tr += "\n"
			}
			desc = strings.TrimSuffix(desc, tr)
		}
		params.SvcDescs[sd.GetFullyQualifiedName()] = desc
	}
	sort.Strings(params.Services)
	for _, methods := range params.Methods {
		sort.Strings(methods)
	}

	// render the template
	var formBuf bytes.Buffer
	if err := webFormTemplate.Execute(&formBuf, params); err != nil {
		panic(err)
	}
	return formBuf.Bytes()
}

// WebFormScript returns the JavaScript that powers the web form returned by
// WebFormContents.
//
// The returned JavaScript requires that jQuery and jQuery UI libraries already
// be loaded in the container HTML page. It includes JavaScript code that relies
// on the "$" symbol.
//
// Note that the script, by default, does not handle CSRF protection. To add
// that, the enclosing page, in which the script is embedded, can use jQuery to
// configure this. For example, you can use the $.ajaxSend() jQuery function to
// intercept RPC invocations and automatically add a CSRF token header. To then
// check the token on the server-side, you will need to create a wrapper handler
// that first verifies the CSRF token header before delegating to a
// RPCInvokeHandler.
func WebFormScript() []byte {
	return webform.Script()
}

// WebFormSampleCSS returns a CSS stylesheet for styling the HTML web form
// returned by WebFormContents. It is possible for uses of the web form to
// supply their own stylesheet, but this makes it simple to use default
// styling.
func WebFormSampleCSS() []byte {
	return webform.SampleCSS()
}
