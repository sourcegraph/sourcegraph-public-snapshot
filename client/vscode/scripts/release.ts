/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

// Update package name from "@sourcegraph/vscode" to sourcegraph for publishing
const originalPackageJson = fs.readFileSync('package.json').toString()

try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const packageJson: any = JSON.parse(originalPackageJson)
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    packageJson.name = 'sourcegraph'
    fs.writeFileSync('package.json', JSON.stringify(packageJson))

    childProcess.execSync('yarn vsce publish patch --pat $VSCODE_MARKETPLACE_TOKEN --yarn --allow-star-activation', {
        stdio: 'inherit',
    })
} finally {
    // Update package name back to "@sourcegraph/vscode"
    fs.writeFileSync('package.json', originalPackageJson)
}
