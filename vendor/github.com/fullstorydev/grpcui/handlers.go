package grpcui

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb" //lint:ignore SA1019 we have to import this because it appears in grpcurl APIs used herein
	"github.com/golang/protobuf/proto"  //lint:ignore SA1019 we have to import this because it appears in grpcurl APIs used herein
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/fullstorydev/grpcui/internal"
	"github.com/fullstorydev/grpcurl"
)

// RPCInvokeHandler returns an HTTP handler that can be used to invoke RPCs. The
// request includes request data, header metadata, and an optional timeout.
//
// The handler accepts POST requests with JSON bodies and returns a JSON payload
// in response. The URI path should name an RPC method ("/service/method"). The
// format of the request and response bodies matches the formats sent and
// expected by the JavaScript client code embedded in WebFormContents.
//
// The returned handler expects to serve "/". If it will instead be handling a
// sub-path (e.g. handling "/rpc/invoke/") then use http.StripPrefix.
//
// Note that the returned handler does not implement any CSRF protection. To
// provide that, you will need to wrap the returned handler with one that first
// enforces CSRF checks.
func RPCInvokeHandler(ch grpc.ClientConnInterface, descs []*desc.MethodDescriptor) http.Handler {
	return RPCInvokeHandlerWithOptions(ch, descs, InvokeOptions{})
}

// InvokeOptions contains optional arguments when creating a gRPCui invocation
// handler.
type InvokeOptions struct {
	// The set of metadata to add to all outgoing RPCs. If the invocation
	// request includes conflicting metadata, these values override, and the
	// values in the request will not be sent.
	ExtraMetadata []string
	// The set of HTTP header names that will be preserved. These are HTTP
	// request headers included in the invocation request that will be added as
	// request metadata when invoking the RPC. If the invocation request
	// includes conflicting metadata, the values in the HTTP request headers
	// will override, and the values in the request will not be sent.
	PreserveHeaders []string
	// If verbosity is greater than zero, the handler may log events, such as
	// cases where the request included metadata that conflicts with the
	// ExtraMetadata and PreserveHeaders fields above. It is an int, instead
	// of a bool "verbose" flag, so that additional logs may be added in the
	// future and the caller control how detailed those logs will be.
	Verbosity int
}

// RPCInvokeHandlerWithOptions is the same as RPCInvokeHandler except that it
// accepts an additional argument, options. This can be used to add extra
// request metadata to all RPCs invoked.
func RPCInvokeHandlerWithOptions(ch grpc.ClientConnInterface, descs []*desc.MethodDescriptor, options InvokeOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Request must be JSON", http.StatusUnsupportedMediaType)
			return
		}

		method := r.URL.Path
		if method[0] == '/' {
			method = method[1:]
		}

		for _, md := range descs {
			if md.GetFullyQualifiedName() == method {
				descSource, err := grpcurl.DescriptorSourceFromFileDescriptors(md.GetFile())
				if err != nil {
					http.Error(w, "Failed to create descriptor source: "+err.Error(), http.StatusInternalServerError)
					return
				}
				results, err := invokeRPC(r.Context(), method, ch, descSource, r.Header, r.Body, &options)
				if err != nil {
					if _, ok := err.(errReadFail); ok {
						http.Error(w, "Failed to read request", 499)
						return
					}
					if _, ok := err.(errBadInput); ok {
						http.Error(w, "Failed to parse JSON: "+err.Error(), http.StatusBadRequest)
						return
					}
					http.Error(w, "Unexpected error: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(results)
				return
			}
		}

		http.NotFound(w, r)
	})
}

