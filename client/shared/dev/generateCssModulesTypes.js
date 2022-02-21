// @ts-check

const { spawn } = require('child_process')
const path = require('path')

// eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
const pnp = require('pnpapi')

// TODO: add Typescript typings
function getPackagePaths(...packages) {
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  const locators = pnp.getAllLocators().filter(locator => packages.some(package => package === locator.name))

  return locators.map(locator =>
    pnp.getPackageInformation(locator).packageLocation.replace(/node_modules.+/, 'node_modules')
  )
}

const packagePaths = getPackagePaths('bootstrap', '@reach/tabs').join(' ')

const REPO_ROOT = path.join(__dirname, '../../..')
const CSS_MODULES_GLOB = path.resolve(__dirname, '../../*/src/**/*.module.scss')
// eslint-disable-next-line @typescript-eslint/restrict-template-expressions
const TSM_COMMAND = `yarn --silent run tsm --quiet --logLevel error "${CSS_MODULES_GLOB}" --includePaths ${packagePaths} client`
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
