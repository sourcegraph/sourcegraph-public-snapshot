# LSIF HTTP server

This is an HTTP server on top of https://github.com/Microsoft/vscode-lsif-extension. Since there's currently no npm release of vscode-lsif-extension, the relevant files have been copied here with only trivial modifications to make it pass type checking and linting in this repository:

- [src/database.ts](src/database.ts)
- [src/json.ts](src/json.ts)
- [src/files.ts](src/files.ts)

The only new file is [src/http-server.ts](src/http-server.ts), which is a Node.js Express HTTP server with the following API:

## API

### `/upload`

Receives a file upload, and stores it on disk. Files that are too big are rejected. If the max disk usage has been reached, old files (based on upload time) get deleted to free up space.

URL query parameters:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit

The request body must be HTML form data with a single file (e.g. `curl -F "data=@file.lsif" ...`).

### `/request`

Performs a `hover`, `definitions`, or `references` request on the LSIF file for the given repository@commit and returns the result. Fails if there is no LSIF data for the given repository@commit. Internally, it maintains an LRU cache of open `Database`s for speed and evicts old ones to avoid running out of memory.

The request body must be a JSON object with these properties:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit
- `method`: `hover`, `definitions`, or `references`

## LSIF file storage details

LSIF files are stored on disk with the following naming convention:

```
base64repository:$BASE_64_REPOSITORY,commit:$40_CHAR_HASH.lsif
```

For example, for `github.com/sourcegraph/codeintellify` at commit `c21c0da7b2a6cacafcbf90c85a81bf432020ad9b`:

```
base64repository:Z2l0aHViLmNvbS9zb3VyY2VncmFwaC9jb2RlaW50ZWxsaWZ5,commit:c21c0da7b2a6cacafcbf90c85a81bf432020ad9b.lsif
```
