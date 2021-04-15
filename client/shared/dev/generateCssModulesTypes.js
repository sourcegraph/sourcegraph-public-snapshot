// @ts-check

const { spawn } = require('child_process')
const path = require('path')

const CSS_MODULES_GLOB = path.resolve(__dirname, '../../*/src/**/*.module.scss')
const TSM_COMMAND = `yarn --silent --ignore-engines --ignore-scripts tsm --logLevel error ${CSS_MODULES_GLOB}`
const [BIN, ...ARGS] = TSM_COMMAND.split(' ')

/**
 * Generates the TypeScript types CSS modules.
 */
function cssModulesTypings() {
  return spawn(BIN, ARGS, {
    stdio: 'inherit',
    shell: true,
  })
}

/**
 * Watch CSS modules and generates the TypeScript types for them.
 */
function watchCSSModulesTypings() {
  return spawn(BIN, [...ARGS, '--watch'], {
    stdio: 'inherit',
    shell: true,
  })
}

module.exports = {
  cssModulesTypings,
  watchCSSModulesTypings,
}
