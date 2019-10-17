# LSIF server endpoints

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

## Queue Endpoints

The following endpoints allow uses to inspect the work queue in order to track LSIF conversion processes. This is necessary in order to triage unprocessable LSIF uploads due to bugs in LSIF indexers or unexpected properties of their output.

### GET `/queue/stats`

Retrieve the counts of jobs in the queue.

### GET `/queue/:jobId`

Return the current state and progress of a job by its ID.

### GET `/queue/active`

Return active jobs.

### GET `/queue/completed`

Return completed jobs.

### GET `/queue/failed`

Return failed jobs.
