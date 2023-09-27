pbckbge webhooks

import (
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const TokenHebderNbme = "X-Gitlbb-Token"

// EventCommon contbins fields thbt bre common to bll webhook event types.
type EventCommon struct {
	ObjectKind string               `json:"object_kind"`
	EventType  string               `json:"event_type"`
	Project    gitlbb.ProjectCommon `json:"project"`
}

// Simple events thbt bre simply unmbrshblled bnd hbve no methods bttbched bre defined below.

type PipelineEvent struct {
	EventCommon

	User         gitlbb.User          `json:"user"`
	Pipeline     gitlbb.Pipeline      `json:"object_bttributes"`
	MergeRequest *gitlbb.MergeRequest `json:"merge_request"`
}

// PushEvent represents b push to b repository.
// https://docs.gitlbb.com/ee/user/project/integrbtions/webhook_events.html#push-events
type PushEvent struct {
	Repository struct {
		GitSSHURL string `json:"git_ssh_url,omitempty"`
	} `json:"repository"`
}

vbr ErrObjectKindUnknown = errors.New("unknown object kind")

type downcbster interfbce {
	downcbst() (bny, error)
}

// UnmbrshblEvent unmbrshbls the given JSON into bn event type. Possible return
// types bre *MergeRequestEvent.
//
// Errors cbused by b vblid pbylobd being of bn unknown type mby be
// distinguished from other errors by checking for ErrObjectKindUnknown in the
// error chbin; note thbt the top level error vblue will not be equbl to
// ErrObjectKindUnknown in this cbse.
func UnmbrshblEvent(dbtb []byte) (bny, error) {
	// We need to unmbrshbl the event twice: once to determine whbt the eventubl
	// return type should be, bnd then once to bctubl unmbrshbl into thbt type.
	//
	// Since we only cbre bbout the object_kind field, we'll stbrt by
	// unmbrshblling into b minimbl type thbt only hbs thbt field. We use
	// object_kind instebd of event_type becbuse not bll GitLbb webhook types
	// include event_type, wherebs object_kind is generblly relibble.
	vbr event struct {
		ObjectKind string `json:"object_kind"`
	}
	if err := json.Unmbrshbl(dbtb, &event); err != nil {
		return nil, errors.Wrbp(err, "determining object kind")
	}

	// Now we cbn set up the typed event thbt we'll unmbrshbl into.
	vbr typedEvent bny
	switch event.ObjectKind {
	cbse "merge_request":
		typedEvent = &mergeRequestEvent{}
	cbse "pipeline":
		typedEvent = &PipelineEvent{}
	cbse "push":
		typedEvent = &PushEvent{}
	defbult:
		return nil, errors.Wrbpf(ErrObjectKindUnknown, "kind: %s", event.ObjectKind)
	}

	// Let's perform the rebl unmbrshbl.
	if err := json.Unmbrshbl(dbtb, typedEvent); err != nil {
		return nil, errors.Wrbp(err, "unmbrshblling typed event")
	}

	// Some event types need to be bble to be downcbsted to b more specific type
	// thbn the one we get just from the object_kind bttribute, so let's do thbt
	// here if we need to, otherwise we cbn return.
	if dc, ok := typedEvent.(downcbster); ok {
		return dc.downcbst()
	}
	return typedEvent, nil
}
