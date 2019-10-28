# LSIF server endpoints

This document outlines the endpoints provided by the HTTP server (running on port 3186 by default). These endpoints are proxied via the frontend under the path `/.api/lsif`. This server should not be directly accessible, as the server relies on the frontend to perform all request authentication.

## LSIF Endpoints

### POST `/upload?repository={repo}&commit={commit}`

- `repository`: the repository name
- `commit`: the 40-character commit hash

Receives an LSIF dump encoded as gzipped JSON lines. The request payload must be uploaded as form data with a single file (e.g. `curl -F "data=@data.lsif.gz" ...`).

Returns `204 No Content` on success.

### POST `/exists?repository={repo}&commit={commit}`

- `repository`: the repository name
- `commit`: the 40-character commit hash
- `file`: the root-relative file path

Determines if an LSIF dump exists that can answer queries for the given file.

### POST `/request?repository={repo}&commit={commit}`

- `repository`: the repository name
- `commit`: the 40-character commit hash

Available only for `reference` requests:

- `limit`: the maximum number of remote dumps to search
- `cursor`: a cursor generated from the previous page of results

Performs an definitions, references, or hover query at a particular position. The request body must be a JSON object with the following properties:

- `path`: the path of the document
- `position`: the zero-based `{ line, character }` hover position
- `method`: `definitions`, `references`, or `hover`

Returns `200 OK` on success with a body containing an LSP-compatible response. Returns `404 Not Found` if no LSIF data exists for this repository.
