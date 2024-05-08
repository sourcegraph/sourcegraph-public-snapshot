package connection

// GlobalConns is the global variable holding a reference to the gitserver connections.
var GlobalConns = &atomicGitServerConns{}
