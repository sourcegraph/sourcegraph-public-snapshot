package sysfs

import (
	"io"
	"io/fs"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/tetratelabs/wazero/internal/fsapi"
	"github.com/tetratelabs/wazero/internal/platform"
	"github.com/tetratelabs/wazero/sys"
)

func newOsFile(openPath string, openFlag int, openPerm fs.FileMode, f *os.File) fsapi.File {
	// Windows cannot read files written to a directory after it was opened.
	// This was noticed in #1087 in zig tests. Use a flag instead of a
	// different type.
	reopenDir := runtime.GOOS == "windows"
	return &osFile{path: openPath, flag: openFlag, perm: openPerm, reopenDir: reopenDir, file: f, fd: f.Fd()}
}

// osFile is a file opened with this package, and uses os.File or syscalls to
// implement api.File.
type osFile struct {
	path string
	flag int
	perm fs.FileMode
	file *os.File
	fd   uintptr

	// reopenDir is true if reopen should be called before Readdir. This flag
	// is deferred until Readdir to prevent redundant rewinds. This could
	// happen if Seek(0) was called twice, or if in Windows, Seek(0) was called
	// before Readdir.
	reopenDir bool

	// closed is true when closed was called. This ensures proper syscall.EBADF
	closed bool

	// cachedStat includes fields that won't change while a file is open.
	cachedSt *cachedStat
}

// cachedStat returns the cacheable parts of fsapi.Stat_t or an error if they
// couldn't be retrieved.
func (f *osFile) cachedStat() (dev uint64, ino sys.Inode, isDir bool, errno syscall.Errno) {
	if f.cachedSt == nil {
		if _, errno = f.Stat(); errno != 0 {
			return
		}
	}
	return f.cachedSt.dev, f.cachedSt.ino, f.cachedSt.isDir, 0
}

// Dev implements the same method as documented on fsapi.File
func (f *osFile) Dev() (uint64, syscall.Errno) {
	dev, _, _, errno := f.cachedStat()
	return dev, errno
}

// Ino implements the same method as documented on fsapi.File
func (f *osFile) Ino() (sys.Inode, syscall.Errno) {
	_, ino, _, errno := f.cachedStat()
	return ino, errno
}

// IsDir implements the same method as documented on fsapi.File
func (f *osFile) IsDir() (bool, syscall.Errno) {
	_, _, isDir, errno := f.cachedStat()
	return isDir, errno
}

// IsAppend implements File.IsAppend
func (f *osFile) IsAppend() bool {
	return f.flag&syscall.O_APPEND == syscall.O_APPEND
}

// SetAppend implements the same method as documented on fsapi.File
func (f *osFile) SetAppend(enable bool) (errno syscall.Errno) {
	if enable {
		f.flag |= syscall.O_APPEND
	} else {
		f.flag &= ^syscall.O_APPEND
	}

	// Clear any create flag, as we are re-opening, not re-creating.
	f.flag &= ^syscall.O_CREAT

	// appendMode (bool) cannot be changed later, so we have to re-open the
	// file. https://github.com/golang/go/blob/go1.20/src/os/file_unix.go#L60
	return fileError(f, f.closed, f.reopen())
}

// compile-time check to ensure osFile.reopen implements reopenFile.
var _ reopenFile = (*fsFile)(nil).reopen

func (f *osFile) reopen() (errno syscall.Errno) {
	// Clear any create flag, as we are re-opening, not re-creating.
	f.flag &= ^syscall.O_CREAT

	_ = f.close()
	f.file, errno = OpenFile(f.path, f.flag, f.perm)
	return
}

// IsNonblock implements the same method as documented on fsapi.File
func (f *osFile) IsNonblock() bool {
	return isNonblock(f)
}

// SetNonblock implements the same method as documented on fsapi.File
func (f *osFile) SetNonblock(enable bool) (errno syscall.Errno) {
	if enable {
		f.flag |= fsapi.O_NONBLOCK
	} else {
		f.flag &= ^fsapi.O_NONBLOCK
	}
	if err := setNonblock(f.fd, enable); err != nil {
		return fileError(f, f.closed, platform.UnwrapOSError(err))
	}
	return 0
}

