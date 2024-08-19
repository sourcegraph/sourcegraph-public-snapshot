// This file contains commands which run in a non daemon mode for testing/debugging.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func debugIndex() *ffcli.Command {
	fs := flag.NewFlagSet("debug index", flag.ExitOnError)
	conf := rootConfig{}
	conf.registerRootFlags(fs)

	return &ffcli.Command{
		Name:       "index",
		ShortUsage: "index [flags] <repository ID>",
		ShortHelp:  "index a repository",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing repository ID")
			}
			s, err := newServer(conf)
			if err != nil {
				return err
			}
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			msg, err := s.forceIndex(uint32(id))
			log.Println(msg)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func debugTrigrams() *ffcli.Command {
	return &ffcli.Command{
		Name:       "trigrams",
		ShortUsage: "trigrams <path/to/shard>",
		ShortHelp:  "list all the trigrams in a shard",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing path to shard")
			}
			return printShardStats(args[0])
		},
	}
}

func debugMeta() *ffcli.Command {
	return &ffcli.Command{
		Name:       "meta",
		ShortUsage: "meta <path/to/shard>",
		ShortHelp:  "output index and repo metadata",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing path to shard")
			}
			return printMetaData(args[0])
		},
	}
}

func debugCmd() *ffcli.Command {
	fs := flag.NewFlagSet("debug", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "debug",
		ShortUsage: "debug <subcommand>",
		ShortHelp:  "a set of commands for debugging and testing",
		LongHelp: `
  Zoekt-sourcegraph-indexserver exposes debug information on the /debug landing page.
  You can use the following wget commands to access this information from the command line.

  wget -q -O - http://localhost:6072/debug/indexed
    list the repositories that are INDEXED by this instance.

  wget -q -O - http://localhost:6072/debug/list[?indexed=TRUE/false]
    list the repositories that are OWNED by this instance. If indexed=true (default), the list may contain repositories
    that this instance holds temporarily, for example during rebalancing.

  wget -q -O - http://localhost:6072/debug/merge
    start a full merge operation in the index directory. You can check the status with
    "wget -q -O - http://localhost:6072/metrics -sS | grep index_shard_merging_running". It is only possible
    to trigger one merge operation at a time.

  wget -q -O - http://localhost:6072/debug/queue
    list the repositories in the indexing queue, sorted by descending priority.

    COLUMN HEADERS
      Position     zero-indexed position of this repository in the indexing queue (sorted by priority).
      Name         name for this repository
      ID           ID for this repository
      IsOnQueue    "true" if this repository has an outstanding indexing job that's enqueued for future work. "false" otherwise.
      Age          amount of time that this repository has spent in the indexing queue since its outstanding indexing job
                   was first added (ignoring any job metadata updates that may have occurred while it was still enqueued).
                   A "-" is printed instead if this repository doesn't have an outstanding job.
      Branches     comma-separated list of branches in $BRANCH_NAME@$COMMIT_HASH format.
                   If the repository has a job on the indexing queue, this list represents the desired set of
                   branches + associated commits that will be process during the next indexing job.
                   However, if the repository  doesn't have a job on the queue, this list represents the set of
                   branches + associated commits that was indexed during its most recent indexing job.`,
		FlagSet: fs,
		Subcommands: []*ffcli.Command{
			debugIndex(),
			debugMeta(),
			debugTrigrams(),
		},
	}
}
