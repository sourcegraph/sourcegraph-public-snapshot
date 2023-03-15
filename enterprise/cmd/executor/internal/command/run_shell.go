//go:build shell

package command

func init() {
	allowedBinaries = append(allowedBinaries, "/bin/sh")
}
