package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

// recordedCommandMaxLimit is the maximum number of recorded commands that can be
// returned in a single query. This limit prevents returning an excessive number of
// recorded commands. It should always be in sync with the default in `cmd/frontend/graphqlbackend/schema.graphql`
const recordedCommandMaxLimit = 40

var MockGetRecordedCommandMaxLimit func() int

func GetRecordedCommandMaxLimit() int {
	if MockGetRecordedCommandMaxLimit != nil {
		return MockGetRecordedCommandMaxLimit()
	}
	return recordedCommandMaxLimit
}

type RecordedCommandsArgs struct {
	Limit  int32
	Offset int32
}

func (r *RepositoryResolver) RecordedCommands(ctx context.Context, args *RecordedCommandsArgs) (gqlutil.SliceConnectionResolver[RecordedCommandResolver], error) {
	// 🚨 SECURITY: Only site admins are allowed to view recorded commands
	err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	if err != nil {
		return nil, err
	}

	offset := int(args.Offset)
	limit := int(args.Limit)
	maxLimit := GetRecordedCommandMaxLimit()
	if limit == 0 || limit > maxLimit {
		limit = maxLimit
	}
	currentEnd := offset + limit

	recordingConf := conf.Get().SiteConfig().GitRecorder
	if recordingConf == nil {
		return gqlutil.NewSliceConnectionResolver([]RecordedCommandResolver{}, 0, currentEnd), nil
	}
	store := rcache.NewFIFOList(redispool.Cache, wrexec.GetFIFOListKey(r.Name()), recordingConf.Size)
	empty, err := store.IsEmpty()
	if err != nil {
		return nil, err
	}
	if empty {
		return gqlutil.NewSliceConnectionResolver([]RecordedCommandResolver{}, 0, currentEnd), nil
	}

	// the FIFO list is zero-indexed, so we need to deduct one from the limit
	// to be able to get the correct amount of data.
	to := currentEnd - 1
	raws, err := store.Slice(ctx, offset, to)
	if err != nil {
		return nil, err
	}

	size, err := store.Size()
	if err != nil {
		return nil, err
	}

	resolvers := make([]RecordedCommandResolver, len(raws))
	for i, raw := range raws {
		command, err := wrexec.UnmarshalCommand(raw)
		if err != nil {
			return nil, err
		}
		resolvers[i] = NewRecordedCommandResolver(command)
	}

	return gqlutil.NewSliceConnectionResolver(resolvers, size, currentEnd), nil
}

type RecordedCommandResolver interface {
	Start() gqlutil.DateTime
	Duration() float64
	Command() string
	Dir() string
	Path() string
	Output() string
	IsSuccess() bool
}

type recordedCommandResolver struct {
	command wrexec.RecordedCommand
}

func NewRecordedCommandResolver(command wrexec.RecordedCommand) RecordedCommandResolver {
	return &recordedCommandResolver{command: command}
}

func (r *recordedCommandResolver) Start() gqlutil.DateTime {
	return *gqlutil.FromTime(r.command.Start)
}

func (r *recordedCommandResolver) Duration() float64 {
	return r.command.Duration
}

func (r *recordedCommandResolver) Command() string {
	return strings.Join(r.command.Args, " ")
}

func (r *recordedCommandResolver) Dir() string {
	return r.command.Dir
}

func (r *recordedCommandResolver) Path() string {
	return r.command.Path
}

func (r *recordedCommandResolver) Output() string {
	return r.command.Output
}

func (r *recordedCommandResolver) IsSuccess() bool {
	return r.command.IsSuccess
}

func (r *RepositoryResolver) IsRecordingEnabled() bool {
	recordingConf := conf.Get().SiteConfig().GitRecorder
	if recordingConf != nil && len(recordingConf.Repos) > 0 {
		if recordingConf.Repos[0] == "*" {
			return true
		}

		for _, repo := range recordingConf.Repos {
			if strings.EqualFold(repo, r.Name()) {
				return true
			}
		}
	}
	return false
}
