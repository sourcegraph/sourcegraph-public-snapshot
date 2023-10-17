// @ts-check

const gulp = require('gulp')

const { graphQlOperations, schema, watchGraphQlOperations } = require('./client/shared/gulpfile')
const { webpack: webWebpack, developmentServer, generate, watchGenerators } = require('./client/web/gulpfile')
const { buildSvelteKit } = require('./client/web-sveltekit/gulpfile.cjs')

/**
 * Generates files needed for builds whenever files change.
 */
const watchGenerate = gulp.series(generate, watchGenerators)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webWebpack)

const tasks = [watchGenerators, developmentServer]

if (process.env.SVELTEKIT) {
  tasks.push(buildSvelteKit)
}

/**
 * Watches everything and rebuilds on file changes.
 */
const development = gulp.series(generate, gulp.parallel(...tasks))

module.exports = {
  generate,
  watchGenerate,
  build,
  dev: development,
  schema,
  graphQlOperations,
  watchGraphQlOperations,
}
