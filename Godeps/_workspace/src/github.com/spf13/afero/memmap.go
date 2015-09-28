// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package afero

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var mux = &sync.Mutex{}

type MemMapFs struct {
	data  map[string]File
	mutex *sync.RWMutex
}

func (m *MemMapFs) lock() {
	mx := m.getMutex()
	mx.Lock()
}
func (m *MemMapFs) unlock()  { m.getMutex().Unlock() }
func (m *MemMapFs) rlock()   { m.getMutex().RLock() }
func (m *MemMapFs) runlock() { m.getMutex().RUnlock() }

func (m *MemMapFs) getData() map[string]File {
	if m.data == nil {
		m.data = make(map[string]File)
	}
	return m.data
}

func (m *MemMapFs) getMutex() *sync.RWMutex {
	mux.Lock()
	if m.mutex == nil {
		m.mutex = &sync.RWMutex{}
	}
	mux.Unlock()
	return m.mutex
}

type MemDirMap map[string]File

func (m MemDirMap) Len() int      { return len(m) }
func (m MemDirMap) Add(f File)    { m[f.Name()] = f }
func (m MemDirMap) Remove(f File) { delete(m, f.Name()) }
func (m MemDirMap) Files() (files []File) {
	for _, f := range m {
		files = append(files, f)
	}
	return files
}

func (m MemDirMap) Names() (names []string) {
	for x := range m {
		names = append(names, x)
	}
	return names
}

func (MemMapFs) Name() string { return "MemMapFS" }

func (m *MemMapFs) Create(name string) (File, error) {
	m.lock()
	file := MemFileCreate(name)
	m.getData()[name] = file
	m.unlock()
	m.registerDirs(file)
	return file, nil
}

func (m *MemMapFs) registerDirs(f File) {
	var x = f.Name()
	for x != "/" {
		f := m.registerWithParent(f)
		if f == nil {
			break
		}
		x = f.Name()
	}
}

func (m *MemMapFs) unRegisterWithParent(f File) File {
	parent := m.findParent(f)
	pmem := parent.(*InMemoryFile)
	pmem.memDir.Remove(f)
	return parent
}

func (m *MemMapFs) findParent(f File) File {
	dirs, _ := path.Split(f.Name())
	if len(dirs) > 1 {
		_, parent := path.Split(path.Clean(dirs))
		if len(parent) > 0 {
			pfile, err := m.Open(parent)
			if err != nil {
				return pfile
			}
		}
	}
	return nil
}

func (m *MemMapFs) registerWithParent(f File) File {
	if f == nil {
		return nil
	}
	parent := m.findParent(f)
	if parent != nil {
		pmem := parent.(*InMemoryFile)
		pmem.memDir.Add(f)
	} else {
		pdir := filepath.Dir(path.Clean(f.Name()))
		m.Mkdir(pdir, 0777)
	}
	return parent
}

func (m *MemMapFs) Mkdir(name string, perm os.FileMode) error {
	m.rlock()
	_, ok := m.getData()[name]
	m.runlock()
	if ok {
		return ErrFileExists
	} else {
		m.lock()
		item := &InMemoryFile{name: name, memDir: &MemDirMap{}, dir: true}
		m.getData()[name] = item
		m.unlock()
		m.registerDirs(item)
	}
	return nil
}

func (m *MemMapFs) MkdirAll(path string, perm os.FileMode) error {
	return m.Mkdir(path, 0777)
}

func (m *MemMapFs) Open(name string) (File, error) {
	m.rlock()
	f, ok := m.getData()[name]
	ff, ok := f.(*InMemoryFile)
	if ok {
		ff.Open()
	}
	m.runlock()

	if ok {
		return f, nil
	} else {
		return nil, ErrFileNotFound
	}
}

func (m *MemMapFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return m.Open(name)
}

func (m *MemMapFs) Remove(name string) error {
	m.lock()
	defer m.unlock()

	if _, ok := m.getData()[name]; ok {
		delete(m.getData(), name)
	} else {
		return &os.PathError{"remove", name, os.ErrNotExist}
	}
	return nil
}

func (m *MemMapFs) RemoveAll(path string) error {
	m.rlock()
	defer m.runlock()
	for p, _ := range m.getData() {
		if strings.HasPrefix(p, path) {
			m.runlock()
			m.lock()
			delete(m.getData(), p)
			m.unlock()
			m.rlock()
		}
	}
	return nil
}

func (m *MemMapFs) Rename(oldname, newname string) error {
	m.rlock()
	defer m.runlock()
	if _, ok := m.getData()[oldname]; ok {
		if _, ok := m.getData()[newname]; !ok {
			m.runlock()
			m.lock()
			m.getData()[newname] = m.getData()[oldname]
			delete(m.getData(), oldname)
			m.unlock()
			m.rlock()
		} else {
			return ErrDestinationExists
		}
	} else {
		return ErrFileNotFound
	}
	return nil
}

func (m *MemMapFs) Stat(name string) (os.FileInfo, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	return &InMemoryFileInfo{file: f.(*InMemoryFile)}, nil
}

func (m *MemMapFs) Chmod(name string, mode os.FileMode) error {
	f, ok := m.getData()[name]
	if !ok {
		return &os.PathError{"chmod", name, ErrFileNotFound}
	}

	ff, ok := f.(*InMemoryFile)
	if ok {
		m.lock()
		ff.mode = mode
		m.unlock()
	} else {
		return errors.New("Unable to Chmod Memory File")
	}
	return nil
}

func (m *MemMapFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	f, ok := m.getData()[name]
	if !ok {
		return &os.PathError{"chtimes", name, ErrFileNotFound}
	}

	ff, ok := f.(*InMemoryFile)
	if ok {
		m.lock()
		ff.modtime = mtime
		m.unlock()
	} else {
		return errors.New("Unable to Chtime Memory File")
	}
	return nil
}

func (m *MemMapFs) List() {
	for _, x := range m.data {
		y, _ := x.Stat()
		fmt.Println(x.Name(), y.Size())
	}
}
