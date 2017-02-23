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

// RepoConfigureResult is the result of a successful "repo/configure"
// request.
type RepoConfigureResult struct {
	// UpstreamConfigurationRemovedFromRefs is a list of refs
	// (provided as information to the client, to notify the user)
	// whose upstreams resided on remotes that were removed in this
	// "repo/configure" call. The server automatically removed the
	// upstream configuration from each of these refs.
	UpstreamConfigurationRemovedFromRefs []string `json:"upstreamConfigurationRemovedFromRefs,omitempty"`
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

func (p RepoWatchParams) validate() error {
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
// equivalent to a Zap ref named "refs/heads/B".)
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

// RefBaseInfo describes what a ref is based on.
type RefBaseInfo struct {
	// GitBase is the git commit ID that all subsequent ops are based
	// on.
	GitBase string `json:"gitBase"`

	// GitBranch is the git branch (HEAD symbolic ref) that the ref is
	// associated with.
	GitBranch string `json:"gitBranch"`
}

// RefState describes the state of a ref.
type RefState struct {
	// RefBaseInfo is the base information used when the ref was created.
	RefBaseInfo

	// History is all ops that have been applied to this ref (on top
	// of the RefBaseInfo). These ops must be applied to the client's
	// local workspace before the client applies any future ops
	// streamed from the server.
	History []ot.WorkspaceOp `json:"history"`
}

func (s RefState) String() string {
	return fmt.Sprintf("git(%s:%s) history(%v)", s.GitBranch, abbrevGitOID(s.GitBase), s.History)
}

// RefUpdateUpstreamParams contains parameters for the "ref/update"
// request sent from the client to the server (i.e., sent upstream).
type RefUpdateUpstreamParams struct {
	RefIdentifier // the upstream (server) ref this update applies to

	// Current is the current state of the ref, as seen by the client
	// (sender). It must be provided in all cases except when creating
	// a ref.
	Current *RefPointer `json:"current"`

	// Force, if set, causes the downstream's ref state to be update
	// to the value in the State field, regardless of whether Current
	// matches the downstream's current ref state.
	//
	// It is used when the upstream does not know the current ref
	// state on the downstream and wants to clobber it, replacing it
	// with a new value.
	Force bool `json:"force,omitempt"`

	// State indicates that a new ref should be created with the given
	// state (or, if Current is non-nil, the ref's state should be
	// reset to the given state).
	State *RefState `json:"state,omitempty"`

	// Op indicates that the given operation should be applied to the
	// ref. The client's revision number is given in Current.Rev.
	Op *ot.WorkspaceOp `json:"op,omitempty"`

	// Delete indicates that the ref should be deleted.
	Delete bool `json:"delete,omitempty"`

	// Local indicates that if this update causes a ref to be
	// created, it should only be created on the immediate server and
	// not created on the upstream as well.
	Local bool `json:"local,omitempty"`
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
		fmt.Fprintf(&buf, "%s(%v)", verb, p.State)
	}
	if p.Op != nil {
		fmt.Fprintf(&buf, "op@%d(%v)", p.Current.Rev, p.Op)
	}
	if p.Delete {
		fmt.Fprintf(&buf, "delete")
	}
	return buf.String()
}

func (p RefUpdateUpstreamParams) validate() error {
	if p.RefIdentifier == (RefIdentifier{}) {
		return errors.New("repo and ref must both be set")
	}
	if p.Force && p.Current != nil {
		return errors.New("force update ignores current, but current is set")
	}
	if isCreate := p.Current == nil; isCreate {
		if p.State == nil {
			return errors.New("ref initial state must be set when creating a ref")
		}
		if p.Op != nil || p.Delete {
			return errors.New("only the ref initial state can be set when creating a ref")
		}
		return nil
	}
	if (p.State != nil && p.Op != nil) || (p.State != nil && p.Delete) || (p.Op != nil && p.Delete) || (p.State == nil && p.Op == nil && !p.Delete) {
		return errors.New("exactly 1 of (state,op,delete) must be set when updating a ref")
	}
	return nil
}

// RefPointer points to a specific revision number within a ref.
type RefPointer struct {
	RefBaseInfo
	Rev int `json:"rev"` // the client's revision number
}

// RefUpdateDownstreamParams contains parameters for the "ref/update"
// request/notification sent from the server to the client (i.e., sent
// downstream).
type RefUpdateDownstreamParams struct {
	RefIdentifier // the upstream (server) ref this update applies to

	// Current is the current state of the ref, as seen by the server
	// (upstream). It must be provided in all cases except when a ref
	// is created.
	Current *RefBaseInfo `json:"current,omitempty"`

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
		var verb string
		if p.Current == nil {
			verb = "create"
		} else {
			verb = "reset"
		}
		fmt.Fprintf(&buf, "%s(%v)", verb, p.State)
	}
	if p.Op != nil {
		fmt.Fprintf(&buf, "op(%v)", p.Op)
	}
	if p.Delete {
		fmt.Fprintf(&buf, "delete")
	}
	return buf.String()
}

