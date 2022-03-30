package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	funkyLogoFlagSet = flag.NewFlagSet("sg logo", flag.ExitOnError)
	funkyLogoCommand = &ffcli.Command{
		Name:       "logo",
		ShortUsage: "sg logo [classic]",
		ShortHelp:  "Print the sg logo",
		LongHelp:   "Prints the sg logo in different colors. When the 'classic' argument is passed it prints the classic logo.",
		FlagSet:    funkyLogoFlagSet,
		Exec:       logoExec,
	}
)

var styleOrange = output.Fg256Color(202)

func printLogo(out io.Writer) {
	fmt.Fprintf(out, "%s", output.StyleLogo)
	fmt.Fprintln(out, `          _____                    _____`)
	fmt.Fprintln(out, `         /\    \                  /\    \`)
	fmt.Fprintf(out, `        /%s::%s\    \                /%s::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `       /%s::::%s\    \              /%s::::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `      /%s::::::%s\    \            /%s::::::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `     /%s:::%s/\%s:::%s\    \          /%s:::%s/\%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `    /%s:::%s/__\%s:::%s\    \        /%s:::%s/  \%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `    \%s:::%s\   \%s:::%s\    \      /%s:::%s/    \%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `  ___\%s:::%s\   \%s:::%s\    \    /%s:::%s/    / \%s:::%s\    \`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, ` /\   \%s:::%s\   \%s:::%s\    \  /%s:::%s/    /   \%s:::%s\ ___\`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `/%s::%s\   \%s:::%s\   \%s:::%s\____\/%s:::%s/____/  ___\%s:::%s|    |`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `\%s:::%s\   \%s:::%s\   \%s::%s/    /\%s:::%s\    \ /\  /%s:::%s|____|`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, ` \%s:::%s\   \%s:::%s\   \/____/  \%s:::%s\    /%s::%s\ \%s::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `  \%s:::%s\   \%s:::%s\    \       \%s:::%s\   \%s:::%s\ \/____/`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `   \%s:::%s\   \%s:::%s\____\       \%s:::%s\   \%s:::%s\____\`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `    \%s:::%s\  /%s:::%s/    /        \%s:::%s\  /%s:::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `     \%s:::%s\/%s:::%s/    /          \%s:::%s\/%s:::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `      \%s::::::%s/    /            \%s::::::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `       \%s::::%s/    /              \%s::::%s/    /`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintf(out, `        \%s::%s/    /                \%s::%s/____/`, styleOrange, output.StyleLogo, styleOrange, output.StyleLogo)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `         \/____/`)
	fmt.Fprintf(out, "%s", output.StyleReset)
}

func logoExec(ctx context.Context, args []string) error {
	if len(args) == 1 && args[0] == "classic" {
		var logoOut bytes.Buffer
		printLogo(&logoOut)
		stdout.Out.Write(logoOut.String())
		return nil
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	randoColor := func() output.Style { return output.Fg256Color(r1.Intn(256)) }

	var (
		color1a = randoColor()
		color1b = randoColor()
		color1c = randoColor()
		color2  = output.StyleLogo
	)

	times := 20
	for i := 0; i < times; i++ {
		const linesPrinted = 23

		stdout.Out.Writef("%s", color2)
		stdout.Out.Write(`          _____                    _____`)
		stdout.Out.Write(`         /\    \                  /\    \`)
		stdout.Out.Writef(`        /%s::%s\    \                /%s::%s\    \`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`       /%s::::%s\    \              /%s::::%s\    \`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`      /%s::::::%s\    \            /%s::::::%s\    \`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`     /%s:::%s/\%s:::%s\    \          /%s:::%s/\%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`    /%s:::%s/__\%s:::%s\    \        /%s:::%s/  \%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`    \%s:::%s\   \%s:::%s\    \      /%s:::%s/    \%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`  ___\%s:::%s\   \%s:::%s\    \    /%s:::%s/    / \%s:::%s\    \`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(` /\   \%s:::%s\   \%s:::%s\    \  /%s:::%s/    /   \%s:::%s\ ___\`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`/%s::%s\   \%s:::%s\   \%s:::%s\____\/%s:::%s/____/  ___\%s:::%s|    |`, color1a, color2, color1b, color2, color1c, color2, color1a, color2, color1b, color2)
		stdout.Out.Writef(`\%s:::%s\   \%s:::%s\   \%s::%s/    /\%s:::%s\    \ /\  /%s:::%s|____|`, color1a, color2, color1b, color2, color1c, color2, color1a, color2, color1b, color2)
		stdout.Out.Writef(` \%s:::%s\   \%s:::%s\   \/____/  \%s:::%s\    /%s::%s\ \%s::%s/    /`, color1a, color2, color1b, color2, color1c, color2, color1a, color2, color1b, color2)
		stdout.Out.Writef(`  \%s:::%s\   \%s:::%s\    \       \%s:::%s\   \%s:::%s\ \/____/`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`   \%s:::%s\   \%s:::%s\____\       \%s:::%s\   \%s:::%s\____\`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`    \%s:::%s\  /%s:::%s/    /        \%s:::%s\  /%s:::%s/    /`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`     \%s:::%s\/%s:::%s/    /          \%s:::%s\/%s:::%s/    /`, color1a, color2, color1b, color2, color1c, color2, color1a, color2)
		stdout.Out.Writef(`      \%s::::::%s/    /            \%s::::::%s/    /`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`       \%s::::%s/    /              \%s::::%s/    /`, color1a, color2, color1b, color2)
		stdout.Out.Writef(`        \%s::%s/    /                \%s::%s/____/`, color1a, color2, color1b, color2)
		stdout.Out.Write(`         \/____/`)
		stdout.Out.Writef("%s", output.StyleReset)

		time.Sleep(200 * time.Millisecond)

		color1a, color1b, color1c, color2 = randoColor(), color1a, color1b, color1c

		if i != times-1 {
			stdout.Out.MoveUpLines(linesPrinted)
		}
	}

	return nil
}
