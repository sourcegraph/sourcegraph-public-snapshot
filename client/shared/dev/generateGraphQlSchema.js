/**
 * Generates the TypeScript types for the GraphQL schema.
 * These are used by older code, new code should rely on the new query-specific generated types.
 *
 * Usage: <outputFile>
 * - outputFile - filename to write types to
 */
const path = require('path')

const { generateNamespace } = require('@gql2ts/from-schema')
const { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } = require('@gql2ts/language-typescript')
const glob = require('glob')
const { buildSchema, introspectionFromSchema } = require('graphql')
const { mkdir, readFile, writeFile } = require('mz/fs')
const { format, resolveConfig } = require('prettier')

const GRAPHQL_SCHEMA_GLOB = path.join(__dirname, '../../../cmd/frontend/graphqlbackend/*.graphql')

async function main(args) {
  if (args.length !== 1) {
    throw new Error('Usage: <outputFile>')
  }

  const outputFile = args[0]
  await graphQlSchema(outputFile)
}

async function graphQlSchema(outputFile) {
  const schemaFiles = glob.sync(GRAPHQL_SCHEMA_GLOB)
  let combinedSchema = ''
  for (const schemaPath of schemaFiles) {
    const schemaString = await readFile(schemaPath, 'utf8')
    combinedSchema += `\n${schemaString}`
  }
  const schema = buildSchema(combinedSchema)

  const result = introspectionFromSchema(schema)

  const formatOptions = await resolveConfig(__dirname, { config: __dirname + '/../../../prettier.config.js' })
  const typings =
    'export type ID = string\n' +
    'export type GitObjectID = string\n' +
    'export type DateTime = string\n' +
    'export type JSONCString = string\n' +
    '\n' +
    generateNamespace(
      '',
      result,
      {
        typeMap: {
          ...DEFAULT_TYPE_MAP,
          ID: 'ID',
          GitObjectID: 'GitObjectID',
          DateTime: 'DateTime',
          JSONCString: 'JSONCString',
        },
      },
      {
        generateNamespace: (name, interfaces) => interfaces,
        interfaceBuilder: (name, body) => `export ${DEFAULT_OPTIONS.interfaceBuilder(name, body)}`,
        enumTypeBuilder: (name, values) =>
          `export ${DEFAULT_OPTIONS.enumTypeBuilder(name, values).replace(/^const enum/, 'enum')}`,
        typeBuilder: (name, body) => `export ${DEFAULT_OPTIONS.typeBuilder(name, body)}`,
        wrapList: type => `${type}[]`,
        postProcessor: code => format(code, { ...formatOptions, parser: 'typescript' }),
      }
    )
  await mkdir(path.dirname(outputFile), { recursive: true })
  await writeFile(outputFile, typings)
}

// Entry point used by Bazel binary
if (require.main === module) {
  main(process.argv.slice(2)).catch(error => {
    console.error(error)
    process.exit(1)
  })
}

module.exports = {
  graphQlSchema,
}
