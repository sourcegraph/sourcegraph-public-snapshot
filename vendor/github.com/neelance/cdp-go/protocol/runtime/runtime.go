// Runtime domain exposes JavaScript runtime by means of remote evaluation and mirror objects. Evaluation results are returned as mirror object that expose object type, string representation and unique identifier that can be used for further object reference. Original objects are maintained in memory unless they are either explicitly released or are released along with the other objects in their object group.
package runtime

import (
	"github.com/neelance/cdp-go/rpc"
)

// Runtime domain exposes JavaScript runtime by means of remote evaluation and mirror objects. Evaluation results are returned as mirror object that expose object type, string representation and unique identifier that can be used for further object reference. Original objects are maintained in memory unless they are either explicitly released or are released along with the other objects in their object group.
type Client struct {
	*rpc.Client
}

// Unique script identifier.

type ScriptId string

// Unique object identifier.

type RemoteObjectId string

// Primitive value which cannot be JSON-stringified.

type UnserializableValue string

// Mirror object referencing original JavaScript object.

type RemoteObject struct {
	// Object type.
	Type string `json:"type"`

	// Object subtype hint. Specified for <code>object</code> type values only. (optional)
	Subtype string `json:"subtype,omitempty"`

	// Object class (constructor) name. Specified for <code>object</code> type values only. (optional)
	ClassName string `json:"className,omitempty"`

	// Remote object value in case of primitive values or JSON values (if it was requested). (optional)
	Value interface{} `json:"value,omitempty"`

	// Primitive value which can not be JSON-stringified does not have <code>value</code>, but gets this property. (optional)
	UnserializableValue UnserializableValue `json:"unserializableValue,omitempty"`

	// String representation of the object. (optional)
	Description string `json:"description,omitempty"`

	// Unique object identifier (for non-primitive values). (optional)
	ObjectId RemoteObjectId `json:"objectId,omitempty"`

	// Preview containing abbreviated property values. Specified for <code>object</code> type values only. (optional, experimental)
	Preview *ObjectPreview `json:"preview,omitempty"`

	// (optional, experimental)
	CustomPreview *CustomPreview `json:"customPreview,omitempty"`
}

// (experimental)

type CustomPreview struct {
	Header string `json:"header"`

	HasBody bool `json:"hasBody"`

	FormatterObjectId RemoteObjectId `json:"formatterObjectId"`

	BindRemoteObjectFunctionId RemoteObjectId `json:"bindRemoteObjectFunctionId"`

	// (optional)
	ConfigObjectId RemoteObjectId `json:"configObjectId,omitempty"`
}

// Object containing abbreviated remote object value. (experimental)

type ObjectPreview struct {
	// Object type.
	Type string `json:"type"`

	// Object subtype hint. Specified for <code>object</code> type values only. (optional)
	Subtype string `json:"subtype,omitempty"`

	// String representation of the object. (optional)
	Description string `json:"description,omitempty"`

	// True iff some of the properties or entries of the original object did not fit.
	Overflow bool `json:"overflow"`

	// List of the properties.
	Properties []*PropertyPreview `json:"properties"`

	// List of the entries. Specified for <code>map</code> and <code>set</code> subtype values only. (optional)
	Entries []*EntryPreview `json:"entries,omitempty"`
}

// (experimental)

type PropertyPreview struct {
	// Property name.
	Name string `json:"name"`

	// Object type. Accessor means that the property itself is an accessor property.
	Type string `json:"type"`

	// User-friendly property value string. (optional)
	Value string `json:"value,omitempty"`

	// Nested value preview. (optional)
	ValuePreview *ObjectPreview `json:"valuePreview,omitempty"`

	// Object subtype hint. Specified for <code>object</code> type values only. (optional)
	Subtype string `json:"subtype,omitempty"`
}

// (experimental)

type EntryPreview struct {
	// Preview of the key. Specified for map-like collection entries. (optional)
	Key *ObjectPreview `json:"key,omitempty"`

	// Preview of the value.
	Value *ObjectPreview `json:"value"`
}

// Object property descriptor.

