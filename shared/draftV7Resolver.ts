import fs from 'fs'
import { ResolverOptions } from 'json-schema-ref-parser'
import path from 'path'

/**
 * Allow json-schema-ref-parser to resolve the v7 draft of JSON Schema
 * using a local copy of the spec, enabling developers to run/develop Sourcegraph offline
 */
export const draftV7resolver: ResolverOptions = {
    order: 1,
    read: () => fs.readFileSync(path.join(__dirname, '../schema/json-schema-draft-07.schema.json')),
    canRead: file => file.url === 'http://json-schema.org/draft-07/schema',
}
