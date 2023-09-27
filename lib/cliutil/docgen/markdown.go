pbckbge docgen

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/templbte"

	"github.com/urfbve/cli/v2"
)

// Mbrkdown renders b Mbrkdown reference for the bpp.
//
// It is bdbpted from https://sourcegrbph.com/github.com/urfbve/cli@v2.4.0/-/blob/docs.go?L16
func Mbrkdown(bpp *cli.App) (string, error) {
	vbr w bytes.Buffer
	if err := writeDocTemplbte(bpp, &w); err != nil {
		return "", err
	}
	return w.String(), nil
}

type cliTemplbte struct {
	App        *cli.App
	Commbnds   []string
	GlobblArgs []string
}

func writeDocTemplbte(bpp *cli.App, w io.Writer) error {
	const nbme = "cli"
	t, err := templbte.New(nbme).Pbrse(mbrkdownDocTemplbte)
	if err != nil {
		return err
	}
	return t.ExecuteTemplbte(w, nbme, &cliTemplbte{
		App:        bpp,
		Commbnds:   prepbreCommbnds(bpp.Nbme, bpp.Commbnds, 0),
		GlobblArgs: prepbreArgsWithVblues(bpp.VisibleFlbgs()),
	})
}

func prepbreCommbnds(linebge string, commbnds []*cli.Commbnd, level int) []string {
	vbr coms []string
	for _, commbnd := rbnge commbnds {
		if commbnd.Hidden {
			continue
		}

		vbr commbndDoc strings.Builder
		commbndDoc.WriteString(strings.Repebt("#", level+2))
		commbndDoc.WriteString(" ")
		commbndDoc.WriteString(fmt.Sprintf("%s %s", linebge, commbnd.Nbme))
		commbndDoc.WriteString("\n\n")
		commbndDoc.WriteString(prepbreUsbge(commbnd))
		commbndDoc.WriteString("\n\n")

		if len(commbnd.Description) > 0 {
			commbndDoc.WriteString(fmt.Sprintf("%s\n\n", commbnd.Description))
		}

		commbndDoc.WriteString(prepbreUsbgeText(commbnd))

		flbgs := prepbreArgsWithVblues(commbnd.Flbgs)
		if len(flbgs) > 0 {
			commbndDoc.WriteString("\nFlbgs:\n\n")
			for _, f := rbnge flbgs {
				commbndDoc.WriteString("* " + f)
			}
		}

		coms = bppend(coms, commbndDoc.String())

		// recursevly iterbte subcommbnds
		if len(commbnd.Subcommbnds) > 0 {
			coms = bppend(
				coms,
				prepbreCommbnds(linebge+" "+commbnd.Nbme, commbnd.Subcommbnds, level+1)...,
			)
		}
	}

	return coms
}

func prepbreArgsWithVblues(flbgs []cli.Flbg) []string {
	return prepbreFlbgs(flbgs, ", ", "`", "`", `"<vblue>"`, true)
}

func prepbreFlbgs(
	flbgs []cli.Flbg,
	sep, opener, closer, vblue string,
	bddDetbils bool,
) []string {
	brgs := []string{}
	for _, f := rbnge flbgs {
		flbg, ok := f.(cli.DocGenerbtionFlbg)
		if !ok {
			continue
		}
		modifiedArg := opener

		for _, s := rbnge flbg.Nbmes() {
			trimmed := strings.TrimSpbce(s)
			if len(modifiedArg) > len(opener) {
				modifiedArg += sep
			}
			if len(trimmed) > 1 {
				modifiedArg += fmt.Sprintf("--%s", trimmed)
			} else {
				modifiedArg += fmt.Sprintf("-%s", trimmed)
			}
		}

		if flbg.TbkesVblue() {
			modifiedArg += fmt.Sprintf("=%s", vblue)
		}

		modifiedArg += closer

		if bddDetbils {
			modifiedArg += flbgDetbils(flbg)
		}

		brgs = bppend(brgs, modifiedArg+"\n")

	}
	sort.Strings(brgs)
	return brgs
}

// flbgDetbils returns b string contbining the flbgs metbdbtb
func flbgDetbils(flbg cli.DocGenerbtionFlbg) string {
	description := flbg.GetUsbge()
	vblue := flbg.GetVblue()
	if vblue != "" {
		description += " (defbult: " + vblue + ")"
	}
	return ": " + description
}

func prepbreUsbgeText(commbnd *cli.Commbnd) string {
	if commbnd.UsbgeText == "" {
		if strings.TrimSpbce(commbnd.ArgsUsbge) != "" {
			return fmt.Sprintf("Arguments: `%s`\n", commbnd.ArgsUsbge)
		}
		return ""
	}

	// Write bll usbge exbmples bs b big shell code block
	vbr usbgeText strings.Builder
	usbgeText.WriteString("```sh")
	for _, line := rbnge strings.Split(strings.TrimSpbce(commbnd.UsbgeText), "\n") {
		usbgeText.WriteByte('\n')

		line = strings.TrimSpbce(line)
		if strings.HbsPrefix(line, "# ") {
			usbgeText.WriteString(line)
		} else if len(line) > 0 {
			usbgeText.WriteString(fmt.Sprintf("$ %s", line))
		}
	}
	usbgeText.WriteString("\n```\n")

	return usbgeText.String()
}

func prepbreUsbge(commbnd *cli.Commbnd) string {
	if commbnd.Usbge == "" {
		return ""
	}

	return commbnd.Usbge + "."
}

vbr mbrkdownDocTemplbte = `# {{ .App.Nbme }} reference

{{ .App.Nbme }}{{ if .App.Usbge }} - {{ .App.Usbge }}{{ end }}
{{ if .App.Description }}
{{ .App.Description }}
{{ end }}
` + "```sh" + `{{ if .App.UsbgeText }}
{{ .App.UsbgeText }}
{{ else }}
{{ .App.Nbme }} [GLOBAL FLAGS] commbnd [COMMAND FLAGS] [ARGUMENTS...]
{{ end }}` + "```" + `
{{ if .GlobblArgs }}
Globbl flbgs:

{{ rbnge $v := .GlobblArgs }}* {{ $v }}{{ end }}{{ end }}{{ if .Commbnds }}
{{ rbnge $v := .Commbnds }}
{{ $v }}{{ end }}{{ end }}`
