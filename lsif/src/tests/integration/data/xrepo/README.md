# Cross-repository test data

This generates several repositories of three types. Dependencies between repositories are ensured via symlinks in the `node_modules` directory of dependent repositories.

### a

This repository provides the `math-util` package containing definitions for an `add` and a `mul` function. The latter function is used by the former (and thus contains a reference to it).

### b{1,2,3}

These repositories have a dependency on the `math-util` package and imports the `add` and `mul` functions.

### c{1,2,3}

These repositories have a dependency on the `math-util` package and imports only the `add` function.

This generates a repository `main` that produces linked reference results that occur when implementing methods from an interface or base class.

The generated LSIF data covers a regression of a mission feature where the LSIF correlator would not correctly link `item` edges between two reference results (previously, it only correctly linked a reference result and a range or result set).
