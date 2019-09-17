# LSIF code intelligence

[LSIF](https://code.visualstudio.com/blogs/2019/02/19/lsif) is a file format that stores code intelligence information such as hover docstrings, definitions, and references.

Sourcegraph receives and stores LSIF files uploaded using [upload.sh](upload.sh) (usually used in CI, similar to [Codecov's Bash Uploader](https://docs.codecov.io/docs/about-the-codecov-bash-uploader)), then uses that information to provide fast and precise code intelligence when viewing files.

The HTTP [server](src/server.ts) runs behind Sourcegraph (for auth) and services requests for relevant LSP queries. LSIF dump uploads received by the server are stored in a temporary file to be asynchronously processed by the [worker](src/worker.ts). The server and worker communicate via [node-resque](https://github.com/taskrabbit/node-resque), a work broker powered by Redis.

## API

### `/upload`

Receives an LSIF dump encoded as JSON lines.

URL query parameters:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit

The request body must be HTML form data with a single file (e.g. `curl -F "data=@file.lsif" ...`).

### `/request`

Performs a `hover`, a `definitions`, or a `references` request for the given repository@commit and returns the result. Fails if there is no LSIF data for the given repository@commit.

The request body must be a JSON object with these properties:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit
- `method`: `hover`, `definitions`, or `references`
- `path`: the file path in the repository.
- `position`: the zero-based `{ line, character }` in the file at which the request is being made
