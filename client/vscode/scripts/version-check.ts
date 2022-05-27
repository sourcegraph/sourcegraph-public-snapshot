/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

import { version } from '../package.json'

/*
 * Run this script to update changelog and current package version to
 * match the last released version number from the VS Code Marketplace
 * using the vsce cli tool
 * Download at https://code.visualstudio.com/api/working-with-extensions/publishing-extension#installation
 */
// ! NOTE: Make sure you have VSCE CLI TOOL installed locally
// Current package.json and changelog files
const originalPackageJson = fs.readFileSync('package.json').toString()
const originalChangelogFile = fs.readFileSync('CHANGELOG.md').toString()

try {
    // Get the version nubmer of the last release from VS Code Marketplace using the vsce cli tool
    // https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph
    const response = childProcess.execSync('vsce show sourcegraph.sourcegraph --json').toString()
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    const lastReleaseVersion: string = JSON.parse(response).versions[0].version
    if (lastReleaseVersion !== version) {
        try {
            // Update version number in package.json
            const updatedPackageJson = originalPackageJson.replace(
                `"version": "${version}"`,
                `"version": "${lastReleaseVersion}"`
            )
            fs.writeFileSync('package.json', updatedPackageJson)
            // Update Changelog format for next release
            const changelogFile = originalChangelogFile.replace(
                'Latest Release',
                `Next Release\n### Changes\n### Fixes\n## ${lastReleaseVersion}`
            )
            fs.writeFileSync('CHANGELOG.md', changelogFile)

            console.log('CHangelog and package version has been updated successfully to match the last release.')
        } catch (error) {
            console.error('Cannot update files with last release version:', error)
        }
    } else {
        console.log('Currently on the same version number as the last release.')
    }
} catch (error) {
    console.error('Cannot get the last release version#:', error)
    fs.writeFileSync('package.json', originalPackageJson)
    fs.writeFileSync('CHANGELOG.md', originalChangelogFile)
}
