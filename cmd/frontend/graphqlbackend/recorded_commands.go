package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const recordedCommandDefaultLimit = 40

func (r *RepositoryResolver) RecordedCommands(ctx context.Context, args *struct {
	Limit  *int32
	Offset *int32
}) (graphqlutil.SliceConnectionResolver[[]byte, RecordedCommandResolver], error) {
	recordingConf := conf.Get().SiteConfig().GitRecorder
	if recordingConf == nil {
		return graphqlutil.NewSliceConnectionResolver([][]byte{}, 0, 0, recordedCommandTransformer), nil
	}
	store := rcache.NewFIFOList(wrexec.GetFIFOListKey(r.Name()), recordingConf.Size)
	empty, err := store.IsEmpty()
	if err != nil {
		return nil, err
	}
	if empty {
		return graphqlutil.NewSliceConnectionResolver([][]byte{}, 0, 0, recordedCommandTransformer), nil
	}
	raws, err := store.All(ctx)
	if err != nil {
		return nil, err
	}

	if args.Limit == nil || *args.Limit > recordedCommandDefaultLimit {
		args.Limit = pointers.Ptr(int32(recordedCommandDefaultLimit))
	}

	if args.Offset == nil {
		args.Offset = pointers.Ptr(int32(0))
	}

	return graphqlutil.NewSliceConnectionResolver(raws, int(*args.Limit), int(*args.Offset), recordedCommandTransformer), nil
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
