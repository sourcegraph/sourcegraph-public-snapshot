package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/src-cli/internal/exec"
)

func init() {
	usage := `
'src debug kube' invokes kubectl diagnostic commands targeting kubectl's current-context, writing returns to an archive.

Usage:

	src debug kube [command options]

Flags:

	-o				Specify the name of the output zip archive.
	-n				Specify the namespace passed to kubectl commands. If not specified the 'default' namespace is used.
	--no-config		Don't include Sourcegraph configuration json.

Examples:

	$ src debug kube -o debug.zip

	$ src -v debug kube -n ns-sourcegraph -o foo

	$ src debug kube -no-configs -o bar.zip

`

	flagSet := flag.NewFlagSet("kube", flag.ExitOnError)
	var base string
	var namespace string
	var excludeConfigs bool
	flagSet.StringVar(&base, "o", "debug.zip", "The name of the output zip archive")
	flagSet.StringVar(&namespace, "n", "default", "The namespace passed to kubectl commands, if not specified the 'default' namespace is used")
	flagSet.BoolVar(&excludeConfigs, "no-configs", false, "If true, exclude Sourcegraph configuration files. Defaults to false.")

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		// process -o flag to get zipfile and base directory names
		if base == "" {
			return errors.New("empty -o flag")
		}
		base, baseDir := processBaseDir(base)

		// init context
		ctx := context.Background()
		// open pipe to output file
		out, err := os.OpenFile(base, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0666)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %q", base)
		}
		defer out.Close()
		// init zip writer
		zw := zip.NewWriter(out)
		defer zw.Close()

		// Gather data for safety check
		pods, err := selectPods(ctx, namespace)
		if err != nil {
			return errors.Wrap(err, "failed to get pods")
		}
		kubectx, err := exec.CommandContext(ctx, "kubectl", "config", "current-context").CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "failed to get current-context")
		}
		// Safety check user knows what they've targeted with this command
		log.Printf("Archiving kubectl data for %d pods\n SRC_ENDPOINT: %v\n Context: %s Namespace: %v\n Output filename: %v", len(pods.Items), cfg.Endpoint, kubectx, namespace, base)
		if verified, _ := verify("Do you want to start writing to an archive?"); !verified {
			return nil
		}

		err = archiveKube(ctx, zw, *verbose, !excludeConfigs, namespace, baseDir, pods)
		if err != nil {
			return err
		}
		return nil
	}

	debugCommands = append(debugCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

type podList struct {
	Items []struct {
		Metadata struct {
			Name string
		}
		Spec struct {
			Containers []struct {
				Name string
			}
		}
	}
}

// Runs common kubectl functions and archive results to zip file
func archiveKube(ctx context.Context, zw *zip.Writer, verbose, archiveConfigs bool, namespace, baseDir string, pods podList) error {
	// Create a context with a cancel function that we call when returning
	// from archiveKube. This ensures we close all pending go-routines when returning
	// early because of an error.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// setup channel for slice of archive function outputs, as well as throttling semaphore
	ch := make(chan *archiveFile)
	g, ctx := errgroup.WithContext(ctx)
	semaphore := semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))

	run := func(f func() *archiveFile) {
		g.Go(func() error {
			if err := semaphore.Acquire(ctx, 1); err != nil {
				return err
			}
			defer semaphore.Release(1)

			if file := f(); file != nil {
				ch <- file
			}

			return nil
		})
	}

	// create goroutine to get pods
	run(func() *archiveFile { return getPods(ctx, namespace, baseDir) })

	// create goroutine to get kubectl events
	run(func() *archiveFile { return getEvents(ctx, namespace, baseDir) })

	// create goroutine to get persistent volumes
	run(func() *archiveFile { return getPV(ctx, namespace, baseDir) })

	// create goroutine to get persistent volumes claim
	run(func() *archiveFile { return getPVC(ctx, namespace, baseDir) })

	// start goroutine to run kubectl logs for each pod's container's
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			p := pod.Metadata.Name
			c := container.Name
			run(func() *archiveFile { return getPodLog(ctx, p, c, namespace, baseDir) })
		}
	}

	// start goroutine to run kubectl logs --previous for each pod's container's
	// won't write to zip on err, only passes bytes to channel if err not nil
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			p := pod.Metadata.Name
			c := container.Name
			run(func() *archiveFile {
				f := getPastPodLog(ctx, p, c, namespace, baseDir)
				if f.err != nil {
					if verbose {
						fmt.Printf("Could not gather --previous pod logs for %s\n", p)
					}
					return nil
				}
				return f
			})
		}
	}

	// start goroutine for each pod to run kubectl describe pod
	for _, pod := range pods.Items {
		p := pod.Metadata.Name
		run(func() *archiveFile { return getDescribe(ctx, p, namespace, baseDir) })
	}

	// start goroutine for each pod to run kubectl get pod <pod> -o yaml
	for _, pod := range pods.Items {
		p := pod.Metadata.Name
		run(func() *archiveFile { return getManifest(ctx, p, namespace, baseDir) })
	}

	// start goroutine to get external service config
	if archiveConfigs {
		run(func() *archiveFile { return getExternalServicesConfig(ctx, baseDir) })

		run(func() *archiveFile { return getSiteConfig(ctx, baseDir) })
	}

	// close channel when wait group goroutines have completed
	go func() {
		if err := g.Wait(); err != nil {
			fmt.Printf("archiveKube failed to open wait group: %s\n", err)
			os.Exit(1)
		}
		close(ch)
	}()

	// Read binaries from channel and write to archive on host machine
	if err := writeChannelContentsToZip(zw, ch, verbose); err != nil {
		return errors.Wrap(err, "failed to write archives from channel")
	}

	return nil
}

