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

	"github.com/fsnotify/fsnotify"
	"golang.org/x/tools/go/buildutil"
)

var (
	buildTags = flag.String("tags", "", buildutil.TagsFlagDoc)
	verbose   = flag.Bool("v", false, "verbose output")
	timings   = flag.Bool("timings", false, "show timings")
	race      = flag.Bool("race", false, "build with Go race detector")
	ienv      = flag.String("installenv", "", "env vars to pass to `go install` (comma-separated: A=B,C=D)")
	extra     = flag.String("extra-watches", "", "comma-separated path match patterns to also watch (in addition to transitive deps of Go pkg)")
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
		if pkg.Goroot {
			continue // don't watch Go stdlib packages
		}
		if *verbose {
			log.Printf("Watch %s", pkg.Dir)
		}
		if err := w.Add(pkg.Dir); err != nil {
			log.Fatal(err)
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

	if *extra != "" {
		for _, pat := range strings.Split(*extra, ",") {
			matches, err := filepath.Glob(pat)
			if err != nil {
				log.Fatal(err)
			}
			for _, path := range matches {
				if *verbose {
					log.Printf("Watch (extra) %s", path)
				}
				if err := w.Add(path); err != nil {
					log.Fatal(err)
				}
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

	matchFile := func(name string) bool {
		return filepath.Ext(name) == ".go" && !strings.HasPrefix(filepath.Base(name), ".")
	}

	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				break
			}

			go func() {
				switch ev.Op {
				case fsnotify.Create, fsnotify.Rename, fsnotify.Write:
					paths := []string{ev.Name}

					// w.Add is non-recursive if the path is a dir, so
					// we need to scan for the files here.
					if fi, err := os.Stat(ev.Name); err != nil {
						if *verbose {
							log.Println(err)
						}
						return
					} else if fi.Mode().IsDir() {
						err := filepath.Walk(ev.Name, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								if *verbose {
									log.Println(err)
								}
								return nil
							}
							if info.Mode().IsDir() || matchFile(info.Name()) {
								paths = append(paths, path)
							}
							return nil
						})
						if err != nil && *verbose {
							log.Println(err)
						}
					} else if !matchFile(ev.Name) {
						// File did not match.
						return
					}

					for _, path := range paths {
						if *verbose {
							log.Printf("Watch %s", path)
						}
						if err := w.Add(path); err != nil {
							if *verbose {
								log.Println(err)
							}
						}
					}

				case fsnotify.Remove:
					if err := w.Remove(ev.Name); err != nil {
						if *verbose {
							log.Println(err)
						}
					}
				case fsnotify.Chmod:
					return
				}
				if *verbose {
					log.Println(ev)
				}
				install <- struct{}{}
			}()
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