// RPCMetadataHandler returns an HTTP handler that can be used to get metadata
// for a specified method.
//
// The handler accepts GET requests, using a query parameter to indicate the
// method whose schema metadata should be fetched. The response payload will be
// JSON. The format of the response body matches the format expected by the
// JavaScript client code embedded in WebFormContents.
func RPCMetadataHandler(methods []*desc.MethodDescriptor, files []*desc.FileDescriptor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		method := r.URL.Query().Get("method")
		var results *schema
		if method == "*" {
			// This means gather *all* message types. This is used to
			// provide a drop-down for Any messages.
			results = gatherAllMessageMetadata(files)
		} else {
			for _, md := range methods {
				if md.GetFullyQualifiedName() == method {
					r, err := gatherMetadataForMethod(md)
					if err != nil {
						http.Error(w, "Failed to gather metadata for RPC Method", http.StatusUnprocessableEntity)
						return
					}

					results = r
					break
				}
			}
		}

		if results == nil {
			http.Error(w, "Unknown RPC Method", http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)

		enc.SetIndent("", "  ")
		// TODO: what if enc.Encode returns a non-I/O error?
		enc.Encode(results)
	})
}

// TODO(jaime, jhump): schema is playing double duty here. It's both a vehicle for all
//  message and enum metadata. As well as RPC method scoped metadata for a single method.
//  What if we wanted to load metadata for all methods? We should consider splitting this
//  into 2 separate types for metadata to respond with accordingly.
type schema struct {
	RequestType   string                  `json:"requestType"`
	RequestStream bool                    `json:"requestStream"`
	MessageTypes  map[string][]fieldDef   `json:"messageTypes"`
	EnumTypes     map[string][]enumValDef `json:"enumTypes"`
}

type fieldDef struct {
	Name        string      `json:"name"`
	ProtoName   string      `json:"protoName"`
	Type        fieldType   `json:"type"`
	OneOfFields []fieldDef  `json:"oneOfFields"`
	IsMessage   bool        `json:"isMessage"`
	IsEnum      bool        `json:"isEnum"`
	IsArray     bool        `json:"isArray"`
	IsMap       bool        `json:"isMap"`
	IsRequired  bool        `json:"isRequired"`
	DefaultVal  interface{} `json:"defaultVal"`
	Description string      `json:"description"`
}

type enumValDef struct {
	Num  int32  `json:"num"`
	Name string `json:"name"`
}

type fieldType string

const (
	typeString   fieldType = "string"
	typeBytes    fieldType = "bytes"
	typeInt32    fieldType = "int32"
	typeInt64    fieldType = "int64"
	typeSint32   fieldType = "sint32"
	typeSint64   fieldType = "sint64"
	typeUint32   fieldType = "uint32"
	typeUint64   fieldType = "uint64"
	typeFixed32  fieldType = "fixed32"
	typeFixed64  fieldType = "fixed64"
	typeSfixed32 fieldType = "sfixed32"
	typeSfixed64 fieldType = "sfixed64"
	typeFloat    fieldType = "float"
	typeDouble   fieldType = "double"
	typeBool     fieldType = "bool"
	typeOneOf    fieldType = "oneof"
)

var typeMap = map[descriptor.FieldDescriptorProto_Type]fieldType{
	descriptor.FieldDescriptorProto_TYPE_STRING:   typeString,
	descriptor.FieldDescriptorProto_TYPE_BYTES:    typeBytes,
	descriptor.FieldDescriptorProto_TYPE_INT32:    typeInt32,
	descriptor.FieldDescriptorProto_TYPE_INT64:    typeInt64,
	descriptor.FieldDescriptorProto_TYPE_SINT32:   typeSint32,
	descriptor.FieldDescriptorProto_TYPE_SINT64:   typeSint64,
	descriptor.FieldDescriptorProto_TYPE_UINT32:   typeUint32,
	descriptor.FieldDescriptorProto_TYPE_UINT64:   typeUint64,
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  typeFixed32,
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  typeFixed64,
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: typeSfixed32,
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: typeSfixed64,
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    typeFloat,
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   typeDouble,
	descriptor.FieldDescriptorProto_TYPE_BOOL:     typeBool,
}

func gatherAllMessageMetadata(files []*desc.FileDescriptor) *schema {
	result := &schema{
		MessageTypes: map[string][]fieldDef{},
		EnumTypes:    map[string][]enumValDef{},
	}
	for _, fd := range files {
		gatherAllMessages(fd.GetMessageTypes(), result)
	}
	return result
}

