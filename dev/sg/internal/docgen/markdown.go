package docgen

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
)

// Markdown renders a Markdown reference for the app.
//
// It is adapted from https://sourcegraph.com/github.com/urfave/cli@v2.4.0/-/blob/docs.go?L16
func Markdown(app *cli.App) (string, error) {
	var w bytes.Buffer
	if err := writeDocTemplate(app, &w); err != nil {
		return "", err
	}
	return w.String(), nil
}

type cliTemplate struct {
	App        *cli.App
	Commands   []string
	GlobalArgs []string
}

func writeDocTemplate(app *cli.App, w io.Writer) error {
	const name = "cli"
	t, err := template.New(name).Parse(markdownDocTemplate)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, &cliTemplate{
		App:        app,
		Commands:   prepareCommands(app.Name, app.Commands, 0),
		GlobalArgs: prepareArgsWithValues(app.VisibleFlags()),
	})
}

func prepareCommands(lineage string, commands []*cli.Command, level int) []string {
	var coms []string
	for _, command := range commands {
		if command.Hidden {
			continue
		}

		var commandDoc strings.Builder
		commandDoc.WriteString(strings.Repeat("#", level+2))
		commandDoc.WriteString(" ")
		commandDoc.WriteString(fmt.Sprintf("%s %s", lineage, command.Name))
		commandDoc.WriteString("\n\n")
		commandDoc.WriteString(prepareUsage(command))
		commandDoc.WriteString("\n\n")

		if len(command.Description) > 0 {
			commandDoc.WriteString(fmt.Sprintf("%s\n\n", command.Description))
		}

		commandDoc.WriteString(prepareUsageText(command))

		flags := prepareArgsWithValues(command.Flags)
		if len(flags) > 0 {
			commandDoc.WriteString("\nFlags:\n\n")
			for _, f := range flags {
				commandDoc.WriteString("* " + f)
			}
		}

		coms = append(coms, commandDoc.String())

		// recursevly iterate subcommands
		if len(command.Subcommands) > 0 {
			coms = append(
				coms,
				prepareCommands(lineage+" "+command.Name, command.Subcommands, level+1)...,
			)
		}
	}

	return coms
}

func prepareArgsWithValues(flags []cli.Flag) []string {
	return prepareFlags(flags, ", ", "`", "`", `"<value>"`, true)
}

func prepareFlags(
	flags []cli.Flag,
	sep, opener, closer, value string,
	addDetails bool,
) []string {
	args := []string{}
	for _, f := range flags {
		flag, ok := f.(cli.DocGenerationFlag)
		if !ok {
			continue
		}
		modifiedArg := opener

		for _, s := range flag.Names() {
			trimmed := strings.TrimSpace(s)
			if len(modifiedArg) > len(opener) {
				modifiedArg += sep
			}
			if len(trimmed) > 1 {
				modifiedArg += fmt.Sprintf("--%s", trimmed)
			} else {
				modifiedArg += fmt.Sprintf("-%s", trimmed)
			}
		}

		if flag.TakesValue() {
			modifiedArg += fmt.Sprintf("=%s", value)
		}

		modifiedArg += closer

		if addDetails {
			modifiedArg += flagDetails(flag)
		}

		args = append(args, modifiedArg+"\n")

	}
	sort.Strings(args)
	return args
}

// flagDetails returns a string containing the flags metadata
func flagDetails(flag cli.DocGenerationFlag) string {
	description := flag.GetUsage()
	value := flag.GetValue()
	if value != "" {
		description += " (default: " + value + ")"
	}
	return ": " + description
}

func prepareUsageText(command *cli.Command) string {
	if command.UsageText == "" {
		if strings.TrimSpace(command.ArgsUsage) != "" {
			return fmt.Sprintf("Arguments: `%s`\n", command.ArgsUsage)
		}
		return ""
	}

	// Write all usage examples as a big shell code block
	var usageText strings.Builder
	usageText.WriteString("```sh")
	for _, line := range strings.Split(strings.TrimSpace(command.UsageText), "\n") {
		usageText.WriteByte('\n')

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			usageText.WriteString(line)
		} else if len(line) > 0 {
			usageText.WriteString(fmt.Sprintf("$ %s", line))
		}
	}
	usageText.WriteString("\n```\n")

	return usageText.String()
}

func prepareUsage(command *cli.Command) string {
	if command.Usage == "" {
		return ""
	}

	return command.Usage + "."
}

var markdownDocTemplate = `# {{ .App.Name }} reference

{{ .App.Name }}{{ if .App.Usage }} - {{ .App.Usage }}{{ end }}
{{ if .App.Description }}
{{ .App.Description }}
{{ end }}
` + "```sh" + `{{ if .App.UsageText }}
{{ .App.UsageText }}
{{ else }}
{{ .App.Name }} [GLOBAL FLAGS] command [COMMAND FLAGS] [ARGUMENTS...]
{{ end }}` + "```" + `
{{ if .GlobalArgs }}
Global flags:

{{ range $v := .GlobalArgs }}* {{ $v }}{{ end }}{{ end }}{{ if .Commands }}
{{ range $v := .Commands }}
{{ $v }}{{ end }}{{ end }}`
