/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

import * as semver from 'semver'

import { version } from '../package.json'
/**
 * This script is used by the CI to publish the extension to the VS Code Marketplace
 * It is triggered when a commit has been made to the vsce release branch
 */
// Get current package.json
const originalPackageJson = fs.readFileSync('package.json').toString()
/**
 * Build and publish the extension with the updated package name
 * using the token stored in the pipeline to run commands in yarn
 * and allows all events to activate the extension
 */
const isPreRelease = semver.minor(version) % 2 !== 0
const publishToken = process.env.VSCODE_MARKETPLACE_TOKEN
// Assume this is for testing purpose if publishToken cannot be found
// Use vsce package command instead without publishing the extension for testing
const publishCommand = publishToken
    ? `yarn vsce publish ${
          isPreRelease ? '--pre-release' : ''
      } --pat $VSCODE_MARKETPLACE_TOKEN --yarn --allow-star-activation`
    : 'yarn vsce package --yarn --allow-star-activation'
// Publish the extension with the correct extension name "sourcegraph"
try {
    // Get the latest release version nubmer of the last release from VS Code Marketplace using the vsce cli tool
    const response = childProcess.execSync('vsce show sourcegraph.sourcegraph --json').toString()
    const latestVersion: string = JSON.parse(response).versions[0].version
    if (publishToken && (version === latestVersion || !semver.valid(latestVersion) || !semver.valid(version))) {
        throw new Error(
            version === latestVersion
                ? 'Cannot release extension with the same version number.'
                : `invalid version number: ${latestVersion} & ${version}`
        )
    }
    // Update package name from @sourcegraph/vscode to sourcegraph in package.json
    const packageJson = originalPackageJson.replace('@sourcegraph/vscode', 'sourcegraph')
    fs.writeFileSync('package.json', packageJson)
    // Run the publish command
    childProcess.execSync(publishCommand, {
        stdio: 'inherit',
    })
    console.log(`The extension has been ${publishToken ? 'published' : 'packaged'} successfully`)
} catch (error) {
    console.error('Failed to publish VSCE:', error)
    console.error('Do not run this script locally to publish the extension.')
} finally {
    // Revert changes made to package.json
    fs.writeFileSync('package.json', originalPackageJson)
}
