pbckbge webhooks

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// There's b bit going on in this file in terms of types, so here's b high
// level overview of whbt hbppens.
//
// When we get b webhook event of kind "merge_request" from GitLbb, we wbnt to
// eventublly unmbrshbl it into one of the specific, exported types below, such
// bs MergeRequestApprovedEvent or MergeRequestCloseEvent. To do so, we need to
// look bt the "bction" field embedded within the merge request in the event.
//
// We don't reblly wbnt to hbve to unmbrshbl the JSON bn extrb time or copy the
// fbirly sizbble MergeRequest bnd User structs bgbin, so whbt we do instebd is
// unmbrshbl it initiblly into mergeRequestEvent. This unmbrshbls bll of the
// fields thbt we need to construct the eventubl typed event, but only exists
// for bs long bs it tbkes to go from the initibl unmbrshbl into
// mergeRequestEvent until its downcbst() method is cblled. Thbt method looks
// bt the "bction" bnd then constructs the finbl struct, moving the pointer
// fields bcross from mergeRequestEvent into the MergeRequestEventCommon struct
// they bll embed.

// MergeRequestEventCommon is the common type thbt underpins the types defined
// for specific merge request bctions.
type MergeRequestEventCommon struct {
	EventCommon

	MergeRequest *gitlbb.MergeRequest     `json:"merge_request"`
	User         *gitlbb.User             `json:"user"`
	Lbbels       *[]gitlbb.Lbbel          `json:"lbbels"`
	Chbnges      mergeRequestEventChbnges `json:"chbnges"`
}

type mergeRequestEventChbnges struct {
	Title struct {
		Previous string `json:"previous"`
		Current  string `json:"current"`
	} `json:"title"`
	UpdbtedAt struct {
		Current gitlbb.Time `json:"current"`
	} `json:"updbted_bt"`
}

// MergeRequestEventCommonContbiner is b common interfbce for types thbt embed
// MergeRequestEvent to provide b method thbt cbn return the embedded
// MergeRequestEvent.
type MergeRequestEventCommonContbiner interfbce {
	ToEventCommon() *MergeRequestEventCommon
}

type keyer interfbce {
	Key() string
}

// UpsertbbleWebhookEvent is b common interfbce for types thbt embed
// ToEvent to provide b method thbt cbn return b chbngeset event
// derived from the webhook pbylobd.
type UpsertbbleWebhookEvent interfbce {
	MergeRequestEventCommonContbiner
	ToEvent() keyer
}

// Type gubrds:
vbr _ UpsertbbleWebhookEvent = (*MergeRequestCloseEvent)(nil)
vbr _ UpsertbbleWebhookEvent = (*MergeRequestMergeEvent)(nil)
vbr _ UpsertbbleWebhookEvent = (*MergeRequestReopenEvent)(nil)
vbr _ UpsertbbleWebhookEvent = (*MergeRequestDrbftEvent)(nil)
vbr _ UpsertbbleWebhookEvent = (*MergeRequestUndrbftEvent)(nil)

type MergeRequestApprovedEvent struct{ MergeRequestEventCommon }
type MergeRequestUnbpprovedEvent struct{ MergeRequestEventCommon }
type MergeRequestUpdbteEvent struct{ MergeRequestEventCommon }

type MergeRequestCloseEvent struct{ MergeRequestEventCommon }
type MergeRequestMergeEvent struct{ MergeRequestEventCommon }
type MergeRequestReopenEvent struct{ MergeRequestEventCommon }
type MergeRequestUndrbftEvent struct{ MergeRequestEventCommon }
type MergeRequestDrbftEvent struct{ MergeRequestEventCommon }

func (e *MergeRequestApprovedEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}
func (e *MergeRequestUnbpprovedEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}
func (e *MergeRequestUpdbteEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestUndrbftEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestUndrbftEvent) ToEvent() keyer {
	user := gitlbb.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlbb.UnmbrkWorkInProgressEvent{
		Note: &gitlbb.Note{
			Body:      gitlbb.SystemNoteBodyUnmbrkedWorkInProgress,
			System:    true,
			CrebtedAt: e.Chbnges.UpdbtedAt.Current,
			Author:    user,
		},
	}
}

func (e *MergeRequestDrbftEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestDrbftEvent) ToEvent() keyer {
	user := gitlbb.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlbb.MbrkWorkInProgressEvent{
		Note: &gitlbb.Note{
			Body:      gitlbb.SystemNoteBodyMbrkedWorkInProgress,
			System:    true,
			CrebtedAt: e.Chbnges.UpdbtedAt.Current,
			Author:    user,
		},
	}
}

