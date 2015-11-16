# Sourcegraph linux package creation

We have a simple `make` driven process to create redistributable Linux
packages (`rpm`, `deb`) from a `src` binary.

1. Create package with `sgtool package` in the top-level directory.
1. Run `make` or `make VERSION=0.9`.

Packages will appear in `./dist/`. You can test the packages with
`make test`. One should usually run something like

<pre>
$ make clean package test VERSION=0.9
</pre>

The packages installs:

* `/usr/bin/src` - The Sourcegraph binary
* `/usr/bin/daemonize.sourcegraph` - A vendored in version of
  [daemonize](http://software.clapper.org/daemonize/)
* `/etc/init.d/src` - A sysvinit script for starting up `src serve`

The packages themselves have no dependencies once installed.

You can read more about `sgtool` in `README.release.md` in the top-level
directory. This tool depends on docker to generate the packages.
