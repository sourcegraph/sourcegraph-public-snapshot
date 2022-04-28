/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

import { parse as SemVerParse, SemVer, ReleaseType } from 'semver'

// Get current package.json file
const originalPackageJson = fs.readFileSync('package.json').toString()
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const packageJson: any = JSON.parse(originalPackageJson)
const changelogFile = fs.readFileSync('CHANGELOG.md').toString()
// Get release type from pipeline env that is retreived from the commit message we use
// to push to vsce/release branch
// commit should always starts with releaseType follows by the word 'release'
// eg. patch release
const releaseType = String(process.env.VSCODE_RELEASE_TYPE && JSON.parse(process.env.VSCODE_RELEASE_TYPE)).toLowerCase()
// Define release type
function isReleaseType(releaseType: string): releaseType is ReleaseType {
    return ['major', 'minor', 'patch'].includes(releaseType)
}

if (!isReleaseType(releaseType)) {
    throw new Error('Invalid Release Type. Only major, minor, and patch are currently supported.')
}
// Update Changelog
try {
    const newChangelogFile = changelogFile.replace('Next Release', 'Latest Release')
    fs.writeFileSync('CHANGELOG.md', newChangelogFile)
} catch {
    throw new Error('Failed to update CHANGELOG.')
}
// Update package.json
try {
    const currentVersion: SemVer | null = SemVerParse(packageJson.version)
    // Get next version number in SemVer using the release type
    const nextVersion = currentVersion?.inc(releaseType).version
    // Update the package.json with the next release version and package name
    packageJson.version = nextVersion
    packageJson.name = 'sourcegraph'
    fs.writeFileSync('package.json', JSON.stringify(packageJson))
    // Publich the extension with the updated version and packname
    // using the token provided by the pipeline to run yarn and
    // allows all events to activate the extension
    childProcess.execSync('yarn vsce publish --pat $VSCODE_MARKETPLACE_TOKEN --yarn --allow-star-activation', {
        stdio: 'inherit',
    })
} finally {
    // Revert package name back to the original one then save file
    packageJson.name = '@sourcegraph/vscode'
    fs.writeFileSync('package.json', JSON.stringify(packageJson))
}
