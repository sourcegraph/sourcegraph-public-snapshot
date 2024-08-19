package gojs

import (
	"context"
	"path"
	"syscall"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/internal/gojs/custom"
	"github.com/tetratelabs/wazero/internal/gojs/goos"
	"github.com/tetratelabs/wazero/internal/gojs/util"
)

// processState are the mutable fields of the current process.
type processState struct {
	cwd   string
	umask uint32
}

func newJsProcess(proc *processState) *jsVal {
	// Fill fake values for user/group info as we don't support it.
	uidRef := goos.RefValueZero
	gidRef := goos.RefValueZero
	euidRef := goos.RefValueZero
	groupSlice := []interface{}{goos.RefValueZero}

	// jsProcess = js.Global().Get("process") // fs_js.go init
	return newJsVal(goos.RefJsProcess, custom.NameProcess).
		addProperties(map[string]interface{}{
			"pid":  float64(1),        // Get("pid").Int() in syscall_js.go for syscall.Getpid
			"ppid": goos.RefValueZero, // Get("ppid").Int() in syscall_js.go for syscall.Getppid
		}).
		addFunction(custom.NameProcessCwd, &processCwd{proc: proc}).       // syscall.Cwd in fs_js.go
		addFunction(custom.NameProcessChdir, &processChdir{proc: proc}).   // syscall.Chdir in fs_js.go
		addFunction(custom.NameProcessGetuid, getId(uidRef)).              // syscall.Getuid in syscall_js.go
		addFunction(custom.NameProcessGetgid, getId(gidRef)).              // syscall.Getgid in syscall_js.go
		addFunction(custom.NameProcessGeteuid, getId(euidRef)).            // syscall.Geteuid in syscall_js.go
		addFunction(custom.NameProcessGetgroups, returnSlice(groupSlice)). // syscall.Getgroups in syscall_js.go
		addFunction(custom.NameProcessUmask, &processUmask{proc: proc})    // syscall.Umask in syscall_js.go
}

// processCwd implements jsFn for fs.Open syscall.Getcwd in fs_js.go
type processCwd struct {
	proc *processState
}

func (p *processCwd) invoke(_ context.Context, _ api.Module, _ ...interface{}) (interface{}, error) {
	return p.proc.cwd, nil
}

// processChdir implements jsFn for fs.Open syscall.Chdir in fs_js.go
type processChdir struct {
	proc *processState
}

func (p *processChdir) invoke(_ context.Context, mod api.Module, args ...interface{}) (interface{}, error) {
	oldWd := p.proc.cwd
	newWd := util.ResolvePath(oldWd, args[0].(string))

	newWd = path.Clean(newWd)
	if newWd == oldWd { // handle .
		return nil, nil
	}

	if s, err := syscallStat(mod, newWd); err != nil {
		return nil, err
	} else if !s.isDir {
		return nil, syscall.ENOTDIR
	} else {
		p.proc.cwd = newWd
		return nil, nil
	}
}

// processUmask implements jsFn for fs.Open syscall.Umask in fs_js.go
type processUmask struct {
	proc *processState
}

func (p *processUmask) invoke(_ context.Context, _ api.Module, args ...interface{}) (interface{}, error) {
	newUmask := goos.ValueToUint32(args[0])

	oldUmask := p.proc.umask
	p.proc.umask = newUmask

	return oldUmask, nil
}

// getId implements jsFn for syscall.Getuid, syscall.Getgid and syscall.Geteuid in syscall_js.go
type getId goos.Ref

func (i getId) invoke(_ context.Context, _ api.Module, _ ...interface{}) (interface{}, error) {
	return goos.Ref(i), nil
}

// returnSlice implements jsFn for syscall.Getgroups in syscall_js.go
type returnSlice []interface{}

func (s returnSlice) invoke(context.Context, api.Module, ...interface{}) (interface{}, error) {
	return &objectArray{slice: s}, nil
}
