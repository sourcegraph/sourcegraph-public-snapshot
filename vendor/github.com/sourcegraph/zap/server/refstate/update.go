package refstate

import (
	"bytes"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/server/refdb"
	"github.com/sourcegraph/zap/server/repodb"
)

//go:generate stringer -type=Source

// Source describes where a RefUpdate originated from.
type Source int

const (
	invalidSource  Source = iota
	Internal              // from this server (e.g., a workspace file change); do not apply this update
	FromDownstream        // from a downstream repository
	FromUpstream          // from an upstream repository
)

// RefUpdate describes an in-progress ref update (either upstream or
// downstream).
//
// See those types' docstrings for more information.
type RefUpdate struct {
	// Source is where the update originated from.
	Source `json:"source"`

	// Union of fields from RefUpdateDownstreamParams and
	// RefUpdateUpstreamParams.
	zap.RefIdentifier
	Force   bool            `json:"force,omitempty"`
	Current *zap.RefPointer `json:"current,omitempty"`
	State   *zap.RefState   `json:"state,omitempty"`
	Op      *ot.WorkspaceOp `json:"op,omitempty"`
	Rev     uint            `json:"rev"`
	Ack     bool            `json:"ack,omitempty"`
	Delete  bool            `json:"delete,omitempty"`
}

func (u RefUpdate) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, u.Source.String(), ":")
	if u.Force {
		fmt.Fprint(&buf, " force")
	}
	if u.Ack {
		fmt.Fprint(&buf, " ack")
	}
	if u.Current != nil {
		fmt.Fprintf(&buf, " cur(%s)", u.Current)
	}
	if u.State != nil {
		fmt.Fprintf(&buf, " state(%s)", u.State)
	}
	if u.Op != nil {
		var rev string // only Source==FromDownstream updates have a rev
		if u.Source == FromDownstream {
			rev = fmt.Sprintf("@%d", u.Rev)
		}
		fmt.Fprintf(&buf, " op%s(%s)", rev, u.Op)
	}
	if u.Delete {
		fmt.Fprint(&buf, " delete")
	}
	return buf.String()
}

// Noop reports whether the update is a noop. Note that the update
// being a noop doesn't mean the same thing as the op being a noop
// like ot.WorkspaceOp{}. The update being a noop means we should
// pretend like we never received this update at all.
func (u RefUpdate) Noop() bool {
	return u.State == nil && u.Op == nil && !u.Delete
}

// FromUpdateUpstream clears u and sets its fields from the given
// RefUpdateUpstreamParams.
func (u *RefUpdate) FromUpdateUpstream(o zap.RefUpdateUpstreamParams) {
	*u = RefUpdate{
		Source:        FromDownstream, // confusing, but an upstream update is an update from downstream -> upstream
		RefIdentifier: o.RefIdentifier,
		Current:       o.Current,
		Force:         o.Force,
		State:         o.State,
		Op:            o.Op,
		Rev:           o.Rev,
		Delete:        o.Delete,
	}
}

// FromUpdateDownstream clears u and sets its fields from the given
// RefUpdateDownstreamParams.
func (u *RefUpdate) FromUpdateDownstream(o zap.RefUpdateDownstreamParams) {
	*u = RefUpdate{
		Source:        FromUpstream, // confusing, but a downstream update is an update from upstream -> downstream
		RefIdentifier: o.RefIdentifier,
		State:         o.State,
		Op:            o.Op,
		Ack:           o.Ack,
		Delete:        o.Delete,
	}
}

// ToUpdateDownstream returns a RefUpdateDownstreamParams equivalent
// to u.
func (u RefUpdate) ToUpdateDownstream() zap.RefUpdateDownstreamParams {
	return zap.RefUpdateDownstreamParams{
		RefIdentifier: u.RefIdentifier,
		State:         u.State,
		Op:            u.Op,
		Ack:           u.Ack,
		Delete:        u.Delete,
	}
}

// ToUpdateUpstream returns a RefUpdateUpstreamParams equivalent to u.
func (u RefUpdate) ToUpdateUpstream() zap.RefUpdateUpstreamParams {
	return zap.RefUpdateUpstreamParams{
		RefIdentifier: u.RefIdentifier,
		Current:       u.Current,
		Force:         u.Force,
		State:         u.State,
		Op:            u.Op,
		Rev:           u.Rev,
		Delete:        u.Delete,
	}
}

