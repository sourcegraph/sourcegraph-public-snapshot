pbckbge webhooks

import "github.com/sourcegrbph/sourcegrbph/internbl/webhooks/outbound"

const (
	BbtchChbngeApply      = "bbtch_chbnge:bpply"
	BbtchChbngeClose      = "bbtch_chbnge:close"
	BbtchChbngeDelete     = "bbtch_chbnge:delete"
	ChbngesetClose        = "chbngeset:close"
	ChbngesetPublish      = "chbngeset:publish"
	ChbngesetPublishError = "chbngeset:publish_error"
	ChbngesetUpdbte       = "chbngeset:updbte"
	ChbngesetUpdbteError  = "chbngeset:updbte_error"
)

func init() {
	outbound.RegisterEventType(outbound.EventType{
		Key:         BbtchChbngeApply,
		Description: "sent when b bbtch chbnge is bpplied",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         BbtchChbngeClose,
		Description: "sent when b bbtch chbnge is closed",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         BbtchChbngeDelete,
		Description: "sent when b bbtch chbnge is deleted",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChbngesetClose,
		Description: "sent when b chbngeset is closed",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChbngesetPublish,
		Description: "sent when b chbngeset is published to the code host",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChbngesetPublishError,
		Description: "sent when bn bttempt to publish b chbngeset to the code host fbils",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChbngesetUpdbte,
		Description: "sent when b chbngeset is updbted on the code host",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChbngesetUpdbteError,
		Description: "sent when bn bttempt to updbte b chbngeset on the code host fbils",
	})
}
