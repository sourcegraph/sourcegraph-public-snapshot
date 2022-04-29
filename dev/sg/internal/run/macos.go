package run

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bitbucket.org/creachadair/shell"
	"github.com/bitfield/script"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// http://www.manpagez.com/man/8/socketfilterfw/
const firewallCmdPath = "/usr/libexec/ApplicationFirewall/socketfilterfw"

var (
	// RLock => adding to firewall (allow multiple), Lock => restarting firewall (allow one)
	firewallMux sync.RWMutex
	// If cannot acquire, then somebody is already restarting the firewall
	firewallRestartSem = semaphore.NewWeighted(1)
)

// addToMacosFirewall returns a callback that is used to add binaries used by the given
// commands to the MacOS firewall.
func addToMacosFirewall(cmds []Command) func() error {
	return func() error {
		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		stdout.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleWarning, "You may be prompted to enter your password to add exceptions to the firewall."))

		// Add binaries in '.bin' to firewall
		for _, cmd := range cmds {
			// Some commands use env variables that may be from command env or global env,
			// so do substitutions and get the binary we want to work with.
			args, ok := shell.Split(os.Expand(cmd.Cmd, func(key string) string {
				if v, exists := cmd.Env[key]; exists {
					return v
				}
				return os.Getenv(key)
			}))
			if !ok || len(args) == 0 {
				stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleSuggestion, "%s: invalid command", cmd.Cmd))
				continue
			}

			for _, arg := range args {
				if strings.HasPrefix(arg, ".bin/") || strings.HasPrefix(arg, "./.bin/") {
					binary := arg
					binaryPath := filepath.Join(root, binary)

					updateMacOSFirewallException(binary, binaryPath)
				}
			}
		}

		// Arbitrary wait to batch updates together from restarts
		time.Sleep(500 * time.Millisecond)

		if !firewallRestartSem.TryAcquire(1) {
			// Somebody else is restarting the firewall, so we're done, but we do another
			// arbitrary wait before proceeding so that the firewall is hopefully updated
			time.Sleep(500 * time.Millisecond)
			return nil
		}
		defer firewallRestartSem.Release(1)

		// Nobody should be updating the firewall now
		firewallMux.Lock()
		defer firewallMux.Unlock()

		stdout.Out.WriteLine(output.Linef(output.EmojiHourglass, output.StylePending, "Restarting firewall..."))

		s, err := script.
			Exec(shell.Join([]string{"sudo", firewallCmdPath, "-k", "--setglobalstate", "off"})).
			String()
		if err != nil {
			return err
		}
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, strings.TrimSpace(s)))

		s, err = script.
			Exec(shell.Join([]string{"sudo", firewallCmdPath, "--setglobalstate", "on"})).
			String()
		if err != nil {
			return err
		}
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, strings.TrimSpace(s)))

		return nil
	}
}

func updateMacOSFirewallException(binary, binaryPath string) {
	firewallMux.RLock()
	defer firewallMux.RUnlock()

	// Remove exception
	removeException := script.Exec(shell.Join([]string{"sudo", firewallCmdPath, "--remove", binaryPath}))
	if _, err := removeException.String(); err != nil {
		stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%s: %s", binary, err.Error()))
	}

	// Add it back
	msg, err := script.Exec(shell.Join([]string{"sudo", firewallCmdPath, "--add", binaryPath})).String()
	if err != nil {
		stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleBold, "%s: %s", binary, err.Error()))
		return
	}
	// socketfilterfw helpfully always returns status 0, so we need to check
	// the output to determine whether things worked or not. In all cases we
	// don't error out becasue we want other commands to go through the firewall
	// updates regardless.
	switch {
	case strings.Contains(msg, "does not exist"):
		stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%s: %s", binary, strings.TrimSpace(msg)))
		return

	case strings.Contains(msg, "added to firewall"):
		stdout.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%s: added to firewall", binary))

	default:
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "%s: %s", binary, strings.TrimSpace(msg)))
	}

	// Also unblock because we need it?
	msg, err = script.Exec(shell.Join([]string{"sudo", firewallCmdPath, "--unblockapp", binaryPath})).String()
	if err != nil {
		stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleBold, "%s: %s", binary, err.Error()))
		return
	}
	stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "%s: %s", binary, strings.TrimSpace(msg)))
}
