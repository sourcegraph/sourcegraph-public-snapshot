package worker

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// srclibUseDockerExeMethod reports whether Docker should be used to
// run srclib toolchains (if SG_SRCLIB_ENABLE_DOCKER is set and
// `docker info` succeeds).
//
// If true, commands that invoke "src" should use "--methods docker"
// make it run srclib toolchains in Docker containers instead of just
// executing them as local native programs. This execution method can
// add additional security and isolation.
//
// If the env var SG_SRCLIB_REQUIRE_DOCKER is true, then it calls
// log.Fatal unless the Docker daemon is available. This is so you can
// ensure srclib will refuse to run untrusted code (without this
// setting, if the Docker daemon happened to be temporarily
// unavailable or misconfigured, the program would automatically
// switch to running toolchains on the current system, which could be
// undesirable).
func srclibUseDockerExeMethod() bool {
	enableDocker, _ := strconv.ParseBool(os.Getenv("SG_SRCLIB_ENABLE_DOCKER"))
	requireDocker, _ := strconv.ParseBool(os.Getenv("SG_SRCLIB_REQUIRE_DOCKER"))

	if !enableDocker {
		if requireDocker {
			log.Fatal("Docker is required but not enabled (SG_SRCLIB_ENABLE_DOCKER not set).")
		}
		return false
	}

	// If `docker info` fails, the Docker daemon is not installed or
	// not running. If it succeeds, Docker is probably available.
	cmd := exec.Command("docker", "info")
	dockerAvailable := false
	dockerError := cmd.Start()
	if dockerError == nil {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		select {
		case dockerError = <-done:
			if dockerError == nil {
				dockerAvailable = true
			}
		case <-time.After(time.Second):
			// don't wait 30 seconds for timeout of docker client
			dockerError = errors.New("timeout")
			cmd.Process.Kill()
		}
	}

	if requireDocker && !dockerAvailable {
		log.Fatalf("Docker is required but Docker daemon is not accessible (`docker info`: %v).", dockerError)
	}

	return dockerAvailable
}
