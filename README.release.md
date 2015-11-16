# Sourcegraph binary release process

1. `make dist-dep`
2. Ensure you have the AWS credentials set so that the AWS CLI (`aws`) can write to the `sourcegraph-release` bucket.
3. Run `sgtool release --public VERSION`, where `VERSION` is a number like `0.1.2`.

If you just want to build and package the binaries (and not publish a new release), use `sgtool package`.

## Deploy (dev)

To deploy Sourcegraph to https://src.sourcegraph.com, install the sgtool command (`godep go install ./sgtool`).

Run `aws s3 ls s3://sourcegraph-release/src/ | ./dev/src_version.py inc_patch` to get a new version number for your Sourcegraph release (that command figures out the current latest version, and increments the patch value by one).

Then run:

```bash
$ sgtool release {new-version-number}
$ sgtool deploy {new-version-number}
```

Important: always run `sgtool release` first or else you will upload an old binary to `src.sourcegraph.com`, which will most likely fail (in combination with updated config files).

On Linux kernels 2.6.17 or newer, if you encounter any `Connection reset by peer` or `Broken pipe` errors during `sgtool release` at the `aws s3 sync` command, you may need to [use this fix](http://scie.nti.st/2008/3/14/amazon-s3-and-connection-reset-by-peer/)

## Linux Packages

See `package/README.md`.
