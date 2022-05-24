/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

/*
 * Get release type from pipeline env called $VSCODE_RELEASE_TYPE
 * it is retreived from the commit message we made to push to vsce/release branch
 * The commit to trigger release must be starts with release-types ["major", "minors", "patch"],
 * follows by the word 'release': eg. patch release
 */
// ! Future Plan: Commit the changes we made for the release (version number and Changelog) to both main and release branch
const releaseType = String(process.env.VSCODE_RELEASE_TYPE && JSON.parse(process.env.VSCODE_RELEASE_TYPE))
    .toLowerCase()
    .replace(' ', '')
const isReleaseType = ['major', 'minor', 'patch'].includes(releaseType)
// Only proceed the release steps if a valid release type is provided
if (isReleaseType) {
    try {
        // Get current package.json and changelog files
        const originalPackageJson = fs.readFileSync('package.json').toString()
        const originalChangelogFile = fs.readFileSync('CHANGELOG.md').toString()
        try {
            // Update package name from @sourcegraph/vscode to sourcegraph in package.json
            const packageJson = originalPackageJson.replace('@sourcegraph/vscode', 'sourcegraph')
            fs.writeFileSync('package.json', packageJson)
            // Changelog
            const changelogFile = originalChangelogFile.replace('Next Release', 'Latest Release')
            fs.writeFileSync('CHANGELOG.md', changelogFile)
            // Build and publish the extension with the updated package name and CHANGELOG
            // using the token provided by the pipeline to run yarn and
            // allows all events to activate the extension
            childProcess.execSync(
                // Use vsce package command for testing this script without publishing the extension
                // eg: 'yarn vsce package $VSCODE_RELEASE_TYPE --yarn --allow-star-activation',
                'yarn build && yarn vsce publish $VSCODE_RELEASE_TYPE --pat $VSCODE_MARKETPLACE_TOKEN --yarn --allow-star-activation',
                {
                    stdio: 'inherit',
                }
            )
        } catch (error) {
            console.error(`Publish VSCE with ${releaseType} failed:`, error)
            // Make sure any changes made are reverted
            fs.writeFileSync('package.json', originalPackageJson)
            fs.writeFileSync('CHANGELOG.md', originalChangelogFile)
        } finally {
            // Revert changes made to package.json & Changelog
            fs.writeFileSync('package.json', originalPackageJson)
            fs.writeFileSync('CHANGELOG.md', originalChangelogFile)
        }
    } catch (error) {
        console.error('Failed to process with releasing the VS Code Extension:', error)
    }
} else {
    throw new Error(`Release type retreived is invalid:  ${releaseType}`)
}
