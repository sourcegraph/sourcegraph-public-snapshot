# Life of an LSIF upload

This document describes how an LSIF data file is uploaded to a Sourcegraph instance and processed. This document does **not** cover the data file generation, which is covered in the [user docs](https://docs.sourcegraph.com/user/code_intelligence/lsif), on [lsif.dev](https://lsif.dev), and in the documentation for individual indexers.

## Uploading

Data files are uploaded via the [lsif upload](https://sourcegraph.com/github.com/sourcegraph/src-cli/-/blob/cmd/src/lsif_upload.go) command in the Sourcegraph command line utility. This command gzip-encodes the file on-disk and sends it to the Sourcegraph instance via an unauthenticated HTTP POST. This request is handled by a [proxy handler](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22func+uploadProxyHandler%28%22), which will redirect the file to [lsif-server](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/lsif) via the [lsif-server client](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22%29+Upload%28%22+file:lsifserver/.*.go).

This handler exists only in the enterprise version of the product. The OSS version does not register this route and will return 404 for all requests.

Prior to proxying the payload to the lsif-server, the frontend will ensure that the target repository is cloned and the target [commit exists](https://sourcegraph.com/search?q=repo:^github\.com/sourcegraph/sourcegraph%24+"%29+ResolveRev%28"). This latter operation may may cause a remote fetch to occur in gitserver.

Additionally, this endpoint will authorize a request via a code host token when `LsifEnforceAuth` is true in the site settings. This is enabled in particular for the dot-com deployment so that LSIF uploads to a public repository are only allowed from requests using an access token with collaborator access to that repository. It is not generally expected for a private instance to enable this setting. See an example of the auth flow [here](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22func+enforceAuthGithub%28%22).

## Processing

Once the upload payload is received via the lsif-server [upload endpoint](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22%27/upload%27%22+file:lsif/.*/routes/.*.ts), it is written to disk and an _unprocessed_ LSIF [upload record](https://sourcegraph.com/search?q=repo:^github\.com/sourcegraph/sourcegraph%24+"class+LsifUpload"+file:lsif/.*.ts) is added to the `lsif_uploads` table in Postgres.

Each upload record has a state which can be one of the following:

- queued
- processing
- completed
- errored

The [lsif-dump-processor](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22Selected+upload+to+convert%22) process will poll for _queued_ uploads. Once it selects (and locks) an upload for processing, it sets its state temporarily to _processing_, converts the raw LSIF input on disk into a SQLite database that can be used by the lsif-server to answer code intelligence queries. On success, the upload's state is set to _completed_. On failure, the upload's state is set to _errored_ along with an error message and a stacktrace. An upload in the _completed_ is visible to the lsif-server to answer queries.

See [life of a code intelligence query](life-of-a-code-intelligence-query.md) for additional documentation on how the SQLite data file is read.
