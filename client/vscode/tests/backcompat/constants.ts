// Oldest version of Sourcegraph that v1 of VS Code extension search works with.
export const OLDEST_VERSION_TAG = 'v3.32.0'
export const DECOMPRESSED_SCHEMA_PATH = __dirname + `/${OLDEST_VERSION_TAG}-schema.graphql`
export const COMPRESSED_SCHEMA_PATH = `${DECOMPRESSED_SCHEMA_PATH}.gz`
