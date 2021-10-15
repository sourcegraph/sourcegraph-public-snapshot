# Quickstart step 2: Get the code

Run the following command in a folder where you want to keep a copy of the code. Command will create a new sub-folder (`sourcegraph`) in this folder.

```bash
git clone https://github.com/sourcegraph/sourcegraph.git
```

## For Sourcegraph employees: clone shared configuration

In order to run the local development environment as a Sourcegraph employee, you'll need to clone another repository: [`sourcegraph/dev-private`](https://github.com/sourcegraph/dev-private). It contains convenient preconfigured settings and code host connections.

It needs to be cloned into the same folder as `sourcegraph/sourcegraph`, so they sit alongside each other. To illustrate:

```
/dir
 |-- dev-private
 +-- sourcegraph
```

> NOTE: Ensure that you periodically pull the latest changes from [`sourcegraph/dev-private`](https://github.com/sourcegraph/dev-private) as the secrets are updated from time to time.

[< Previous](quickstart_1_install_dependencies.md) | [Next >](quickstart_3_install_sg.md)
