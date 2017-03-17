package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func export_upstart(cfg *config, path string) error {
	keys := make([]string, len(procs))
	i := 0
	for k := range procs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for i, n := range keys {
		f, err := os.Create(filepath.Join(path, "app-"+n+".conf"))
		if err != nil {
			return err
		}

		fmt.Fprintf(f, "start on starting app-%s\n", n)
		fmt.Fprintf(f, "stop on stopping app-%s\n", n)
		fmt.Fprintf(f, "respawn\n")
		fmt.Fprintf(f, "\n")

		env := map[string]string{}
		procfile, err := filepath.Abs(cfg.Procfile)
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(filepath.Join(filepath.Dir(procfile), ".env"))
		if err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				token := strings.SplitN(line, "=", 2)
				if len(token) != 2 {
					continue
				}
				if strings.HasPrefix(token[0], "export ") {
					token[0] = token[0][7:]
				}
				token[0] = strings.TrimSpace(token[0])
				token[1] = strings.TrimSpace(token[1])
				env[token[0]] = token[1]
			}
		}

		fmt.Fprintf(f, "env PORT=%d\n", cfg.BasePort+uint(i))
		for k, v := range env {
			fmt.Fprintf(f, "env %s='%s'\n", k, strings.Replace(v, "'", "\\'", -1))
		}
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "setuid app\n")
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "chdir %s\n", filepath.ToSlash(filepath.Dir(procfile)))
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "exec %s\n", procs[n].cmdline)

		f.Close()
	}
	return nil
}

// command: export.
func export(cfg *config, format, path string) error {
	err := readProcfile(cfg)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	switch format {
	case "upstart":
		return export_upstart(cfg, path)
	}
	return nil
}
