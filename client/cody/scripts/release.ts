/* eslint-disable no-sync */
import childProcess from 'child_process'

import * as semver from 'semver'

/**
 * This script is used by the CI to publish the extension to the VS Code Marketplace.
 * Stable release is triggered when a commit has been made to the `cody/release` branch
 * Nightly release is triggered by CI nightly and is build on top of the main version
 *
 * Refer to our CONTRIBUTION docs to learn more about our release process.
 */

/**
 * Build and publish the extension with the updated package name using the tokens stored in the
 * pipeline to run commands in pnpm and allows all events to activate the extension
 */
const releaseType = process.env.CODY_RELEASE_TYPE || 'stable'

// Tokens are stored in CI pipeline
const tokens = {
    vscode: process.env.VSCODE_MARKETPLACE_TOKEN,
    openvsx: process.env.VSCODE_OPENVSX_TOKEN,
}

// Assume this is for testing purpose if tokens are not found
const hasTokens = tokens.vscode !== undefined && tokens.openvsx !== undefined

export const commands = {
    // Get the latest release version number of the last release from VS Code Marketplace
    vscode_info: 'vsce show sourcegraph.cody-ai --json',
    // Stable: publish to VS Code Marketplace
    vscode_publish: 'vsce publish --packagePath dist/cody.vsix --pat $VSCODE_MARKETPLACE_TOKEN',
    // Pre-release: minor bump the current version - pre-release should always be the next minor version
    vscode_pre_release: 'vsce publish minor --pre-release --packagePath dist/cody.vsix --pat $VSCODE_MARKETPLACE_TOKEN',
    // Nightly release: patch the pre-release version
    vscode_nightly: 'vsce publish {nightly} --pre-release --packagePath dist/cody.vsix --pat $VSCODE_MARKETPLACE_TOKEN',
    // To publish to the open-vsx registry
    openvsx_publish: 'npx ovsx publish dist/cody.vsix --pat $VSCODE_OPENVSX_TOKEN',
}

if (!hasTokens) {
    throw new Error('Cannot publish extension without tokens.')
}

// Build and bundle the extension
childProcess.execSync('pnpm run download-rg', { stdio: 'inherit' })
childProcess.execSync('pnpm run vsce:package', { stdio: 'inherit' })

// Run the publish commands based on the release type
switch (releaseType) {
    case 'openvsx':
        childProcess.execSync(commands.openvsx_publish, { stdio: 'inherit' })
        break
    case 'pre-release':
        // Bump the minor version number of the current number in package.json
        childProcess.execSync(commands.vscode_pre_release, { stdio: 'inherit' })
        break
    case 'nightly': {
        // Update the number from the latest pre-released version by 1 patch version manually
        const nextNightlyVersion = verifyNightly()
        if (!nextNightlyVersion) {
            throw new Error('Failed to publish nightly.')
        }
        const nightlyCommand = commands.vscode_nightly.replace('{nightly}', nextNightlyVersion)
        childProcess.execSync(nightlyCommand, { stdio: 'inherit' })
        break
    }
    default:
        // Publish to VS Code Marketplace as the version number listed in package.json
        // Publish to Open VSX Marketplace
        // Then publish to VS Code Marketplace again with a minor bump as a pre-release so that nightly builds can patch on top of it
        childProcess.execSync(commands.vscode_publish, { stdio: 'inherit' })
        childProcess.execSync(commands.openvsx_publish, { stdio: 'inherit' })
        childProcess.execSync(commands.vscode_pre_release, { stdio: 'inherit' })
}

console.log('The extension has been published successfully.')

// Get the latest release version number from VS Code Marketplace using the vsce cli tool
/*
    `vscode_info` command output is extension info prepended by command name and arguments, e.g.
    $ vsce show sourcegraph.cody-ai --json
    {
        ...json
    }
    We cut everything before the json object so that we can parse it as json.
*/
function getPublishedVersion(): string {
    const response = childProcess.execSync(commands.vscode_info).toString()
    const trimmedResponse = response.slice(Math.max(0, response.indexOf('{')))
    const latestVersion = JSON.parse(trimmedResponse).versions[0].version as string
    return latestVersion
}

// Check if the last release version number has a minor version number that is odd
// A pre-release version has odd minor number
function verifyNightly(): string {
    const marketplaceVersion = getPublishedVersion()
    const isLastPublishedVersionPreRelease = semver.minor(marketplaceVersion) % 2 !== 0
    const nextNightlyVersion = semver.inc(marketplaceVersion, 'patch') || '0.0.0'
    const isNextNightlyVersionPreRelease = semver.minor(nextNightlyVersion) % 2 !== 0
    if (!isLastPublishedVersionPreRelease || isNextNightlyVersionPreRelease) {
        throw new Error('Nightly build should only build on top of pre-released version.')
    }
    return nextNightlyVersion
}
