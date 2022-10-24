# Hax

This branch is a hack to stream LSIF indexes from the Dotcom codeintel-db into a GCS bucket that we can run PageRank over. To run, set the following envvars:

```bash
export SUPER_SECRET_FRONTEND_DSN='postgres://dev-readonly:{URL_ENCODED_PASSWORD}@localhost:5444/sg'
export SUPER_SECRET_CODEINTEL_DSN='postgres://dev-readonly:{URL_ENCODED_PASSWORD}@localhost:5333/sg'
```

Then run the binary in this directory:

```bash
go run main.go
```