type PropertyDescriptor struct {
	// Property name or symbol description.
	Name string `json:"name"`

	// The value associated with the property. (optional)
	Value *RemoteObject `json:"value,omitempty"`

	// True if the value associated with the property may be changed (data descriptors only). (optional)
	Writable bool `json:"writable,omitempty"`

	// A function which serves as a getter for the property, or <code>undefined</code> if there is no getter (accessor descriptors only). (optional)
	Get *RemoteObject `json:"get,omitempty"`

	// A function which serves as a setter for the property, or <code>undefined</code> if there is no setter (accessor descriptors only). (optional)
	Set *RemoteObject `json:"set,omitempty"`

	// True if the type of this property descriptor may be changed and if the property may be deleted from the corresponding object.
	Configurable bool `json:"configurable"`

	// True if this property shows up during enumeration of the properties on the corresponding object.
	Enumerable bool `json:"enumerable"`

	// True if the result was thrown during the evaluation. (optional)
	WasThrown bool `json:"wasThrown,omitempty"`

	// True if the property is owned for the object. (optional)
	IsOwn bool `json:"isOwn,omitempty"`

	// Property symbol object, if the property is of the <code>symbol</code> type. (optional)
	Symbol *RemoteObject `json:"symbol,omitempty"`
}

// Object internal property descriptor. This property isn't normally visible in JavaScript code.

type InternalPropertyDescriptor struct {
	// Conventional property name.
	Name string `json:"name"`

	// The value associated with the property. (optional)
	Value *RemoteObject `json:"value,omitempty"`
}

// Represents function call argument. Either remote object id <code>objectId</code>, primitive <code>value</code>, unserializable primitive value or neither of (for undefined) them should be specified.

type CallArgument struct {
	// Primitive value. (optional)
	Value interface{} `json:"value,omitempty"`

	// Primitive value which can not be JSON-stringified. (optional)
	UnserializableValue UnserializableValue `json:"unserializableValue,omitempty"`

	// Remote object handle. (optional)
	ObjectId RemoteObjectId `json:"objectId,omitempty"`
}

// Id of an execution context.

type ExecutionContextId int

// Description of an isolated world.

type ExecutionContextDescription struct {
	// Unique id of the execution context. It can be used to specify in which execution context script evaluation should be performed.
	Id ExecutionContextId `json:"id"`

	// Execution context origin.
	Origin string `json:"origin"`

	// Human readable name describing given context.
	Name string `json:"name"`

	// Embedder-specific auxiliary data. (optional)
	AuxData interface{} `json:"auxData,omitempty"`
}

// Detailed information about exception (or error) that was thrown during script compilation or execution.

type ExceptionDetails struct {
	// Exception id.
	ExceptionId int `json:"exceptionId"`

	// Exception text, which should be used together with exception object when available.
	Text string `json:"text"`

	// Line number of the exception location (0-based).
	LineNumber int `json:"lineNumber"`

	// Column number of the exception location (0-based).
	ColumnNumber int `json:"columnNumber"`

	// Script ID of the exception location. (optional)
	ScriptId ScriptId `json:"scriptId,omitempty"`

	// URL of the exception location, to be used when the script was not reported. (optional)
	URL string `json:"url,omitempty"`

	// JavaScript stack trace if available. (optional)
	StackTrace *StackTrace `json:"stackTrace,omitempty"`

	// Exception object if available. (optional)
	Exception *RemoteObject `json:"exception,omitempty"`

	// Identifier of the context where exception happened. (optional)
	ExecutionContextId ExecutionContextId `json:"executionContextId,omitempty"`
}

// Number of milliseconds since epoch.

type Timestamp float64

// Stack entry for runtime errors and assertions.

type CallFrame struct {
	// JavaScript function name.
	FunctionName string `json:"functionName"`

	// JavaScript script id.
	ScriptId ScriptId `json:"scriptId"`

	// JavaScript script name or url.
	URL string `json:"url"`

	// JavaScript script line number (0-based).
	LineNumber int `json:"lineNumber"`

	// JavaScript script column number (0-based).
	ColumnNumber int `json:"columnNumber"`
}

// Call frames for assertions or error messages.

type StackTrace struct {
	// String label of this stack trace. For async traces this may be a name of the function that initiated the async call. (optional)
	Description string `json:"description,omitempty"`

	// JavaScript function name.
	CallFrames []*CallFrame `json:"callFrames"`

	// Asynchronous JavaScript stack trace that preceded this stack, if available. (optional)
	Parent *StackTrace `json:"parent,omitempty"`

	// Creation frame of the Promise which produced the next synchronous trace when resolved, if available. (optional, experimental)
	PromiseCreationFrame *CallFrame `json:"promiseCreationFrame,omitempty"`
}

type EvaluateRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Evaluates expression on global object.
func (d *Client) Evaluate() *EvaluateRequest {
	return &EvaluateRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Expression to evaluate.
func (r *EvaluateRequest) Expression(v string) *EvaluateRequest {
	r.opts["expression"] = v
	return r
}

// Symbolic group name that can be used to release multiple objects. (optional)
func (r *EvaluateRequest) ObjectGroup(v string) *EvaluateRequest {
	r.opts["objectGroup"] = v
	return r
}

// Determines whether Command Line API should be available during the evaluation. (optional)
func (r *EvaluateRequest) IncludeCommandLineAPI(v bool) *EvaluateRequest {
	r.opts["includeCommandLineAPI"] = v
	return r
}

// In silent mode exceptions thrown during evaluation are not reported and do not pause execution. Overrides <code>setPauseOnException</code> state. (optional)
func (r *EvaluateRequest) Silent(v bool) *EvaluateRequest {
	r.opts["silent"] = v
	return r
}

// Specifies in which execution context to perform evaluation. If the parameter is omitted the evaluation will be performed in the context of the inspected page. (optional)
func (r *EvaluateRequest) ContextId(v ExecutionContextId) *EvaluateRequest {
	r.opts["contextId"] = v
	return r
}

// Whether the result is expected to be a JSON object that should be sent by value. (optional)
func (r *EvaluateRequest) ReturnByValue(v bool) *EvaluateRequest {
	r.opts["returnByValue"] = v
	return r
}

// Whether preview should be generated for the result. (optional, experimental)
func (r *EvaluateRequest) GeneratePreview(v bool) *EvaluateRequest {
	r.opts["generatePreview"] = v
	return r
}

// Whether execution should be treated as initiated by user in the UI. (optional, experimental)
func (r *EvaluateRequest) UserGesture(v bool) *EvaluateRequest {
	r.opts["userGesture"] = v
	return r
}

// Whether execution should wait for promise to be resolved. If the result of evaluation is not a Promise, it's considered to be an error. (optional)
func (r *EvaluateRequest) AwaitPromise(v bool) *EvaluateRequest {
	r.opts["awaitPromise"] = v
	return r
}

type EvaluateResult struct {
	// Evaluation result.
	Result *RemoteObject `json:"result"`

	// Exception details. (optional)
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

func (r *EvaluateRequest) Do() (*EvaluateResult, error) {
	var result EvaluateResult
	err := r.client.Call("Runtime.evaluate", r.opts, &result)
	return &result, err
}

type AwaitPromiseRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Add handler to promise with given promise object id.
func (d *Client) AwaitPromise() *AwaitPromiseRequest {
	return &AwaitPromiseRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the promise.
func (r *AwaitPromiseRequest) PromiseObjectId(v RemoteObjectId) *AwaitPromiseRequest {
	r.opts["promiseObjectId"] = v
	return r
}

// Whether the result is expected to be a JSON object that should be sent by value. (optional)
func (r *AwaitPromiseRequest) ReturnByValue(v bool) *AwaitPromiseRequest {
	r.opts["returnByValue"] = v
	return r
}

// Whether preview should be generated for the result. (optional)
func (r *AwaitPromiseRequest) GeneratePreview(v bool) *AwaitPromiseRequest {
	r.opts["generatePreview"] = v
	return r
}

type AwaitPromiseResult struct {
	// Promise result. Will contain rejected value if promise was rejected.
	Result *RemoteObject `json:"result"`

	// Exception details if stack strace is available. (optional)
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

func (r *AwaitPromiseRequest) Do() (*AwaitPromiseResult, error) {
	var result AwaitPromiseResult
	err := r.client.Call("Runtime.awaitPromise", r.opts, &result)
	return &result, err
}

type CallFunctionOnRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Calls function with given declaration on the given object. Object group of the result is inherited from the target object.
func (d *Client) CallFunctionOn() *CallFunctionOnRequest {
	return &CallFunctionOnRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the object to call function on.
func (r *CallFunctionOnRequest) ObjectId(v RemoteObjectId) *CallFunctionOnRequest {
	r.opts["objectId"] = v
	return r
}

// Declaration of the function to call.
func (r *CallFunctionOnRequest) FunctionDeclaration(v string) *CallFunctionOnRequest {
	r.opts["functionDeclaration"] = v
	return r
}

// Call arguments. All call arguments must belong to the same JavaScript world as the target object. (optional)
func (r *CallFunctionOnRequest) Arguments(v []*CallArgument) *CallFunctionOnRequest {
	r.opts["arguments"] = v
	return r
}

// In silent mode exceptions thrown during evaluation are not reported and do not pause execution. Overrides <code>setPauseOnException</code> state. (optional)
func (r *CallFunctionOnRequest) Silent(v bool) *CallFunctionOnRequest {
	r.opts["silent"] = v
	return r
}

// Whether the result is expected to be a JSON object which should be sent by value. (optional)
func (r *CallFunctionOnRequest) ReturnByValue(v bool) *CallFunctionOnRequest {
	r.opts["returnByValue"] = v
	return r
}

// Whether preview should be generated for the result. (optional, experimental)
func (r *CallFunctionOnRequest) GeneratePreview(v bool) *CallFunctionOnRequest {
	r.opts["generatePreview"] = v
	return r
}

// Whether execution should be treated as initiated by user in the UI. (optional, experimental)
func (r *CallFunctionOnRequest) UserGesture(v bool) *CallFunctionOnRequest {
	r.opts["userGesture"] = v
	return r
}

// Whether execution should wait for promise to be resolved. If the result of evaluation is not a Promise, it's considered to be an error. (optional)
func (r *CallFunctionOnRequest) AwaitPromise(v bool) *CallFunctionOnRequest {
	r.opts["awaitPromise"] = v
	return r
}

type CallFunctionOnResult struct {
	// Call result.
	Result *RemoteObject `json:"result"`

	// Exception details. (optional)
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

func (r *CallFunctionOnRequest) Do() (*CallFunctionOnResult, error) {
	var result CallFunctionOnResult
	err := r.client.Call("Runtime.callFunctionOn", r.opts, &result)
	return &result, err
}

type GetPropertiesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Returns properties of a given object. Object group of the result is inherited from the target object.
func (d *Client) GetProperties() *GetPropertiesRequest {
	return &GetPropertiesRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the object to return properties for.
func (r *GetPropertiesRequest) ObjectId(v RemoteObjectId) *GetPropertiesRequest {
	r.opts["objectId"] = v
	return r
}

// If true, returns properties belonging only to the element itself, not to its prototype chain. (optional)
func (r *GetPropertiesRequest) OwnProperties(v bool) *GetPropertiesRequest {
	r.opts["ownProperties"] = v
	return r
}

// If true, returns accessor properties (with getter/setter) only; internal properties are not returned either. (optional, experimental)
func (r *GetPropertiesRequest) AccessorPropertiesOnly(v bool) *GetPropertiesRequest {
	r.opts["accessorPropertiesOnly"] = v
	return r
}

// Whether preview should be generated for the results. (optional, experimental)
func (r *GetPropertiesRequest) GeneratePreview(v bool) *GetPropertiesRequest {
	r.opts["generatePreview"] = v
	return r
}

type GetPropertiesResult struct {
	// Object properties.
	Result []*PropertyDescriptor `json:"result"`

	// Internal object properties (only of the element itself). (optional)
	InternalProperties []*InternalPropertyDescriptor `json:"internalProperties"`

	// Exception details. (optional)
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

func (r *GetPropertiesRequest) Do() (*GetPropertiesResult, error) {
	var result GetPropertiesResult
	err := r.client.Call("Runtime.getProperties", r.opts, &result)
	return &result, err
}

type ReleaseObjectRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Releases remote object with given id.
func (d *Client) ReleaseObject() *ReleaseObjectRequest {
	return &ReleaseObjectRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Identifier of the object to release.
func (r *ReleaseObjectRequest) ObjectId(v RemoteObjectId) *ReleaseObjectRequest {
	r.opts["objectId"] = v
	return r
}

func (r *ReleaseObjectRequest) Do() error {
	return r.client.Call("Runtime.releaseObject", r.opts, nil)
}

type ReleaseObjectGroupRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Releases all remote objects that belong to a given group.
func (d *Client) ReleaseObjectGroup() *ReleaseObjectGroupRequest {
	return &ReleaseObjectGroupRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Symbolic object group name.
func (r *ReleaseObjectGroupRequest) ObjectGroup(v string) *ReleaseObjectGroupRequest {
	r.opts["objectGroup"] = v
	return r
}

func (r *ReleaseObjectGroupRequest) Do() error {
	return r.client.Call("Runtime.releaseObjectGroup", r.opts, nil)
}

type RunIfWaitingForDebuggerRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Tells inspected instance to run if it was waiting for debugger to attach.
func (d *Client) RunIfWaitingForDebugger() *RunIfWaitingForDebuggerRequest {
	return &RunIfWaitingForDebuggerRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *RunIfWaitingForDebuggerRequest) Do() error {
	return r.client.Call("Runtime.runIfWaitingForDebugger", r.opts, nil)
}

type EnableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Enables reporting of execution contexts creation by means of <code>executionContextCreated</code> event. When the reporting gets enabled the event will be sent immediately for each existing execution context.
func (d *Client) Enable() *EnableRequest {
	return &EnableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *EnableRequest) Do() error {
	return r.client.Call("Runtime.enable", r.opts, nil)
}

type DisableRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Disables reporting of execution contexts creation.
func (d *Client) Disable() *DisableRequest {
	return &DisableRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DisableRequest) Do() error {
	return r.client.Call("Runtime.disable", r.opts, nil)
}

type DiscardConsoleEntriesRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Discards collected exceptions and console API calls.
func (d *Client) DiscardConsoleEntries() *DiscardConsoleEntriesRequest {
	return &DiscardConsoleEntriesRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *DiscardConsoleEntriesRequest) Do() error {
	return r.client.Call("Runtime.discardConsoleEntries", r.opts, nil)
}

type SetCustomObjectFormatterEnabledRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// (experimental)
func (d *Client) SetCustomObjectFormatterEnabled() *SetCustomObjectFormatterEnabledRequest {
	return &SetCustomObjectFormatterEnabledRequest{opts: make(map[string]interface{}), client: d.Client}
}

func (r *SetCustomObjectFormatterEnabledRequest) Enabled(v bool) *SetCustomObjectFormatterEnabledRequest {
	r.opts["enabled"] = v
	return r
}

func (r *SetCustomObjectFormatterEnabledRequest) Do() error {
	return r.client.Call("Runtime.setCustomObjectFormatterEnabled", r.opts, nil)
}

type CompileScriptRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Compiles expression.
func (d *Client) CompileScript() *CompileScriptRequest {
	return &CompileScriptRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Expression to compile.
func (r *CompileScriptRequest) Expression(v string) *CompileScriptRequest {
	r.opts["expression"] = v
	return r
}

// Source url to be set for the script.
func (r *CompileScriptRequest) SourceURL(v string) *CompileScriptRequest {
	r.opts["sourceURL"] = v
	return r
}

// Specifies whether the compiled script should be persisted.
func (r *CompileScriptRequest) PersistScript(v bool) *CompileScriptRequest {
	r.opts["persistScript"] = v
	return r
}

// Specifies in which execution context to perform script run. If the parameter is omitted the evaluation will be performed in the context of the inspected page. (optional)
func (r *CompileScriptRequest) ExecutionContextId(v ExecutionContextId) *CompileScriptRequest {
	r.opts["executionContextId"] = v
	return r
}

type CompileScriptResult struct {
	// Id of the script. (optional)
	ScriptId ScriptId `json:"scriptId"`

	// Exception details. (optional)
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

func (r *CompileScriptRequest) Do() (*CompileScriptResult, error) {
	var result CompileScriptResult
	err := r.client.Call("Runtime.compileScript", r.opts, &result)
	return &result, err
}

type RunScriptRequest struct {
	client *rpc.Client
	opts   map[string]interface{}
}

// Runs script with given id in a given context.
func (d *Client) RunScript() *RunScriptRequest {
	return &RunScriptRequest{opts: make(map[string]interface{}), client: d.Client}
}

// Id of the script to run.
func (r *RunScriptRequest) ScriptId(v ScriptId) *RunScriptRequest {
	r.opts["scriptId"] = v
	return r
}

// Specifies in which execution context to perform script run. If the parameter is omitted the evaluation will be performed in the context of the inspected page. (optional)
func (r *RunScriptRequest) ExecutionContextId(v ExecutionContextId) *RunScriptRequest {
	r.opts["executionContextId"] = v
	return r
}

// Symbolic group name that can be used to release multiple objects. (optional)
func (r *RunScriptRequest) ObjectGroup(v string) *RunScriptRequest {
	r.opts["objectGroup"] = v
	return r
}

// In silent mode exceptions thrown during evaluation are not reported and do not pause execution. Overrides <code>setPauseOnException</code> state. (optional)
func (r *RunScriptRequest) Silent(v bool) *RunScriptRequest {
	r.opts["silent"] = v
	return r
}

// Determines whether Command Line API should be available during the evaluation. (optional)
func (r *RunScriptRequest) IncludeCommandLineAPI(v bool) *RunScriptRequest {
	r.opts["includeCommandLineAPI"] = v
	return r
}

// Whether the result is expected to be a JSON object which should be sent by value. (optional)
func (r *RunScriptRequest) ReturnByValue(v bool) *RunScriptRequest {
	r.opts["returnByValue"] = v
	return r
}

// Whether preview should be generated for the result. (optional)
func (r *RunScriptRequest) GeneratePreview(v bool) *RunScriptRequest {
	r.opts["generatePreview"] = v
	return r
}

// Whether execution should wait for promise to be resolved. If the result of evaluation is not a Promise, it's considered to be an error. (optional)
func (r *RunScriptRequest) AwaitPromise(v bool) *RunScriptRequest {
	r.opts["awaitPromise"] = v
	return r
}

type RunScriptResult struct {
	// Run result.
	Result *RemoteObject `json:"result"`

	// Exception details. (optional)
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

func (r *RunScriptRequest) Do() (*RunScriptResult, error) {
	var result RunScriptResult
	err := r.client.Call("Runtime.runScript", r.opts, &result)
	return &result, err
}

func init() {
	rpc.EventTypes["Runtime.executionContextCreated"] = func() interface{} { return new(ExecutionContextCreatedEvent) }
	rpc.EventTypes["Runtime.executionContextDestroyed"] = func() interface{} { return new(ExecutionContextDestroyedEvent) }
	rpc.EventTypes["Runtime.executionContextsCleared"] = func() interface{} { return new(ExecutionContextsClearedEvent) }
	rpc.EventTypes["Runtime.exceptionThrown"] = func() interface{} { return new(ExceptionThrownEvent) }
	rpc.EventTypes["Runtime.exceptionRevoked"] = func() interface{} { return new(ExceptionRevokedEvent) }
	rpc.EventTypes["Runtime.consoleAPICalled"] = func() interface{} { return new(ConsoleAPICalledEvent) }
	rpc.EventTypes["Runtime.inspectRequested"] = func() interface{} { return new(InspectRequestedEvent) }
}

// Issued when new execution context is created.
type ExecutionContextCreatedEvent struct {
	// A newly created execution context.
	Context *ExecutionContextDescription `json:"context"`
}

// Issued when execution context is destroyed.
type ExecutionContextDestroyedEvent struct {
	// Id of the destroyed context
	ExecutionContextId ExecutionContextId `json:"executionContextId"`
}

// Issued when all executionContexts were cleared in browser
type ExecutionContextsClearedEvent struct {
}

// Issued when exception was thrown and unhandled.
type ExceptionThrownEvent struct {
	// Timestamp of the exception.
	Timestamp Timestamp `json:"timestamp"`

	ExceptionDetails *ExceptionDetails `json:"exceptionDetails"`
}

// Issued when unhandled exception was revoked.
type ExceptionRevokedEvent struct {
	// Reason describing why exception was revoked.
	Reason string `json:"reason"`

	// The id of revoked exception, as reported in <code>exceptionUnhandled</code>.
	ExceptionId int `json:"exceptionId"`
}

// Issued when console API was called.
type ConsoleAPICalledEvent struct {
	// Type of the call.
	Type string `json:"type"`

	// Call arguments.
	Args []*RemoteObject `json:"args"`

	// Identifier of the context where the call was made.
	ExecutionContextId ExecutionContextId `json:"executionContextId"`

	// Call timestamp.
	Timestamp Timestamp `json:"timestamp"`

	// Stack trace captured when the call was made. (optional)
	StackTrace *StackTrace `json:"stackTrace"`

	// Console context descriptor for calls on non-default console context (not console.*): 'anonymous#unique-logger-id' for call on unnamed context, 'name#unique-logger-id' for call on named context. (optional, experimental)
	Context string `json:"context"`
}

// Issued when object should be inspected (for example, as a result of inspect() command line API call).
type InspectRequestedEvent struct {
	Object *RemoteObject `json:"object"`

	Hints interface{} `json:"hints"`
}
