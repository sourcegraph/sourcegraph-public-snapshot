package run

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bitbucket.org/creachadair/shell"
	"github.com/bitfield/script"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var firewallMux sync.RWMutex

// addToMacosFirewall returns a callback that is used to add binaries used by the given
// commands to the MacOS firewall.
func addToMacosFirewall(cmds []Command) func() error {
	return func() error {
		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		stdout.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleWarning, "You may be prompted to enter your password to add exceptions to the firewall."))

		// http://www.manpagez.com/man/8/socketfilterfw/
		firewallCmdPath := "/usr/libexec/ApplicationFirewall/socketfilterfw"

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

					// add it back
					firewallMux.RLock()
					addException := script.
						Exec(shell.Join([]string{"sudo", firewallCmdPath, "--add", binaryPath})).
						FilterLine(func(msg string) string {
							// socketfilterfw helpfully always returns status 0, so we need to check
							// the output to determine whether things worked or not. In all cases we
							// don't error out becasue we want other commands to go through the firewall
							// updates regardless.
							switch {
							case strings.Contains(msg, "does not exist"):
								stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, "%s: %s", binary, strings.TrimSpace(msg)))

							case strings.Contains(msg, "added to firewall"):
								stdout.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%s: added to firewall", binary))

							default:
								stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "%s: %s", binary, strings.TrimSpace(msg)))
							}
							return ""
						}).
						Exec(shell.Join([]string{"sudo", firewallCmdPath, "--unblockapp", binaryPath})).
						FilterLine(func(msg string) string {
							stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "%s: %s", binary, strings.TrimSpace(msg)))
							return ""
						})
					addException.Wait()
					firewallMux.RUnlock()

					// Check for errors
					if err := addException.Error(); err != nil {
						stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleBold, "%s: %s", binary, err.Error()))
						continue
					}

				}
			}
		}

		firewallMux.Lock()
		defer firewallMux.Unlock()
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Restarting firewall..."))

		s, err := script.
			Exec(shell.Join([]string{"sudo", firewallCmdPath, "-k", "--setglobalstate", "off"})).
			String()
		if err != nil {
			return err
		}
		time.Sleep(3 * time.Second)
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, strings.TrimSpace(s)))

		s, err = script.
			Exec(shell.Join([]string{"sudo", firewallCmdPath, "--setglobalstate", "on"})).
			String()
		if err != nil {
			return err
		}
		time.Sleep(3 * time.Second)
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, strings.TrimSpace(s)))

		return nil
	}
}
