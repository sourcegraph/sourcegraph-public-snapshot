import childProcess from 'child_process'
import fs from 'fs'

const originalPackageJson = fs.readFileSync('package.json').toString()

try {
    childProcess.execSync('pnpm build-inline-extensions && pnpm build', { stdio: 'inherit' })
    const packageJson: any = JSON.parse(originalPackageJson)
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    packageJson.name = 'sourcegraph'
    fs.writeFileSync('package.json', JSON.stringify(packageJson))

    childProcess.execSync('vsce package --no-dependencies --allow-star-activation -o dist', { stdio: 'inherit' })
} finally {
    fs.writeFileSync('package.json', originalPackageJson)
}
