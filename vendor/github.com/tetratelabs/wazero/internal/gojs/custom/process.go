package custom

const (
	NameProcess          = "process"
	NameProcessArgv0     = "argv0"
	NameProcessCwd       = "cwd"
	NameProcessChdir     = "chdir"
	NameProcessGetuid    = "getuid"
	NameProcessGetgid    = "getgid"
	NameProcessGeteuid   = "geteuid"
	NameProcessGetgroups = "getgroups"
	NameProcessUmask     = "umask"
)

// ProcessNameSection are the functions defined in the object named NameProcess.
// Results here are those set to the current event object, but effectively are
// results of the host function.
var ProcessNameSection = map[string]*Names{
	NameProcessArgv0: {
		Name:        NameProcessArgv0,
		ParamNames:  []string{},
		ResultNames: []string{"argv0"},
	},
	NameProcessCwd: {
		Name:        NameProcessCwd,
		ParamNames:  []string{},
		ResultNames: []string{"cwd"},
	},
	NameProcessChdir: {
		Name:        NameProcessChdir,
		ParamNames:  []string{"path"},
		ResultNames: []string{"err"},
	},
	NameProcessGetuid: {
		Name:        NameProcessGetuid,
		ParamNames:  []string{},
		ResultNames: []string{"uid"},
	},
	NameProcessGetgid: {
		Name:        NameProcessGetgid,
		ParamNames:  []string{},
		ResultNames: []string{"gid"},
	},
	NameProcessGeteuid: {
		Name:        NameProcessGeteuid,
		ParamNames:  []string{},
		ResultNames: []string{"euid"},
	},
	NameProcessGetgroups: {
		Name:        NameProcessGetgroups,
		ParamNames:  []string{},
		ResultNames: []string{"groups"},
	},
	NameProcessUmask: {
		Name:        NameProcessUmask,
		ParamNames:  []string{"mask"},
		ResultNames: []string{"oldmask"},
	},
}
