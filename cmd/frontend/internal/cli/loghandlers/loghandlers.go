// Package loghandlers contains log15 handlers/filters used by the sourcegraph
// cli
package loghandlers

import (
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

// Trace returns a filter for the given traces that run longer than threshold
func Trace(types []string, threshold time.Duration) func(*log15.Record) bool {
	all := false
	valid := map[string]bool{}
	for _, t := range types {
		valid[t] = true
		if t == "all" {
			all = true
		}
	}
	return func(r *log15.Record) bool {
		if r.Lvl != log15.LvlDebug {
			return true
		}
		if !strings.HasPrefix(r.Msg, "TRACE ") {
			return true
		}
		if !all && !valid[r.Msg[6:]] {
			return false
		}
		for i := 1; i < len(r.Ctx); i += 2 {
			if r.Ctx[i-1] != "duration" {
				continue
			}
			d, ok := r.Ctx[i].(time.Duration)
			return !ok || d >= threshold
		}
		return true
	}
}

// NotNoisey filters out high firing and low signal debug logs
func NotNoisey(r *log15.Record) bool {
	if r.Lvl != log15.LvlDebug {
		return true
	}
	noiseyPrefixes := []string{"repoUpdater: RefreshVCS"}
	for _, prefix := range noiseyPrefixes {
		if strings.HasPrefix(r.Msg, prefix) {
			return false
		}
	}
	if !strings.HasPrefix(r.Msg, "TRACE backend") || len(r.Ctx) < 2 {
		return true
	}
	rpc, ok := r.Ctx[1].(string)
	if !ok {
		return true
	}
	for _, n := range noiseyRPC {
		if rpc == n {
			return false
		}
	}
	return true
}

var noiseyRPC = []string{"MirrorRepos.RefreshVCS"}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_327(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
