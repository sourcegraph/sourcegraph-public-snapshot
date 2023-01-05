// @ts-check

const path = require('path')

const gulp = require('gulp')
const { readFile } = require('mz/fs')

const { cssModulesTypings, watchCSSModulesTypings } = require('./dev/generateCssModulesTypes')
const { generateGraphQlOperations, ALL_DOCUMENTS_GLOB } = require('./dev/generateGraphQlOperations')
const { generateSchema } = require('./dev/generateSchema')

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
  if (process.env.DEV_WEB_BUILDER_UNSAFE_FAST) {
    // Setting the env var DEV_WEB_BUILDER_UNSAFE_FAST skips various operations in frontend dev.
    // It's not safe, but if you know what you're doing, go ahead and use it. (CI will catch any
    // issues you forgot about.)
    return
  }

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
 * Generates the TypeScript types for the JSON schemas.
 *
 * @returns {Promise<void>}
 */
async function schema() {
  await Promise.all(
    ['json-schema-draft-07', 'settings', 'site', 'batch_spec'].map(async name => {
      await generateSchema(name)
    })
  )
}

function watchSchema() {
  return gulp.watch(path.join(__dirname, '../schema/*.schema.json'), schema)
}

module.exports = {
  watchSchema,
  schema,
  graphQlOperations,
  watchGraphQlOperations,
  cssModulesTypings,
  watchCSSModulesTypings,
}
