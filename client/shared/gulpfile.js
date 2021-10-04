// @ts-check

const path = require('path')

const { generateNamespace } = require('@gql2ts/from-schema')
const { DEFAULT_OPTIONS, DEFAULT_TYPE_MAP } = require('@gql2ts/language-typescript')
const glob = require('glob')
const { buildSchema, introspectionFromSchema } = require('graphql')
const gulp = require('gulp')
const { compile: compileJSONSchema } = require('json-schema-to-typescript')
const { readFile, writeFile, mkdir } = require('mz/fs')
const { format, resolveConfig } = require('prettier')

const { cssModulesTypings, watchCSSModulesTypings } = require('./dev/generateCssModulesTypes')
const { generateGraphQlOperations, ALL_DOCUMENTS_GLOB } = require('./dev/generateGraphQlOperations')

const GRAPHQL_SCHEMA_GLOB = path.join(__dirname, '../../cmd/frontend/graphqlbackend/*.graphql')

/**
 * Generates the TypeScript types for the GraphQL schema.
 * These are used by older code, new code should rely on the new query-specific generated types.
 *
 * @returns {Promise<void>}
 */
async function graphQlSchema() {
  const schemaFiles = glob.sync(GRAPHQL_SCHEMA_GLOB)
  let combinedSchema = ''
  for (const schemaPath of schemaFiles) {
    const schemaString = await readFile(schemaPath, 'utf8')
    combinedSchema += `\n${schemaString}`
  }
  const schema = buildSchema(combinedSchema)

  const result = introspectionFromSchema(schema)

  const formatOptions = await resolveConfig(__dirname, { config: __dirname + '/../../prettier.config.js' })
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
  await writeFile(__dirname + '/src/graphql/schema.ts', typings)
}

/**
 * Generates the legacy graphql.ts types on file changes.
 */
async function watchGraphQlSchema() {
  await new Promise((resolve, reject) => {
    gulp.watch(GRAPHQL_SCHEMA_GLOB, graphQlSchema).on('error', reject)
  })
}

/**
 * Determine whether to regenerate GraphQL operations based on the given
 * Chokidar event. If we can determine that the file being modified is a
 * non-GraphQL-using TypeScript or JavaScript file, then we can skip the
 * expensive generation step.
 *
 * @param {string} type
 * @param {string} name
 * @returns bool
 */
async function shouldRegenerateGraphQlOperations(type, name) {
  if (type === 'unlink' || type === 'unlinkDir') {
    // For all deletions, we'll regenerate, since we don't know if the file(s)
    // that were deleted were used when generating the GraphQL operations.
    return true
  }

  // If we're watching a JavaScript or TypeScript file, then we should only
  // regenerate if there are gql-tagged strings. But first, we have to figure
  // out if it is that type of file.
  const isJS = ['.tsx', '.ts', '.jsx', '.js'].reduce((previous, extension) => {
    if (previous) {
      return previous
    }
    return name.endsWith(extension)
  }, false)
  if (isJS) {
    // Look for the tagged string in the most naïve way imaginable.
    return (await readFile(name)).includes('gql`')
  }

  // Finally, for non-JavaScript/TypeScript files, we'll be safe and always
  // regenerate.
  return true
}

/**
 * Generates the new query-specific types on file changes.
 */
function watchGraphQlOperations() {
  // Although graphql-codegen has watching capabilities, they don't appear to
  // use chokidar correctly and rely on polling. Instead, let's get gulp to
  // watch for us, since we know it'll do it more efficiently, and then we can
  // trigger the code generation more selectively.
  return gulp
    .watch(ALL_DOCUMENTS_GLOB, {
      ignored: /** @param {string} name */ name => name.endsWith('graphql-operations.ts'),
    })
    .on('all', (type, name) => {
      ;(async () => {
        if (await shouldRegenerateGraphQlOperations(type, name)) {
          console.log('Regenerating GraphQL types')
          await generateGraphQlOperations()
          console.log('Done regenerating GraphQL types')
        }
      })().catch(error => {
        console.error(error)
      })
    })
}

/**
 * Generates the new query-specific types.
 */
async function graphQlOperations() {
  await generateGraphQlOperations()
}

/**
 * Allow json-schema-ref-parser to resolve the v7 draft of JSON Schema
 * using a local copy of the spec, enabling developers to run/develop Sourcegraph offline
 */
const draftV7resolver = {
  order: 1,
  read: () => readFile(path.join(__dirname, '../../schema/json-schema-draft-07.schema.json')),
  canRead: file => file.url === 'http://json-schema.org/draft-07/schema',
}

/**
 * Generates the TypeScript types for the JSON schemas.
 *
 * @returns {Promise<void>}
 */
async function schema() {
  const outputDirectory = path.join(__dirname, '..', 'web', 'src', 'schema')
  await mkdir(outputDirectory, { recursive: true })
  const schemaDirectory = path.join(__dirname, '..', '..', 'schema')
  await Promise.all(
    ['json-schema-draft-07', 'settings', 'site', 'batch_spec'].map(async file => {
      let schema = await readFile(path.join(schemaDirectory, `${file}.schema.json`), 'utf8')
      // HACK: Rewrite absolute $refs to be relative. They need to be absolute for Monaco to resolve them
      // when the schema is in a oneOf (to be merged with extension schemas).
      schema = schema.replace(
        /https:\/\/sourcegraph\.com\/v1\/settings\.schema\.json#\/definitions\//g,
        '#/definitions/'
      )

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
      await writeFile(path.join(outputDirectory, `${file}.schema.d.ts`), types)
    })
  )
}

function watchSchema() {
  return gulp.watch(path.join(__dirname, '../schema/*.schema.json'), schema)
}

module.exports = {
  watchSchema,
  schema,
  graphQlSchema,
  watchGraphQlSchema,
  graphQlOperations,
  watchGraphQlOperations,
  cssModulesTypings,
  watchCSSModulesTypings,
}
