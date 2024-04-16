# API

## `/`

- `POST` to `/` with `Content-Type: application/json`. The following fields are required:
  - `filepath` string, e.g. `the/file.go` or `file.go` or `Dockerfile`, see "Supported file extensions" section below.
  - `theme` string, e.g. `Solarized (dark)`, see "Embedded themes" section below.
  - `code` string, i.e. the literal code to highlight.
- The response is a JSON object of either:
  - A successful response (`data` field):
    - `data` string with syntax highlighted response. The input `code` string [is properly escaped](https://github.com/sourcegraph/syntect_server/blob/ee3810f70e5701b961b7249393dbac8914c162ce/syntect/src/html.rs#L6) and as such can be directly rendered in the browser safely.
    - `plaintext` boolean indicating whether a syntax could not be found for the file and instead it was rendered as plain text.
  - An error response (`error` field), one of:
    - `{"error": "resource not found", "code": "resource_not_found"}`
- `GET` to `/health` to receive an `OK` health check response / ensure the service is alive.

## `/lsif`

Returns base64-encoded SCIP document.

To be deprecated and removed

## `/scip`

Returns base64-encoded SCIP document
