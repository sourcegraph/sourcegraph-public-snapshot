package standalone

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path"
)

// WebFormContainerTemplateData is the param type for templates that embed the webform HTML.
// If you use WithIndexTemplate to provide an alternate HTML template for Handler, the template
// should expect a value of this type.
type WebFormContainerTemplateData struct {
	// Target is the name of the machine we are making requests to (for display purposes).
	Target string

	// WebFormContents is the generated form HTML from your ServiceDescriptors.
	WebFormContents template.HTML

	// AddlResources are additional CSS and JS files, in the form of <link> and <script>
	// tags, that we want to append to the HEAD of the index template.
	AddlResources []template.HTML
}

// HandlerOption instances allow for configuration of the standalone Handler.
type HandlerOption interface {
	apply(opts *handlerOptions)
}

// WithIndexTemplate replace the default HTML template used in Handler with the one
// given. The template will be provided an instance of WebFormContainerTemplateData
// as the data to render.
func WithIndexTemplate(tmpl *template.Template) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.indexTmpl = tmpl
	})
}

// WithCSS entirely replaces the WebFormCSS bytes used by default in Handler.
func WithCSS(css []byte) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.css = css
		opts.cssPublic = false
	})
}

// AddJS adds a JS file to Handler, serving the supplied contents at the URI
// "/s/<filename>" with a Content-Type of "text/javascript; charset=UTF-8". It
// will also be added to the AddlResources field of the WebFormContainerTemplateData
// so that it is rendered into the HEAD of the HTML page.
//
// It is safe to pass in multiple AddJS Opts to Handler. Each will be rendered in
// the order they are passed.
func AddJS(filename string, js []byte) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.tmplResources = append(opts.tmplResources, newResource(path.Join("/s", filename), js, "text/javascript; charset=utf-8", false))
	})
}

// AddJSFile is like AddJS except that the contents are provided in the form of
// a function that is used to "open" the file to read. This means that the
// contents of the file need not be eagerly loaded into memory. Each time a
// request is received for this file, the function is called.
func AddJSFile(filename string, open func() (io.ReadCloser, error)) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.tmplResources = append(opts.tmplResources, newDeferredResource(path.Join("/s", filename), open, "text/javascript; charset=utf-8"))
	})
}

// AddCSS adds a CSS file to Handler, serving the supplied contents at the URI
// "/s/<filename>" with a Content-Type of "text/css; charset=UTF-8". It
// will also be added to the AddlResources field of the WebFormContainerTemplateData
// so that it is rendered into the HEAD of the HTML page.
//
// It is safe to pass in multiple AddCSS Opts to Handler. Each will be rendered in
// the order they are passed.
func AddCSS(filename string, css []byte) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.tmplResources = append(opts.tmplResources, newResource(path.Join("/s", filename), css, "text/css; charset=utf-8", false))
	})
}

// AddCSSFile is like AddCSS except that the contents are provided in the form
// of a function that is used to "open" the file to read. This means that the
// contents of the file need not be eagerly loaded into memory. Each time a
// request is received for this file, the function is called.
func AddCSSFile(filename string, open func() (io.ReadCloser, error)) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.tmplResources = append(opts.tmplResources, newDeferredResource(path.Join("/s", filename), open, "text/css; charset=utf-8"))
	})
}

// ServeAsset will add an additional file to Handler, serving the supplied contents
// at the URI "/s/<filename>" with a Content-Type that is computed based on the given
// filename's extension.
//
// These assets could be images or other files referenced by a custom index template.
// Unlike files added via AddJS or AddCSS, they will NOT be provided to the template
// in the AddlResources field of the WebFormContainerTemplateData.
func ServeAsset(filename string, contents []byte) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.servedOnlyResources = append(opts.servedOnlyResources, newResource(path.Join("/s", filename), contents, "", false))
	})
}

// ServeAssetFile is like ServeAsset except that the contents are provided in
// the form of a function that is used to "open" the file to read. This means
// that the contents of the file need not be eagerly loaded into memory. Each
// time a request is received for this file, the function is called.
func ServeAssetFile(filename string, open func() (io.ReadCloser, error)) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.servedOnlyResources = append(opts.servedOnlyResources, newDeferredResource(path.Join("/s", filename), open, ""))
	})
}

