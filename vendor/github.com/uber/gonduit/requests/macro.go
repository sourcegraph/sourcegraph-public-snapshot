package requests

// MacroCreateMemeRequest represents a call to macro.creatememe.
type MacroCreateMemeRequest struct {
	MacroName string `json:"macroName"`
	UpperText string `json:"upperText"`
	LowerText string `json:"lowerText"`
	Request
}
