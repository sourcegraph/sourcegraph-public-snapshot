# vendored std regexp/syntax

This package contains a vendored copy of std regexp/syntax. However, it only
contains the code for converting syntax.Regexp into a String. It is the
version of the code at a recent go commit, but with a commit which introduces
a significant performance regression reverted.

At the time of writing regexp.String on go1.22 is taking 40% of CPU at
Sourcegraph. This should return to ~0% with this vendored code.

https://github.com/sourcegraph/sourcegraph/issues/61462

## Vendored commit

```
commit 2e1003e2f7e42efc5771812b9ee6ed264803796c
Author: Daniel Martí <mvdan@mvdan.cc>
Date:   Tue Mar 26 22:59:41 2024 +0200

    cmd/go: replace reflect.DeepEqual with slices.Equal and maps.Equal

    All of these maps and slices are made up of comparable types,
    so we can avoid the overhead of reflection entirely.

    Change-Id: If77dbe648a336ba729c171e84c9ff3f7e160297d
    Reviewed-on: https://go-review.googlesource.com/c/go/+/574597
    Reviewed-by: Than McIntosh <thanm@google.com>
    LUCI-TryBot-Result: Go LUCI <golang-scoped@luci-project-accounts.iam.gserviceaccount.com>
    Reviewed-by: Ian Lance Taylor <iant@google.com>
```

## Reverted commit

```
commit 98c9f271d67b501ecf2ce995539abd2cdc81d505
Author: Russ Cox <rsc@golang.org>
Date:   Wed Jun 28 17:45:26 2023 -0400

    regexp/syntax: use more compact Regexp.String output

    Compact the Regexp.String output. It was only ever intended for debugging,
    but there are at least some uses in the wild where regexps are built up
    using regexp/syntax and then formatted using the String method.
    Compact the output to help that use case. Specifically:

     - Compact 2-element character class ranges: [a-b] -> [ab].
     - Aggregate flags: (?i:A)(?i:B)*(?i:C)|(?i:D)?(?i:E) -> (?i:AB*C|D?E).

    Fixes #57950.

    Change-Id: I1161d0e3aa6c3ae5a302677032bb7cd55caae5fb
    Reviewed-on: https://go-review.googlesource.com/c/go/+/507015
    TryBot-Result: Gopher Robot <gobot@golang.org>
    Reviewed-by: Than McIntosh <thanm@google.com>
    Run-TryBot: Russ Cox <rsc@golang.org>
    Reviewed-by: Rob Pike <r@golang.org>
    Auto-Submit: Russ Cox <rsc@golang.org>
```
