# App troubleshooting guide

This page helps with troubleshooting steps to run before filing App bugs on the [issue tracker](https://github.com/sourcegraph/sourcegraph/issues) or asking in our [Discord](https://discord.gg/s2qDtYGnAE).

## Installation

### App failing to start

If you are seeing the following in the log output:

```
The target schema is marked as dirty and no other migration operation is seen running on this schema. The last migration operation over this schema has failed (or, at least, the migrator instance issuing that migration has died). Please contact support@sourcegraph.com for further assistance.
```
You could have a corrupted database that will block the App instance from running. 

To fix this:

1. Stop/kill the running instance of Sourcegraph App. This can be done by quitting the terminal.
2. Remove all Sourcegraph data by running:
### In MacOS
```
rm -rf $HOME/.sourcegraph-psql
rm -rf $HOME/Library/Application\ Support/sourcegraph-sp
rm -rf $HOME/Library/Caches/sourcegraph-sp
```
### In Linux
```
rm -rf $HOME/.sourcegraph-psql
rm -rf $XDG_CACHE_HOME/sourcegraph-sp
rm -rf $XDG_CONFIG_HOME/sourcegraph-sp
rm -rf $HOME/.cache/sourcegraph-sp
rm -rf $HOME/.config/sourcegraph-sp
```
3. Re-install App using the binary package downloaded earlier.

>Note: Ensure you have a running Docker daemon on your machine.