/* eslint-disable no-sync */
import childProcess from 'child_process'

import * as semver from 'semver'

// eslint-disable-next-line  @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
const { version } = require('../package.json')

/**
 * This script is used by the CI to publish the extension to the VS Code Marketplace. It is
 * triggered when a commit has been made to the `cody/release` branch
 *
 * Refer to our CONTRIBUTION docs to learn more about our release process.
 */

/**
 * Build and publish the extension with the updated package name using the tokens stored in the
 * pipeline to run commands in pnpm and allows all events to activate the extension
 */
const isPreRelease = semver.minor(version) % 2 !== 0 ? '--pre-release' : ''

// Tokens are stored in CI pipeline
const tokens = {
    vscode: process.env.VSCODE_MARKETPLACE_TOKEN,
    openvsx: process.env.VSCODE_OPENVSX_TOKEN,
}

// Assume this is for testing purpose if tokens are not found
const hasTokens = tokens.vscode !== undefined && tokens.openvsx !== undefined

const commands = {
    vscode_info: 'vsce show sourcegraph.cody-ai --json',
    // To publish to VS Code Marketplace
    vscode_publish: `vsce publish ${isPreRelease} --packagePath dist/cody.vsix --pat $VSCODE_MARKETPLACE_TOKEN`,
    // To publish to the open-vsx registry
    openvsx_publish: 'npx ovsx publish dist/cody.vsix --pat $VSCODE_OPENVSX_TOKEN',
}

childProcess.execSync('pnpm run download-rg', { stdio: 'inherit' })
childProcess.execSync('pnpm run vsce:package', { stdio: 'inherit' })

const latestVersion = getPublishedVersion()

if (version === latestVersion || !semver.valid(latestVersion) || !semver.valid(version)) {
    throw new Error(
        version === latestVersion
            ? 'Cannot release extension with the same version number.'
            : `Invalid version number: ${latestVersion} & ${version}`
    )
}

if (!hasTokens) {
    throw new Error('Cannot publish extension without tokens.')
}

// Run the publish commands
childProcess.execSync(commands.vscode_publish, { stdio: 'inherit' })
childProcess.execSync(commands.openvsx_publish, { stdio: 'inherit' })

console.log('The extension has been published successfully.')

function getPublishedVersion(): string {
    // Get the latest release version number of the last release from VS Code Marketplace using the vsce cli tool
    const response = childProcess.execSync(commands.vscode_info).toString()
    /*
          `vscode_info` command output is extension info prepended by command name and arguments, e.g.
          $ vsce show sourcegraph.sourcegraph --json
          {
              ...json
          }
          We cut everything before the json object so that we can parse it as json.
      */
    const trimmedResponse = response.slice(Math.max(0, response.indexOf('{')))
    const latestVersion: string = JSON.parse(trimmedResponse).versions[0].version
    return latestVersion
}
