# TODOs

## Simplify build data

Build data is currently only ever stored on local filesystems, and
it's unlikely that will change (because any virtual/remote access
would be to the store, not to the build data tree). That means it's
probably unnecessary to use a VFS interface to access it.
