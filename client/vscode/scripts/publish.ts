/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

/**
 * This script is used by the CI to publish the extension to the VS Code Marketplace
 * It is triggered when a commit has been made to the vsce release branch
 */

try {
    // Get current package.json and changelog files
    const originalPackageJson = fs.readFileSync('package.json').toString()
    try {
        // Update package name from @sourcegraph/vscode to sourcegraph in package.json
        const packageJson = originalPackageJson.replace('@sourcegraph/vscode', 'sourcegraph')
        fs.writeFileSync('package.json', packageJson)
        // Build and publish the extension with the updated package name
        // using the token provided by the pipeline to run yarn and
        // allows all events to activate the extension
        childProcess.execSync(
            // Use vsce package command for testing this script without publishing the extension
            // 'yarn vsce package $VSCODE_RELEASE_TYPE --yarn --allow-star-activation',
            'yarn vsce publish $VSCODE_RELEASE_TYPE --pat $VSCODE_MARKETPLACE_TOKEN --yarn --allow-star-activation',
            {
                stdio: 'inherit',
            }
        )
    } catch (error) {
        console.error('Failed to publish VSCE to Marketplace:', error)
    } finally {
        // Revert changes made to package.json & Changelog
        fs.writeFileSync('package.json', originalPackageJson)
    }
} catch (error) {
    console.error('Failed to process with releasing the VS Code Extension:', error)
}