func gatherAllMessages(msgs []*desc.MessageDescriptor, result *schema) {
	for _, md := range msgs {
		result.visitMessage(md)
		gatherAllMessages(md.GetNestedMessageTypes(), result)
	}
}

func gatherMetadataForMethod(md *desc.MethodDescriptor) (*schema, error) {
	msg := md.GetInputType()
	result := &schema{
		RequestType:   msg.GetFullyQualifiedName(),
		RequestStream: md.IsClientStreaming(),
		MessageTypes:  map[string][]fieldDef{},
		EnumTypes:     map[string][]enumValDef{},
	}

	result.visitMessage(msg)

	return result, nil
}

func (s *schema) visitMessage(md *desc.MessageDescriptor) {
	if _, ok := s.MessageTypes[md.GetFullyQualifiedName()]; ok {
		// already visited
		return
	}

	fields := make([]fieldDef, 0, len(md.GetFields()))
	s.MessageTypes[md.GetFullyQualifiedName()] = fields

	oneOfsSeen := map[*desc.OneOfDescriptor]struct{}{}
	for _, fd := range md.GetFields() {
		ood := fd.GetOneOf()
		if ood != nil {
			if _, ok := oneOfsSeen[ood]; ok {
				// already processed this one
				continue
			}
			oneOfsSeen[ood] = struct{}{}
			fields = append(fields, s.processOneOf(ood))
		} else {
			fields = append(fields, s.processField(fd))
		}
	}

	s.MessageTypes[md.GetFullyQualifiedName()] = fields
}

func (s *schema) processField(fd *desc.FieldDescriptor) fieldDef {
	def := fieldDef{
		Name:       fd.GetJSONName(),
		ProtoName:  fd.GetName(),
		IsEnum:     fd.GetEnumType() != nil,
		IsMessage:  fd.GetMessageType() != nil,
		IsArray:    fd.IsRepeated() && !fd.IsMap(),
		IsMap:      fd.IsMap(),
		IsRequired: fd.IsRequired(),
		DefaultVal: fd.GetDefaultValue(),
	}

	if def.IsMap {
		// fd.GetDefaultValue returns empty map[interface{}]interface{}
		// as the default for map fields, but "encoding/json" refuses
		// to encode a map with interface{} keys (even if it's empty).
		// So we fix up the key type here.
		def.DefaultVal = map[string]interface{}{}
	}

	// 64-bit int values are represented as strings in JSON
	if i, ok := def.DefaultVal.(int64); ok {
		def.DefaultVal = fmt.Sprintf("%d", i)
	} else if u, ok := def.DefaultVal.(uint64); ok {
		def.DefaultVal = fmt.Sprintf("%d", u)
	} else if b, ok := def.DefaultVal.([]byte); ok && b == nil {
		// bytes fields may have []byte(nil) as default value, but
		// that gets rendered as JSON null, not empty array
		def.DefaultVal = []byte{}
	}

	switch fd.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		def.Type = fieldType(fd.GetEnumType().GetFullyQualifiedName())
		s.visitEnum(fd.GetEnumType())
		// DefaultVal will be int32 for enums, but we want to instead
		// send enum name as string
		if val, ok := def.DefaultVal.(int32); ok {
			valDesc := fd.GetEnumType().FindValueByNumber(val)
			if valDesc != nil {
				def.DefaultVal = valDesc.GetName()
			}
		}

	case descriptor.FieldDescriptorProto_TYPE_GROUP, descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		def.Type = fieldType(fd.GetMessageType().GetFullyQualifiedName())
		s.visitMessage(fd.GetMessageType())

	default:
		def.Type = typeMap[fd.GetType()]
	}

	desc, err := protoPrinter.PrintProtoToString(fd)
	if err != nil {
		// generate simple description with no comments or options
		var label string
		if fd.IsRequired() {
			label = "required "
		} else if fd.IsRepeated() {
			label = "repeated "
		} else if fd.IsProto3Optional() || !fd.GetFile().IsProto3() {
			label = "optional "
		}
		desc = fmt.Sprintf("%s%s %s = %d;", label, def.Type, fd.GetName(), fd.GetNumber())
	}
	def.Description = desc

	return def
}

