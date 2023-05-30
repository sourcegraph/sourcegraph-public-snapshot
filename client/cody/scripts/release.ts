/* eslint-disable no-sync */
import childProcess from 'child_process'

import * as semver from 'semver'

// eslint-disable-next-line  @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
const { version } = require('../package.json')

/**
 * This script is used by the CI to publish the extension to the VS Code Marketplace.
 * Stable release is triggered when a commit has been made to the `cody/release` branch
 * Nightly release is triggered by CI nightly and is built from the `main` branch
 *
 * Refer to our CONTRIBUTION docs to learn more about our release process.
 */

/**
 * Build and publish the extension with the updated package name using the tokens stored in the
 * pipeline to run commands in pnpm and allows all events to activate the extension.
 *
 * releaseType avilable in CI: stable, nightly
 */
const releaseType = process.env.CODY_RELEASE_TYPE

// Tokens are stored in CI pipeline
const tokens = {
    vscode: releaseType === 'dry-run' ? 'dry-run' : process.env.VSCODE_MARKETPLACE_TOKEN,
    openvsx: releaseType === 'dry-run' ? 'dry-run' : process.env.VSCODE_OPENVSX_TOKEN,
}

// Assume this is for testing purpose if tokens are not found
const hasTokens = tokens.vscode !== undefined && tokens.openvsx !== undefined
// Get today's date for nightly build. Example: 2021-01-01 = 20210101
const today = new Date().toISOString().slice(0, 10).replace(/-/g, '')
// Set the version number for today's nightly build.
// The minor number should be the current minor number plus 1.
// The patch number should be today's date while major and minor should reminds the same as package.json version.
// Example: 1.0.0 in package.json -> 1.1.today's date -> 1.1.20210101
const currentVersion = semver.valid(version)
if (!currentVersion) {
    throw new Error('Cannot get the current version number from package.json')
}
const tonightVersion = semver.inc(currentVersion, 'minor')?.replace(/\.\d+$/, `.${today}`)
if (!tonightVersion || semver.minor(tonightVersion) - semver.minor(currentVersion) !== 1) {
    throw new Error("Could not populate the current version number for tonight's build.")
}

export const commands = {
    // Get the latest release version number of the last release from VS Code Marketplace
    vscode_info: 'vsce show sourcegraph.cody-ai --json',
    // Stable: publish to VS Code Marketplace
    vscode_publish: 'vsce publish --packagePath dist/cody.vsix --pat $VSCODE_MARKETPLACE_TOKEN',
    // Nightly release: publish to VS Code Marketplace with today's date as patch number
    vscode_nightly: `vsce publish ${tonightVersion} --pre-release --packagePath dist/cody.vsix --pat $VSCODE_MARKETPLACE_TOKEN`,
    // To publish to the open-vsx registry
    openvsx_publish: 'npx ovsx publish dist/cody.vsix --pat $VSCODE_OPENVSX_TOKEN',
}

if (!hasTokens) {
    throw new Error('Cannot publish extension without tokens.')
}

if (releaseType !== 'dry-run') {
    // Build and bundle the extension
    childProcess.execSync('pnpm run download-rg', { stdio: 'inherit' })
    childProcess.execSync('pnpm run vsce:package', { stdio: 'inherit' })
}

// Run the publish commands based on the release type
switch (releaseType) {
    case 'dry-run': {
        console.log(
            !semver.valid(tonightVersion)
                ? 'Not a valid version number: ' + tonightVersion
                : `Current version is ${currentVersion} and the pre-release number for tonight's build is ${tonightVersion}`
        )
        break
    }
    case 'openvsx':
        childProcess.execSync(commands.openvsx_publish, { stdio: 'inherit' })
        break
    case 'nightly':
        // check if tonightVersion is a valid semv version number
        if (!tonightVersion || !semver.valid(tonightVersion) || semver.valid(tonightVersion) === currentVersion) {
            throw new Error('Cannot publish nightly build with an invalid version number: ' + tonightVersion)
        }
        // Publish to VS Code Marketplace with today's date as patch number
        childProcess.execSync(commands.vscode_nightly, { stdio: 'inherit' })
        break
    case 'stable':
        // Publish to VS Code Marketplace as the version number listed in package.json
        // Publish to Open VSX Marketplace
        childProcess.execSync(commands.vscode_publish, { stdio: 'inherit' })
        childProcess.execSync(commands.openvsx_publish, { stdio: 'inherit' })
    default:
        throw new Error(`Invalid release type: ${releaseType}`)
}

console.log(releaseType === 'dry-run' ? 'Dry run completed.' : 'The extension has been published successfully.')
