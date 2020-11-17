# Sourcegraph release tool

This directory contains scripts and code to automate our releases. Refer to
[the handbook](https://about.sourcegraph.com/handbook/engineering/releases) for details
on our release process and how this tool is used.

To see all available steps:

```sh
yarn build
yarn run release help # add 'all' to see test commands as well
```

Run `yarn run build` to build the tool (_required_ alongside `yarn run release` to make
sure you are using the latest version of the tool), and `yarn run watch` to build on any
changes to files.

Before using this tool, please also verify that the [release configuration](./release-config.jsonc)
is set up correctly.
