package shared

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
)

func maybePostgresProcFile() (string, error) {
	// PG is already configured
	if os.Getenv("PGHOST") != "" || os.Getenv("PGDATASOURCE") != "" {
		return "", nil
	}

	// Postgres needs to be able to write to run
	var output bytes.Buffer
	e := execer{Out: &output}
	e.Command("mkdir", "-p", "/run/postgresql")
	e.Command("chown", "-R", "postgres", "/run/postgresql")
	if err := e.Error(); err != nil {
		log.Printf("Setting up postgres failed:\n%s", output.String())
		return "", err
	}

	dataDir := os.Getenv("DATA_DIR")
	path := filepath.Join(dataDir, "postgresql")

	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		if verbose {
			log.Printf("Setting up PostgreSQL at %s", path)
		}
		log.Println("✱ Sourcegraph is initializing the internal database... (may take 15-20 seconds)")

		var output bytes.Buffer
		e := execer{Out: &output}
		e.Command("mkdir", "-p", path)
		e.Command("chown", "postgres", path)
		// initdb --nosync saves ~3-15s on macOS during initial startup. By the time actual data lives in the
		// DB, the OS should have had time to fsync.
		e.Command("su-exec", "postgres", "initdb", "-D", path, "--nosync")
		e.Command("su-exec", "postgres", "pg_ctl", "-D", path, "-o -c listen_addresses=127.0.0.1", "-l", "/tmp/pgsql.log", "-w", "start")
		e.Command("su-exec", "postgres", "createdb", "sourcegraph")
		e.Command("su-exec", "postgres", "pg_ctl", "-D", path, "-m", "fast", "-l", "/tmp/pgsql.log", "-w", "stop")
		if err := e.Error(); err != nil {
			log.Printf("Setting up postgres failed:\n%s", output.String())
			os.RemoveAll(path)
			return "", err
		}
	} else {
		// Between restarts the owner of the volume may have changed. Ensure
		// postgres can still read it.
		var output bytes.Buffer
		e := execer{Out: &output}
		e.Command("chown", "-R", "postgres", path)
		if err := e.Error(); err != nil {
			log.Printf("Adjusting fs owners for postgres failed:\n%s", output.String())
			return "", err
		}

		// Version of Postgres we are running. Note that this is different
		// from our *Minimum Supported Version* which dictates the features
		// we can depend on.
		// https://github.com/sourcegraph/sourcegraph/blob/master/doc/dev/postgresql.md#version-requirements
		const version = "11"

		if err = maybeUpgradePostgres(path, version); err != nil {
			return "", err
		}
	}

	// Set PGHOST to default to 127.0.0.1, NOT localhost, as localhost does not correctly resolve in some environments
	// (see https://github.com/sourcegraph/issues/issues/34 and https://github.com/sourcegraph/sourcegraph/issues/9129).
	SetDefaultEnv("PGHOST", "127.0.0.1")
	SetDefaultEnv("PGUSER", "postgres")
	SetDefaultEnv("PGDATABASE", "sourcegraph")
	SetDefaultEnv("PGSSLMODE", "disable")

	return "postgres: su-exec postgres sh -c 'postgres -c listen_addresses=127.0.0.1 -D " + path + "' 2>&1 | grep -v 'database system was shut down' | grep -v 'MultiXact member wraparound' | grep -v 'database system is ready' | grep -v 'autovacuum launcher started' | grep -v 'the database system is starting up'", nil
}

