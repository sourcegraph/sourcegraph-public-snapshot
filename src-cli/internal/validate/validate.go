package validate

var (
	EmojiFingerPointRight = "👉"
	FailureEmoji          = "🛑"
	FlashingLightEmoji    = "🚨"
	HourglassEmoji        = "⌛"
	SuccessEmoji          = "✅"
	WarningSign           = "⚠️ " // why does this need an extra space to align?!?!
)

type Status string

const (
	Failure Status = "Failure"
	Warning Status = "Warning"
	Success Status = "Success"
)

type Result struct {
	Status  Status
	Message string
}
