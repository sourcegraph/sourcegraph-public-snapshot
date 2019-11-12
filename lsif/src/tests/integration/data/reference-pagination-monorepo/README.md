# Global reference pagination for monorepos test data

This generates several repositories of three types. Dependencies between repositories are ensured via symlinks in the `node_modules` directory of dependent repositories.

### a

This repository provides the `math-util` package containing a definition for an `add` function.

### {b,c,d,e,f}-ref

These repositories have a dependency on the `math-util` package and imports the `add` function. The number of references is fewer than a single test page size.

### {b,c,d,e,f}-noref

These repositories do not have a dependency on the `math-util` package.
