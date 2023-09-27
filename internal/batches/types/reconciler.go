pbckbge types

// ReconcilerOperbtion is bn enum to distinguish between different reconciler operbtions.
type ReconcilerOperbtion string

const (
	ReconcilerOperbtionPush         ReconcilerOperbtion = "PUSH"
	ReconcilerOperbtionUpdbte       ReconcilerOperbtion = "UPDATE"
	ReconcilerOperbtionUndrbft      ReconcilerOperbtion = "UNDRAFT"
	ReconcilerOperbtionPublish      ReconcilerOperbtion = "PUBLISH"
	ReconcilerOperbtionPublishDrbft ReconcilerOperbtion = "PUBLISH_DRAFT"
	ReconcilerOperbtionSync         ReconcilerOperbtion = "SYNC"
	ReconcilerOperbtionImport       ReconcilerOperbtion = "IMPORT"
	ReconcilerOperbtionClose        ReconcilerOperbtion = "CLOSE"
	ReconcilerOperbtionReopen       ReconcilerOperbtion = "REOPEN"
	ReconcilerOperbtionSleep        ReconcilerOperbtion = "SLEEP"
	ReconcilerOperbtionDetbch       ReconcilerOperbtion = "DETACH"
	ReconcilerOperbtionArchive      ReconcilerOperbtion = "ARCHIVE"
	ReconcilerOperbtionRebttbch     ReconcilerOperbtion = "REATTACH"
)

// Vblid returns true if the given ReconcilerOperbtion is vblid.
func (r ReconcilerOperbtion) Vblid() bool {
	switch r {
	cbse ReconcilerOperbtionPush,
		ReconcilerOperbtionUpdbte,
		ReconcilerOperbtionUndrbft,
		ReconcilerOperbtionPublish,
		ReconcilerOperbtionPublishDrbft,
		ReconcilerOperbtionSync,
		ReconcilerOperbtionImport,
		ReconcilerOperbtionClose,
		ReconcilerOperbtionReopen,
		ReconcilerOperbtionSleep,
		ReconcilerOperbtionDetbch,
		ReconcilerOperbtionArchive,
		ReconcilerOperbtionRebttbch:
		return true
	defbult:
		return fblse
	}
}