// maybeUpgradePostgres upgrades the Postgres data files in path to the given version
// if they're not already upgraded. It requires access to the host's Docker socket.
func maybeUpgradePostgres(path, newVersion string) error {
	dataDir := filepath.Dir(path)

	bs, err := ioutil.ReadFile(filepath.Join(path, "PG_VERSION"))
	if err != nil {
		return errors.Wrap(err, "failed to detect version of existing Postgres data")
	}
	oldVersion := strings.TrimSpace(string(bs))

	id, err := containerID()
	if err != nil {
		return errors.Wrap(err, "failed to determine running container id")
	}

	// Use a fairly old Docker version for maximum compatibility.
	cli, err := docker.NewClientWithOpts(client.FromEnv, client.WithVersion("1.28"))
	if err != nil {
		return errors.Wrap(err, "failed to initialise docker client")
	}

	ctx := context.Background()
	hostDataDir, err := hostMountPoint(ctx, cli, id, dataDir)
	if err != nil {
		return errors.Wrap(err, "failed to determine host mount point")
	}

	upgradeDir := filepath.Join(dataDir, fmt.Sprintf("postgres-%s-upgrade", newVersion))
	hostUpgradeDir := filepath.Join(hostDataDir, filepath.Base(upgradeDir))
	statusFile := filepath.Join(upgradeDir, "status")

	if err := os.MkdirAll(upgradeDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create upgrade dir")
	}

	status, statusVersion, err := readStatus(statusFile)
	if err != nil {
		return errors.Wrap(err, "failed to read status file")
	}

	// e.g: ~/.sourcegraph/data/postgresql
	hostPath := filepath.Join(hostDataDir, filepath.Base(path))
	hostPathNew := hostPath + "-" + newVersion

	if status == "started" {
		log.Printf("✱ Sourcegraph was previously interrupted while upgrading its internal database.")
		log.Printf("✱ To try again, start the container after running these commands (safe):\n")

		if oldVersion == newVersion {
			hostPathOld := hostPath + "-" + statusVersion
			log.Printf(
				"$ mv %s %s\n$ mv %s %s\n$ rm -rf %s",
				hostPath, hostPathNew+".$(date +%s).bak",
				hostPathOld, hostPath,
				hostUpgradeDir,
			)
		} else {
			log.Printf(
				"$ mv %s %s\n$ rm -rf %s",
				hostPathNew, hostPathNew+".$(date +%s).bak",
				hostUpgradeDir,
			)
		}
		return errors.New("Interrupted internal database upgrade detected")
	}

	// Nothing to do, already upgraded.
	if oldVersion == newVersion {
		return nil
	}

	log.Printf("✱ Sourcegraph is upgrading its internal database. Please don't interrupt this operation.")
	defer log.Printf("✱ Sourcegraph finished upgrading its internal database.")

	if err = writeStatus(statusFile, "started", oldVersion); err != nil {
		return err
	}

	err = upgradePostgres(ctx, cli, upgradeParams{
		oldVersion:     oldVersion,
		newVersion:     newVersion,
		upgradeDir:     upgradeDir,
		hostUpgradeDir: hostUpgradeDir,
		path:           path,
		hostPath:       hostPath,
		hostPathNew:    hostPathNew,
	})

	if err != nil {
		return err
	}

	return writeStatus(statusFile, "done", oldVersion)
}

type upgradeParams struct {
	oldVersion     string // e.g. 9.6
	newVersion     string // e.g. 11
	upgradeDir     string // e.g. /var/opt/sourcegraph/postgres-11-upgrade
	hostUpgradeDir string // e.g. ~/.sourcegraph/data/postgres-11-upgrade
	path           string // e.g. /var/opt/sourcegraph/postgresql
	hostPath       string // e.g. ~/.sourcegraph/data/postgresql
	hostPathNew    string // e.g. ~/.sourcegraph/data/postgresql-11
}

