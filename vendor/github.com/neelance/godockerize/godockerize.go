package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:    "godockerize",
		Usage:   "build Docker images from Go packages",
		Version: "0.0.1",
		Commands: []*cli.Command{
			{
				Name:      "build",
				Usage:     "build a Docker image from a Go package",
				ArgsUsage: "[package]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Usage:   "output Docker image name and optionally a tag in the 'name:tag' format",
					},
					&cli.StringFlag{
						Name:  "base",
						Usage: "base Docker image name",
						Value: "alpine:3.4",
					},
				},
				Action: doBuild,
			},
		},
	}
	app.Run(os.Args)
}

func doBuild(c *cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := c.Args()
	if args.Len() != 1 {
		return errors.New(`"godockerize build" requires exactly 1 argument`)
	}

	pkg, err := build.Import(args.First(), wd, 0)
	if err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir("", "godockerize")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	binname := path.Base(pkg.ImportPath)

	fset := token.NewFileSet()
	env := []string{}
	expose := []string{}
	install := []string{"ca-certificates", "mailcap"} // mailcap is for /etc/mime.types
	run := []string{}
	for _, name := range pkg.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, name), nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if strings.HasPrefix(c.Text, "//docker:") {
					parts := strings.SplitN(c.Text[9:], " ", 2)
					switch parts[0] {
					case "env":
						env = append(env, strings.Fields(parts[1])...)
					case "expose":
						expose = append(expose, strings.Fields(parts[1])...)
					case "install":
						install = append(install, strings.Fields(parts[1])...)
					case "run":
						run = append(run, parts[1])
					default:
						return fmt.Errorf("%s: invalid docker comment: %s", fset.Position(c.Pos()), c.Text)
					}
				}
			}
		}
	}

	var dockerfile bytes.Buffer
	fmt.Fprintf(&dockerfile, "  FROM %s\n", c.String("base"))

	for _, pkg := range install {
		if strings.HasSuffix(pkg, "@edge") {
			fmt.Fprintf(&dockerfile, "  RUN echo -e \"@edge http://dl-cdn.alpinelinux.org/alpine/edge/main\\n@edge http://dl-cdn.alpinelinux.org/alpine/edge/community\" >> /etc/apk/repositories\n")
			break
		}
	}
	if len(install) != 0 {
		fmt.Fprintf(&dockerfile, "  RUN apk add --no-cache %s\n", strings.Join(sortedStringSet(install), " "))
	}

	for _, cmd := range run {
		fmt.Fprintf(&dockerfile, "  RUN %s\n", cmd)
	}
	if len(env) != 0 {
		fmt.Fprintf(&dockerfile, "  ENV %s\n", strings.Join(sortedStringSet(env), " "))
	}
	if len(expose) != 0 {
		fmt.Fprintf(&dockerfile, "  EXPOSE %s\n", strings.Join(sortedStringSet(expose), " "))
	}
	fmt.Fprintf(&dockerfile, "  ENTRYPOINT [\"/usr/local/bin/%s\"]\n", binname)
	fmt.Fprintf(&dockerfile, "  ADD %s /usr/local/bin/\n", binname)

	fmt.Println("godockerize: Generated Dockerfile:")
	fmt.Print(dockerfile.String())

	ioutil.WriteFile(filepath.Join(tmpdir, "Dockerfile"), dockerfile.Bytes(), 0777)
	if err != nil {
		return err
	}

	fmt.Println("godockerize: Building Go binary...")
	cmd := exec.Command("go", "build", "-o", binname, pkg.ImportPath)
	cmd.Dir = tmpdir
	cmd.Env = []string{
		"GOARCH=amd64",
		"GOOS=linux",
		"GOROOT=" + build.Default.GOROOT,
		"GOPATH=" + build.Default.GOPATH,
		"CGO_ENABLED=0",
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("godockerize: Building Docker image...")
	dockerArgs := []string{"build"}
	if tag := c.String("tag"); tag != "" {
		dockerArgs = append(dockerArgs, "-t", tag)
	}
	dockerArgs = append(dockerArgs, ".")
	cmd = exec.Command("docker", dockerArgs...)
	cmd.Dir = tmpdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func sortedStringSet(in []string) []string {
	set := make(map[string]struct{})
	for _, s := range in {
		set[s] = struct{}{}
	}
	var out []string
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
