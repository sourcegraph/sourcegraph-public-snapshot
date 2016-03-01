package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/buildutil"
	"gopkg.in/fsnotify.v1"
)

var (
	buildTags = flag.String("tags", "", buildutil.TagsFlagDoc)
	verbose   = flag.Bool("v", false, "verbose output")
	timings   = flag.Bool("timings", false, "show timings")
	race      = flag.Bool("race", false, "build with Go race detector")
	ienv      = flag.String("installenv", "", "env vars to pass to `go install` (comma-separated: A=B,C=D)")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("must provide package path")
	}

	pkgPath := flag.Arg(0)
	cmdArgs := flag.Args()[1:]

	var installEnv []string
	if *ienv != "" {
		installEnv = append(os.Environ(), strings.Split(*ienv, ",")...)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	pkg, err := build.Import(pkgPath, wd, 0)
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		log.Printf("Watching package %s", pkg.ImportPath)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	pkgs := []*build.Package{pkg}
	seenPkgs := map[string]struct{}{}
	for i := 0; i < len(pkgs); i++ {
		pkg := pkgs[i]
		if err := w.Add(pkg.Dir); err != nil {
			log.Fatal(err)
		}
		for _, file := range pkg.GoFiles {
			srcFile := filepath.Join(pkg.Dir, file)
			if *verbose {
				log.Printf("Watch %s", srcFile)
			}
			if err := w.Add(srcFile); err != nil {
				log.Fatal(err)
			}
		}
		for _, imp := range pkg.Imports {
			if _, seen := seenPkgs[imp]; !seen {
				if imp == "C" {
					continue
				}
				impPkg, err := build.Import(imp, pkg.Dir, 0)
				if err != nil {
					log.Fatal(err)
				}
				pkgs = append(pkgs, impPkg)
				seenPkgs[imp] = struct{}{}
			}
		}
	}

	restart := make(chan struct{})
	go func() {
		var proc *os.Process
		for v := range restart {
			_ = v
			if proc != nil {
				if err := proc.Signal(os.Interrupt); err != nil {
					log.Println(err)
					proc.Kill()
				}
				proc.Wait()
			}
			cmd := exec.Command(filepath.Join(pkg.BinDir, filepath.Base(pkg.ImportPath)), cmdArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if *verbose {
				log.Println(cmd.Args)
			}
			if err := cmd.Start(); err != nil {
				log.Println(err)
			}
			proc = cmd.Process
		}
	}()

	nrestarts := 0
	installAndRestart := func() {
		s := "\x1b[37;1m\x1b[44m .. \x1b[0m"
		del := len(s)
		fmt.Fprint(os.Stderr, s)

		cmd := exec.Command("go", "install", "-tags="+*buildTags)
		if *race {
			cmd.Args = append(cmd.Args, "-race")
		}
		cmd.Args = append(cmd.Args, pkg.ImportPath)
		cmd.Env = installEnv
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if *verbose {
			log.Println(cmd.Args)
			if installEnv != nil {
				log.Println("# with env:", installEnv)
			}
		}
		start := time.Now()
		if err := cmd.Run(); err == nil {
			var word string
			if nrestarts == 0 {
				word = "starting"
			} else {
				word = "restarting"
			}
			nrestarts++
			fmt.Fprint(os.Stderr, strings.Repeat("\b", del))
			log.Println("\x1b[37;1m\x1b[42m ok \x1b[0m", word)
			if *timings {
				log.Println("compilation took", time.Since(start))
			}
			restart <- struct{}{}
		} else {
			log.Println("\x1b[37;1m\x1b[41m!!!!\x1b[0m", "compilation failed")
		}
	}

	install := make(chan struct{})
	go func() {
		needsInstall := 0
		for {
			var timerChan <-chan time.Time
			if needsInstall > 0 {
				timerChan = time.After(200 * time.Millisecond)
			} else {
				timerChan = make(chan time.Time) // never sent on, blocks indefinitely
			}
			select {
			case <-install:
				needsInstall++
				continue
			case <-timerChan:
				needsInstall = 0
				installAndRestart()
			}
		}
	}()
	install <- struct{}{}

	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				break
			}

			if ext := filepath.Ext(ev.Name); ext != ".go" {
				continue
			}
			if strings.HasPrefix(filepath.Base(ev.Name), ".") {
				continue
			}

			switch ev.Op {
			case fsnotify.Create:
				if err := w.Add(ev.Name); err != nil {
					log.Println(err)
				}
			case fsnotify.Remove:
				if err := w.Remove(ev.Name); err != nil {
					log.Println(err)
				}
			case fsnotify.Chmod:
				continue
			}
			if *verbose {
				log.Println(ev)
			}
			install <- struct{}{}
		case err, ok := <-w.Errors:
			if !ok {
				break
			}
			if ok {
				log.Fatal(err)
			}
		}
	}
}
