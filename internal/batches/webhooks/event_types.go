package webhooks

import "github.com/sourcegraph/sourcegraph/internal/webhooks/outbound"

const (
	BatchChangeApply      = "batch_change:apply"
	BatchChangeClose      = "batch_change:close"
	BatchChangeDelete     = "batch_change:delete"
	ChangesetClose        = "changeset:close"
	ChangesetPublish      = "changeset:publish"
	ChangesetPublishError = "changeset:publish_error"
	ChangesetUpdate       = "changeset:update"
	ChangesetUpdateError  = "changeset:update_error"
)

func init() {
	outbound.RegisterEventType(outbound.EventType{
		Key:         BatchChangeApply,
		Description: "sent when a batch change is applied",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         BatchChangeClose,
		Description: "sent when a batch change is closed",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         BatchChangeDelete,
		Description: "sent when a batch change is deleted",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChangesetClose,
		Description: "sent when a changeset is closed",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChangesetPublish,
		Description: "sent when a changeset is published to the code host",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChangesetPublishError,
		Description: "sent when an attempt to publish a changeset to the code host fails",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChangesetUpdate,
		Description: "sent when a changeset is updated on the code host",
	})

	outbound.RegisterEventType(outbound.EventType{
		Key:         ChangesetUpdateError,
		Description: "sent when an attempt to update a changeset on the code host fails",
	})
}
