package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	flagSet := flag.NewFlagSet("doc", flag.ExitOnError)
	var (
		outputFlag = flagSet.String("o", "", "Output directory")
	)

	usage := `
'src doc' is an internal command that generates the Markdown reference
documentation used within Sourcegraph.

Usage:

	src doc -o DIR
	
Examples:

    $ src doc -o ~/sourcegraph/doc/integration/cli/reference
	`

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		if *outputFlag == "" {
			out.WriteLine(output.Line(output.EmojiFailure, output.StyleWarning, "output directory must be set via -o"))
			flagSet.Usage()
			return cmderrors.ExitCode(1, nil)
		}

		dr, err := newDocRenderer()
		if err != nil {
			return err
		}

		commanders := map[string]*commander{
			"":             &commands,
			"batch":        &batchCommands,
			"config":       &configCommands,
			"extensions":   &extensionsCommands,
			"extsvc":       &extsvcCommands,
			"code-intel":   &codeintelCommands,
			"orgs":         &orgsCommands,
			"orgs members": &orgsMembersCommands,
			"repos":        &reposCommands,
			"users":        &usersCommands,
		}

		pending := out.Pending(output.Line("", output.StylePending, "Rendering Markdown..."))
		count := 0
		defer func() {
			pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "%d files rendered under %s", count, *outputFlag))
		}()
		for groupName, cmdr := range commanders {
			subcommands := map[string]string{}

			for _, cmd := range *cmdr {
				name := cmd.flagSet.Name()

				// Figure out the fully qualified name of this command.
				fqcn := strings.TrimSpace(strings.Join([]string{
					groupName,
					name,
				}, " "))

				if fqcn == "doc" || fqcn == "publish" {
					continue
				}
				pending.Update(fqcn)

				// If this name appears in our commanders map, then this isn't a
				// real command, and we'll handle it differently.
				if _, ok := commanders[fqcn]; !ok {
					content, err := dr.RenderCommand(fqcn, cmd)
					if err != nil {
						return err
					}

					file, err := openDocFile(*outputFlag, fqcn)
					if err != nil {
						return err
					}
					defer file.Close()

					if _, err := file.WriteString(content); err != nil {
						return err
					}
					count++
					subcommands[name] = name + ".md"
				} else {
					subcommands[name] = name + "/index.md"
				}
			}

			content, err := dr.RenderGroup(groupName, subcommands)
			if err != nil {
				return err
			}

			file, err := openDocFile(*outputFlag, groupName+" index")
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := file.WriteString(content); err != nil {
				return err
			}
			count++
		}

		return nil
	}

	commands = append(commands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintln(flag.CommandLine.Output(), usage)
		},
	})
}

func openDocFile(base, fqcn string) (interface {
	io.StringWriter
	io.WriteCloser
}, error) {
	cmdPath := path.Join(append([]string{base}, strings.Split(fqcn, " ")...)...)
	if err := os.MkdirAll(path.Dir(cmdPath), 0755); err != nil {
		return nil, err
	}

	return os.Create(cmdPath + ".md")
}

type docRenderer struct {
	commandTemplate *template.Template
	groupTemplate   *template.Template
}

func newDocRenderer() (*docRenderer, error) {
	ct := template.New("command")
	ct = ct.Funcs(template.FuncMap{
		"sanitise": func(v string) (string, error) {
			// Just replacing the home directory will probably cover off most
			// cases, but we'll do the user cache and config as well just to be
			// safe.
			cache, err := os.UserCacheDir()
			if err != nil {
				return "", errors.Wrap(err, "getting user cache dir")
			}

			config, err := os.UserConfigDir()
			if err != nil {
				return "", errors.Wrap(err, "getting user config dir")
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return "", errors.Wrap(err, "getting home dir")
			}

			v = strings.ReplaceAll(v, cache, "~/.cache")
			v = strings.ReplaceAll(v, config, "~/.config")
			v = strings.ReplaceAll(v, home, "~")

			return v, nil
		},
	})

	ct, err := ct.Parse(docCommandTemplate)
	if err != nil {
		return nil, err
	}

	gt, err := template.New("group").Parse(docGroupTemplate)
	if err != nil {
		return nil, err
	}

	return &docRenderer{
		commandTemplate: ct,
		groupTemplate:   gt,
	}, nil
}

func (dr *docRenderer) RenderCommand(fqcn string, cmd *command) (string, error) {
	flags := []*flag.Flag{}
	cmd.flagSet.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f)
	})

	// The usage functions are, unfortunately, a bit of a mess right now. Some
	// output to flag.CommandLine.Output(), some to os.Stderr, some to
	// os.Stdout. So let's replace the stdout and stderr variables temporarily
	// to capture whatever output we get, then we can put everything back after.
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = w
	os.Stderr = w
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()
	out := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, r); err != nil {
			// This shouldn't ever panic in normal operation.
			panic(err)
		}
		out <- buf.String()
	}()
	if cmd.usageFunc != nil {
		cmd.usageFunc()
	} else if cmd.flagSet.Usage != nil {
		cmd.flagSet.Usage()
	}
	w.Close()
	usage := <-out

	buf := &bytes.Buffer{}
	if err := dr.commandTemplate.Execute(buf, &docCommandContext{
		FQCN:  fqcn,
		Usage: usage,
		Flags: flags,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (dr *docRenderer) RenderGroup(fqcn string, commands map[string]string) (string, error) {
	buf := &bytes.Buffer{}
	if err := dr.groupTemplate.Execute(buf, &docGroupContext{
		FQCN:     fqcn,
		Commands: commands,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type docCommandContext struct {
	FQCN  string
	Usage string
	Flags []*flag.Flag
}

type docGroupContext struct {
	FQCN     string
	Commands map[string]string
}

const (
	docCommandTemplate = `# ` + "`" + `src {{ .FQCN }}` + "`" + `

{{ if .Flags }}
## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
{{- range .Flags -}}
{{- "\n" -}}
| ` + "`" + `-{{ .Name }}` + "`" + ` | {{ .Usage }} | {{ if .DefValue }}` + "`" + `{{ sanitise .DefValue }}` + "`" + `{{ end }} |
{{- end }}
{{ end }}

## Usage

` + "```" + `
{{ sanitise .Usage }}
` + "```" + `
	`

	docGroupTemplate = `# ` + "`" + `src{{ if .FQCN }} {{ .FQCN }}{{ end }}` + "`" + `

## Subcommands

{{ range $name, $link := .Commands -}}
{{- "\n" -}}
* [` + "`" + `{{ $name }}` + "`" + `]({{ $link }})
{{- end }}
	`
)
