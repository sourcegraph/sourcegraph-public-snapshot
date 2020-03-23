 # Non-Git code hosts (Perforce, Mercurial, Subversion, raw text, etc.)

Sourcegraph natively supports all Git-based Version Control Systems (VCSs) and code hosts. For non-Git code hosts, Sourcegraph provides a tool called `src-expose` to periodically sync and continuously serve local directories as Git repositories over HTTP. 

**Guides for specific non-Git code hosts**:
- [Perforce](../repo/perforce.md)

## Installing `src-expose`

Navigate to the directory that contains the Git repositories that you want to serve, then run:

``` shell
wget https://storage.googleapis.com/sourcegraph-artifacts/src-expose/latest/darwin-amd64/src-expose
# For linux comment the above and uncomment the below
# wget https://storage.googleapis.com/sourcegraph-artifacts/src-expose/latest/linux-amd64/src-expose

chmod +x src-expose
```

You can run `src-expose -h` any time for help. 

## Using `src-expose`

`src-expose` provides two main functions:

- Serving local Git repositories over the network, and making them available to Sourcegraph (as if they were available on a traditional code host). See [serving repositories](#serving-repositories), or run `src-expose serve -h`.
- Periodically running a command to sync changes to the code, and then combining those changes into a new Git commit in the local repositories. See [syncing repositories](#syncing-repositories), or run `src-expose sync -h`.

By default, if you exclude the first argument to `src-expose` (i.e., you run `src-expose` without `sync` or `serve`), it will perform both of these functions simultaneously.

### Serving repositories

`src-expose` serves a list of local Git repositories over HTTP, making them available to Sourcegraph. In addition to simply providing a Git endpoint, it also provides a repository listing API that Sourcegraph expects a code host to have. 

If you wish to serve a local directory without running any syncing commands automatically, you can run `src-expose serve` (instead of the default `src-expose`) to only perform this function.

### Syncing repositories

In addition to serving a local directory, `src-expose` will periodically run a command of your choice to fetch changes from a remote and combine them into a new Git commit. 

For example, if your provided configuration YAML file contains:

```
# before is a command run before sync. before is run from root.
before: p4 sync
# duration defines how often sync should happen. Defaults to 10s.
duration: 10s
```

Then Sourcegraph will run `p4 sync` every 10 seconds, and combine all of the fetched changes into a new Git commit. The new Git commit's author will be `src-expose`, and will contain all changes since the last time the syncing command was run.

While this syncing functionality means that the original change history will be lost, it eliminates any slow and costly Perforce-to-Git or Hg-to-Git or similar conversions that would otherwise be required. If you prefer to retain history, see [choosing the right src-expose setup](#choosing-the-rigth-src-expose-setup).

## Quickstart

1. Start up a Sourcegraph instance (using our [Quickstart](../../../index.md) or our [full installation documentation](../../install/index.md)).

1. [Install `src-expose`](#installing-src-expose)

1. Pick the directory you want to export from, then run:

``` shell
./src-expose dir1 dir2 dir3
```

or

``` shell
./src-expose serve dir1 dir2 dir3
```

depending on whether you want to automatically sync changes and serve the local directories, or just serve the local directories.

1. `src-expose` will output a configuration to use. It may scroll by quickly due to logging, so if so, just scroll up. However, this configuration should work:

``` json
 {
    // url is the http url to src-expose (listening on 127.0.0.1:3434)
    // url should be reachable by Sourcegraph.
    // "http://host.docker.internal:3434" works from Sourcegraph when using Docker for Desktop.
    "url": "http://host.docker.internal:3434",
    // repos should have the special value ("src-expose") below, and it will pull all of the repositories that src-expose is serving.
    "repos": ["src-expose"]
}
```

**IMPORTANT:** If you are using a Linux host machine, replace `host.docker.internal` in the above with the IP address of your actual host machine because `host.docker.internal` [does not work on Linux](https://github.com/docker/for-linux/issues/264). You should use the network-accessible IP shown by `ifconfig` (not e.g. 127.0.0.1 or localhost).

Go to **Admin > Manage Repositories > Add repositories > Single Git repositories**. Input the above configuration. Your directories should now be syncing in Sourcegraph.

### Next steps: advanced configuration

Please consult `src-expose -help` to learn more about the options available. 

For more complex setups, configure your `src-expose` by providing a local configuration file:

``` shell
src-expose -snapshot-config config.yaml
```

See [an example YAML file containing available configuration options](https://github.com/sourcegraph/sourcegraph/blob/master/dev/src-expose/examples/example.yaml). 

## Choosing the right `src-expose` setup

`src-expose` provides a spectrum of ways to access non-Git code in Sourcegraph, trading off freshness of the code against completeness of the code history. 

**Retaining code change history**

If you have a small enough code base, it is possible to use a tool for converting the full non-Git code host's history into Git history, and to keep them synced. For example, for Perforce, by using `git p4 sync` (which converts the full Perforce change history into Git commits) rather than using `p4 sync` in conjunction with `src-expose` to squash all changes into a single new commit (see the [Syncing repositories](#syncing-repositories) section above).

To achieve this 

**Retaining code change history**

