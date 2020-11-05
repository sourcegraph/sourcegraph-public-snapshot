# Sourcegraph release tool

This directory contains scripts and code to automate our releases. Run `yarn run release`
to see a list of automated steps. Refer to
[the handbook](https://about.sourcegraph.com/handbook/engineering/releases) for details
on our release process and how this tool is used.

Run `yarn run build` to build the tool (_required_ alongside `yarn run release` to make
sure you are using the latest version of the tool), and `yarn run watch` to build on any
changes to files. Note that configuration changes requires rebuilds as well!
