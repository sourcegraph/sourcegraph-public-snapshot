// @ts-check

const { spawn } = require('child_process')
const gulp = require('gulp')
const path = require('path')

function build() {
  return spawn('yarn', ['-s', 'run', 'build'], {
    stdio: 'inherit',
    shell: true,
    env: { ...process.env, NODE_OPTIONS: '--max_old_space_size=8192' },
  })
}

function watch() {
  return spawn('yarn', ['-s', 'run', 'dev'], {
    stdio: 'inherit',
    shell: true,
    env: { ...process.env, NODE_OPTIONS: '--max_old_space_size=8192' },
  })
}

const INTEGRATION_FILES = path.join(__dirname, './build/integration/**')
const INTEGRATION_ASSETS_DIRECTORY = path.join(__dirname, '../ui/assets/extension')

/**
 * Copies the phabricator extension over to the ui/assets folder so they can be served by the webapp. The package
 * is published from ./browser.
 */
function copyIntegrationAssets() {
  return gulp.src(INTEGRATION_FILES).pipe(gulp.dest(INTEGRATION_ASSETS_DIRECTORY))
}

const watchIntegrationAssets = gulp.series(copyIntegrationAssets, async function watchIntegrationAssets() {
  await new Promise((_, reject) => {
    gulp.watch(INTEGRATION_FILES, copyIntegrationAssets).on('error', reject)
  })
})

module.exports = { watchIntegrationAssets, copyIntegrationAssets, build, watch }