// Stat implements the same method as documented on fsapi.File
func (f *osFile) Stat() (sys.Stat_t, syscall.Errno) {
	if f.closed {
		return sys.Stat_t{}, syscall.EBADF
	}

	st, errno := statFile(f.file)
	switch errno {
	case 0:
		f.cachedSt = &cachedStat{dev: st.Dev, ino: st.Ino, isDir: st.Mode&fs.ModeDir == fs.ModeDir}
	case syscall.EIO:
		errno = syscall.EBADF
	}
	return st, errno
}

// Read implements the same method as documented on fsapi.File
func (f *osFile) Read(buf []byte) (n int, errno syscall.Errno) {
	if len(buf) == 0 {
		return 0, 0 // Short-circuit 0-len reads.
	}
	if NonBlockingFileIoSupported && f.IsNonblock() {
		n, errno = readFd(f.fd, buf)
	} else {
		n, errno = read(f.file, buf)
	}
	if errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)
	}
	return
}

// Pread implements the same method as documented on fsapi.File
func (f *osFile) Pread(buf []byte, off int64) (n int, errno syscall.Errno) {
	if n, errno = pread(f.file, buf, off); errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)
	}
	return
}

// Seek implements the same method as documented on fsapi.File
func (f *osFile) Seek(offset int64, whence int) (newOffset int64, errno syscall.Errno) {
	if newOffset, errno = seek(f.file, offset, whence); errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)

		// If the error was trying to rewind a directory, re-open it. Notably,
		// seeking to zero on a directory doesn't work on Windows with Go 1.18.
		if errno == syscall.EISDIR && offset == 0 && whence == io.SeekStart {
			errno = 0
			f.reopenDir = true
		}
	}
	return
}

// PollRead implements the same method as documented on fsapi.File
func (f *osFile) PollRead(timeout *time.Duration) (ready bool, errno syscall.Errno) {
	fdSet := platform.FdSet{}
	fd := int(f.fd)
	fdSet.Set(fd)
	nfds := fd + 1 // See https://man7.org/linux/man-pages/man2/select.2.html#:~:text=condition%20has%20occurred.-,nfds,-This%20argument%20should
	count, err := _select(nfds, &fdSet, nil, nil, timeout)
	if errno = platform.UnwrapOSError(err); errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)
	}
	return count > 0, errno
}

// Readdir implements File.Readdir. Notably, this uses "Readdir", not
// "ReadDir", from os.File.
func (f *osFile) Readdir(n int) (dirents []fsapi.Dirent, errno syscall.Errno) {
	if f.reopenDir { // re-open the directory if needed.
		f.reopenDir = false
		if errno = adjustReaddirErr(f, f.closed, f.reopen()); errno != 0 {
			return
		}
	}

	if dirents, errno = readdir(f.file, f.path, n); errno != 0 {
		errno = adjustReaddirErr(f, f.closed, errno)
	}
	return
}

// Write implements the same method as documented on fsapi.File
func (f *osFile) Write(buf []byte) (n int, errno syscall.Errno) {
	if n, errno = write(f.file, buf); errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)
	}
	return
}

// Pwrite implements the same method as documented on fsapi.File
func (f *osFile) Pwrite(buf []byte, off int64) (n int, errno syscall.Errno) {
	if n, errno = pwrite(f.file, buf, off); errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)
	}
	return
}

// Truncate implements the same method as documented on fsapi.File
func (f *osFile) Truncate(size int64) (errno syscall.Errno) {
	if errno = platform.UnwrapOSError(f.file.Truncate(size)); errno != 0 {
		// Defer validation overhead until we've already had an error.
		errno = fileError(f, f.closed, errno)
	}
	return
}

// Sync implements the same method as documented on fsapi.File
func (f *osFile) Sync() syscall.Errno {
	return fsync(f.file)
}

// Datasync implements the same method as documented on fsapi.File
func (f *osFile) Datasync() syscall.Errno {
	return datasync(f.file)
}

// Utimens implements the same method as documented on fsapi.File
func (f *osFile) Utimens(times *[2]syscall.Timespec) syscall.Errno {
	if f.closed {
		return syscall.EBADF
	}

	err := futimens(f.fd, times)
	return platform.UnwrapOSError(err)
}

// Close implements the same method as documented on fsapi.File
func (f *osFile) Close() syscall.Errno {
	if f.closed {
		return 0
	}
	f.closed = true
	return f.close()
}

func (f *osFile) close() syscall.Errno {
	return platform.UnwrapOSError(f.file.Close())
}