// upgradePostgres upgrades the existing Postgres data in ps.path to be compatible
// with ps.newVersion.
func upgradePostgres(ctx context.Context, cli *docker.Client, ps upgradeParams) error {
	var output io.Writer
	if verbose {
		output = os.Stdout
	} else {
		output = &bytes.Buffer{}
	}

	img := fmt.Sprintf("tianon/postgres-upgrade:%s-to-%s", ps.oldVersion, ps.newVersion)

	out, err := cli.ImagePull(ctx, img, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to pull %q", img)
	}

	if _, err := io.Copy(output, out); err != nil {
		return errors.Wrap(err, "failed to read output of docker pull")
	}

	config := container.Config{Image: img, WorkingDir: "/tmp/upgrade"}
	hostConfig := container.HostConfig{
		Binds: []string{
			// The *.sql and *.sh scripts generated by pg_upgrade will be stored in this directory
			// so that we can access them in the current container when running /postgres-optimize.sh
			// after pg_upgrade finished.
			fmt.Sprintf("%s:%s", ps.hostUpgradeDir, config.WorkingDir),
			fmt.Sprintf("%s:/var/lib/postgresql/%s/data", ps.hostPath, ps.oldVersion),
			fmt.Sprintf("%s:/var/lib/postgresql/%s/data", ps.hostPathNew, ps.newVersion),
		},
	}

	now := time.Now()
	name := fmt.Sprintf("sourcegraph-postgres-upgrade-%d", now.Unix())
	resp, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, name)
	if err != nil {
		return errors.Wrapf(err, "failed to create %q", name)
	}

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "failed to start %q", name)
	}

	statusch, errch := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errch:
		if err != nil {
			return errors.Wrap(err, "failed to upgrade postgres")
		}
	case <-statusch:
	}

	out, err = cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve %q logs", name)
	}

	if _, err = stdcopy.StdCopy(output, output, out); err != nil {
		return errors.Wrap(err, "failed to copy logs to output")
	}

	// Run the /postgres-optimize.sh in the same dir as the *.sql and *.sh scripts
	// left behind by pg_upgrade.
	e := execer{Out: output, Dir: ps.upgradeDir}
	pathNew := ps.path + "-" + ps.newVersion
	pathOld := ps.path + "-" + ps.oldVersion
	e.Command("mv", ps.path, pathOld)
	e.Command("mv", pathNew, ps.path)
	e.Command("chown", "-R", "postgres", ps.path)
	e.Command("su-exec", "postgres", "/postgres-optimize.sh", ps.path)

	if err := e.Error(); err != nil {
		if b, ok := output.(*bytes.Buffer); ok && !verbose {
			log.Print(b.String())
		}
		return errors.Wrap(err, "postgres upgrade failed after running pg_upgrade")
	}

	return nil
}

// hostMountPoint finds the Docker host mountpoint corresponding to the given path
// in the container with the given id, if any.
func hostMountPoint(ctx context.Context, cli *docker.Client, id, path string) (string, error) {
	c, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", errors.Wrapf(err, "failed to inspect container %q", id)
	}

	for _, bind := range c.HostConfig.Binds {
		if ps := strings.SplitN(bind, ":", 2); len(ps) == 2 && ps[1] == path {
			return ps[0], nil
		}
	}

	for _, mount := range c.Mounts {
		if mount.Destination == path {
			return mount.Source, nil
		}
	}

	return "", fmt.Errorf("couldn't find host mountpoint of %q on container %q", path, id)
}

// containerID retrieves the Docker container id of the running container
func containerID() (string, error) {
	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", errors.Wrap(err, "failed to read /proc/self/cgroup to determine container id")
	}
	defer f.Close()

	r := bufio.NewReader(f)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", errors.Wrap(err, "failed to read first line of /proc/self/cgroup")
	}

	// e.g. 11:hugetlb:/docker/ed70f86d8e5cb2e94975d29d0185c90dd56621c05444e5d7ae0891f290255ce9
	ps := strings.SplitN(line, "/", 3)
	if len(ps) != 3 {
		return "", errors.New("failed to parse /proc/self/cgroup")
	}

	return strings.TrimSpace(ps[2]), nil
}

func writeStatus(path, status, oldVersion string) error {
	if err := ioutil.WriteFile(path, []byte("from:"+oldVersion+":"+status), 0755); err != nil {
		return errors.Wrap(err, "failed to write status file")
	}
	return nil
}

func readStatus(path string) (status, version string, err error) {
	bs, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", "", errors.Wrap(err, "failed to read status file")
	}

	ps := strings.SplitN(string(bytes.TrimSpace(bs)), ":", 3)
	if len(ps) != 3 {
		return "", "", nil
	}

	return ps[2], ps[1], nil
}