func (s *schema) processOneOf(ood *desc.OneOfDescriptor) fieldDef {
	choices := make([]fieldDef, len(ood.GetChoices()))
	for i, fd := range ood.GetChoices() {
		choices[i] = s.processField(fd)
	}
	return fieldDef{
		Name:        ood.GetName(),
		Type:        typeOneOf,
		OneOfFields: choices,
	}
}

func (s *schema) visitEnum(ed *desc.EnumDescriptor) {
	if _, ok := s.EnumTypes[ed.GetFullyQualifiedName()]; ok {
		// already visited
		return
	}

	enumVals := make([]enumValDef, len(ed.GetValues()))
	for i, evd := range ed.GetValues() {
		enumVals[i] = enumValDef{
			Num:  evd.GetNumber(),
			Name: evd.GetName(),
		}
	}

	s.EnumTypes[ed.GetFullyQualifiedName()] = enumVals
}

type errBadInput struct {
	err error
}

func (e errBadInput) Error() string {
	return e.err.Error()
}

type errReadFail struct {
	err error
}

func (e errReadFail) Error() string {
	return e.err.Error()
}

func invokeRPC(ctx context.Context, methodName string, ch grpc.ClientConnInterface, descSource grpcurl.DescriptorSource, reqHdrs http.Header, body io.Reader, options *InvokeOptions) (*rpcResult, error) {
	js, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errReadFail{err: err}
	}

	var input rpcInput
	if err := json.Unmarshal(js, &input); err != nil {
		return nil, errBadInput{err: err}
	}

	reqStats := rpcRequestStats{
		Total: len(input.Data),
	}
	requestFunc := func(m proto.Message) error {
		if len(input.Data) == 0 {
			return io.EOF
		}
		reqStats.Sent++
		req := input.Data[0]
		input.Data = input.Data[1:]
		if err := jsonpb.Unmarshal(bytes.NewReader(req), m); err != nil {
			return status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil
	}

	webFormHdrs := make(metadata.MD, len(input.Metadata))
	for _, hdr := range input.Metadata {
		webFormHdrs.Append(hdr.Name, hdr.Value)
	}
	invokeHdrs := options.computeHeaders(reqHdrs, webFormHdrs)

	if input.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		timeout := time.Duration(input.TimeoutSeconds * float32(time.Second))
		// If the timeout is too huge that it overflows int64, cap it off.
		if timeout < 0 {
			timeout = time.Duration(math.MaxInt64)
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	result := rpcResult{
		descSource: descSource,
		Requests:   &reqStats,
	}
	if err := grpcurl.InvokeRPC(ctx, descSource, ch, methodName, invokeHdrs, &result, requestFunc); err != nil {
		return nil, err
	}

	return &result, nil
}

func (opts *InvokeOptions) overrideHeaders(reqHdrs http.Header) metadata.MD {
	hdrs := grpcurl.MetadataFromHeaders(opts.ExtraMetadata)
	for _, name := range opts.PreserveHeaders {
		vals := reqHdrs.Values(name)
		if opts.Verbosity > 0 {
			if existing := hdrs.Get(name); len(existing) > 0 {
				internal.LogInfof("preserving HTTP header %q, which overrides given extra header", name)
			}
		}
		hdrs.Set(name, vals...)
	}
	return hdrs
}

func (opts *InvokeOptions) computeHeaders(reqHdrs http.Header, webFormHdrs metadata.MD) []string {
	// add extra headers, overriding whatever was in the web form
	overrideHdrs := opts.overrideHeaders(reqHdrs)
	for k, v := range overrideHdrs {
		if opts.Verbosity > 0 {
			if existing := webFormHdrs.Get(k); len(existing) > 0 {
				internal.LogInfof("web form included metadata %q, but it will be ignored due to given extra/preserved headers", k)
			}
		}
		webFormHdrs[k] = v
	}

	// convert them back to the form that the grpcurl lib expects
	totalLen := 0
	keys := make([]string, 0, len(webFormHdrs))
	for k, vs := range webFormHdrs {
		totalLen += len(vs)
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]string, 0, totalLen)
	for _, k := range keys {
		for _, v := range webFormHdrs[k] {
			result = append(result, fmt.Sprintf("%s: %s", k, v))
		}
	}
	return result
}

type rpcMetadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type rpcInput struct {
	TimeoutSeconds float32           `json:"timeout_seconds"`
	Metadata       []rpcMetadata     `json:"metadata"`
	Data           []json.RawMessage `json:"data"`
}

type rpcResponseElement struct {
	Data    json.RawMessage `json:"message"`
	IsError bool            `json:"isError"`
}

type rpcRequestStats struct {
	Total int `json:"total"`
	Sent  int `json:"sent"`
}

type rpcError struct {
	Code    uint32               `json:"code"`
	Name    string               `json:"name"`
	Message string               `json:"message"`
	Details []rpcResponseElement `json:"details"`
}

type rpcResult struct {
	descSource grpcurl.DescriptorSource
	Headers    []rpcMetadata        `json:"headers"`
	Error      *rpcError            `json:"error"`
	Responses  []rpcResponseElement `json:"responses"`
	Requests   *rpcRequestStats     `json:"requests"`
	Trailers   []rpcMetadata        `json:"trailers"`
}

func (*rpcResult) OnResolveMethod(*desc.MethodDescriptor) {}

func (*rpcResult) OnSendHeaders(metadata.MD) {}

func (r *rpcResult) OnReceiveHeaders(md metadata.MD) {
	r.Headers = responseMetadata(md)
}

func (r *rpcResult) OnReceiveResponse(m proto.Message) {
	r.Responses = append(r.Responses, responseToJSON(r.descSource, m))
}

func (r *rpcResult) OnReceiveTrailers(stat *status.Status, md metadata.MD) {
	r.Trailers = responseMetadata(md)
	r.Error = toRpcError(r.descSource, stat)
}

func responseMetadata(md metadata.MD) []rpcMetadata {
	keys := make([]string, 0, len(md))
	for k := range md {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ret := make([]rpcMetadata, 0, len(md))
	for _, k := range keys {
		vals := md[k]
		for _, v := range vals {
			if strings.HasSuffix(k, "-bin") {
				v = base64.StdEncoding.EncodeToString([]byte(v))
			}
			ret = append(ret, rpcMetadata{Name: k, Value: v})
		}
	}
	return ret
}

func toRpcError(descSource grpcurl.DescriptorSource, stat *status.Status) *rpcError {
	if stat.Code() == codes.OK {
		return nil
	}

	details := stat.Proto().Details
	msgs := make([]rpcResponseElement, len(details))
	for i, d := range details {
		msgs[i] = responseToJSON(descSource, d)
	}
	return &rpcError{
		Code:    uint32(stat.Code()),
		Name:    stat.Code().String(),
		Message: stat.Message(),
		Details: msgs,
	}
}

func responseToJSON(descSource grpcurl.DescriptorSource, msg proto.Message) rpcResponseElement {
	anyResolver := grpcurl.AnyResolverFromDescriptorSourceWithFallback(descSource)
	jsm := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, Indent: "  ", AnyResolver: anyResolver}
	var b bytes.Buffer
	if err := jsm.Marshal(&b, msg); err == nil {
		return rpcResponseElement{Data: json.RawMessage(b.Bytes())}
	} else {
		b, err := json.Marshal(err.Error())
		if err != nil {
			// unable to marshal err message to JSON?
			// should never happen... here's a dumb fallback
			b = []byte(strconv.Quote(err.Error()))
		}
		return rpcResponseElement{Data: b, IsError: true}
	}
}
