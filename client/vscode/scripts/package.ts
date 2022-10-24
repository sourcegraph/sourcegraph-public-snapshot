/* eslint-disable @typescript-eslint/no-unsafe-assignment */
import childProcess from 'child_process'
import fs from 'fs'

const originalPackageJson = fs.readFileSync('package.json').toString()

try {
    childProcess.execSync('yarn build-inline-extensions && yarn build', { stdio: 'inherit' })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const packageJson: any = JSON.parse(originalPackageJson)
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    packageJson.name = 'sourcegraph'
    fs.writeFileSync('package.json', JSON.stringify(packageJson))

    childProcess.execSync('vsce package --yarn --allow-star-activation -o dist', { stdio: 'inherit' })
} finally {
    fs.writeFileSync('package.json', originalPackageJson)
}
