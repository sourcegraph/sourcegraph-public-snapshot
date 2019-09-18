# LSIF endpoints

This document outlines the endpoints provided by the HTTP server (running on port 3186 by default). These endpoints are proxied via the frontend under the path `/.api/lsif`. This server should not be directly accessibly, as the server relies on the frontend to perform all request authentication.

**Note:** The commit parameters are currently ignored; each repository has at most one LSIF data attached to it and that will be used regardless of the source or queried commit. This will soon be resolved.

## LSIF Endpoints

### POST `/upload?repository={repo}&commit={commit}`

- `repository`: the repository name
- `commit`: the 40-character commit hash

Receives an LSIF dump encoded as gzipped JSON lines. The request payload must be uploaded as form data with a single file (e.g. `curl -F "data=@data.lsif.gz" ...`). This endpoint is targeted by the [upload.sh](../upload.sh) script.

Returns `204 No Content` on success.

### POST `/request?repository={repo}&commit={commit}`

- `repository`: the repository name
- `commit`: the 40-character commit hash

Performs a query for a particular hover position. The request body must be a JSON object with the following properties:

- `method`: `definitions`, `references`, or `hover`
- `path`: the path of the document
- `position`: the zero-based `{ line, character }` hover position

Returns `200 OK` on success with a body containing an LSPS-compatible response. Returns `404 Not Found` if no LSIF data exists for this repository.

## Control Endpoints

The following endpoints allow uses to inspect the work queue in order to track LSIF conversion processes. This is necessary in order to triage unprocessable LSIF uploads due to bugs in LSIF indexers or unexpected properties of their output.

### GET `/active`

Retrieve the list of jobs that have been accepted by a worker but have not yet been completed (or failed).

### GET `/queued`

Retrieve the jobs which have been queued but have not yet been accepted by a worker.

### GET `/failed`

Retrieve the list of jobs that have been attempted but failed to process. This payload contains the date of failure as well as the error/exception message that occurred during processing.
