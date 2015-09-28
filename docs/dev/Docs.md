+++
title = "Documentation policy"
+++

All packages and exported identifiers should be documented.

To measure documentation coverage, run:

```
go install src.sourcegraph.com/sourcegraph/dev/doccover
doccover ./...
```

Run with the `-all` flag to print the names of undocumented packages
and identifiers.