// Update receives a ref update (either upstream or downstream),
// recording and applying it to the ref's state.
//
// It may modify params (e.g., if params.Op needs to be transformed to
// be consecutive to the ref's history).
//
// It should be kept in sync with libzap's refState.update func.
func Update(logger log.Logger, repo *repodb.OwnedRepo, ref *refdb.OwnedRef, params *RefUpdate) error {
	level.Debug(logger).Log("refstate.Update", params.RefIdentifier, "update", fmt.Sprintf("%+v", params))

	if params.Source == invalidSource {
		panic("invalid params.Source: " + params.Source.String())
	}
	if repo == nil || repo.Repo == nil {
		panic("repo is nil")
	}
	if repo.Path != params.RefIdentifier.Repo {
		panic("repo path mismatch: " + repo.Path + " != " + params.RefIdentifier.Repo)
	}
	if ref.Ref != nil && ref.Ref.Name != params.RefIdentifier.Ref {
		panic("ref name mismatch: " + ref.Ref.Name + " != " + params.RefIdentifier.Ref)
	}

	if params.Ack {
		if params.Source != FromUpstream {
			panic("ref update ack received from unexpected source " + params.Source.String())
		}
		if params.Delete {
			return nil // nothing to do
		}
		if ref.Ref == nil {
			return zap.Errorf(zap.ErrorCodeRefNotExists, "ack ref %v: ref does not exist (update: %v)", params.RefIdentifier, params)
		}
		u := ref.Ref.Data.(RefState).Upstream
		if u == nil {
			return zap.Errorf(zap.ErrorCodeRefUpdateInvalid, "ack ref %v: ref has no upstream configured (update: %v)", params.RefIdentifier, params)
		}
		if err := u.ackFromUpstream(); err != nil {
			return zap.Errorf(zap.ErrorCodeRefUpdateInvalid, "ack ref %v: %s (update: %v)", params.RefIdentifier, err, params)
		}
		return nil
	}

	if params.Delete {
		if err := deleteRef(repo.RefDB, ref, *params); err != nil {
			return err
		}
	} else {
		// Validate.
		if params.Force && params.Op != nil {
			return zap.Errorf(zap.ErrorCodeRefUpdateInvalid, "update ref %v: force is incompatible with op updates", params.RefIdentifier)
		}

		// Treat internal and upstream updates as "always force" for
		// now. TODO(nick): When needed, decide on how the downstream
		// should decide when to accept a reset or other conflicting
		// change from the upstream.
		force := params.Force || (params.Source == Internal || params.Source == FromUpstream)
		if err := prepareRefToUpdate(logger, *repo, ref, params.RefIdentifier, params.Current, force); err != nil {
			return err
		}

		// Perform update.
		refState := ref.Ref.Data.(RefState)
		switch {
		case params.State != nil:
			// Copy to obtain exclusive ownership over this. At least
			// in tests, our mocks sometimes actually reuse the same
			// object in memory if we don't do this.
			refState.RefState = params.State.DeepCopy()

			if refState.Upstream != nil {
				if params.Source == FromUpstream {
					if err := refState.Upstream.recvFromUpstream(params); err != nil {
						return err
					}
				} else {
					// We received a reset from downstream (or from local
					// workspace).  The server hasn't yet acked anything
					// here, so we're back to rev 0.
					//
					// We keep Wait and Buf, though, instead of clearing
					// them immediately. We keep Wait because otherwise
					// the server's next ack (which is for the current
					// Wait) would be interpreted as an ack of the wrong
					// Wait value. We keep Buf because it'll be
					// clobbered anyway by our params in the call to
					// (*Upstream).sendUpdate at the end of this func.
					refState.Upstream.Rev = 0
				}
			}

		case params.Op != nil:
			if refState.IsSymbolic() {
				return zap.Errorf(zap.ErrorCodeRefConflict, "update symbolic ref %v: can't recv op updates directly (use a non-symbolic ref)", params.RefIdentifier)
			}
			data := refState.Data
			rev := params.Rev

			// Transform the op so that it can be recorded at the
			// current point in time (in our OT history).
			switch params.Source {
			case FromDownstream, Internal:
				// Transform it so it can be appended to our view of the
				// history. The op will be modified.
				if uint(len(data.History)) < rev {
					return zap.Errorf(zap.ErrorCodeRefUpdateInvalid, "ref update %v: revision %d not in history", params.RefIdentifier, rev)
				}
				var err error
				for _, other := range data.History[rev:] {
					if *params.Op, _, err = ot.TransformWorkspaceOps(*params.Op, other); err != nil {
						return err
					}
				}

			case FromUpstream:
				if u := refState.Upstream; u != nil {
					if u.Rev > uint(len(data.History)) {
						panic(fmt.Sprintf("invalid upstream rev > len(history) (%d > %d)", u.Rev, len(data.History)))
					}
					if err := u.recvFromUpstream(params); err != nil {
						return err
					}
				}

			default:
				panic("unhandled ref update source " + params.Source.String())
			}

			// Persist the op. (Unless the update became a noop after
			// being transformed with our history.)
			if !params.Noop() {
				data.History = append(data.History, *params.Op)
			}
		}

		ref.Ref.Data = refState
		if err := repo.RefDB.Write(*ref); err != nil {
			return err
		}
	}

	// Send upstream. (Unless the update became a noop.)
	if params.Source != FromUpstream && !params.Noop() {
		if u := ref.Ref.Data.(RefState).Upstream; u != nil {
			level.Debug(logger).Log("send-upstream", params.RefIdentifier)
			if err := u.sendUpdate(params.ToUpdateUpstream()); err != nil {
				return err
			}
		} else {
			level.Debug(logger).Log("no-upstream", params.RefIdentifier)
		}
	}

	return nil
}

