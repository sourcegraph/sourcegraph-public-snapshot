/**
 * Generates the TypeScript types for the JSON schemas.
 *
 * Usage: <schemaName1> <schemaName2> ...
 *
 * schemaName - extensionless name of the schema.json file to generate types for
 */

import { mkdir, readFile, writeFile } from 'fs/promises'
import path from 'path'

import { ResolverOptions } from 'json-schema-ref-parser'
import { compile as compileJSONSchema } from 'json-schema-to-typescript'

const schemaDirectory = path.join(__dirname, '..', '..', '..', 'schema')
const outputDirectory = path.join(__dirname, '..', 'src', 'schema')

/**
 * Allow json-schema-ref-parser to resolve the v7 draft of JSON Schema
 * using a local copy of the spec, enabling developers to run/develop Sourcegraph offline
 */
const draftV7resolver: ResolverOptions = {
    order: 1,
    read: () => readFile(path.join(schemaDirectory, 'json-schema-draft-07.schema.json')),
    canRead: file => file.url === 'http://json-schema.org/draft-07/schema',
}

export async function generateSchema(schemaName: string): Promise<void> {
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    let schema = await readFile(path.join(schemaDirectory, `${schemaName}.schema.json`), 'utf8')

    // HACK: Rewrite absolute $refs to be relative. They need to be absolute for Monaco to resolve them
    // when the schema is in a oneOf (to be merged with extension schemas).
    schema = schema.replaceAll('https://sourcegraph.com/v1/settings.schema.json#/definitions/', '#/definitions/')

    const types = await compileJSONSchema(JSON.parse(schema), 'settings.schema', {
        cwd: schemaDirectory,
        $refOptions: {
            resolve: {
                draftV7resolver,
                // there should be no reason to make network calls during this process,
                // and if there are we've broken env for offline devs/increased dev startup time
                http: false as any,
            },
        },
    })

    await mkdir(outputDirectory, { recursive: true })
    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    await writeFile(path.join(outputDirectory, `${schemaName}.schema.d.ts`), types)
}

if (require.main === module) {
    async function main(schemas: string[]) {
        if (schemas.length === 0) {
            throw new Error('Usage: <schemaName1> <schemaName2> ...')
        }

        for (const schema of schemas) {
            await generateSchema(schema)
        }
    }
    main(process.argv.slice(2)).catch(error => {
        console.error(error)
        process.exit(1)
    })
}
