# gitserver

gitserver exposes an "exec" API over HTTP for running git commands against
clones of repositories. gitserver also exposes APIs for the management of
clones.

The management of clones comprises most of the complexity in gitserver since:

- We want to avoid concurrent clones and fetches of the same repository.
- We want to limit the number of concurrent clones and fetches.
- When adding/removing/modifying a clone, concurrent attempts to run commands
  needs to be gracefully dealt with.
- We need to be robust against the many ways git clones can degrade. (gc,
  interrupted clones)

Additionally we have invested heavily in the observability of
gitserver. Nearly every operation Sourcegraph does runs one or more git
commands. So we have detailed oberservability in prometheus, net/event,
lightstep, honeycomb and stderr logs.

We normalize repository names when storing them on disk. Always use
`protocol.NormalizeRepo`. The `$GIT_DIR` of a repository is at
`reposRoot/normalized_name/.git`.

When doing an operation on a file or directory which may be concurrently
read/written please use atomic filesystem patterns. This usually involves
heavy use of `os.Rename`. Search for existing uses of `os.Rename` to see
examples.
