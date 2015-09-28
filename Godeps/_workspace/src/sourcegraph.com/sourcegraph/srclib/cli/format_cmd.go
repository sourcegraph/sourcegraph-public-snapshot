package cli

import (
	"bytes"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/srclib/toolchain"
)

func init() {
	_, err := CLI.AddCommand("fmt",
		"format an object (def, ref, doc, etc)",
		"The fmt command takes an object and formats it.",
		&fmtCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

var fmtOpt = "fmt"

type FmtCmd struct {
	UnitType   string `short:"u" long:"unit-type" description:"unit type" required:"yes"`
	ObjectType string `short:"t" long:"object-type" description:"Object type ('def', 'doc')" required:"yes"`
	Format     string `short:"f" long:"format" description:"Format to output ('full', 'decl')" default:"full"`

	Object string `long:"object" description:"Object to format, serialized as JSON" required:"yes"`
}

var fmtCmd FmtCmd

func (c *FmtCmd) Get() (string, error) {
	t, err := toolchain.ChooseTool(fmtOpt, c.UnitType)
	if err != nil {
		return "", err
	}
	// Only call as a program for now.
	tool, err := toolchain.OpenTool(t.Toolchain, t.Subcmd, toolchain.AsProgram)
	if err != nil {
		return "", err
	}
	cmd, err := tool.Command()
	if err != nil {
		return "", err
	}
	out := &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Args = append(cmd.Args, "--unit-type", c.UnitType, "--object-type", c.ObjectType, "--format", c.Format, "--object", c.Object)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (c *FmtCmd) Execute(args []string) error {
	out, err := c.Get()
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