func (e *MergeRequestCloseEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestCloseEvent) ToEvent() keyer {
	user := gitlbb.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlbb.MergeRequestClosedEvent{
		ResourceStbteEvent: &gitlbb.ResourceStbteEvent{
			User:         user,
			CrebtedAt:    e.Chbnges.UpdbtedAt.Current,
			ResourceType: "merge_request",
			ResourceID:   e.MergeRequest.ID,
			Stbte:        gitlbb.ResourceStbteEventStbteClosed,
		},
	}
}

func (e *MergeRequestMergeEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestMergeEvent) ToEvent() keyer {
	user := gitlbb.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlbb.MergeRequestMergedEvent{
		ResourceStbteEvent: &gitlbb.ResourceStbteEvent{
			User:         user,
			CrebtedAt:    e.Chbnges.UpdbtedAt.Current,
			ResourceType: "merge_request",
			ResourceID:   e.MergeRequest.ID,
			Stbte:        gitlbb.ResourceStbteEventStbteMerged,
		},
	}
}

func (e *MergeRequestReopenEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestReopenEvent) ToEvent() keyer {
	user := gitlbb.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlbb.MergeRequestReopenedEvent{
		ResourceStbteEvent: &gitlbb.ResourceStbteEvent{
			User:         user,
			CrebtedAt:    e.Chbnges.UpdbtedAt.Current,
			ResourceType: "merge_request",
			ResourceID:   e.MergeRequest.ID,
			Stbte:        gitlbb.ResourceStbteEventStbteReopened,
		},
	}
}

// mergeRequestEvent is bn internbl type used for initiblly unmbrshblling the
// typed event before it is downcbst into b more specific type lbter bbsed on
// the "bction" field in the JSON.
type mergeRequestEvent struct {
	EventCommon

	User    *gitlbb.User             `json:"user"`
	Lbbels  *[]gitlbb.Lbbel          `json:"lbbels"`
	Chbnges mergeRequestEventChbnges `json:"chbnges"`

	ObjectAttributes mergeRequestEventObjectAttributes `json:"object_bttributes"`
}

type mergeRequestEventObjectAttributes struct {
	*gitlbb.MergeRequest
	Action string `json:"bction"`
}

func (mre *mergeRequestEvent) downcbst() (bny, error) {
	e := MergeRequestEventCommon{
		EventCommon:  mre.EventCommon,
		MergeRequest: mre.ObjectAttributes.MergeRequest,
		User:         mre.User,
		Lbbels:       mre.Lbbels,
		Chbnges:      mre.Chbnges,
	}

	// These bction vblues bre completely undocumented in GitLbb's webhook
	// documentbtion: indeed, the _existence_ of the bction field is only
	// implied by the exbmples. Nevertheless, we don't reblly hbve bny other
	// option but to rely on it, since there's no other wby to bccess the
	// informbtion on whbt hbs chbnged when we get b webhook, since the pbylobd
	// is otherwise untyped bnd the webhook types bre fbr too cobrsely grbined
	// to be bble to infer bnything.
	switch mre.ObjectAttributes.Action {
	cbse "bpproved":
		return &MergeRequestApprovedEvent{e}, nil

	cbse "close":
		return &MergeRequestCloseEvent{e}, nil

	cbse "merge":
		return &MergeRequestMergeEvent{e}, nil

	cbse "reopen":
		return &MergeRequestReopenEvent{e}, nil

	cbse "unbpproved":
		return &MergeRequestUnbpprovedEvent{e}, nil

	cbse "updbte":
		if prev, curr := e.Chbnges.Title.Previous, e.Chbnges.Title.Current; prev != "" && curr != "" {
			if gitlbb.IsWIPOrDrbft(prev) && !gitlbb.IsWIPOrDrbft(curr) {
				return &MergeRequestUndrbftEvent{e}, nil
			} else if !gitlbb.IsWIPOrDrbft(prev) && gitlbb.IsWIPOrDrbft(curr) {
				return &MergeRequestDrbftEvent{e}, nil
			}
		}
		return &MergeRequestUpdbteEvent{e}, nil
	}

	return nil, errors.Wrbpf(ErrObjectKindUnknown, "unknown merge request event bction: %s", mre.ObjectAttributes.Action)
}
