package custom

import "io/fs"

const (
	NameFs          = "fs"
	NameFsOpen      = "open"
	NameFsStat      = "stat"
	NameFsFstat     = "fstat"
	NameFsLstat     = "lstat"
	NameFsClose     = "close"
	NameFsWrite     = "write"
	NameFsRead      = "read"
	NameFsReaddir   = "readdir"
	NameFsMkdir     = "mkdir"
	NameFsRmdir     = "rmdir"
	NameFsRename    = "rename"
	NameFsUnlink    = "unlink"
	NameFsUtimes    = "utimes"
	NameFsChmod     = "chmod"
	NameFsFchmod    = "fchmod"
	NameFsChown     = "chown"
	NameFsFchown    = "fchown"
	NameFsLchown    = "lchown"
	NameFsTruncate  = "truncate"
	NameFsFtruncate = "ftruncate"
	NameFsReadlink  = "readlink"
	NameFsLink      = "link"
	NameFsSymlink   = "symlink"
	NameFsFsync     = "fsync"
)

// FsNameSection are the functions defined in the object named NameFs. Results
// here are those set to the current event object, but effectively are results
// of the host function.
var FsNameSection = map[string]*Names{
	NameFsOpen: {
		Name:        NameFsOpen,
		ParamNames:  []string{"path", "flags", "perm", NameCallback},
		ResultNames: []string{"err", "fd"},
	},
	NameFsStat: {
		Name:        NameFsStat,
		ParamNames:  []string{"path", NameCallback},
		ResultNames: []string{"err", "stat"},
	},
	NameFsFstat: {
		Name:        NameFsFstat,
		ParamNames:  []string{"fd", NameCallback},
		ResultNames: []string{"err", "stat"},
	},
	NameFsLstat: {
		Name:        NameFsLstat,
		ParamNames:  []string{"path", NameCallback},
		ResultNames: []string{"err", "stat"},
	},
	NameFsClose: {
		Name:        NameFsClose,
		ParamNames:  []string{"fd", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsRead: {
		Name:        NameFsRead,
		ParamNames:  []string{"fd", "buf", "offset", "byteCount", "fOffset", NameCallback},
		ResultNames: []string{"err", "n"},
	},
	NameFsWrite: {
		Name:        NameFsWrite,
		ParamNames:  []string{"fd", "buf", "offset", "byteCount", "fOffset", NameCallback},
		ResultNames: []string{"err", "n"},
	},
	NameFsReaddir: {
		Name:        NameFsReaddir,
		ParamNames:  []string{"path", NameCallback},
		ResultNames: []string{"err", "dirents"},
	},
	NameFsMkdir: {
		Name:        NameFsMkdir,
		ParamNames:  []string{"path", "perm", NameCallback},
		ResultNames: []string{"err", "fd"},
	},
	NameFsRmdir: {
		Name:        NameFsRmdir,
		ParamNames:  []string{"path", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsRename: {
		Name:        NameFsRename,
		ParamNames:  []string{"from", "to", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsUnlink: {
		Name:        NameFsUnlink,
		ParamNames:  []string{"path", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsUtimes: {
		Name:        NameFsUtimes,
		ParamNames:  []string{"path", "atime", "mtime", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsChmod: {
		Name:        NameFsChmod,
		ParamNames:  []string{"path", "mode", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsFchmod: {
		Name:        NameFsFchmod,
		ParamNames:  []string{"fd", "mode", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsChown: {
		Name:        NameFsChown,
		ParamNames:  []string{"path", "uid", "gid", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsFchown: {
		Name:        NameFsFchown,
		ParamNames:  []string{"fd", "uid", "gid", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsLchown: {
		Name:        NameFsLchown,
		ParamNames:  []string{"path", "uid", "gid", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsTruncate: {
		Name:        NameFsTruncate,
		ParamNames:  []string{"path", "length", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsFtruncate: {
		Name:        NameFsFtruncate,
		ParamNames:  []string{"fd", "length", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsReadlink: {
		Name:        NameFsReadlink,
		ParamNames:  []string{"path", NameCallback},
		ResultNames: []string{"err", "dst"},
	},
	NameFsLink: {
		Name:        NameFsLink,
		ParamNames:  []string{"path", "link", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsSymlink: {
		Name:        NameFsSymlink,
		ParamNames:  []string{"path", "link", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
	NameFsFsync: {
		Name:        NameFsFsync,
		ParamNames:  []string{"fd", NameCallback},
		ResultNames: []string{"err", "ok"},
	},
}

// mode constants from syscall_js.go
const (
	S_IFSOCK = uint32(0o000140000)
	S_IFLNK  = uint32(0o000120000)
	S_IFREG  = uint32(0o000100000)
	S_IFBLK  = uint32(0o000060000)
	S_IFDIR  = uint32(0o000040000)
	S_IFCHR  = uint32(0o000020000)
	S_IFIFO  = uint32(0o000010000)

	S_ISUID = uint32(0o004000)
	S_ISGID = uint32(0o002000)
	S_ISVTX = uint32(0o001000)
)

// ToJsMode is required because the mode property read in `GOOS=js` is
// incompatible with normal go. Particularly the directory flag isn't the same.
func ToJsMode(fm fs.FileMode) (jsMode uint32) {
	switch {
	case fm.IsRegular():
		jsMode = S_IFREG
	case fm.IsDir():
		jsMode = S_IFDIR
	case fm&fs.ModeSymlink != 0:
		jsMode = S_IFLNK
	case fm&fs.ModeDevice != 0:
		// Unlike ModeDevice and ModeCharDevice, S_IFCHR and S_IFBLK are set
		// mutually exclusively.
		if fm&fs.ModeCharDevice != 0 {
			jsMode = S_IFCHR
		} else {
			jsMode = S_IFBLK
		}
	case fm&fs.ModeNamedPipe != 0:
		jsMode = S_IFIFO
	case fm&fs.ModeSocket != 0:
		jsMode = S_IFSOCK
	default: // unknown
		jsMode = 0
	}

	jsMode |= uint32(fm & fs.ModePerm)

	if fm&fs.ModeSetgid != 0 {
		jsMode |= S_ISGID
	}
	if fm&fs.ModeSetuid != 0 {
		jsMode |= S_ISUID
	}
	if fm&fs.ModeSticky != 0 {
		jsMode |= S_ISVTX
	}
	return
}

// FromJsMode is required because the mode property read in `GOOS=js` is
// incompatible with normal go. Particularly the directory flag isn't the same.
func FromJsMode(jsMode, umask uint32) (fm fs.FileMode) {
	switch {
	case jsMode&S_IFREG != 0: // zero
	case jsMode&S_IFDIR != 0:
		fm = fs.ModeDir
	case jsMode&S_IFLNK != 0:
		fm = fs.ModeSymlink
	case jsMode&S_IFCHR != 0:
		fm = fs.ModeDevice | fs.ModeCharDevice
	case jsMode&S_IFBLK != 0:
		fm = fs.ModeDevice
	case jsMode&S_IFIFO != 0:
		fm = fs.ModeNamedPipe
	case jsMode&S_IFSOCK != 0:
		fm = fs.ModeSocket
	default: // unknown
		fm = 0
	}

	fm |= fs.FileMode(jsMode) & fs.ModePerm

	if jsMode&S_ISGID != 0 {
		fm |= fs.ModeSetgid
	}
	if jsMode&S_ISUID != 0 {
		fm |= fs.ModeSetuid
	}
	if jsMode&S_ISVTX != 0 {
		fm |= fs.ModeSticky
	}
	fm &= ^(fs.FileMode(umask))
	return
}