// newRef populates a ref's initial state for a ref/update (if the ref
// doesn't already exist). It modifies the ref argument.
//
// It does not add ref to the refdb. That should be done after the
// update is applied to the ref's state, to ensure that the update is
// persisted.
func newRef(ref *refdb.OwnedRef, refName string) error {
	if ref.Ref != nil {
		panic("createRef: ref already exists")
	}
	ref.Ref = &refdb.Ref{
		Name: refName,
		Data: RefState{},
	}
	return nil
}

// prepareRefToUpdate prepares ref to be created or updated. The ref
// is checked for consistency (against the current value), and the ref
// is created if it doesn't exist.
//
// The ref arg is modified.
//
// The sendToUpstream is a channel whose values are sent upstream.
func prepareRefToUpdate(logger log.Logger, repo repodb.OwnedRepo, ref *refdb.OwnedRef, refID zap.RefIdentifier, current *zap.RefPointer, force bool) error {
	logger = log.With(logger, "prepare-ref-to-update", "")
	// Check consistency if the ref exists. (Force updates bypass this check.)
	if ref.Ref != nil && !force {
		if current == nil {
			return zap.Errorf(zap.ErrorCodeRefExists, "update existing ref %v: requires force == true or current != nil", refID)
		} else if err := compareRefPointerInfo(*current, ref.Ref.Data.(RefState).RefState); err != nil {
			return zap.Errorf(zap.ErrorCodeRefConflict, "update existing ref %v: inconsistent: %s", refID, err)
		}
	}

	// Create the ref in the refdb if it doesn't already exist.
	if ref.Ref == nil {
		if current != nil {
			return zap.Errorf(zap.ErrorCodeRefNotExists, "update ref %v: ref does not exist but current != nil", refID)
		}
		if err := newRef(ref, refID.Ref); err != nil {
			return err
		}

		// Set up upstream state, if there is an upstream.
		if repo.Repo.SendRefUpdateUpstream != nil {
			level.Debug(logger).Log("send-ref-update-upstream", "true")
			remoteRepo := repo.Config.Remote.Repo
			if remoteRepo == "" {
				remoteRepo = repo.Path
			}

			state := ref.Ref.Data.(RefState)
			state.Upstream = &Upstream{
				Send:       repo.Repo.SendRefUpdateUpstream,
				RemoteRepo: remoteRepo,
			}
			ref.Ref.Data = state
		} else {
			level.Debug(logger).Log("send-ref-update-upstream", "false")
		}
	}

	if repo.Repo.SendRefUpdateUpstream != nil && (ref.Ref.Data.(RefState).Upstream == nil || ref.Ref.Data.(RefState).Upstream.Send == nil) {
		panic("repo.Repo.SendRefUpdateUpstream is set but ref upstream is not")
	}

	return nil
}

// deleteRef deletes a ref for a ref/update with Delete == true.
//
// The force and current parameters should be the values of
// RefUpdate{Downstream,Upstream}Params.{Current,Force}.
func deleteRef(refdb *refdb.SyncRefDB, ref *refdb.OwnedRef, params RefUpdate) error {
	if ref.Ref == nil {
		return zap.Errorf(zap.ErrorCodeRefNotExists, "downstream sent ref deletion for nonexistent ref %q", params.RefIdentifier)
	}
	if params.Source != FromUpstream {
		if !params.Force {
			if params.Current == nil {
				return zap.Errorf(zap.ErrorCodeRefUpdateInvalid, "delete ref %v: requires force == true or current != nil", params.RefIdentifier)
			}
			if err := compareRefPointerInfo(*params.Current, ref.Ref.Data.(RefState).RefState); err != nil {
				return zap.Errorf(zap.ErrorCodeRefConflict, "delete rev %v: inconsistent current: %s", params.RefIdentifier, err)
			}
		}
	}

	// Delete the ref.
	if err := refdb.Delete(*ref); err != nil {
		return err
	}
	return nil
}