func (p RefUpdateDownstreamParams) validate() error {
	if p.RefIdentifier == (RefIdentifier{}) {
		return errors.New("repo and ref must both be set")
	}
	if isCreate := p.Current == nil; isCreate {
		if p.State == nil {
			return errors.New("ref initial state must be set when creating a ref")
		}
		if p.Op != nil || p.Delete {
			return errors.New("only the ref initial state can be set when creating a ref")
		}
		return nil
	}
	if (p.State != nil && p.Op != nil) || (p.State != nil && p.Delete) || (p.Op != nil && p.Delete) {
		return errors.New("exactly 1 of (state,op,delete) must be set when updating a ref")
	}
	return nil
}

// RefUpdateSymbolicParams contains the parameters for the
// "ref/updateSymbolic" request/notification.
type RefUpdateSymbolicParams struct {
	RefIdentifier        // the symbolic ref to update
	Target        string `json:"target"`              // the new target
	OldTarget     string `json:"oldTarget,omitempty"` // for consistency, the old target (if any)

	Ack bool `json:"ack,omitempty"` // if this is a server ack of the client's update
}

func (p RefUpdateSymbolicParams) String() string {
	return p.string(false)
}

func (p RefUpdateSymbolicParams) string(includeRefIdentifier bool) string {
	var buf bytes.Buffer
	if includeRefIdentifier {
		fmt.Fprintf(&buf, "%s: ", p.RefIdentifier)
	}
	if p.Ack {
		fmt.Fprint(&buf, "ack:")
	}
	if p.OldTarget != "" {
		fmt.Fprintf(&buf, "%s🡒%s", p.OldTarget, p.Target)
	} else {
		fmt.Fprintf(&buf, "↦%s", p.Target)
	}
	return buf.String()
}

// RefConfigureParams contains the parameters for the "ref/configure"
// request.
type RefConfigureParams struct {
	// RefIdentifier specifies the ref to configure.
	RefIdentifier

	RefConfiguration
}

// RefConfiguration describes the configuration for a ref.
type RefConfiguration struct {
	// Upstream, if set, indicates that the ref should track the ref
	// with the same name on the given remote (which must have already
	// been configured with repo/configure).
	Upstream string `json:"upstream"`

	// Overwrite, if set, indicates that if this ref diverges from its
	// upstream, the server should clobber the upstream and replace
	// the upstream with this ref's state.
	Overwrite bool `json:"overwrite"`
}

// RefInfoResult is the result from the remote "ref/info" request.
type RefInfoResult struct {
	State  *RefState `json:"state,omitempty"`  // the state of the ref (symbolic refs are NOT resolved)
	Target string    `json:"target,omitempty"` // the target of the ref (for symbolic refs)

	// Extra diagnostics
	Wait, Buf         *ot.WorkspaceOp
	UpstreamRevNumber int
}

// RefInfo describes a ref.
type RefInfo struct {
	RefIdentifier // the repo and ref name
	RefBaseInfo
	Rev      int      `json:"rev"`      // number of ops received by the server
	Target   string   `json:"target"`   // if symbolic ref, the target ref
	Watchers []string `json:"watchers"` // IDs of clients watching this ref or repo
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
