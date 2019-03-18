# Sourcegraph JSON Schemas

[JSON Schema](http://json-schema.org/) is a way to define the structure of a JSON document. It enables typechecking and code intelligence on JSON documents.

Sourcegraph uses the following JSON Schemas:

- [`settings.schema.json`](./settings.schema.json)
- [`site.schema.json`](./site.schema.json)
- [`extension.schema.json`](https://github.com/sourcegraph/extensions-client-common/blob/master/src/schema/extension.schema.json) is manually copied to this directory as needed. Only the subset of properties and definitions used by our Go code is needed. The web app uses the `extension.schema.json` file from the `@sourcegraph/extensions-client-common` npm package (the Go code currently doesn't use the file from this npm package because that would require running `yarn` in all Go tests in CI, which would be slow).

# Modifying a schema

1.  Edit the `*.schema.json` file in this directory.
1.  Run `go generate` to update the `*_stringdata.json` file.
1.  Commit the changes to both files.
1.  When the change is ready for release, [update the documentation](https://github.com/sourcegraph/website/blob/master/README.md#documentation-pages).

## Known issues

- The JSON Schema IDs (URIs) are of the form `https://sourcegraph.com/v1/*.schema.json#`, but these are not actually valid URLs. This means you generally need to supply them to JSON Schema validation libraries manually instead of having the validator fetch the schema from the web.
