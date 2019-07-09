# LSIF code intelligence

BEFOER MERGING consider moving some of this to docs now or later

[LSIF](https://code.visualstudio.com/blogs/2019/02/19/lsif) is a file format that stores code intelligence information such as hover docstrings, definitions, and references.

Sourcegraph receives and stores LSIF files uploaded using [upload-lsif.sh](upload-lsif.sh), then uses that information to provide fast and precise code intelligence when viewing files.

In this directory:

- [upload-lsif.sh](upload-lsif.sh): a script that uploads an LSIF file to Sourcegraph (usually used in CI, similar to [Codecov's Bash Uploader](https://docs.codecov.io/docs/about-the-codecov-bash-uploader))
- [server/](server/): an HTTP server which runs inside of Sourcegraph (for auth), receives and stores LSIF file uploads, and services requests for hovers/defs/refs
- [extension/](extension/): a Sourcegraph extension which sends requests to the Sourcegraph instance at `.api/lsif` for hovers/defs/refs
