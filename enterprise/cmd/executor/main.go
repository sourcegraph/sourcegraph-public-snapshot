package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/coreos/go-iptables/iptables"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	// This import is required to force a binary hash change when the src-cli version is bumped.
	_ "github.com/sourcegraph/sourcegraph/internal/src-cli"
)

func main() {
	config := &Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()

	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("executor", "the executor service polls the public frontend API for work to perform")

	// TODO: Make executor a CLI program.
	// Add commands to:
	// Validate the installation, including the config, CNI, ignite setup etc.
	// Prepare the installation, by getting the required ignite modules etc. This might replace a good
	// chunk of the install.sh file we currently got for the executor VMs.
	// Run a debug VM. This is meant to replace `ignite run XXX`, which in our executor
	// images currently relies on an ignite config file, the fact that iptables are preconfigured
	// and the cni config to be written to disk, otherwise the VM that ignite
	// gives us is very different from the one that the executor runs.
	// We currently cannot ask a customer to "simply spawn a VM and tell us if it works",
	// this command shall change that, and potentially print additional debug info
	// in the future. TODO: Make sure this supports creating a loop device that is
	// mounted as well to ensure that this also works correctly.
	// This command might also be used to verify the executor images before they're released.

	app := &cli.App{
		Version: version.Version(),
		Commands: []*cli.Command{
			{
				Name: "prepare",
				Action: func(ctx *cli.Context) error {
					// if err := config.Validate(); err != nil {
					// 	return errors.Wrap(err, "failed to validate config")
					// }
					return prepareFirecracker(ctx.Context, logger, config.FirecrackerOptions())
				},
			},
		},
		Name:  "boom",
		Usage: "make an explosive entrance",
		Action: func(*cli.Context) error {
			fmt.Println("boom! I say!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	// TODO: validate docker is installed.
	// TODO: validate git is installed.
	// TODO: validate src-cli is installed and a good version, helper for that at the end of the file..

	if err := config.Validate(); err != nil {
		logger.Error("failed to read config", log.Error(err))
		os.Exit(1)
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log.Scoped("service", "executor service"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Determine telemetry data.
	telemetryOptions := func() apiclient.TelemetryOptions {
		// Run for at most 5s to get telemetry options.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return apiclient.NewTelemetryOptions(ctx, config.UseFirecracker)
	}()
	logger.Info("Telemetry information gathered", log.String("info", fmt.Sprintf("%+v", telemetryOptions)))

	gatherer := metrics.MakeExecutorMetricsGatherer(log.Scoped("executor-worker.metrics-gatherer", ""), prometheus.DefaultGatherer, options.NodeExporterEndpoint, options.DockerRegistryNodeExporterEndpoint)
	queueStore := apiclient.New(options.ClientOptions, gatherer, observationContext)

	nameSet := janitor.NewNameSet()
	ctx, cancel := context.WithCancel(context.Background())
	worker := worker.NewWorker(nameSet, queueStore, config.APIWorkerOptions(telemetryOptions), observationContext)

	routines := []goroutine.BackgroundRoutine{
		worker,
	}

	if config.UseFirecracker {
		routines = append(routines, janitor.NewOrphanedVMJanitor(
			config.VMPrefix,
			nameSet,
			config.CleanupTaskInterval,
			janitor.NewMetrics(observationContext),
		))

		mustRegisterVMCountMetric(observationContext, config.VMPrefix)

		// If this causes harm, we can disable it.
		if _, ok := os.LookupEnv("EXECUTOR_SKIP_FIRECRACKER_SETUP"); !ok {
			if err := prepareFirecracker(ctx, logger, config.FirecrackerOptions()); err != nil {
				logger.Error("failed to prepare firecracker environment", log.Error(err))
				os.Exit(1)
			}
		}
	}

	go func() {
		// Block until the worker has exited. The executor worker is unique
		// in that we want a maximum runtime and/or number of jobs to be
		// executed by a single instance, after which the service should shut
		// down without error.
		worker.Wait()

		// Once the worker has finished its current set of jobs and stops
		// the dequeue loop, we want to finish off the rest of the sibling
		// routines so that the service can shut down.
		cancel()
	}()

	goroutine.MonitorBackgroundRoutines(ctx, routines...)
}

func makeWorkerMetrics(queueName string) workerutil.WorkerMetrics {
	observationContext := &observation.Context{
		Logger:     log.Scoped("executor_processor", "executor worker processor"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return workerutil.NewMetrics(observationContext, "executor_processor",
		// derived from historic data, ideally we will use spare high-res histograms once they're a reality
		// 										 30s 1m	 2.5m 5m   7.5m 10m  15m  20m	30m	  45m	1hr
		workerutil.WithDurationBuckets([]float64{30, 60, 150, 300, 450, 600, 900, 1200, 1800, 2700, 3600}),
		workerutil.WithLabels(map[string]string{
			"queue": queueName,
		}),
	)
}

func mustRegisterVMCountMetric(observationContext *observation.Context, prefix string) {
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_executor_vms_total",
		Help: "Total number of running VMs.",
	}, func() float64 {
		runningVMsByName, err := ignite.ActiveVMsByName(context.Background(), prefix, false)
		if err != nil {
			log15.Error("Failed to determine number of running VMs", "error", err)
		}

		return float64(len(runningVMsByName))
	}))
}

// prepareFirecracker makes sure all resources required to run firecracker VMs exist.
// If they do, this function is a noop. Otherwise, it will start pulling and importing
// images.
func prepareFirecracker(ctx context.Context, logger log.Logger, c command.FirecrackerOptions) error {
	// Make sure the image exists. When ignite imports these at runtime, there can
	// be a race condition and it is imported multiple times. Also, this would
	// happen for the first job, which is not desirable.
	logger.Info("Ensuring VM image is imported", log.String("image", c.Image))
	cmd := exec.CommandContext(ctx, "ignite", "image", "import", "--runtime", "docker", c.Image)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "importing ignite VM base image: %s", out)
	}

	// Make sure the kernel image exists.
	cmd = exec.CommandContext(ctx, "ignite", "kernel", "import", "--runtime", "docker", c.KernelImage)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "importing ignite kernel: %s", out)
	}

	// Also preload the runtime image.
	cmd = exec.CommandContext(ctx, "docker", "pull", c.SandboxImage)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "importing ignite isolation image: %s", out)
	}

	cniSubnetCIDR := "10.61.0.0/16"
	_, net, err := net.ParseCIDR(cniSubnetCIDR)
	if err != nil {
		return err
	}
	// TODO: Use net below instead of hard coded CIDRs.

	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return err
	}

	// Ensure the chain exists.
	if ok, err := ipt.ChainExists("filter", "CNI-ADMIN"); err != nil {
		return err
	} else if !ok {
		if err := ipt.NewChain("filter", "CNI-ADMIN"); err != nil {
			return err
		}
	}

	// Explicitly allow DNS traffic (currently, the DNS server lives in the private
	// networks for GCP and AWS. Ideally we'd want to use an internet-only DNS server
	// to prevent leaking any network details).
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-p udp --dport 53 -j ACCEPT"); err != nil {
		return err
	}

	// Disallow any host-VM network traffic from the guests, except connections made
	// FROM the host (to ssh into the guest).
	if err := ipt.AppendUnique("filter", "INPUT", "-d 10.61.0.0/16 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "INPUT", "-s 10.61.0.0/16 -j DROP"); err != nil {
		return err
	}

	// Disallow any inter-VM traffic.
	// But allow to reach the gateway for internet access.
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-s 10.61.0.1/32 -d 10.61.0.0/16 -j ACCEPT"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-d 10.61.0.0/16 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-s 10.61.0.0/16 -d 10.61.0.0/16 -j DROP"); err != nil {
		return err
	}

	// Disallow local networks access.
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-s 10.61.0.0/16 -d 10.0.0.0/8 -p tcp -j DROP"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-s 10.61.0.0/16 -d 192.168.0.0/16 -p tcp -j DROP"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-s 10.61.0.0/16 -d 172.16.0.0/12 -p tcp -j DROP"); err != nil {
		return err
	}
	// Disallow link-local traffic, too. This usually contains cloud provider
	// resources that we don't want to expose.
	if err := ipt.AppendUnique("filter", "CNI-ADMIN", "-s 10.61.0.0/16 -d 169.254.0.0/16 -j DROP"); err != nil {
		return err
	}

	return nil
}

