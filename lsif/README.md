# LSIF code intelligence

[LSIF](https://code.visualstudio.com/blogs/2019/02/19/lsif) is a file format that stores code intelligence information such as hover docstrings, definitions, and references.

Sourcegraph receives and stores LSIF files uploaded using [upload-lsif.sh](upload-lsif.sh) (usually used in CI, similar to [Codecov's Bash Uploader](https://docs.codecov.io/docs/about-the-codecov-bash-uploader)), then uses that information to provide fast and precise code intelligence when viewing files.

The HTTP [server](src) runs behind Sourcegraph (for auth) and receives and stores LSIF dump uploads and services requests for relevant LSP queries.

## API

_`/upload`_

Receives an LSIF dump encoded as JSON lines.

URL query parameters:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit

The request body must be HTML form data with a single file (e.g. `curl -F "data=@file.lsif" ...`).

---

The remaining methods perform LSP-like requests for the provided repository@commit and returns the result. Fails if there is no LSIF data for the given repository@commit. The request body must be a JSON object with these properties:

- `repository`: the name of the repository (e.g. `github.com/sourcegraph/codeintellify`)
- `commit`: the 40 character hash of the commit
- `path`: the file path in the repository.
- `position`: the `{ line, character }` in the file at which the request is being made

The reponse payload is in the `data` field of the JSON response.

_`/definitions`_

Performs a definitions request for the given repository@commit and returns the result.

_`/references`_

Performs a references request for the given repository@commit and returns the result. This method additionally takes a `page` query parameter indicating the page of results when searching for global references. The `nextPage` key of the JSON response indicates the value of the subsequent page in the result set. If this key is `null`, then no pages remain.

_`/hover`_

Performs a hover request for the given repository@commit and returns the result.
