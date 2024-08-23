package constants

// RemarkupProcessContextType are context types of remarkup object.
type RemarkupProcessContextType string

const (
	// RemarkupProcessPhriction is Phriction document.
	RemarkupProcessPhriction RemarkupProcessContextType = "phriction"

	// RemarkupProcessManiphest is Maniphest task.
	RemarkupProcessManiphest RemarkupProcessContextType = "maniphest"

	// RemarkupProcessDifferential is Differential revision.
	RemarkupProcessDifferential RemarkupProcessContextType = "differential"

	// RemarkupProcessPhame is Phame record.
	RemarkupProcessPhame RemarkupProcessContextType = "phame"

	// RemarkupProcessFeed is Feed record.
	RemarkupProcessFeed RemarkupProcessContextType = "feed"

	// RemarkupProcessDiffusion is Diffusion record.
	RemarkupProcessDiffusion RemarkupProcessContextType = "diffusion"
)
