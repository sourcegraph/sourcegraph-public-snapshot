Issues
======

This is an example app built on the Sourcegraph platform.

It registers an HTTP handler on the Repo Tab integration point in the platform, which lets it inject its own HTML directly into the Sourcegraph UI as a tab on the repository page.

### Running

Import this package for side effects in the Sourcegraph main repository in the package`sourcegraph.com/sourcegraph/sourcegraph/sgx`. When the Sourcegraph binary is run, an Issues tab should appear on every repository page.

During development, frontend assets are accessed from disk directly. Run:

```
cd $GOPATH/src/sourcegraph.com/sourcegraph/issues
go build -tags=dev
```

For production, run the following to generate assets that will be compiled into the Sourcegraph binary:
```
go generate sourcegraph.com/sourcegraph/issues/...
go build sourcegraph.com/sourcegraph/issues
```
