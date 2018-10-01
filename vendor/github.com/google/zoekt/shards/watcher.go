// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shards

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type shardLoader interface {
	// Load a new file. Should be safe for concurrent calls.
	load(filename string)
	drop(filename string)
}

type shardWatcher struct {
	dir        string
	timestamps map[string]time.Time
	loader     shardLoader
	quit       chan<- struct{}
}

func (sw *shardWatcher) Close() error {
	if sw.quit != nil {
		close(sw.quit)
		sw.quit = nil
	}
	return nil
}

func NewDirectoryWatcher(dir string, loader shardLoader) (io.Closer, error) {
	quitter := make(chan struct{}, 1)
	sw := &shardWatcher{
		dir:        dir,
		timestamps: map[string]time.Time{},
		loader:     loader,
		quit:       quitter,
	}
	if err := sw.scan(); err != nil {
		return nil, err
	}

	if err := sw.watch(quitter); err != nil {
		return nil, err
	}

	return sw, nil
}

func (s *shardWatcher) String() string {
	return fmt.Sprintf("shardWatcher(%s)", s.dir)
}

func (s *shardWatcher) scan() error {
	fs, err := filepath.Glob(filepath.Join(s.dir, "*.zoekt"))
	if err != nil {
		return err
	}

	ts := map[string]time.Time{}
	for _, fn := range fs {
		fi, err := os.Lstat(fn)
		if err != nil {
			continue
		}

		ts[fn] = fi.ModTime()
	}

	var toLoad []string
	for k, mtime := range ts {
		if t, ok := s.timestamps[k]; !ok || t != mtime {
			toLoad = append(toLoad, k)
			s.timestamps[k] = mtime
		}
	}

	var toDrop []string
	// Unload deleted shards.
	for k := range s.timestamps {
		if _, ok := ts[k]; !ok {
			toDrop = append(toDrop, k)
			delete(s.timestamps, k)
		}
	}

	for _, t := range toDrop {
		log.Printf("unloading: %s", t)
		s.loader.drop(t)
	}

	var wg sync.WaitGroup
	for _, t := range toLoad {
		wg.Add(1)
		go func(k string) {
			s.loader.load(k)
			wg.Done()
		}(t)
	}
	wg.Wait()

	return nil
}

func (s *shardWatcher) watch(quitter <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(s.dir); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-watcher.Events:
				s.scan()
			case err := <-watcher.Errors:
				if err != nil {
					log.Println("watcher error:", err)
				}
			case <-quitter:
				watcher.Close()
				return
			}
		}
	}()
	return nil
}