// ServeAssetDirectory is similar to ServeAssetFile except the give name is the
// root of a subtree, which can be used to serve a directory of assets. When a
// request is received, the remaining relative path is provided to the open
// function, to indicate which path in the subtree to open. For example, if the
// given name is "foo/bar" and a request is made for "foo/bar/baz/buzz", then
// the open function will be called with "baz/buzz" as the argument.
//
// If a given path does not exist or is a directory, not a file, the open function
// should return an error that can be classified via os.IsNotExist, so that the
// server can return a "404 Not Found" status. Any other error will result in the
// server sending a "500 Internal Server Error" status.
func ServeAssetDirectory(dirname string, open func(filename string) (io.ReadCloser, error)) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.servedOnlyResources = append(opts.servedOnlyResources, newDeferredResourceFolder(path.Join("/s", dirname), open))
	})
}

// WithExamples will add examples to the UI. This function returns an error if
// it is unable to convert the given examples to JSON, which could happen if an
// example request message is not a valid request and contains invalid data or
// types (such as a channel or function).
func WithExamples(examples ...Example) (HandlerOption, error) {
	data, err := json.Marshal(examples)
	if err != nil {
		return nil, fmt.Errorf("failed to encode examples to json: %v", err)
	}
	return optFunc(func(opts *handlerOptions) {
		opts.examples = data
	}), nil
}

// WithExampleData will add examples to the UI. The given data must be valid
// JSON that can be unmarshalled as a []Example. If not, this function will
// return an error.
func WithExampleData(data []byte) (HandlerOption, error) {
	var examples []Example
	err := json.Unmarshal(data, &examples)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json to examples: %v", err)
	}
	return optFunc(func(opts *handlerOptions) {
		opts.examples = data
	}), nil
}

// WithDefaultMetadata sets the default metadata in the web form to the given
// values. Each string should be in the form "name: value".
func WithDefaultMetadata(headers []string) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.defaultMetadata = headers
	})
}

// WithMetadata adds extra request metadata that will be included when an RPC
// in invoked. Each string should be in the form "name: value". If the web
// form includes conflicting metadata, the web form input will be ignored and
// the metadata supplied to this option will be sent instead.
func WithMetadata(headers []string) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.extraMetadata = headers
	})
}

// PreserveHeaders instructs the Handler to preserve the named HTTP headers
// if they are included in the invocation request, and send them as request
// metadata when invoking the RPC. If the web form includes conflicting
// metadata, the web form input will be ignored and the matching header
// value in the HTTP request will be sent instead.
func PreserveHeaders(headerNames []string) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.preserveHeaders = headerNames
	})
}

// WithInvokeVerbosity indicates the level of log output from the gRPC UI server
// handler that performs RPC invocations.
func WithInvokeVerbosity(verbosity int) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.invokeVerbosity = verbosity
	})
}

// WithDebug enables console logging in the client JS. This prints extra
// information as the UI processes user input.
//
// Deprecated: Use WithClientDebug instead.
func WithDebug(debug bool) HandlerOption {
	return WithClientDebug(debug)
}

// WithClientDebug enables console logging in the client JS. This prints extra
// information as the UI processes user input.
func WithClientDebug(debug bool) HandlerOption {
	return optFunc(func(opts *handlerOptions) {
		opts.debug = &debug
	})
}

// optFunc implements HandlerOption
type optFunc func(opts *handlerOptions)

func (f optFunc) apply(opts *handlerOptions) {
	f(opts)
}

type handlerOptions struct {
	indexTmpl           *template.Template
	css                 []byte
	cssPublic           bool
	examples            []byte
	tmplResources       []*resource
	servedOnlyResources []*resource
	defaultMetadata     []string
	extraMetadata       []string
	preserveHeaders     []string
	invokeVerbosity     int
	debug               *bool
}

func (opts *handlerOptions) addlServedResources() []*resource {
	return append(opts.tmplResources, opts.servedOnlyResources...)
}
