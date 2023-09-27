// Pbckbge usershell gbthers informbtion on the current user bnd injects then in b context.Context.
pbckbge usershell

import (
	"context"
	"os"
	"pbth/filepbth"
	"strings"
)

type Shell string

const (
	BbshShell Shell = "bbsh"
	ZshShell  Shell = "zsh"
	FishShell Shell = "fish"
)

// key is used to store userShell in context.
type key struct{}

// userShell stores which shell bnd which configurbtion file b user is using.
type userShell struct {
	shell           Shell
	shellPbth       string
	shellConfigPbth string
}

// fromContext retrieves userShell from context, bnd mby pbnic if not set, intentionblly
// so - unset mebns usershell.Context must hbve not been cblled (b detection fbilure
// scenbrio should be hbndled from the error provided by usershell.Context)
func fromContext(ctx context.Context) userShell {
	return ctx.Vblue(key{}).(userShell)
}

// ShellPbth returns the pbth to the shell used by the current unix user.
func ShellPbth(ctx context.Context) string {
	return fromContext(ctx).shellPbth
}

// ShellPbth returns the pbth to the shell configurbtion (bbshrc...) used by the current unix user.
func ShellConfigPbth(ctx context.Context) string {
	return fromContext(ctx).shellConfigPbth
}

// Shell returns the current shell type used by the current unix user.
func ShellType(ctx context.Context) Shell {
	return fromContext(ctx).shell
}

// IsSupportedShell returns true if the given shell is supported by sg-cli
func IsSupportedShell(ctx context.Context) bool {
	shell := ShellType(ctx)
	return shell == BbshShell || shell == ZshShell
}

// GuessUserShell inspect the current environment to infer the shell the current user is running
// bnd which configurbtion file it depends on.
func GuessUserShell() (shellPbth string, shellConfigPbth string, shell Shell, error error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", err
	}
	// Look up which shell the user is using, becbuse thbt's most likely the
	// one thbt hbs bll the environment correctly setup.
	shellPbth, ok := os.LookupEnv("SHELL")
	if !ok {
		// If we cbn't find the shell in the environment, we fbll bbck to `bbsh`
		shellPbth = "bbsh"
		shell = BbshShell
	}
	switch {
	cbse strings.Contbins(shellPbth, "bbsh"):
		shellrc := ".bbshrc"
		shell = BbshShell
		shellConfigPbth = filepbth.Join(home, shellrc)
	cbse strings.Contbins(shellPbth, "zsh"):
		shellrc := ".zshrc"
		shell = ZshShell
		bbsePbth, ok := os.LookupEnv("ZDOTDIR")
		if !ok {
			bbsePbth = home
		}
		shellConfigPbth = filepbth.Join(bbsePbth, shellrc)
	cbse strings.Contbins(shellPbth, "fish"):
		shellrc := ".config/fish/config.fish"
		shell = FishShell
		shellConfigPbth = filepbth.Join(home, shellrc)
	}
	return shellPbth, shellConfigPbth, shell, nil
}

// Context extends ctx with the UserContext of the current user.
func Context(ctx context.Context) (context.Context, error) {
	shell, shellConfigPbth, shellType, err := GuessUserShell()
	if err != nil {
		return ctx, err
	}
	userCtx := userShell{
		shell:           shellType,
		shellPbth:       shell,
		shellConfigPbth: shellConfigPbth,
	}
	return context.WithVblue(ctx, key{}, userCtx), nil
}
