package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type LogMessage struct {
	Body_ string
}

func (v LogMessage) Body() string { return v.Body_ }

type LogMessageInput struct {
	Body string `json:"body"`
}

func FromLogMessageInput(input ...LogMessageInput) []LogMessage {
	out := make([]LogMessage, len(input))
	for i, in := range input {
		out[i] = LogMessage{Body_: in.Body}
	}
	return out
}

type LogMessageConnectionArgs struct {
	graphqlutil.ConnectionArgs
}

type hasLogMessages interface {
	LogMessages(context.Context, *LogMessageConnectionArgs) (LogMessageConnection, error)
}

// LogMessageConnection implements the LogMessageConnection GraphQL type.
type LogMessageConnection []LogMessage

func (c LogMessageConnection) Nodes(context.Context) ([]LogMessage, error) {
	return []LogMessage(c), nil
}

func (c LogMessageConnection) TotalCount(context.Context) (int32, error) { return int32(len(c)), nil }

func (c LogMessageConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
