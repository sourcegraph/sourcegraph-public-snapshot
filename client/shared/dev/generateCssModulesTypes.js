// @ts-check

const { spawn } = require('child_process')
const path = require('path')

const CSS_MODULES_GLOB = path.resolve(__dirname, '../../*/src/**/*.module.scss')
const TSM_LOG_LEVEL = '--logLevel error'

/**
 * Generates the TypeScript types CSS modules.
 */
function cssModulesTypings() {
  return spawn('tsm', [TSM_LOG_LEVEL, CSS_MODULES_GLOB], {
    stdio: 'inherit',
    shell: true,
  })
}

/**
 * Watch CSS modules and generates the TypeScript types for them.
 */
function watchCSSModulesTypings() {
  return spawn('tsm', ['--watch', TSM_LOG_LEVEL, CSS_MODULES_GLOB], {
    stdio: 'inherit',
    shell: true,
  })
}

module.exports = {
  cssModulesTypings,
  watchCSSModulesTypings,
}
