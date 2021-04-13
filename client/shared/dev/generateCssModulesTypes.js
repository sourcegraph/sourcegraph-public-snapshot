// @ts-check

const { spawn } = require('child_process')
const path = require('path')

const CSS_MODULES_GLOB = path.resolve(__dirname, '../../*/src/**/*.module.scss')
const TSM_LOG_LEVEL = '--logLevel error'
const TSM_COMMAND = 'yarn --silent --ignore-engines --ignore-scripts tsm'

/**
 * Generates the TypeScript types CSS modules.
 */
function cssModulesTypings() {
  return spawn(TSM_COMMAND, [TSM_LOG_LEVEL, CSS_MODULES_GLOB], {
    stdio: 'inherit',
    shell: true,
  })
}

/**
 * Watch CSS modules and generates the TypeScript types for them.
 */
function watchCSSModulesTypings() {
  return spawn(TSM_COMMAND, ['--watch', TSM_LOG_LEVEL, CSS_MODULES_GLOB], {
    stdio: 'inherit',
    shell: true,
  })
}

module.exports = {
  cssModulesTypings,
  watchCSSModulesTypings,
}
