# Importing the Sourcegraph API module

In order to use the Sourcegraph extension API, your extension needs to import
the `sourcegraph` module at the top of the file.

This import will look like this:

```typescript
import * as sourcegraph from 'sourcegraph'
```

Or, with the equivalent `require` syntax:

```typescript
const sourcegraph = require('sourcegraph')
```

## How the import works

Unlike a regular module that is installed with `npm` or `yarn` and is packaged
or bundled with your project, the `sourcegraph` module is actually injected
dynamically when your extension is loaded within a Sourcegraph instance.

The `sourcegraph` module that you install as a `package.json` dependency is
actually a set of TypeScript type definitions that allow you to use the
Sourcegraph API's interfaces. The module doesn't contain the implementations of
these interfaces.

## Testing

Because the `sourcegraph` module doesn't actually contain any implementation
code, a separate helper module is provided to create stubs of the API for
testing:
[@sourcegraph/extension-api-stubs](https://github.com/sourcegraph/extension-api-stubs).

See [Testing extensions](testing_extensions.md) for more about writing automated
tests for your extension.
