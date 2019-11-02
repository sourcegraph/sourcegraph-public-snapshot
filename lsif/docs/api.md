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

### GET `/dumps/{repository}?query={query}&limit={limit}&offset={offset}`

- `repository`: the repository name
- `query`: a search query (compares commits and root fields)
- `limit`: the maximum number of dumps to return
- `offset`: the number of dumps seen previously

Returns all dumps for a given repository.

### GET `/dumps/{repository}/{dumpId}

- `repo`: the repository name
- `dumpId`: the dump identifier

Returns a particular dump by its identifier. The resulting dump must belong to the given repository.

### GET `/jobs/stats`

Retrieve the current counts of jobs in each state.

### GET `/jobs/{state}?query={query}&limit={limit}&offset={offset}`

- `state`: the job state (`processing`, `errored`, `completed`, `queued`, or `scheduled`)
- `query`: a search query
- `limit`: the maximum number of jobs to return
- `offset`: the number of jobs seen previously

Returns the jobs with the given state. If a search term is given, then only jobs matching that search term are returned. If no search term is given, then the response will also contain a total count.

### GET `/jobs/{id}`

Returns a particular job by its identifier.
