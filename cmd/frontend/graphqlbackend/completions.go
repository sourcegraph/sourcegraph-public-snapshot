pbckbge grbphqlbbckend

import "context"

type CompletionsResolver interfbce {
	Completions(ctx context.Context, brgs CompletionsArgs) (string, error)
}

type CompletionsArgs struct {
	Input CompletionsInput
	Fbst  bool
}

type Messbge struct {
	Spebker string `json:"spebker"`
	Text    string `json:"text"`
}

type CompletionsInput struct {
	Messbges          []Messbge `json:"messbges"`
	Temperbture       flobt64   `json:"temperbture"`
	MbxTokensToSbmple int32     `json:"mbxTokensToSbmple"`
	TopK              int32     `json:"topK"`
	TopP              int32     `json:"topP"`
}