// validateSrcCLIVersion queries the latest recommended version of src-cli and makes sure it
// matches what is installed. If not, a warning message recommending to use a different
// version is logged.
func validateSrcCLIVersion(ctx context.Context, logger log.Logger, client *apiclient.Client) {
	latestVersion, err := client.LatestSrcCLIVersion(ctx)
	if err != nil {
		logger.Error("cannot retrieve latest compatible src-cli version", log.Error(err))
		return
	}
	cmd := exec.CommandContext(ctx, "src", "version", "-client-only")
	out, err := cmd.Output()
	if err != nil {
		logger.Error("failed to get src-cli version, is it installed?", log.Error(err))
	}
	actualVersion := string(out)
	actualVersion = strings.TrimSpace(actualVersion)
	actualVersion = strings.TrimPrefix(actualVersion, "Current version: ")
	if version.IsDev(actualVersion) {
		return
	}
	actual, err := semver.NewVersion(actualVersion)
	if err != nil {
		logger.Error("failed to parse src-cli version", log.Error(err))
	}
	latest, err := semver.NewVersion(latestVersion)
	if err != nil {
		logger.Error("failed to parse latest src-cli version", log.Error(err))
	}
	if actual.LessThan(latest) {
		logger.Warn("installed src-cli is not the latest recommended version, consider upgrading", log.String("actual", actual.String()), log.String("latest", latest.String()))
	} else if actual.Major() != latest.Major() || actual.Minor() != latest.Minor() {
		logger.Warn("installed src-cli is not the latest recommended version, consider switching", log.String("actual", actual.String()), log.String("recommended", latest.String()))
	}
}
