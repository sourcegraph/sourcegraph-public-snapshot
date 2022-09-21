// @ts-check

const { spawn } = require('child_process')
const path = require('path')

const REPO_ROOT = path.join(__dirname, '../../..')
const CSS_MODULES_GLOB = path.resolve(__dirname, '../../*/src/**/*.module.scss')
const JETBRAINS_CSS_MODULES_GLOB = path.resolve(__dirname, '../../jetbrains/webview/**/*.module.scss')
const TSM_COMMAND = `yarn tsm --logLevel error "{${CSS_MODULES_GLOB},${JETBRAINS_CSS_MODULES_GLOB}}" --includePaths node_modules client`
const [BIN, ...ARGS] = TSM_COMMAND.split(' ')

/**
 * Generates the TypeScript types CSS modules.
 */
function cssModulesTypings() {
  return spawn(BIN, ARGS, {
    stdio: 'inherit',
    shell: true,
    cwd: REPO_ROOT,
  })
}

/**
 * Watch CSS modules and generates the TypeScript types for them.
 */
function watchCSSModulesTypings() {
  return spawn(BIN, [...ARGS, '--watch'], {
    stdio: 'inherit',
    shell: true,
    cwd: REPO_ROOT,
  })
}

module.exports = {
  cssModulesTypings,
  watchCSSModulesTypings,
}
