package parser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cmdutil"
)

func vslog(out ...string) {
	os.Stderr.WriteString(strings.Join(out, "") + "\n")
}

// cmdOutput is a helper around c.Output which logs the command, how long it
// took to run, and a nice error in the event of failure.
//
// The specified env is set as cmd.Env (because we do this at ALL callsites
// today anyway).
func cmdOutput(ctx context.Context, env []string, c *exec.Cmd) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, c.Args[0])
	defer span.Finish()
	span.SetTag("command", strings.Join(c.Args, " "))
	if len(env) > 0 {
		span.SetTag("env", strings.Join(env, "; "))
		c.Env = env
	}
	start := time.Now()
	stdout, err := cmdutil.Output(c)
	log.Printf("TIME: %v '%s'\n", time.Since(start), strings.Join(c.Args, " "))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return stdout, nil
}

type Tag struct {
	File          string
	DefLinePrefix string
	Name          string

	// Extension fields
	Access         string // "private", "public"
	FileScope      string // ?
	Inheritance    string // ?
	Kind           string // "class"
	Language       string // "Java"
	Implementation string // ?
	Line           int    // 23
	Scope          string // "enum:gl::foobar"
	Signature      string // "(rtclass,objtype,obj,hr)"
	Type           string // ?
}

type TagsParser struct {
	// input
	config *Config

	// output
	tags      []Tag
	langFiles map[string][]string

	// temporary state
	curFile string
}

func NewParser(ctx context.Context) (*TagsParser, error) {
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &TagsParser{
		langFiles: make(map[string][]string),
		config:    cfg,
	}, nil
}

func (p *TagsParser) Tags() []Tag {
	return p.tags
}

func (p *TagsParser) Parse(r *bufio.Reader) error {
	p.curFile = ""

	line, err := r.ReadString('\n')
	for ; err == nil; line, err = r.ReadString('\n') {
		if err := p.parseLine(strings.TrimRight(line, "\r\n")); err != nil {
			return err
		}
	}
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (p *TagsParser) parseLine(line string) error {
	if len(strings.TrimSpace(line)) == 0 || strings.HasPrefix(line, "!") {
		return nil
	}

	t1 := strings.Index(line, "\t")
	if t1 == -1 {
		return fmt.Errorf("expected tab-delimited line with at least 4 fields, but got %q", line)
	}
	name := line[0:t1]

	offset := strings.Index(line[t1+1:], "\t")
	if offset == -1 {
		return fmt.Errorf("expected tab-delimited line with at least 4 fields, but got %q", line)
	}
	t2 := t1 + 1 + offset
	file := line[t1+1 : t2]

	offset = strings.LastIndex(line[t2+1:], `;"`)
	if offset == -1 {
		return fmt.Errorf(`expected find command to terminate with ';"', but got %q`, line)
	}
	t3 := offset + 2 + t2 + 1
	if len(line) <= t3 || line[t3] != '\t' {
		return fmt.Errorf(`expected tab immediately following ';"', but got %q, line: was %q`, line[t3:t3+1], line)
	}
	findCmd := line[t2+1 : t3]

	extFields_ := strings.Split(line[t3+1:], "\t")
	extFields := make(map[string]string)
	for _, extField := range extFields_ {
		s := strings.Index(extField, ":")
		key, val := extField[0:s], extField[s+1:]
		extFields[key] = val
	}
	lineno, err := strconv.Atoi(extFields["line"])
	if err != nil {
		return fmt.Errorf("could not parse line number, line was %q", line)
	}

	p.tags = append(p.tags, Tag{
		Name:          name,
		File:          file,
		DefLinePrefix: findCmdToDefLinePrefix(findCmd),
		Access:        extFields["access"],
		// FileScope:      string,
		// Inheritance:    string,
		Kind:     extFields["kind"],
		Language: extFields["language"],
		// Implementation: string,
		Line:      lineno,
		Scope:     extFields["scope"],
		Signature: extFields["signature"],
		Type:      extFields["typeref"],
	})
	return nil
}

func findCmdToDefLinePrefix(findCmd string) string {
	def := strings.TrimSuffix(strings.TrimPrefix(findCmd, `/^`), `/;"`)
	if strings.HasSuffix(def, "$") {
		def = strings.TrimSuffix(def, "$")
	}
	return def
}

var ignoreFiles = []string{".srclib-cache", "node_modules", "vendor", "dist"}

func Parse(ctx context.Context, rootDir string, files []string) (*TagsParser, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "parse ctags file")
	span.SetTag("rootDir", rootDir)
	defer span.Finish()

	tagsFilename := path.Join(rootDir, "tags")

	// Reuse an existing ctags file if we have it, otherwise generate one.
	if _, err := os.Stat(tagsFilename); os.IsNotExist(err) {
		args := []string{"-f", tagsFilename, "--fields=*", "--excmd=pattern"}
		if len(files) == 0 {
			args = append(args, "-R")
		} else {
			args = append(args, files...)
		}
		excludeArgs := make([]string, 0, len(ignoreFiles))
		for _, ignoreFile := range ignoreFiles {
			excludeArgs = append(excludeArgs, fmt.Sprintf("--exclude=%s", ignoreFile))
		}
		args = append(args, excludeArgs...)

		cmd := exec.Command("ctags", args...)
		cmd.Dir = rootDir

		vslog("...running ctags")
		out, err := cmdOutput(ctx, nil, cmd)
		vslog(string(out))
		if err != nil {
			return nil, err
		}
		vslog("...done running ctags")
	}

	tagsFile, err := os.Open(tagsFilename)
	if err != nil {
		return nil, err
	}
	fileInfo, err := tagsFile.Stat()
	if err != nil {
		return nil, err
	}
	defer tagsFile.Close()

	r := bufio.NewReader(tagsFile)
	p, err := NewParser(ctx)
	if err != nil {
		return nil, err
	}
	span.SetTag("ctags file size", fileInfo.Size())
	if err := p.Parse(r); err != nil {
		return nil, err
	}
	return p, nil
}
