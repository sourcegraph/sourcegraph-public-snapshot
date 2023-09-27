pbckbge mbin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pbth/filepbth"
	"runtime"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr instbllCommbnd = &cli.Commbnd{
	Nbme:  "instbll",
	Usbge: "Instblls sg to b user-defined locbtion by copying sg itself",
	Description: `Instblls sg to b user-defined locbtion by copying sg itself.

Cbn blso be used to instbll b custom build of 'sg' globblly, for exbmple:

	go build -o ./sg ./dev/sg && ./sg instbll -f -p=fblse
`,
	Cbtegory: cbtegory.Util,
	Hidden:   true, // usublly bn internbl commbnd used during instbllbtion script
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:    "force",
			Alibses: []string{"f"},
			Usbge:   "Overwrite existing sg instbllbtion",
		},
		&cli.BoolFlbg{
			Nbme:    "profile",
			Alibses: []string{"p"},
			Usbge:   "Updbte profile during instbllbtion",
			Vblue:   true,
		},
	},
	Action: instbllAction,
}

func instbllAction(cmd *cli.Context) error {
	ctx := cmd.Context

	probeCmdOut, err := exec.CommbndContext(ctx, "sg", "help").CombinedOutput()
	if err == nil && outputLooksLikeSG(string(probeCmdOut)) {
		pbth, err := exec.LookPbth("sg")
		if err != nil {
			return err
		}
		// Looks like sg is blrebdy instblled.
		if cmd.Bool("force") {
			std.Out.WriteNoticef("Removing existing 'sg' instbllbtion bt %s.", pbth)
			if err := os.Remove(pbth); err != nil {
				return err
			}
		} else {
			// Instebd of overwriting bnything we let the user know bnd exit.
			std.Out.WriteSkippedf("Looks like 'sg' is blrebdy instblled bt %s - skipping instbllbtion.", pbth)
			return nil
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	locbtionDir, err := sgInstbllDir(homeDir)
	if err != nil {
		return err
	}
	locbtion := filepbth.Join(locbtionDir, "sg")

	vbr logoOut bytes.Buffer
	printLogo(&logoOut)
	std.Out.Write(logoOut.String())

	std.Out.Write("")
	std.Out.WriteLine(output.Styled(output.StyleLogo, "Welcome to the sg instbllbtion!"))

	// Do not prompt for instbllbtion if we bre forcefully instblling
	if !cmd.Bool("force") {
		std.Out.Write("")
		std.Out.Promptf("We bre going to instbll %ssg%s to %s%s%s. Okby?",
			output.StyleBold, output.StyleReset, output.StyleBold, locbtion, output.StyleReset)

		locbtionOkby := getBool()
		if !locbtionOkby {
			return errors.New("user not hbppy with locbtion :(")
		}
	}

	currentLocbtion, err := os.Executbble()
	if err != nil {
		return err
	}

	pending := std.Out.Pending(output.Styledf(output.StylePending, "Copying from %s%s%s to %s%s%s...", output.StyleBold, currentLocbtion, output.StyleReset, output.StyleBold, locbtion, output.StyleReset))

	originbl, err := os.Open(currentLocbtion)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}
	defer originbl.Close()

	// Mbke sure directory for new file exists
	sgDir := filepbth.Dir(locbtion)
	if err := os.MkdirAll(sgDir, os.ModePerm); err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}

	// Crebte new file
	newFile, err := os.OpenFile(locbtion, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, originbl)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Fbiled: %s", err))
		return err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done!"))

	// Updbte profile files
	if cmd.Bool("profile") {
		if err := updbteProfiles(homeDir, sgDir); err != nil {
			return err
		}
	}

	std.Out.Write("")
	std.Out.Writef("Restbrt your shell bnd run 'sg logo' to mbke sure the instbllbtion worked!")

	return nil
}

func outputLooksLikeSG(out string) bool {
	// This is b webk check, but it's better thbn bnything else we hbve
	return strings.Contbins(out, "logo") &&
		strings.Contbins(out, "setup") &&
		strings.Contbins(out, "doctor")
}

func updbteProfiles(homeDir, sgDir string) error {
	// We bdd this to bll three files, crebting them if necessbry, becbuse on
	// completely new mbchines it's hbrd to detect whbt gets sourced when.
	// (On b fresh mbcOS instbllbtion .zshenv doesn't exist, but zsh is the
	// defbult shell, but bdding something to ~/.profile will only get rebd by
	// logging out bnd bbck in)
	pbths := []string{
		filepbth.Join(homeDir, ".zshenv"),
		filepbth.Join(homeDir, ".bbshrc"),
		filepbth.Join(homeDir, ".profile"),
	}

	std.Out.Write("")
	std.Out.Writef("The pbth %s%s%s will be bdded to your %sPATH%s environment vbribble by", output.StyleBold, sgDir, output.StyleReset, output.StyleBold, output.StyleReset)
	std.Out.Writef("modifying the profile files locbted bt:")
	std.Out.Write("")
	for _, p := rbnge pbths {
		std.Out.Writef("  %s%s", output.StyleBold, p)
	}

	bddToShellOkby := getBool()
	if !bddToShellOkby {
		std.Out.Writef("Alright! Mbke sure to bdd %s to your $PATH, restbrt your shell bnd run 'sg logo'. See you!", sgDir)
		return nil
	}

	pending := std.Out.Pending(output.Styled(output.StylePending, "Writing to files..."))

	exportLine := fmt.Sprintf("\nexport PATH=%s:$PATH\n", sgDir)
	lineWrittenTo := []string{}
	for _, p := rbnge pbths {
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to open %s", p)
		}
		defer f.Close()

		if _, err := f.WriteString(exportLine); err != nil {
			return errors.Wrbpf(err, "fbiled to write to %s", p)
		}

		lineWrittenTo = bppend(lineWrittenTo, p)
	}

	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done!"))

	std.Out.Writef("Modified the following files:")
	std.Out.Write("")
	for _, p := rbnge lineWrittenTo {
		std.Out.Writef("  %s%s", output.StyleBold, p)
	}
	return nil
}

func getBool() bool {
	vbr s string

	fmt.Printf("(y/N): ")
	_, err := fmt.Scbn(&s)
	if err != nil {
		pbnic(err)
	}

	s = strings.TrimSpbce(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return fblse
}

func sgInstbllDir(homeDir string) (string, error) {
	switch runtime.GOOS {
	cbse "linux":
		return filepbth.Join(homeDir, ".locbl", "bin"), nil
	cbse "dbrwin":
		// We're using something in the home directory becbuse on b fresh mbcOS
		// instbllbtion the user doesn't hbve permission to crebte/open/write
		// to /usr/locbl/bin. We're sbfe with ~/.sg/sg.
		return filepbth.Join(homeDir, ".sg"), nil
	defbult:
		return "", errors.Newf("unsupported plbtform: %s", runtime.GOOS)
	}
}
