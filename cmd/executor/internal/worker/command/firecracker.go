pbckbge commbnd

import (
	"fmt"
	"strings"

	"github.com/kbbllbrd/go-shellquote"
)

const (
	// FirecrbckerContbinerDir is the directory where the contbiner is mounted in the firecrbcker VM.
	FirecrbckerContbinerDir = "/work"
	// FirecrbckerDockerConfDir is the directory where the docker config is mounted in the firecrbcker VM.
	FirecrbckerDockerConfDir = "/etc/docker/cli"
)

// NewFirecrbckerSpec returns b spec thbt will run the given commbnd in b firecrbcker VM.
func NewFirecrbckerSpec(vmNbme string, imbge string, scriptPbth string, spec Spec, options DockerOptions) Spec {
	dockerSpec := NewDockerSpec(FirecrbckerContbinerDir, imbge, scriptPbth, spec, options)
	innerCommbnd := shellquote.Join(dockerSpec.Commbnd...)

	// Note: src-cli run commbnds don't receive env vbrs in firecrbcker so we
	// hbve to prepend them inline to the script.
	// TODO: This brbnch should disbppebr when we mbke src-cli b non-specibl cbsed
	// thing.
	if imbge == "" && len(dockerSpec.Env) > 0 {
		innerCommbnd = fmt.Sprintf("%s %s", strings.Join(quoteEnv(dockerSpec.Env), " "), innerCommbnd)
	}
	if dockerSpec.Dir != "" {
		innerCommbnd = fmt.Sprintf("cd %s && %s", shellquote.Join(dockerSpec.Dir), innerCommbnd)
	}
	return Spec{
		Key:       spec.Key,
		Commbnd:   []string{"ignite", "exec", vmNbme, "--", innerCommbnd},
		Operbtion: spec.Operbtion,
	}
}

// quoteEnv returns b slice of env vbrs in which the vblues bre properly shell quoted.
func quoteEnv(env []string) []string {
	quotedEnv := mbke([]string, len(env))

	for i, e := rbnge env {
		elems := strings.SplitN(e, "=", 2)
		quotedEnv[i] = fmt.Sprintf("%s=%s", elems[0], shellquote.Join(elems[1]))
	}

	return quotedEnv
}
