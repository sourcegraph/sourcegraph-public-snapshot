/**
 * Generates the TypeScript types for the JSON schemas.
 *
 * Usage: <schemaName>
 *
 *   schemaName - extensionless name of the schema.json file to generate types for
 */

const { compile: compileJSONSchema } = require('json-schema-to-typescript')
const { mkdir, readFile, writeFile } = require('mz/fs')
const path = require('path')

const schemaDirectory = path.join(__dirname, '..', '..', '..', '..', 'schema')
const outputDirectory = __dirname

/**
 * Allow json-schema-ref-parser to resolve the v7 draft of JSON Schema
 * using a local copy of the spec, enabling developers to run/develop Sourcegraph offline
 */
const draftV7resolver = {
  order: 1,
  read: () => readFile(path.join(schemaDirectory, 'json-schema-draft-07.schema.json')),
  canRead: file => file.url === 'http://json-schema.org/draft-07/schema',
}

async function main(args) {
  if (args.length !== 1) {
    console.error('Usage: node generateSchema.js <schemaName>')
    process.exit(1)
  }

  const schemaName = args[0]

  let schema = await readFile(path.join(schemaDirectory, `${schemaName}.schema.json`), 'utf8')

  // HACK: Rewrite absolute $refs to be relative. They need to be absolute for Monaco to resolve them
  // when the schema is in a oneOf (to be merged with extension schemas).
  schema = schema.replace(/https:\/\/sourcegraph\.com\/v1\/settings\.schema\.json#\/definitions\//g, '#/definitions/')

  const types = await compileJSONSchema(JSON.parse(schema), 'settings.schema', {
    cwd: schemaDirectory,
    $refOptions: {
      resolve: /** @type {import('json-schema-ref-parser').Options['resolve']} */ ({
        draftV7resolver,
        // there should be no reason to make network calls during this process,
        // and if there are we've broken env for offline devs/increased dev startup time
        http: false,
      }),
    },
  })

  await mkdir(outputDirectory, { recursive: true })
  await writeFile(path.join(outputDirectory, `${schemaName}.schema.d.ts`), types)
}

; (async () => await main(process.argv.slice(2)))()
