package zap

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap/ot"
)

// InitializeParams holds parameters for the "initialize" request.
type InitializeParams struct {
	ID                    string             `json:"id"` // client ID
	Capabilities          ClientCapabilities `json:"capabilities"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Trace                 string             `json:"trace,omitempty"` // "off" | "messages" | "verbose"
}

// InitializeResult is the result returned from an "initialize"
// request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ClientCapabilities describes the capabilities provided by a Zap
// client.
type ClientCapabilities struct {
	E2ETestable bool `json:"e2eTestable,omitempty"` // whether the client supports Zap end-to-end tests (see test/README.md)
}

// ServerCapabilities describes the capabilities provided by a Zap
// server.
type ServerCapabilities struct {
	// WorkspaceOperationalTransformation indicates that the server
	// provides support for operational transformation on a workspace.
	WorkspaceOperationalTransformation bool `json:"workspaceOperationalTransformation"`

	// Type is either "remote" (if this server is a remote server) or
	// "local" (if this server is a local server that interfaces with
	// an editor).
	Type string `json:"type"`

	Remote *RemoteServerCapabilities    `json:"remoteServer,omitempty"`
	Local  *WorkspaceServerCapabilities `json:"workspaceServer,omitempty"`
}

// ShowStatusParams holds parameters for the "window/showStatus"
// request.
type ShowStatusParams struct {
	Message string     `json:"message"` // the local status text
	Type    StatusType `json:"type"`    // the local status type
}

// StatusType is a code that specifies the type of status.
type StatusType int

const (
	// StatusTypeError indicates there is an error that requires user
	// intervention.
	StatusTypeError = 1

	// StatusTypeWarning indicates there is a potential error but zap
	// may be able to resolve it (i.e., by retrying).
	StatusTypeWarning = 2

	// StatusTypeOK indicates that everything is OK.
	StatusTypeOK = 3
)

func (v StatusType) String() string {
	switch v {
	case StatusTypeError:
		return "error"
	case StatusTypeWarning:
		return "warning"
	case StatusTypeOK:
		return "ok"
	default:
		return "unknown"
	}
}

// ShowMessageParams holds parameters for the "window/showMessage"
// request. It is equivalent to the LSP request parameters of the same
// name
// (https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#showmessage-request).
type ShowMessageParams struct {
	Message string      `json:"message"` // the message text
	Type    MessageType `json:"type"`    // the message type
}

// MessageType is a code that specifies the type of message. It
// matches LSP's MessageType
// (https://github.com/Microsoft/language-server-protocol/blob/master/protocol.md#showmessage-notification).
type MessageType int

const (
	// MessageTypeError is an error message.
	MessageTypeError = 1

	// MessageTypeWarning is a warning.
	MessageTypeWarning = 2

	// MessageTypeInfo is an informational message.
	MessageTypeInfo = 3

	// MessageTypeLog is a log message.
	MessageTypeLog = 4
)

// RemoteServerCapabilities describes the remote capabilities provided
// by a remote Zap server.
type RemoteServerCapabilities struct{}

// RepoInfoParams contains the parameters for the "repo/info" request.
type RepoInfoParams struct {
	Repo string `json:"repo"` // the repo to get info for
}

// RepoInfoResult describes the configuration for a repo.
type RepoInfoResult struct {
	RepoConfiguration
}

// RepoConfigureParams contains the parameters for the
// "repo/configure" request.
type RepoConfigureParams struct {
	Repo string `json:"repo"` // the repo to configure

	// Remotes contains the configuration for repository remotes. The
	// map keys are the names of the remotes.
	Remotes map[string]RepoRemoteConfiguration `json:"remotes"`
}

// RepoRemoteConfiguration describes the configuration of a repository
// remote.
type RepoRemoteConfiguration struct {
	Endpoint string   `json:"endpoint"` // the endpoint URL of the server that hosts the remote repository
	Repo     string   `json:"repo"`     // the name of the repo on the remote server
	Refspecs []string `json:"refspecs"` // the refspecs describing which remote refs to sync (see (RepoWatchParams).Refspecs for spec)
}

// EquivalentTo returns whether the other RepoRemoteConfiguration is
// equivalent to c.
func (c RepoRemoteConfiguration) EquivalentTo(other RepoRemoteConfiguration) bool {
	return c.Endpoint == other.Endpoint && c.Repo == other.Repo && reflect.DeepEqual(c.Refspecs, other.Refspecs)
}

func (c RepoRemoteConfiguration) IsEmpty() bool {
	return c.Endpoint == "" && c.Repo == "" && len(c.Refspecs) == 0
}

func (c RepoRemoteConfiguration) String() string {
	if c.IsEmpty() {
		return "<none>"
	}
	return fmt.Sprintf("%s:%s refspecs(%s)", c.Endpoint, c.Repo, c.Refspecs)
}

// RepoWatchParams contains the parameters for the "repo/watch"
// request.
type RepoWatchParams struct {
	Repo string `json:"repo"` // the repo to watch

	// Refspecs are patterns that can each match any number of
	// refs. The "*" character is a wildcard when matching ref names
	// (e.g., "a*c" matches "abc"). An empty list matches no refs. To
	// watch all refs, pass a refspec "*" as one of the items.
	//
	// Each call to repo/watch overwrites all previous refspecs.
	//
	// Only non-symbolic refs are eligible to be matched.
	Refspecs []string `json:"refspecs"`
}

func (p RepoWatchParams) Validate() error {
	seen := map[string]struct{}{}
	for _, refspec := range p.Refspecs {
		if _, seen := seen[refspec]; seen {
			return fmt.Errorf("duplicate refspec %q", refspec)
		}
		seen[refspec] = struct{}{}
	}
	return nil
}

// RefIdentifier identifies a Zap ref. (A Zap branch named "B" is
// equivalent to a Zap ref named "branch/B".)
type RefIdentifier struct {
	Repo string `json:"repo"` // the repo that contains the Zap ref
	Ref  string `json:"ref"`  // the Zap ref name
}

func (r RefIdentifier) String() string {
	if r.Repo == "" {
		r.Repo = "<no-repo>"
	}
	if r.Ref == "" {
		r.Ref = "<no-ref>"
	}
	return r.Repo + ":" + r.Ref
}

// RefListParams contains the parameters for the "ref/list" request.
type RefListParams struct {
	Repo string `json:"repo"` // the repo to watch
}

// RefState represents the state of a ref (either symbolic or
// non-symbolic).
type RefState struct {
	Data   *RefData `json:"data,omitempty"`   // ref data (for non-symbolic refs only)
	Target string   `json:"target,omitempty"` // name of target ref (for symbolic refs only)
}

// DeepCopy makes a deep copy of r.
func (r RefState) DeepCopy() RefState {
	r2 := RefState{
		Target: r.Target,
	}
	if r.Data != nil {
		tmp := r.Data.DeepCopy()
		r2.Data = &tmp
	}
	return r2
}

func (r RefState) panicIfInvalid() {
	if (r.Target != "") == (r.Data != nil) {
		panic("RefState: exactly 1 of Target and Data must be set")
	}
}

// IsSymbolic reports whether r represents a symbolic ref.
func (r RefState) IsSymbolic() bool {
	r.panicIfInvalid()
	return r.Target != ""
}

func (r RefState) String() string {
	switch {
	case r.Target != "":
		return fmt.Sprintf("target(%s)", r.Target)
	case r.Data != nil:
		return fmt.Sprintf("git(%s:%s) history(%v)", r.Data.GitBranch, abbrevOID(r.Data.GitBase), r.Data.History)
	default:
		return "invalid"
	}
}

func abbrevOID(oid string) string {
	if len(oid) == 40 {
		return oid[:6]
	}
	return oid
}

// RefData describes the state of a non-symbolic ref.
type RefData struct {
	RefBase

	// History is the op sequence applied on top of RefBase to modify
	// this ref's state.
	History []ot.WorkspaceOp `json:"history"`
}

// DeepCopy makes a deep copy of r.
func (r RefData) DeepCopy() RefData {
	r2 := RefData{
		RefBase: r.RefBase,
	}
	if r.History != nil {
		r2.History = make([]ot.WorkspaceOp, len(r.History))
		for i, op := range r.History {
			r2.History[i] = op.DeepCopy()
		}
	}
	return r2
}

// RefBase describes the base Git revision and branch that the ref is
// derived from.
type RefBase struct {
	GitBase   string `json:"gitBase"`   // the 40-char SHA of the Git commit that this ref is derived from
	GitBranch string `json:"gitBranch"` // the name of the Git branch that this ref is derived from
}

func (p *RefBase) deepCopy() *RefBase {
	if p == nil {
		return nil
	}
	tmp := *p
	return &tmp
}

// RefUpdateUpstreamParams contains parameters for the "ref/update"
// request sent from the client to the server (i.e., sent upstream).
type RefUpdateUpstreamParams struct {
	RefIdentifier // the upstream (server) ref this update applies to

	// Current is the current state of the ref, as seen by the client
	// (sender). It must be provided in all cases except when creating
	// a ref when there is no existing ref by the same name, or when
	// Force is true.
	//
	// If (RefUpdateUpstreamParams).Op != nil (op update), then
	// Current.Rev must be set. See the docstring on that field for
	// more information.
	//
	// The upstream server uses it to: (1) determine at what revision
	// an op should be inserted, or (2) determine if the downstream's
	// state is stale (in which case the upstream would reject the
	// update).
	//
	// If the downstream doesn't know a valid RefPointer value to send
	// to the upstream, it can send a RefUpdateUpstreamParams with
	// Force == true instead of providing the RefPointer.
	Current *RefPointer `json:"current,omitempty"`

	// Force, if set, causes the downstream's ref state to be update
	// to the value in the State field, regardless of whether Current
	// matches the downstream's current ref state.
	//
	// It is used when the upstream does not know the current ref
	// state on the downstream and wants to clobber it, replacing it
	// with a new value.
	Force bool `json:"force,omitempty"`

	// State indicates that a new ref should be created with the given
	// state (or, if Current is non-nil, the ref's state should be
	// reset to the given state).
	State *RefState `json:"state,omitempty"`

	// Op indicates that the given operation should be applied to the
	// ref. The client's revision number must given in Rev.
	Op *ot.WorkspaceOp `json:"op,omitempty"`

	// Rev is the revision number of the server's last-acked revision,
	// from the POV of the downstream client that is sending this
	// value.
	//
	// It only needs to be set when RefUpdateUpstreamParams.Op !=
	// nil. (It is ignored for other types of updates.)
	Rev uint `json:"rev"`

	// Delete indicates that the ref should be deleted.
	Delete bool `json:"delete,omitempty"`
}

func (p RefUpdateUpstreamParams) String() string {
	var buf bytes.Buffer
	// fmt.Fprintf(&buf, "%s: ", p.RefIdentifier)
	if p.State != nil {
		var verb string
		if p.Force {
			verb = "force"
		} else if p.Current == nil {
			verb = "create"
		} else {
			verb = "reset"
		}
		fmt.Fprintf(&buf, "%s(", verb)
		if p.State != nil {
			fmt.Fprint(&buf, p.State)
		}
		fmt.Fprint(&buf, ")")
	}
	if p.Op != nil {
		fmt.Fprintf(&buf, "op@%d(%v)", p.Rev, p.Op)
	}
	if p.Delete {
		fmt.Fprintf(&buf, "delete")
	}
	return buf.String()
}

func (p RefUpdateUpstreamParams) Validate() error {
	if p.RefIdentifier == (RefIdentifier{}) {
		return errors.New("repo and ref must both be set")
	}
	if p.Force && p.Current != nil {
		return errors.New("force update ignores current, but current is set")
	}
	if isCreate := p.Current == nil && !p.Delete; isCreate {
		if p.State == nil {
			return errors.New("ref initial state must be set when creating a ref")
		}
		if p.Op != nil {
			return errors.New("only the ref initial state can be set when creating a ref")
		}
		return nil
	}
	if (p.State != nil && p.Op != nil) || (p.State != nil && p.Delete) || (p.Op != nil && p.Delete) || (p.State == nil && p.Op == nil && !p.Delete) {
		return errors.New("exactly 1 of (state,op,delete) must be set when updating a ref")
	}
	return nil
}

// DeepCopy creates a deep copy of p.
func (p RefUpdateUpstreamParams) DeepCopy() RefUpdateUpstreamParams {
	p2 := p
	p2.Current = p.Current.deepCopy()
	if p.State != nil {
		tmp := p.State.DeepCopy()
		p2.State = &tmp
	}
	if p.Op != nil {
		tmp := p.Op.DeepCopy()
		p2.Op = &tmp
	}
	return p2
}

// RefPointer describes the expected prior state of an
// ref from the peer's POV.
//
// When the upstream receives a RefUpdateUpstreamParams, it checks the
// Current field (which is a *RefPointer) to ensure the downstream's
// update pertains to the same type of ref. And vice-versa for when
// the downstream receives an update.
//
// It is only used inside RefUpdateDownstreamParams and
// RefUpdateUpstreamParams.
type RefPointer struct {
	Base   *RefBase `json:"base,omitempty"`   // if the upstream ref is a non-symbolic ref
	Target string   `json:"target,omitempty"` // if the upstream ref is a symbolic-ref
}

func (p *RefPointer) deepCopy() *RefPointer {
	if p == nil {
		return nil
	}
	return &RefPointer{
		Base:   p.Base.deepCopy(),
		Target: p.Target,
	}
}

func (p RefPointer) String() string {
	switch {
	case p.Base != nil:
		return fmt.Sprintf("base(%s)", p.Base)
	case p.Target != "":
		return fmt.Sprintf("target(%s)", p.Target)
	default:
		return "<empty>"
	}
}

// RefPointerFrom is a helper function that creates a *RefPointer
// derived from the provided state.
func RefPointerFrom(state RefState) *RefPointer {
	var p RefPointer
	switch {
	case state.Target != "":
		p.Target = state.Target
	case state.Data != nil:
		p.Base = &state.Data.RefBase
	default:
		panic("invalid state (Target == \"\" && Data == nil)")
	}
	return &p
}

// RefUpdateDownstreamParams contains parameters for the "ref/update"
// request/notification sent from the server to the client (i.e., sent
// downstream).
type RefUpdateDownstreamParams struct {
	RefIdentifier // the upstream (server) ref this update applies to

	// State indicates that a new ref was created with the given state
	// (or, if Current is non-nil, the ref's state was reset to the
	// given state).
	State *RefState `json:"state,omitempty"`

	// Op indicates that the given operation was applied to the ref.
	Op *ot.WorkspaceOp `json:"op,omitempty"`

	// Ack indicates that the server received and acknowledges the
	// last op sent by the client that receives this request.
	Ack bool `json:"ack,omitempty"`

	// Delete indicates that the ref was deleted.
	Delete bool `json:"delete,omitempty"`
}

// IsFastForward reports whether p is a fast-forward update.
func (p RefUpdateDownstreamParams) IsFastForward() bool {
	return p.Op != nil
}

func (p RefUpdateDownstreamParams) String() string {
	return p.string(false)
}

func (p RefUpdateDownstreamParams) string(includeRefIdentifier bool) string {
	var buf bytes.Buffer
	if includeRefIdentifier {
		fmt.Fprintf(&buf, "%s: ", p.RefIdentifier)
	}
	if p.Ack {
		fmt.Fprint(&buf, "ack:")
	}
	if p.State != nil {
		fmt.Fprintf(&buf, "state(%v)", p.State)
	}
	if p.Op != nil {
		fmt.Fprintf(&buf, "op(%v)", p.Op)
	}
	if p.Delete {
		fmt.Fprintf(&buf, "delete")
	}
	return buf.String()
}

func (p RefUpdateDownstreamParams) Validate() error {
	if p.RefIdentifier == (RefIdentifier{}) {
		return errors.New("repo and ref must both be set")
	}
	if (p.State != nil && p.Op != nil) || (p.State != nil && p.Delete) || (p.Op != nil && p.Delete) {
		return errors.New("exactly 1 of (state,op,delete) must be set when updating a ref")
	}
	return nil
}

// RefInfoParams contains the parameters for the "ref/info" request.
type RefInfoParams struct {
	RefIdentifier

	// Fuzzy is whether (RefInfoParams).RefIdentifier.Ref should be
	// treated as a fuzzy ref name. Example: A fuzzy ref name "foo" is
	// resolved to "branch/foo" if no ref named "foo" exists.
	Fuzzy bool `json:"fuzzy,omitempty"`
}

// RefInfo is the result from the remote "ref/info" request.
type RefInfo struct {
	RefIdentifier          // the repo and ref name (omitted by API method that return only a single RefInfo)
	RefState               // the state of the ref
	Watchers      []string `json:"watchers"` // names of clients that are watching this ref
}

// IsSymbolic reports whether r is a symbolic ref.
func (r RefInfo) IsSymbolic() bool {
	return r.Target != ""
}

// DebugLogParams contains the parameters for the "debug/log" request.
type DebugLogParams struct {
	Header bool   `json:"header,omitempty"` // render this log message as a section header in the server log
	Text   string `json:"text"`             // log message
}

//go:generate stringer -type=ErrorCode

// ErrorCode is a JSON-RPC 2.0 error code used in the Zap protocol.
type ErrorCode int64

const (
	errorCodeInvalid                     ErrorCode = iota
	ErrorCodeNotInitialized                        // the connection is not yet initialized
	ErrorCodeAlreadyInitialized                    // the connection is already initialized
	ErrorCodeRepoNotExists                         // the specified repo does not exist
	ErrorCodeRepoExists                            // the specified repo already exists
	ErrorCodeRefNotExists                          // the specified ref does not exist
	ErrorCodeRefExists                             // the specified ref exists
	ErrorCodeRefConflict                           // ref base/state conflict
	ErrorCodeRemoteNotExists                       // the specified remote does not exist
	ErrorCodeInvalidConfig                         // the repo or ref configuration is invalid
	ErrorCodeSymbolicRefInvalid                    // when a symbolic ref was given but a non-symbolic ref was required
	ErrorCodeWorkspaceNotExists                    // workspace does not exist
	ErrorCodeWorkspaceExists                       // workspace has already been added
	ErrorCodeWorkspaceIdentifierRequired           // workspace identifier is required in params
	ErrorCodeWorkspaceStateConflict                // workspace state does not match Zap branch's initial state
	ErrorCodeWorkspaceAlreadyOnBranch              // workspace is already on a branch
	ErrorCodeRefUpdateInvalid                      // an invalid ref update (e.g., trying to update a remote tracking ref)
	ErrorCodeInvalidOp                             // an invalid operation (e.g., edit with incorrect base length)
)

// Code returns the Zap-specific error code for error, or 0 if err has
// no Zap error code.
func Code(err error) ErrorCode {
	if e, ok := err.(*jsonrpc2.Error); ok {
		return ErrorCode(e.Code)
	}
	return 0
}

// Errorf formats an error message according to a format specifier and
// returns a Zap error with the given code.
func Errorf(code ErrorCode, format string, a ...interface{}) error {
	return &jsonrpc2.Error{
		Code:    int64(code),
		Message: fmt.Sprintf(format, a...),
	}
}
