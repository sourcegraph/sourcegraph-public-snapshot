package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

const recordedCommandMaxLimit = 40

type RecordedCommandsArgs struct {
	Limit  int32
	Offset int32
}

func (r *RepositoryResolver) RecordedCommands(ctx context.Context, args *RecordedCommandsArgs) (graphqlutil.SliceConnectionResolver[RecordedCommandResolver], error) {
	var resolvers []RecordedCommandResolver
	recordingConf := conf.Get().SiteConfig().GitRecorder
	if recordingConf == nil {
		return graphqlutil.NewSliceConnectionResolver(resolvers, 0, 0), nil
	}
	store := rcache.NewFIFOList(wrexec.GetFIFOListKey(r.Name()), recordingConf.Size)
	empty, err := store.IsEmpty()
	if err != nil {
		return nil, err
	}
	if empty {
		return graphqlutil.NewSliceConnectionResolver(resolvers, 0, 0), nil
	}

	offset := int(args.Offset)
	limit := int(args.Limit)
	if limit <= 0 || limit > recordedCommandMaxLimit {
		limit = recordedCommandMaxLimit
	}
	raws, err := store.Slice(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	size, err := store.Size()
	if err != nil {
		return nil, err
	}

	resolvers = make([]RecordedCommandResolver, len(raws))
	for i, raw := range raws {
		command, err := wrexec.UnmarshalCommand(raw)
		if err != nil {
			return nil, err
		}
		resolvers[i] = NewRecordedCommandResolver(command)
	}

	return graphqlutil.NewSliceConnectionResolver(resolvers, size, offset+limit), nil
}

func recordedCommandTransformer(raw []byte) (RecordedCommandResolver, error) {
	command, err := wrexec.UnmarshalCommand(raw)
	if err != nil {
		return nil, err
	}
	return NewRecordedCommandResolver(command), nil
}

type RecordedCommandResolver interface {
	Start() gqlutil.DateTime
	Duration() float64
	Command() string
	Dir() string
	Path() string
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
