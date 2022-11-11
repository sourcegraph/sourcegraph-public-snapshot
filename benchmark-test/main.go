package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"
	"time"
)


var (
	lastCheckAt    = make(map[string]time.Time)
	lastCheckMutex sync.Mutex
)

func Debounce(name string, since time.Duration) bool {
	lastCheckMutex.Lock()
	defer lastCheckMutex.Unlock()
	if t, ok := lastCheckAt[name]; ok && time.Now().Before(t.Add(since)) {
		return false
	}
	lastCheckAt[name] = time.Now()
	return true
}


type GitDir string

func (dir GitDir) Path(elem ...string) string {
	return filepath.Join(append([]string{string(dir)}, elem...)...)
}


func RepoLastFetched(dir GitDir) (time.Time, error) {
	fi, err := os.Stat(("FETCH_HEAD"))
	if os.IsNotExist(err) {
		fi, err = os.Stat(dir.Path("HEAD"))
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}


func main() {
	// fmt.Println(Calculate(1, 2))
	fmt.Println("hello world")
}
