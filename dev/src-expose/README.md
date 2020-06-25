```bash
✗ podman run sourcegraph/src-expose:latest -h
```

```
USAGE
  src-expose [flags] <src1> [<src2> ...]

Periodically sync directories src1, src2, ... and serve them.

See "src-expose -h" for the flags that can be passed.

For more advanced uses specify --config pointing to a yaml file.
See https://github.com/sourcegraph/sourcegraph/tree/master/dev/src-expose/examples

SUBCOMMANDS
  serve  Serve git repos for Sourcegraph to list and clone.
  sync   Do a one-shot sync of directories

FLAGS
  -addr :3434     address on which to serve (end with : for unused port)
  -before ...     A command to run before sync. It is run from the current working directory.
  -config ...     If set will be used instead of command line arguments to specify configuration.
  -quiet false    
  -repos-dir ...  src-expose's git directories. src-expose creates a git repo per directory synced. The git repo is then served to Sourcegraph. The repositories are stored and served relative to this directory. Default: ~/.sourcegraph/src-expose-repos
  -verbose false
```
