+++
title = "Ignoring Generated Files"
+++

If your code-base has a lot of auto-generated files that need to be updated
along with your changes, but do not need to be reviewed, it can clutter up the
diff and make it hard to review the important changes.

Sourcegraph has a special `.srcignore` file which can be used to list files that
should by default have their diffs suppressed in changesets for easier review.

# srcignore Location

Sourcegraph looks for the file in two different locations, and falls back to
the default if there is no such file:

1. The repository root itself (`myrepo/.srcignore`).
2. If such file exists in the repository, the global file at `$SGPATH/.srcignore`
   (i.e. `~/.sourcegraph/.srcignore`) will be used.

# Default srcignore

Sourcegraph will use the following `.srcignore` file if there is no repository
specific or global one available:

```
# Lines starting with a hash are ignored.

# Go
Godeps/*
*.gen.go
*.pb.go
*.pb_mock.go
```

# Syntax

The file paths are always '/'-separated ones, and the syntax simply causes `*`
to match any string, including the empty string and strings containing slashes.
For example:

| Pattern      | Description                                           |
|--------------|-------------------------------------------------------|
| `foo/*`      | Matches all files under the `foo` directory.          |
| `*.go`       | Matches all files with the `.go` extension.           |
| `*hello.txt` | Matches all files named `hello.txt` in any directory. |
