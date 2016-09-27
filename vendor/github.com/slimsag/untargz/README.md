# Untargz

Is a simple library to untar a tar.gz file from an `io.Reader`.

# Usage

```
import "github.com/slimsag/untargz"

...
err := untargz.Extract(r, "destination/folder", nil)
...
```

# Credit

This is really just a small subset of the code in https://github.com/mholt/archiver -- because it doesn't support providing an `io.Reader` (it requires an on-disk filename). As such, all credit for this goes to @mholt.
