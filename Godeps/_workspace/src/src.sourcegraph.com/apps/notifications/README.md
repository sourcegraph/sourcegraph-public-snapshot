# notifications

Notifications is a notification center app.

Installation
------------

```bash
go get -u src.sourcegraph.com/apps/notifications/...
```

Development
-----------

This project relies on `go generate` directives to process and statically embed assets. For development only, you'll need extra dependencies:

```bash
go get -u -d -tags=generate src.sourcegraph.com/apps/notifications/...
go get -u -d -tags=js src.sourcegraph.com/apps/notifications/...
```

Afterwards, you can build and run in development mode, where all assets are always read and processed from disk:

```bash
go build -tags=dev something/that/uses/notifications
```

When you're done with development, you should run `go generate` and commit that:

```bash
go generate src.sourcegraph.com/apps/notifications/...
```

License
-------

-	TODO.
