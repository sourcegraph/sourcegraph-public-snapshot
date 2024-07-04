package main

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/dev/bazel-execlog/proto"
)

func main() {
	old, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer old.Close()

	new, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer new.Close()

	oldTargetSpawns := make(Map)
	newTargetSpawns := make(Map)

	fmt.Println("------------------ OLD ------------------")
	r := NewSpawnLogReconstructor(old)
	for {
		exec, err := r.GetSpawnExec()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		oldTargetSpawns[exec.TargetLabel] = append(oldTargetSpawns[exec.TargetLabel], exec)
	}

	fmt.Println("------------------ NEW ------------------")
	r = NewSpawnLogReconstructor(new)
	for {
		exec, err := r.GetSpawnExec()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		newTargetSpawns[exec.TargetLabel] = append(newTargetSpawns[exec.TargetLabel], exec)
	}

	// newOnly := newTargetSpawns.Minus(oldTargetSpawns)
	// oldOnly := oldTargetSpawns.Minus(newTargetSpawns)

	opts := cmpopts.IgnoreUnexported(proto.SpawnExec{}, proto.EnvironmentVariable{}, proto.Platform{}, proto.Platform_Property{}, proto.File{}, proto.Digest{}, proto.SpawnMetrics{}, durationpb.Duration{}, timestamppb.Timestamp{})

	// for _, v := range newOnly {
	// 	fmt.Println(cmp.Diff(v, nil, opts))
	// }
	// for _, v := range oldOnly {
	// 	fmt.Println(cmp.Diff(nil, v, opts))
	// }

	oldShared, newShared := oldTargetSpawns.Intersection(newTargetSpawns)
	// cmp.Comparer(func(t1, t2 map[string][]*proto.SpawnExec) bool {
	// 	return false
	// })

	// fmt.Println(len(oldTargetSpawns), len(newTargetSpawns))

	reporter := &DiffReporter{}
	if diff := cmp.Diff(oldShared, newShared, opts, cmp.Reporter(reporter)); diff != "" {
		// fmt.Println(diff)
		fmt.Println(reporter.String())
	}
}