// Calls kubectl get pods and returns a list of pods to be processed
func selectPods(ctx context.Context, namespace string) (podList, error) {
	// Declare buffer type var for kubectl pipe
	var podsBuff bytes.Buffer

	// Get all pod names as json
	podsCmd := exec.CommandContext(
		ctx,
		"kubectl", "-n", namespace, "get", "pods", "-l", "deploy=sourcegraph", "-o=json",
	)
	podsCmd.Stdout = &podsBuff
	podsCmd.Stderr = os.Stderr
	err := podsCmd.Run()
	if err != nil {
		errors.Wrap(err, "failed to aquire pods for subcommands with err")
	}

	//Decode json from podList
	var pods podList
	if err := json.Unmarshal(podsBuff.Bytes(), &pods); err != nil {
		return pods, errors.Wrap(err, "failed to unmarshall get pods json")
	}

	return pods, err
}

// runs archiveFileFromCommand with arg get pods
func getPods(ctx context.Context, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "getPods.txt"),
		"kubectl", "-n", namespace, "get", "pods", "-o", "wide",
	)
}

// runs archiveFileFromCommand with arg get events
func getEvents(ctx context.Context, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "events.txt"),
		"kubectl", "-n", namespace, "get", "events",
	)
}

// runs archiveFileFromCommand with arg get pv
func getPV(ctx context.Context, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "persistent-volumes.txt"),
		"kubectl", "-n", namespace, "get", "pv",
	)
}

// runs archiveFileFromCommand with arg get pvc
func getPVC(ctx context.Context, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "persistent-volume-claims.txt"),
		"kubectl", "-n", namespace, "get", "pvc",
	)
}

// runs archiveFileFromCommand with arg logs $POD -c $CONTAINER
func getPodLog(ctx context.Context, podName, containerName, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "pods", podName, fmt.Sprintf("%s.log", containerName)),
		"kubectl", "-n", namespace, "logs", podName, "-c", containerName,
	)
}

// runs archiveFileFromCommand with arg logs --previous $POD -c $CONTAINER
func getPastPodLog(ctx context.Context, podName, containerName, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "pods", podName, fmt.Sprintf("prev-%s.log", containerName)),
		"kubectl", "-n", namespace, "logs", "--previous", podName, "-c", containerName,
	)
}

// runs archiveFileFromCommand with arg describe pod $POD
func getDescribe(ctx context.Context, podName, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "pods", podName, fmt.Sprintf("describe-%s.txt", podName)),
		"kubectl", "-n", namespace, "describe", "pod", podName,
	)
}

// runs archiveFileFromCommand with arg get pod $POD -o yaml
func getManifest(ctx context.Context, podName, namespace, baseDir string) *archiveFile {
	return archiveFileFromCommand(
		ctx,
		filepath.Join(baseDir, "kubectl", "pods", podName, fmt.Sprintf("manifest-%s.yaml", podName)),
		"kubectl", "-n", namespace, "get", "pod", podName, "-o", "yaml",
	)
}
